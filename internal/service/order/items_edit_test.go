package order

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/internal/storage/repository/repository_products"
)

func orderItem(product, variant uuid.UUID, price string, qty int64) *models.OrderItem {
	p := decimal.RequireFromString(price)
	return &models.OrderItem{
		ID:        uuid.New(),
		ProductID: uuid.NullUUID{UUID: product, Valid: true},
		VariantID: uuid.NullUUID{UUID: variant, Valid: true},
		Name:      "item",
		UnitPrice: p,
		Quantity:  qty,
		LineTotal: p.Mul(decimal.NewFromInt(qty)),
	}
}

func catalogueRow(product, variant uuid.UUID, price string, stock int64) *repository_products.GetOrderLinesByVariantIDsRow {
	return &repository_products.GetOrderLinesByVariantIDsRow{
		ProductID:      product,
		VariantID:      variant,
		PriceRetail:    decimal.RequireFromString(price),
		StockQuantity:  stock,
		Subtract:       true,
		Minimum:        1,
		ProductEnabled: true,
		VariantEnabled: true,
		Name:           "new item",
	}
}

func TestPlanItemsEdit(t *testing.T) {
	pA, vA := uuid.New(), uuid.New()
	pB, vB := uuid.New(), uuid.New()
	pC, vC := uuid.New(), uuid.New()

	items := []*models.OrderItem{
		orderItem(pA, vA, "100", 3),
		orderItem(pB, vB, "50", 2),
	}

	t.Run("no change", func(t *testing.T) {
		plan, err := planItemsEdit(items, []dto.OrderLineEditDTO{{VariantID: vA, Quantity: 3}, {VariantID: vB, Quantity: 2}}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if plan.changed() {
			t.Fatalf("expected no change: %+v", plan)
		}
	})

	t.Run("decrease, remove and add", func(t *testing.T) {
		// Catalogue price of A moved to 120 — a kept line must keep its 100.
		rows := map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow{
			vC: catalogueRow(pC, vC, "30", 10),
		}
		plan, err := planItemsEdit(items, []dto.OrderLineEditDTO{
			{VariantID: vA, Quantity: 1},
			{VariantID: vC, Quantity: 4},
		}, rows, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.update) != 1 || plan.update[0].quantity != 1 {
			t.Fatalf("update: %+v", plan.update)
		}
		if len(plan.remove) != 1 || plan.remove[0].VariantID.UUID != vB {
			t.Fatalf("remove: %+v", plan.remove)
		}
		if len(plan.add) != 1 || !plan.add[0].unitPrice.Equal(decimal.NewFromInt(30)) {
			t.Fatalf("add: %+v", plan.add)
		}
		if got := sumSubtotal(plan.final); !got.Equal(decimal.NewFromInt(100 + 120)) {
			t.Fatalf("subtotal = %s", got)
		}
		want := map[uuid.UUID]int64{pA: -2, pB: -2, pC: 4}
		for p, d := range want {
			if plan.stockDelta[p] != d {
				t.Errorf("stock delta %s = %d, want %d", p, plan.stockDelta[p], d)
			}
		}
	})

	t.Run("increase keeps sold-at price", func(t *testing.T) {
		row := catalogueRow(pA, vA, "120", 10)
		plan, err := planItemsEdit(items, []dto.OrderLineEditDTO{
			{VariantID: vA, Quantity: 5},
			{VariantID: vB, Quantity: 2},
		}, map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow{vA: row}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := sumSubtotal(plan.final); !got.Equal(decimal.NewFromInt(500 + 100)) {
			t.Fatalf("subtotal = %s", got)
		}
		if plan.stockDelta[pA] != 2 {
			t.Fatalf("delta = %d", plan.stockDelta[pA])
		}
	})

	t.Run("two variants of one product net out", func(t *testing.T) {
		vA2 := uuid.New()
		rows := map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow{vA2: catalogueRow(pA, vA2, "100", 0)}
		plan, err := planItemsEdit(items, []dto.OrderLineEditDTO{
			{VariantID: vA2, Quantity: 3},
			{VariantID: vB, Quantity: 2},
		}, rows, nil)
		if err != nil {
			t.Fatal(err)
		}
		if plan.stockDelta[pA] != 0 {
			t.Fatalf("swap within a product should not move stock, delta = %d", plan.stockDelta[pA])
		}
	})

	t.Run("added variant must be orderable", func(t *testing.T) {
		row := catalogueRow(pC, vC, "30", 10)
		row.VariantEnabled = false
		_, err := planItemsEdit(items, []dto.OrderLineEditDTO{{VariantID: vC, Quantity: 1}},
			map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow{vC: row}, nil)
		var le *LineError
		if !errors.As(err, &le) || le.Reason != LineUnavailable {
			t.Fatalf("want unavailable, got %v", err)
		}

		_, err = planItemsEdit(items, []dto.OrderLineEditDTO{{VariantID: uuid.New(), Quantity: 1}}, nil, nil)
		if !errors.As(err, &le) || le.Reason != LineUnavailable {
			t.Fatalf("unknown variant: want unavailable, got %v", err)
		}
	})

	t.Run("shrinking a line of a since-deleted product is allowed", func(t *testing.T) {
		plan, err := planItemsEdit(items, []dto.OrderLineEditDTO{{VariantID: vA, Quantity: 2}}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !plan.changed() {
			t.Fatal("expected a change")
		}
	})

	t.Run("minimum applies to lines that grow", func(t *testing.T) {
		row := catalogueRow(pC, vC, "30", 10)
		row.Minimum = 5
		_, err := planItemsEdit(items, []dto.OrderLineEditDTO{{VariantID: vC, Quantity: 2}},
			map[uuid.UUID]*repository_products.GetOrderLinesByVariantIDsRow{vC: row}, nil)
		var le *LineError
		if !errors.As(err, &le) || le.Reason != LineBelowMinimum {
			t.Fatalf("want below minimum, got %v", err)
		}
	})
}

func TestMergeEditLines(t *testing.T) {
	v1, v2 := uuid.New(), uuid.New()
	got, err := mergeEditLines([]dto.OrderLineEditDTO{{VariantID: v1, Quantity: 1}, {VariantID: v2, Quantity: 2}, {VariantID: v1, Quantity: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].VariantID != v1 || got[0].Quantity != 4 || got[1].Quantity != 2 {
		t.Fatalf("%+v", got)
	}
	if _, err := mergeEditLines(nil); !errors.Is(err, ErrNoLines) {
		t.Fatalf("want ErrNoLines, got %v", err)
	}
}
