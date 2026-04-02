package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stickpro/go-store/internal/delivery/http/request/cart_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/cart_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// getCart
//
//	@Summary		Get cart
//	@Tags			Cart
//	@Description	Returns the current cart for authenticated user or guest session
//	@Produce		json
//	@Success		200	{object}	response.Result[cart_response.CartResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Router			/api/v1/cart [get]
func (h *Handler) getCart(c fiber.Ctx) error {
	owner, err := h.cartOwner(c)
	if err != nil {
		return err
	}

	cartDTO, err := h.services.CartService.GetCart(c.Context(), owner)
	if err != nil {
		return h.handleError(err, "cart")
	}

	return c.JSON(response.OkByData(cart_response.NewFromDTO(cartDTO)))
}

// addCartItem
//
//	@Summary		Add item to cart
//	@Tags			Cart
//	@Description	Adds a product variant to the cart. If the variant is already present, its quantity is increased.
//	@Accept			json
//	@Produce		json
//	@Param			body	body		cart_request.AddCartItemRequest	true	"Item to add"
//	@Success		200		{object}	response.Result[cart_response.CartResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/api/v1/cart/items [post]
func (h *Handler) addCartItem(c fiber.Ctx) error {
	owner, err := h.cartOwner(c)
	if err != nil {
		return err
	}
	req := &cart_request.AddCartItemRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.RequestToAddCartItemDTO(req)
	cartDTO, err := h.services.CartService.AddItem(c.Context(), owner, d)
	if err != nil {
		return h.handleError(err, "cart")
	}

	return c.JSON(response.OkByData(cart_response.NewFromDTO(cartDTO)))
}

// updateCartItemQuantity
//
//	@Summary		Update cart item quantity
//	@Tags			Cart
//	@Description	Sets a new quantity for the specified variant in the cart
//	@Accept			json
//	@Produce		json
//	@Param			variantId	path		string								true	"Variant UUID"
//	@Param			body		body		cart_request.UpdateCartItemRequest	true	"New quantity"
//	@Success		200			{object}	response.Result[cart_response.CartResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/api/v1/cart/items/{variantId} [patch]
func (h *Handler) updateCartItemQuantity(c fiber.Ctx) error {
	owner, err := h.cartOwner(c)
	if err != nil {
		return err
	}
	req := &cart_request.UpdateCartItemRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	variantID, err := uuid.Parse(c.Params("variantId"))

	cartDTO, err := h.services.CartService.UpdateQuantity(c.Context(), owner, variantID, req.Quantity)
	if err != nil {
		return h.handleError(err, "cart")
	}

	return c.JSON(response.OkByData(cart_response.NewFromDTO(cartDTO)))
}

// removeCartItem
//
//	@Summary		Remove item from cart
//	@Tags			Cart
//	@Description	Removes the specified variant from the cart
//	@Produce		json
//	@Param			variantId	path		string	true	"Variant UUID"
//	@Success		200			{object}	response.Result[cart_response.CartResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Router			/api/v1/cart/items/{variantId} [delete]
func (h *Handler) removeCartItem(c fiber.Ctx) error {
	owner, err := h.cartOwner(c)
	if err != nil {
		return err
	}
	variantID, err := uuid.Parse(c.Params("variantId"))

	cartDTO, err := h.services.CartService.RemoveItem(c.Context(), owner, variantID)
	if err != nil {
		return h.handleError(err, "cart")
	}

	return c.JSON(response.OkByData(cart_response.NewFromDTO(cartDTO)))
}

// clearCart
//
//	@Summary		Clear cart
//	@Tags			Cart
//	@Description	Removes all items from the cart
//	@Produce		json
//	@Success		200	{object}	response.Result[any]
//	@Failure		400	{object}	apierror.Errors
//	@Router			/api/v1/cart [delete]
func (h *Handler) clearCart(c fiber.Ctx) error {
	owner, err := h.cartOwner(c)
	if err != nil {
		return err
	}

	err = h.services.CartService.ClearCart(c.Context(), owner)
	if err != nil {
		return h.handleError(err, "cart")
	}

	return c.JSON(response.OkByMessage("cart cleared"))
}

func (h *Handler) initCartRoutes(v1 fiber.Router) {
	c := v1.Group("/cart")
	c.Get("/", h.getCart)
	c.Post("/items", h.addCartItem)
	c.Patch("/items/:variantId", h.updateCartItemQuantity)
	c.Delete("/items/:variantId", h.removeCartItem)
	c.Delete("/", h.clearCart)
}

// cartOwner builds a CartOwner from the request context.
// Prefers the authenticated user; falls back to session cookie,
// then X-Session-ID header. Returns an error if neither is present.
func (h *Handler) cartOwner(c fiber.Ctx) (dto.Owner, error) {
	if user, err := loadAuthUser(c); err == nil {
		return dto.Owner{UserID: &user.ID}, nil
	}

	sessionStr := c.Cookies("session_id")
	if sessionStr == "" {
		sessionStr = c.Get("X-Session-ID")
	}

	if sessionStr == "" {
		return dto.Owner{}, apierror.New().
			AddError(errors.New("session_id cookie or X-Session-ID header is required")).
			SetHttpCode(fiber.StatusBadRequest)
	}

	sessionID, err := uuid.Parse(sessionStr)
	if err != nil {
		return dto.Owner{}, apierror.New().
			AddError(errors.New("session_id must be a valid UUID")).
			SetHttpCode(fiber.StatusBadRequest)
	}

	return dto.Owner{SessionID: &sessionID}, nil
}
