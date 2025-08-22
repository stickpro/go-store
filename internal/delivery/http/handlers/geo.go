package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/geo_response"
	"github.com/stickpro/go-store/internal/tools/apierror"

	// swag gen import
	_ "github.com/stickpro/go-store/internal/models"
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
		return c.JSON(response.OkByData(geo_response.GeoResponse{City: "Москва"}))
	}

	return c.JSON(response.OkByData(geo_response.GeoResponse{City: location}))
}

// findCity is a function find city by name
//
//	@Summary		Find city
//	@Description	Find city by name
//	@Tags			Geo
//	@Accept			json
//	@Produce		json
//	@Param			city	query		string	true	"City name"
//	@Success		200		{object}	response.Result[[]models.City]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/geo/city/find [get]
func (h *Handler) findCity(c fiber.Ctx) error {
	city := c.Query("city")
	if city == "" {
		return apierror.New().AddError(fmt.Errorf("city is requered")).SetHttpCode(fiber.StatusBadRequest)
	}
	location, err := h.services.SearchService.Search(constant.CitiesIndex, city, 10, 0)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(location.Hits))
}

// getPopularCity is a function return most popular city
//
//	@Summary		Get popular city
//	@Description	Get most popular city
//	@Tags			Geo
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[geo_response.CityResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/geo/city/popular [get]
func (h *Handler) getPopularCity(c fiber.Ctx) error {
	cities, err := h.services.GeoService.GetPopularCity(c.Context())
	if err != nil {
		return h.handleError(err, "cities")
	}
	return c.JSON(response.OkByData(geo_response.NewFromModels(cities)))
}

func (h *Handler) initGeoRoutes(v1 fiber.Router) {
	g := v1.Group("/geo")
	g.Get("/city", h.getGeoLocation)
	g.Get("/city/find", h.findCity)
	g.Get("/city/popular", h.getPopularCity)
}
