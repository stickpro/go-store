package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// InvoiceDTO is an issued invoice for a "pay by invoice" order. Number is the
// printed number (with the configured prefix).
type InvoiceDTO struct {
	ID       uuid.UUID
	OrderID  uuid.UUID
	Number   string
	Status   string
	Amount   decimal.Decimal
	Currency string
	IssuedAt time.Time
	DueAt    time.Time
}

// InvoicePDF is a rendered invoice document.
type InvoicePDF struct {
	FileName string
	Content  []byte
}
