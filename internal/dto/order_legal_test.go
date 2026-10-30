package dto

import (
	"errors"
	"testing"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/delivery/http/request/order_request"
)

func sp(s string) *string { return &s }

func TestLegalEntityFromRequest(t *testing.T) {
	cases := []struct {
		name string
		req  order_request.CreateOrderRequest
		want error
		typ  string
	}{
		{"individual by default", order_request.CreateOrderRequest{}, nil, constant.CustomerIndividual},
		{"legal without block", order_request.CreateOrderRequest{CustomerType: "legal"}, ErrLegalEntityRequired, ""},
		{"company ok", order_request.CreateOrderRequest{CustomerType: "legal", LegalEntity: &order_request.LegalEntityRequest{
			Name: "ООО", INN: "7707083893", KPP: sp("773601001"), OGRN: sp("1027700132195"), Address: "Москва",
		}}, nil, constant.CustomerLegal},
		{"company without kpp", order_request.CreateOrderRequest{CustomerType: "legal", LegalEntity: &order_request.LegalEntityRequest{
			Name: "ООО", INN: "7707083893", Address: "Москва",
		}}, ErrInvalidLegalEntity, ""},
		{"company with ogrnip", order_request.CreateOrderRequest{CustomerType: "legal", LegalEntity: &order_request.LegalEntityRequest{
			Name: "ООО", INN: "7707083893", KPP: sp("773601001"), OGRN: sp("304500116000157"), Address: "Москва",
		}}, ErrInvalidLegalEntity, ""},
		{"sole proprietor ok", order_request.CreateOrderRequest{CustomerType: "legal", LegalEntity: &order_request.LegalEntityRequest{
			Name: "ИП Петров", INN: "500100732259", OGRN: sp("304500116000157"), Address: "Москва",
		}}, nil, constant.CustomerLegal},
		{"sole proprietor with kpp", order_request.CreateOrderRequest{CustomerType: "legal", LegalEntity: &order_request.LegalEntityRequest{
			Name: "ИП Петров", INN: "500100732259", KPP: sp("773601001"), Address: "Москва",
		}}, ErrInvalidLegalEntity, ""},
		{"blank kpp counts as empty", order_request.CreateOrderRequest{CustomerType: "legal", LegalEntity: &order_request.LegalEntityRequest{
			Name: "ИП Петров", INN: "500100732259", KPP: sp("  "), Address: "Москва",
		}}, nil, constant.CustomerLegal},
	}
	for _, c := range cases {
		typ, le, err := legalEntityFromRequest(&c.req)
		if !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
			continue
		}
		if err == nil && typ != c.typ {
			t.Errorf("%s: type = %q, want %q", c.name, typ, c.typ)
		}
		if err == nil && (typ == constant.CustomerLegal) != (le != nil) {
			t.Errorf("%s: legal entity presence mismatch", c.name)
		}
	}
}
