package order

import (
	"testing"

	"github.com/stickpro/go-store/internal/constant"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
)

func TestQuickOrderToCreateOrderDTO(t *testing.T) {
	email := "guest@example.com"
	q := dto.CreateQuickOrderDTO{
		Name:  "Иван Петров",
		Phone: "+79001234567",
		Email: &email,
	}

	d := q.ToCreateOrderDTO()

	if !d.Quick {
		t.Fatal("Quick must be set")
	}
	if d.ShipRecipient != "Иван Петров" {
		t.Fatalf("ShipRecipient = %q, want the customer name", d.ShipRecipient)
	}
	if d.Phone == nil || *d.Phone != "+79001234567" {
		t.Fatalf("Phone = %v, want the request phone", d.Phone)
	}
	if d.Email != email {
		t.Fatalf("Email = %q, want %q", d.Email, email)
	}
	if d.ExpectedTotal != nil || d.DeliveryMethodCode != nil || d.ShipProvider != nil {
		t.Fatal("quick order carries no delivery choice or price guard")
	}
}

func TestQuickOrderAccountEmailWins(t *testing.T) {
	reqEmail := "typed@example.com"
	q := dto.CreateQuickOrderDTO{
		User:  &models.User{Email: "account@example.com"},
		Phone: "+79000000000",
		Email: &reqEmail,
	}
	if got := q.ToCreateOrderDTO().Email; got != "account@example.com" {
		t.Fatalf("Email = %q, want the account email", got)
	}
}

func TestQuickOrderGuestWithoutEmail(t *testing.T) {
	q := dto.CreateQuickOrderDTO{Name: "No Email", Phone: "+79000000000"}
	if got := q.ToCreateOrderDTO().Email; got != "" {
		t.Fatalf("Email = %q, want empty (phone is the contact)", got)
	}
}

func TestInitialStatusAndSource(t *testing.T) {
	if initialStatus(dto.CreateOrderDTO{Quick: true}) != constant.OrderNew {
		t.Error("quick order must start in 'new'")
	}
	if initialStatus(dto.CreateOrderDTO{}) != constant.OrderPending {
		t.Error("normal checkout must start in 'pending'")
	}
	if orderSource(dto.CreateOrderDTO{Quick: true}) != constant.OrderSourceQuick {
		t.Error("quick order source must be 'quick'")
	}
	if orderSource(dto.CreateOrderDTO{}) != constant.OrderSourceCheckout {
		t.Error("normal checkout source must be 'checkout'")
	}
}

func TestQuickOrderTransitions(t *testing.T) {
	if !canTransition(constant.OrderNew, constant.OrderPending) {
		t.Error("new -> pending must be allowed (manager confirms the quick order)")
	}
	if !canTransition(constant.OrderNew, constant.OrderCancelled) {
		t.Error("new -> cancelled must be allowed")
	}
	if canTransition(constant.OrderNew, constant.OrderPaid) {
		t.Error("new -> paid must be rejected (goes through pending first)")
	}
}
