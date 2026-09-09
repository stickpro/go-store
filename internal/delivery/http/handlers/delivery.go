package handlers

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/shopspring/decimal"
	"github.com/stickpro/go-store/internal/delivery/http/request/delivery_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/delivery_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/shipping"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// listDeliveryPoints returns a carrier's cached pickup points in one common
// shape. The carrier is the {provider} path segment ("cdek", "yandex_delivery",
// "pochta"); every carrier's full directory is refreshed in the background, so
// this only filters an in-memory list.
//
//	@Summary		List carrier pickup points
//	@Description	Cached pickup points for one delivery carrier, mapped to a common shape (carrier-specific fields live under `details`). min_lat/max_lat/min_lon/max_lon narrow the result to a map viewport and must be supplied together; latitude+longitude (with optional radius_km) narrow it to a radius around a point, nearest first.
//	@Tags			Delivery
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string										true	"Delivery carrier"	Enums(cdek, yandex_delivery, pochta)
//	@Param			request		query		delivery_request.ListDeliveryPointsRequest	true	"Filters"
//	@Success		200			{object}	response.Result[[]delivery_response.DeliveryPointResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/delivery/{provider}/points [get]
func (h *Handler) listDeliveryPoints(c fiber.Ctx) error {
	provider, ok := h.services.Shipping.Get(c.Params("provider"))
	if !ok {
		return apierror.New().AddError(fmt.Errorf("unknown delivery provider %q", c.Params("provider"))).SetHttpCode(fiber.StatusNotFound)
	}

	req := &delivery_request.ListDeliveryPointsRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	filter := dto.DeliveryPointsFilter{
		Type:      req.Type,
		Locality:  req.Locality,
		Region:    req.Region,
		Index:     req.Index,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		RadiusKM:  req.RadiusKM,
	}
	if req.HasBBox() {
		filter.BBox = &dto.GeoBBox{
			MinLat: *req.MinLat, MaxLat: *req.MaxLat,
			MinLon: *req.MinLon, MaxLon: *req.MaxLon,
		}
	}

	points := shipping.Filter(provider.Points(), filter)
	return c.JSON(response.OkByData(delivery_response.NewList(points)))
}

// listDeliveryProviders returns the delivery carriers this store supports and
// whether each is currently enabled.
//
//	@Summary	List delivery carriers
//	@Tags		Delivery
//	@Produce	json
//	@Success	200	{object}	response.Result[[]delivery_response.ProviderResponse]
//	@Router		/v1/delivery/providers [get]
func (h *Handler) listDeliveryProviders(c fiber.Ctx) error {
	out := make([]delivery_response.ProviderResponse, 0, len(h.services.Shipping.All()))
	for _, p := range h.services.Shipping.All() {
		out = append(out, delivery_response.ProviderResponse{Code: p.Code(), Enabled: p.Enabled()})
	}
	return c.JSON(response.OkByData(out))
}

// listDeliveryMethods returns the checkout delivery-method catalogue: one entry
// per method tab, with the pickup-point / price availability flags the frontend
// needs. Pass the entry's `code` back as delivery_method_code at checkout.
//
//	@Summary	List delivery methods
//	@Tags		Delivery
//	@Produce	json
//	@Success	200	{object}	response.Result[[]delivery_response.DeliveryMethodResponse]
//	@Router		/v1/delivery/methods [get]
func (h *Handler) listDeliveryMethods(c fiber.Ctx) error {
	return c.JSON(response.OkByData(delivery_response.NewMethodList(h.services.Shipping.Methods())))
}

// calculateDeliveryRates returns shipping-cost options for a parcel from every
// carrier that supports rate calculation, cheapest first.
//
//	@Summary		Calculate shipping rates (all carriers)
//	@Description	Shipping cost for a parcel from every enabled carrier. Destination is a pickup point (to_point_code) or a postal code (to_postal_code). Parcel weight (kg) / dimensions (cm) are optional: omit them and the parcel is derived from the caller's cart (auth or X-Session-ID), falling back to the store default. Results are cached server-side per route.
//	@Tags			Delivery
//	@Accept			json
//	@Produce		json
//	@Param			request	body		delivery_request.CalculateRatesRequest	true	"Parcel and route"
//	@Success		200		{object}	response.Result[[]delivery_response.ShippingRateResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/delivery/rates [post]
func (h *Handler) calculateDeliveryRates(c fiber.Ctx) error {
	q, err := h.parseRateQuery(c)
	if err != nil {
		return err
	}

	rates, errs := h.services.Shipping.QuoteAll(c.Context(), q)
	for _, e := range errs {
		h.logger.Errorw("shipping: rate calculation failed", "error", e)
	}
	return c.JSON(response.OkByData(delivery_response.NewRateList(rates)))
}

// calculateProviderDeliveryRates returns one carrier's shipping-cost options.
//
//	@Summary	Calculate shipping rates (one carrier)
//	@Tags		Delivery
//	@Accept		json
//	@Produce	json
//	@Param		provider	path		string									true	"Delivery carrier"	Enums(cdek, yandex_delivery, pochta)
//	@Param		request		body		delivery_request.CalculateRatesRequest	true	"Parcel and route"
//	@Success	200			{object}	response.Result[[]delivery_response.ShippingRateResponse]
//	@Failure	400			{object}	apierror.Errors
//	@Failure	404			{object}	apierror.Errors
//	@Failure	422			{object}	apierror.Errors
//	@Router		/v1/delivery/{provider}/rates [post]
func (h *Handler) calculateProviderDeliveryRates(c fiber.Ctx) error {
	code := c.Params("provider")
	if _, ok := h.services.Shipping.Rater(code); !ok {
		return apierror.New().AddError(fmt.Errorf("carrier %q does not support rate calculation", code)).SetHttpCode(fiber.StatusNotFound)
	}

	q, err := h.parseRateQuery(c)
	if err != nil {
		return err
	}

	rates, err := h.services.Shipping.Quote(c.Context(), code, q)
	if errors.Is(err, shipping.ErrRatesNotSupported) {
		return apierror.New().AddError(fmt.Errorf("carrier %q does not support rate calculation", code)).SetHttpCode(fiber.StatusNotFound)
	}
	if err != nil {
		return h.handleError(err, "shipping rates")
	}
	return c.JSON(response.OkByData(delivery_response.NewRateList(rates)))
}

// parseRateQuery binds and validates the request body and turns it into a
// shipping.RateQuery. The parcel comes from an explicit body field set, or —
// failing that — the caller's cart, or the store default.
func (h *Handler) parseRateQuery(c fiber.Ctx) (shipping.RateQuery, error) {
	req := &delivery_request.CalculateRatesRequest{}
	if err := c.Bind().Body(req); err != nil {
		return shipping.RateQuery{}, err
	}

	defaults := h.services.Shipping.ParcelDefaults()
	var parcel shipping.Parcel
	switch {
	case req.HasParcel():
		parcel = shipping.NewParcel(
			dec(req.WeightKG), dec(req.LengthCM), dec(req.WidthCM), dec(req.HeightCM), dec(req.DeclaredValue),
			defaults,
		)
	default:
		if items := h.cartItems(c); len(items) > 0 {
			parcel = shipping.ParcelFromCart(items, defaults)
		} else {
			parcel = shipping.NewParcel(decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, decimal.Zero, defaults)
		}
	}

	return shipping.RateQuery{
		FromPostalCode: req.FromPostalCode,
		ToPostalCode:   req.ToPostalCode,
		ToPointCode:    req.ToPointCode,
		DeliveryType:   req.DeliveryType,
		Parcel:         parcel,
	}, nil
}

// cartItems returns the caller's cart lines, or nil when there is no
// identifiable cart (no auth and no session).
func (h *Handler) cartItems(c fiber.Ctx) []dto.CartItemsDTO {
	owner, err := h.cartOwner(c)
	if err != nil {
		return nil
	}
	cart, err := h.services.CartService.GetCart(c.Context(), owner)
	if err != nil || cart == nil {
		return nil
	}
	return cart.Items
}

// dec parses a decimal string, treating an empty or invalid value as zero.
func dec(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

func (h *Handler) initDeliveryRoutes(v1 fiber.Router) {
	g := v1.Group("/delivery")
	g.Get("/providers", h.listDeliveryProviders)
	g.Get("/methods", h.listDeliveryMethods)
	g.Get("/:provider/points", h.listDeliveryPoints)
	g.Post("/rates", h.calculateDeliveryRates)
	g.Post("/:provider/rates", h.calculateProviderDeliveryRates)
}
