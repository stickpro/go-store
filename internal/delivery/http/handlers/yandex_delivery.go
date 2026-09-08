package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/yandex_delivery_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// listYandexDeliveryPoints returns the cached Yandex Delivery pickup points
// (ПВЗ + postomats), optionally filtered.
//
//	@Summary		List Yandex Delivery pickup points
//	@Description	List cached Yandex Delivery pickup points (ПВЗ and postomats). The list is refreshed in the background once a day; use the query params to narrow it down. min_lat/max_lat/min_lon/max_lon narrow the result to a map viewport and must be supplied together.
//	@Tags			YandexDelivery
//	@Accept			json
//	@Produce		json
//	@Param			geo_id		query		int		false	"Yandex geo id (locality)"
//	@Param			locality	query		string	false	"Locality name"
//	@Param			type		query		string	false	"Pickup point type"	Enums(pickup_point, terminal)
//	@Param			min_lat		query		number	false	"Map viewport min latitude (requires max_lat, min_lon, max_lon)"
//	@Param			max_lat		query		number	false	"Map viewport max latitude (requires min_lat, min_lon, max_lon)"
//	@Param			min_lon		query		number	false	"Map viewport min longitude (requires min_lat, max_lat, max_lon)"
//	@Param			max_lon		query		number	false	"Map viewport max longitude (requires min_lat, max_lat, min_lon)"
//	@Success		200			{object}	response.Result[[]yandex_delivery_response.DeliveryPointResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/yandex-delivery/delivery-points [get]
func (h *Handler) listYandexDeliveryPoints(c fiber.Ctx) error {
	filter := dto.YandexDeliveryPointsFilter{
		Locality: c.Query("locality"),
		Type:     c.Query("type"),
	}

	if raw := c.Query("geo_id"); raw != "" {
		geoID, err := strconv.Atoi(raw)
		if err != nil {
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
		}
		filter.GeoID = &geoID
	}

	bbox, err := parseViewportBBox(c)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	filter.BBox = bbox

	points, err := h.services.YandexDeliveryService.ListDeliveryPoints(c.Context(), filter)
	if err != nil {
		return h.handleError(err, "yandex delivery points")
	}

	return c.JSON(response.OkByData(yandex_delivery_response.NewListFromDTO(points)))
}

func (h *Handler) initYandexDeliveryRoutes(v1 fiber.Router) {
	g := v1.Group("/yandex-delivery")
	g.Get("/delivery-points", h.listYandexDeliveryPoints)
}
