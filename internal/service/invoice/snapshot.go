package invoice

import (
	"strconv"
	"time"

	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/pkg/moneywords"
)

// Snapshot is everything printed on an invoice, frozen at issue time and
// stored in invoices.data. The PDF is always rendered from it, so a later
// change to the order, the catalogue or the seller config never alters an
// issued invoice.
type Snapshot struct {
	Number      string    `json:"number"`
	IssuedAt    time.Time `json:"issued_at"`
	DueAt       time.Time `json:"due_at"`
	OrderNumber int64     `json:"order_number"`
	Currency    string    `json:"currency"`

	Seller Party `json:"seller"`
	Buyer  Party `json:"buyer"`

	Lines    []Line          `json:"lines"`
	Discount decimal.Decimal `json:"discount"`
	Total    decimal.Decimal `json:"total"`
	// VATRate is nil when the seller is not a VAT payer ("Без НДС").
	VATRate       *decimal.Decimal `json:"vat_rate,omitempty"`
	VATAmount     decimal.Decimal  `json:"vat_amount"`
	AmountInWords string           `json:"amount_in_words"`
}

// Party is the seller or buyer block. Bank fields are filled for the seller only.
type Party struct {
	Name        string  `json:"name"`
	INN         string  `json:"inn"`
	KPP         *string `json:"kpp,omitempty"`
	OGRN        *string `json:"ogrn,omitempty"`
	Address     string  `json:"address"`
	Phone       *string `json:"phone,omitempty"`
	BankName    string  `json:"bank_name,omitempty"`
	BIK         string  `json:"bik,omitempty"`
	Account     string  `json:"account,omitempty"`
	CorrAccount string  `json:"corr_account,omitempty"`
	Director    string  `json:"director,omitempty"`
	Accountant  string  `json:"accountant,omitempty"`
}

// Line is one invoice row.
type Line struct {
	No       int             `json:"no"`
	Name     string          `json:"name"`
	Quantity int64           `json:"quantity"`
	Unit     string          `json:"unit"`
	Price    decimal.Decimal `json:"price"`
	Total    decimal.Decimal `json:"total"`
}

const (
	unitPiece   = "шт"
	unitService = "усл"
)

// buildSnapshot assembles the invoice content for an order. o must be a legal
// order (LegalEntity set). Shipping, when charged, is a separate line.
func buildSnapshot(cfg config.InvoiceConfig, vat *decimal.Decimal, o *dto.OrderDTO, number int64, issuedAt, dueAt time.Time) Snapshot {
	lines := make([]Line, 0, len(o.Items)+1)
	for _, it := range o.Items {
		lines = append(lines, Line{
			No:       len(lines) + 1,
			Name:     it.Name,
			Quantity: it.Quantity,
			Unit:     unitPiece,
			Price:    it.UnitPrice,
			Total:    it.LineTotal,
		})
	}
	if o.ShippingTotal.IsPositive() {
		lines = append(lines, Line{
			No:       len(lines) + 1,
			Name:     "Доставка",
			Quantity: 1,
			Unit:     unitService,
			Price:    o.ShippingTotal,
			Total:    o.ShippingTotal,
		})
	}

	s := cfg.Seller
	seller := Party{
		Name:        s.Name,
		INN:         s.INN,
		KPP:         optional(s.KPP),
		OGRN:        optional(s.OGRN),
		Address:     s.Address,
		Phone:       optional(s.Phone),
		BankName:    s.BankName,
		BIK:         s.BIK,
		Account:     s.Account,
		CorrAccount: s.CorrAccount,
		Director:    s.Director,
		Accountant:  s.Accountant,
	}
	if seller.Accountant == "" {
		seller.Accountant = seller.Director
	}

	var buyer Party
	if le := o.LegalEntity; le != nil {
		buyer = Party{Name: le.Name, INN: le.INN, KPP: le.KPP, OGRN: le.OGRN, Address: le.Address, Phone: o.Phone}
	}

	return Snapshot{
		Number:        cfg.NumberPrefix + strconv.FormatInt(number, 10),
		IssuedAt:      issuedAt,
		DueAt:         dueAt,
		OrderNumber:   o.Number,
		Currency:      o.Currency,
		Seller:        seller,
		Buyer:         buyer,
		Lines:         lines,
		Discount:      o.DiscountTotal,
		Total:         o.GrandTotal,
		VATRate:       vat,
		VATAmount:     vatIncluded(o.GrandTotal, vat),
		AmountInWords: moneywords.Rubles(o.GrandTotal),
	}
}

// vatIncluded is the VAT contained in a VAT-inclusive total:
// total * rate / (100 + rate), rounded to kopecks.
func vatIncluded(total decimal.Decimal, rate *decimal.Decimal) decimal.Decimal {
	if rate == nil || !rate.IsPositive() {
		return decimal.Zero
	}
	hundred := decimal.NewFromInt(100)
	return total.Mul(*rate).Div(hundred.Add(*rate)).Round(2)
}

// sameContent reports whether two snapshots print the same document apart
// from number and dates — i.e. whether an issued invoice is still valid for
// the order as it stands.
func sameContent(a, b Snapshot) bool {
	if !a.Total.Equal(b.Total) || !a.Discount.Equal(b.Discount) || a.Currency != b.Currency ||
		len(a.Lines) != len(b.Lines) || !partyEqual(a.Seller, b.Seller) || !partyEqual(a.Buyer, b.Buyer) {
		return false
	}
	if (a.VATRate == nil) != (b.VATRate == nil) || (a.VATRate != nil && !a.VATRate.Equal(*b.VATRate)) {
		return false
	}
	for i := range a.Lines {
		x, y := a.Lines[i], b.Lines[i]
		if x.Name != y.Name || x.Quantity != y.Quantity || x.Unit != y.Unit ||
			!x.Price.Equal(y.Price) || !x.Total.Equal(y.Total) {
			return false
		}
	}
	return true
}

func partyEqual(a, b Party) bool {
	return a.Name == b.Name && a.INN == b.INN && strEq(a.KPP, b.KPP) && strEq(a.OGRN, b.OGRN) &&
		a.Address == b.Address && strEq(a.Phone, b.Phone) && a.BankName == b.BankName && a.BIK == b.BIK &&
		a.Account == b.Account && a.CorrAccount == b.CorrAccount && a.Director == b.Director &&
		a.Accountant == b.Accountant
}

func strEq(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
