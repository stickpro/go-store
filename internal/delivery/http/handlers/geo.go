package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/geo_response"
)

// getGeoLocation is a function get city by IP address
//
//	@Summary		Geo city
//	@Description	Get city by IP address
//	@Tags			Geo
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[geo_response.GeoResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/geo/city/ [get]
func (h *Handler) getGeoLocation(c fiber.Ctx) error {
	ip := c.IP()

	location, err := h.services.GeoService.GetCityByIP(ip)
	if err != nil {
		return h.handleError(err, "geo")
	}

	return c.JSON(response.OkByData(geo_response.GeoResponse{City: location}))
}

func (h *Handler) initGeoRoutes(v1 fiber.Router) {
	g := v1.Group("/geo")
	g.Get("/city", h.getGeoLocation)
}
