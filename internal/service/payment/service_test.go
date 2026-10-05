package payment

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/models"
)

func paymentRow(provider, status, amount string) *models.Payment {
	return &models.Payment{
		Provider: provider,
		Status:   status,
		Amount:   decimal.RequireFromString(amount),
	}
}

func TestIsReusablePending(t *testing.T) {
	cases := []struct {
		name     string
		latest   *models.Payment
		provider string
		amount   string
		want     bool
	}{
		{
			name:     "same provider, pending, same amount is reusable",
			latest:   paymentRow("tbank", StatusPending.String(), "1500.00"),
			provider: "tbank",
			amount:   "1500.00",
			want:     true,
		},
		{
			name:     "confirmed attempt is not reusable",
			latest:   paymentRow("tbank", StatusConfirmed.String(), "1500.00"),
			provider: "tbank",
			amount:   "1500.00",
			want:     false,
		},
		{
			name:     "failed attempt is not reusable",
			latest:   paymentRow("tbank", StatusFailed.String(), "1500.00"),
			provider: "tbank",
			amount:   "1500.00",
			want:     false,
		},
		{
			name:     "different provider is not reusable",
			latest:   paymentRow("tbank", StatusPending.String(), "1500.00"),
			provider: "yookassa",
			amount:   "1500.00",
			want:     false,
		},
		{
			name:     "changed amount is not reusable",
			latest:   paymentRow("tbank", StatusPending.String(), "1500.00"),
			provider: "tbank",
			amount:   "1600.00",
			want:     false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := isReusablePending(c.latest, c.provider, decimal.RequireFromString(c.amount))
			if got != c.want {
				t.Errorf("isReusablePending() = %v, want %v", got, c.want)
			}
		})
	}
}
