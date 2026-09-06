package dashboard_request

// OverviewRequest is the optional period for the admin dashboard. Both bounds
// are RFC3339 (e.g. "2026-01-01T00:00:00Z"); an empty request defaults to the
// current day in the store timezone.
type OverviewRequest struct {
	From *string `json:"from" query:"from"`
	To   *string `json:"to" query:"to"`
} //	@name	DashboardOverviewRequest
