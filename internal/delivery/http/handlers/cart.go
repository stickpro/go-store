package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
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

func (h *Handler) addCartItem(c fiber.Ctx) error {
	// TODO implement me
	return nil
}

func (h *Handler) updateCartItemQuantity(c fiber.Ctx) error {
	// TODO implement me
	return nil
}

func (h *Handler) removeCartItem(c fiber.Ctx) error {
	// TODO implement me
	return nil
}

func (h *Handler) clearCart(c fiber.Ctx) error {
	// TODO implement me
	return nil
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
func (h *Handler) cartOwner(c fiber.Ctx) (dto.CartOwner, error) {
	if user, err := loadAuthUser(c); err == nil {
		return dto.CartOwner{UserID: &user.ID}, nil
	}

	sessionStr := c.Cookies("session_id")
	if sessionStr == "" {
		sessionStr = c.Get("X-Session-ID")
	}

	if sessionStr == "" {
		return dto.CartOwner{}, apierror.New().
			AddError(errors.New("session_id cookie or X-Session-ID header is required")).
			SetHttpCode(fiber.StatusBadRequest)
	}

	sessionID, err := uuid.Parse(sessionStr)
	if err != nil {
		return dto.CartOwner{}, apierror.New().
			AddError(errors.New("session_id must be a valid UUID")).
			SetHttpCode(fiber.StatusBadRequest)
	}

	return dto.CartOwner{SessionID: &sessionID}, nil
}
