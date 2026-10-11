package topics

const (
	Products = "store.products"
	// Orders carries order events from the store to the ERP.
	Orders = "store.orders"
	// ErpOrders carries order updates (status, tracking) from the ERP to the store.
	ErpOrders = "erp.orders"
)
