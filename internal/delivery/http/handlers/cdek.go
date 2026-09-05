package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/cdek_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// listCDEKDeliveryPoints is a function that returns the cached CDEK delivery
// points (offices + postomats), optionally filtered.
//
//	@Summary		List CDEK delivery points
//	@Description	List cached CDEK delivery points (offices and postomats). The list is refreshed in the background once a day; use the query params to narrow it down. min_lat/max_lat/min_lon/max_lon narrow the result to a map viewport and must be supplied together.
//	@Tags			CDEK
//	@Accept			json
//	@Produce		json
//	@Param			city_code	query		int		false	"CDEK city code"
//	@Param			postal_code	query		string	false	"Postal code"
//	@Param			type		query		string	false	"Delivery point type"	Enums(PVZ, POSTAMAT)
//	@Param			min_lat		query		number	false	"Map viewport min latitude (requires max_lat, min_lon, max_lon)"
//	@Param			max_lat		query		number	false	"Map viewport max latitude (requires min_lat, min_lon, max_lon)"
//	@Param			min_lon		query		number	false	"Map viewport min longitude (requires min_lat, max_lat, max_lon)"
//	@Param			max_lon		query		number	false	"Map viewport max longitude (requires min_lat, max_lat, min_lon)"
//	@Success		200			{object}	response.Result[[]cdek_response.DeliveryPointResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/cdek/delivery-points [get]
func (h *Handler) listCDEKDeliveryPoints(c fiber.Ctx) error {
	filter := dto.CDEKDeliveryPointsFilter{
		PostalCode: c.Query("postal_code"),
		Type:       c.Query("type"),
	}

	if raw := c.Query("city_code"); raw != "" {
		cityCode, err := strconv.Atoi(raw)
		if err != nil {
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
		}
		filter.CityCode = &cityCode
	}

	bbox, err := parseCDEKBBox(c)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	filter.BBox = bbox

	points, err := h.services.CDEKService.ListDeliveryPoints(c.Context(), filter)
	if err != nil {
		return h.handleError(err, "cdek delivery points")
	}

	return c.JSON(response.OkByData(cdek_response.NewListFromDTO(points)))
}

// parseCDEKBBox parses the min_lat/max_lat/min_lon/max_lon query params into a
// bbox filter. It returns nil (no error) if none of them are set, and an
// error if only some are set or the bounds are invalid.
func parseCDEKBBox(c fiber.Ctx) (*dto.CDEKDeliveryPointsBBox, error) {
	raw := [4]string{c.Query("min_lat"), c.Query("max_lat"), c.Query("min_lon"), c.Query("max_lon")}

	set := 0
	for _, v := range raw {
		if v != "" {
			set++
		}
	}
	if set == 0 {
		return nil, nil
	}
	if set != len(raw) {
		return nil, fmt.Errorf("min_lat, max_lat, min_lon and max_lon must all be provided together")
	}

	names := [4]string{"min_lat", "max_lat", "min_lon", "max_lon"}
	vals := [4]float64{}
	for i, v := range raw {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", names[i], err)
		}
		vals[i] = f
	}

	bbox := &dto.CDEKDeliveryPointsBBox{MinLat: vals[0], MaxLat: vals[1], MinLon: vals[2], MaxLon: vals[3]}
	if bbox.MinLat > bbox.MaxLat || bbox.MinLon > bbox.MaxLon {
		return nil, fmt.Errorf("min_lat/min_lon must not exceed max_lat/max_lon")
	}

	return bbox, nil
}

func (h *Handler) initCDEKRoutes(v1 fiber.Router) {
	g := v1.Group("/cdek")
	g.Get("/delivery-points", h.listCDEKDeliveryPoints)
}
