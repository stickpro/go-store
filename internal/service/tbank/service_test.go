package tbank

import "testing"

func TestWithOrderParam(t *testing.T) {
	cases := []struct {
		name  string
		raw   string
		order int64
		want  string
	}{
		{"empty url stays empty", "", 123, ""},
		{"plain url gets the param", "https://shop.ru/payment/success", 123, "https://shop.ru/payment/success?order=123"},
		{"existing query is preserved", "https://shop.ru/payment/success?ref=email", 123, "https://shop.ru/payment/success?order=123&ref=email"},
		{"existing order param is overwritten", "https://shop.ru/payment/success?order=999", 123, "https://shop.ru/payment/success?order=123"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := withOrderParam(c.raw, c.order); got != c.want {
				t.Errorf("withOrderParam(%q, %d) = %q, want %q", c.raw, c.order, got, c.want)
			}
		})
	}
}
