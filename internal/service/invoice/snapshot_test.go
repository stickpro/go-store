package invoice

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
)

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func testConfig() config.InvoiceConfig {
	return config.InvoiceConfig{
		NumberPrefix: "СЧ-",
		TTL:          120 * time.Hour,
		Seller: config.InvoiceSellerConfig{
			Name: `ООО "Ромашка"`, INN: "7707083893", KPP: "773601001", Address: "Москва, ул. Ленина, 1",
			BankName: "ПАО Сбербанк", BIK: "044525225", Account: "40702810938000000001",
			CorrAccount: "30101810400000000225", Director: "Иванов И.И.",
		},
	}
}

func testOrder() *dto.OrderDTO {
	kpp := "770101001"
	return &dto.OrderDTO{
		ID:       uuid.New(),
		Number:   1042,
		Currency: "RUB",
		LegalEntity: &dto.LegalEntityDTO{
			Name: `ООО "Покупатель"`, INN: "7736207543", KPP: &kpp, Address: "Москва, ул. Льва Толстого, 16",
		},
		Items: []dto.OrderItemDTO{
			{Name: "Чайник", UnitPrice: dec("1500"), Quantity: 2, LineTotal: dec("3000")},
			{Name: "Кружка", UnitPrice: dec("250.50"), Quantity: 4, LineTotal: dec("1002")},
		},
		ShippingTotal: dec("498"),
		GrandTotal:    dec("4500"),
	}
}

func TestBuildSnapshot(t *testing.T) {
	vat := dec("20")
	now := time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC)
	s := buildSnapshot(testConfig(), &vat, testOrder(), 17, now, now.Add(120*time.Hour))

	if s.Number != "СЧ-17" {
		t.Errorf("number = %q", s.Number)
	}
	if len(s.Lines) != 3 || s.Lines[2].Name != "Доставка" || !s.Lines[2].Total.Equal(dec("498")) || s.Lines[2].No != 3 {
		t.Fatalf("lines = %+v, want 2 items + shipping line", s.Lines)
	}
	if !s.VATAmount.Equal(dec("750")) { // 4500 * 20 / 120
		t.Errorf("vat = %s, want 750", s.VATAmount)
	}
	if s.AmountInWords != "Четыре тысячи пятьсот рублей 00 копеек" {
		t.Errorf("words = %q", s.AmountInWords)
	}
	if s.Seller.Accountant != "Иванов И.И." {
		t.Errorf("accountant should default to director, got %q", s.Seller.Accountant)
	}
	if s.Buyer.INN != "7736207543" || s.Buyer.KPP == nil {
		t.Errorf("buyer = %+v", s.Buyer)
	}
}

func TestBuildSnapshotNoVATNoShipping(t *testing.T) {
	o := testOrder()
	o.ShippingTotal = decimal.Zero
	o.GrandTotal = dec("4002")
	s := buildSnapshot(testConfig(), nil, o, 1, time.Now(), time.Now())
	if len(s.Lines) != 2 {
		t.Errorf("no shipping line expected, got %d lines", len(s.Lines))
	}
	if s.VATRate != nil || !s.VATAmount.IsZero() {
		t.Errorf("expected no VAT, got rate=%v amount=%s", s.VATRate, s.VATAmount)
	}
}

func TestSameContent(t *testing.T) {
	cfg := testConfig()
	t1 := time.Now()
	a := buildSnapshot(cfg, nil, testOrder(), 1, t1, t1)
	b := buildSnapshot(cfg, nil, testOrder(), 2, t1.Add(time.Hour), t1.Add(time.Hour))
	if !sameContent(a, b) {
		t.Error("number and dates must not count as a content change")
	}

	o := testOrder()
	o.Items[0].Quantity = 3
	if sameContent(a, buildSnapshot(cfg, nil, o, 1, t1, t1)) {
		t.Error("quantity change must count")
	}

	o = testOrder()
	o.LegalEntity.Address = "другой адрес"
	if sameContent(a, buildSnapshot(cfg, nil, o, 1, t1, t1)) {
		t.Error("buyer change must count")
	}

	vat := dec("20")
	if sameContent(a, buildSnapshot(cfg, &vat, testOrder(), 1, t1, t1)) {
		t.Error("VAT change must count")
	}
}

func TestRenderHTML(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	tmpl, err := parseTemplate("", loc)
	if err != nil {
		t.Fatal(err)
	}
	vat := dec("20")
	issued := time.Date(2026, 10, 30, 22, 30, 0, 0, time.UTC) // already Oct 31 in Moscow
	out, err := renderHTML(tmpl, buildSnapshot(testConfig(), &vat, testOrder(), 17, issued, issued.Add(120*time.Hour)))
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)
	for _, want := range []string{
		"Счёт на оплату № СЧ-17 от 31 октября 2026 г.",
		"044525225",
		"ИНН 7736207543, КПП 770101001",
		"Основание:",
		"Заказ № 1042",
		"4 500,00",
		"В том числе НДС (20%)",
		"Четыре тысячи пятьсот рублей 00 копеек",
		"Оплатить не позднее 05.11.2026",
		"Всего наименований 3",
		"&#34;Покупатель&#34;",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered html is missing %q", want)
		}
	}
}

func TestFormatMoney(t *testing.T) {
	for in, want := range map[string]string{
		"0":         "0,00",
		"12.5":      "12,50",
		"1234":      "1 234,00",
		"1234567.8": "1 234 567,80",
		"-1000":     "-1 000,00",
	} {
		if got := formatMoney(dec(in)); got != want {
			t.Errorf("formatMoney(%s) = %q, want %q", in, got, want)
		}
	}
}
