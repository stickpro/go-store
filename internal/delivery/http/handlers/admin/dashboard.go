package admin

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/stickpro/go-store/internal/delivery/http/request/dashboard_request"
	"github.com/stickpro/go-store/internal/delivery/http/response"
	"github.com/stickpro/go-store/internal/delivery/http/response/dashboard_response"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/service/dashboard"
	"github.com/stickpro/go-store/internal/tools/apierror"
)

// getDashboard returns the admin overview snapshot in one request: order and
// revenue counters, catalogue and customer totals.
//
//	@Summary		Dashboard overview
//	@Description	Aggregated admin overview. Order status/payment buckets and totals are all-time; `today` uses the store-timezone day; `from`/`to` (RFC3339, default today) drive revenue.period and average_order_value. Revenue counts payment_status=paid orders only.
//	@Tags			Admin Dashboard
//	@Accept			json
//	@Produce		json
//	@Param			request	query		dashboard_request.OverviewRequest	false	"Period"
//	@Success		200		{object}	response.Result[dashboard_response.DashboardResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Router			/v1/admin/dashboard [get]
//	@Security		BearerAuth
func (h *Handler) getDashboard(c fiber.Ctx) error {
	req := &dashboard_request.OverviewRequest{}
	if err := c.Bind().Query(req); err != nil {
		return err
	}

	from, to, err := dto.RequestToDashboardPeriod(req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	d, err := h.services.DashboardService.Overview(c.Context(), from, to)
	if err != nil {
		if errors.Is(err, dashboard.ErrInvalidPeriod) {
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
		}
		return h.handleError(err, "dashboard")
	}

	return c.JSON(response.OkByData(dashboard_response.NewFromDTO(d)))
}

func (h *Handler) initDashboardRoutes(v1 fiber.Router) {
	v1.Get("/admin/dashboard", h.getDashboard)
}
