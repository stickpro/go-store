package tbank

import (
	"testing"

	"github.com/goccy/go-json"
)

func TestSign_DeterministicRegardlessOfInputOrder(t *testing.T) {
	a := sign("pwd", map[string]string{"TerminalKey": "TK", "Amount": "100", "OrderId": "o1"})
	b := sign("pwd", map[string]string{"OrderId": "o1", "Amount": "100", "TerminalKey": "TK"})
	if a != b {
		t.Fatalf("sign must not depend on map iteration order: %q != %q", a, b)
	}
}

func TestSign_ChangesWithAnyField(t *testing.T) {
	base := map[string]string{"TerminalKey": "TK", "Amount": "100", "OrderId": "o1"}
	baseSig := sign("pwd", base)

	cases := []map[string]string{
		{"TerminalKey": "TK2", "Amount": "100", "OrderId": "o1"},
		{"TerminalKey": "TK", "Amount": "101", "OrderId": "o1"},
		{"TerminalKey": "TK", "Amount": "100", "OrderId": "o2"},
	}
	for _, c := range cases {
		if sign("pwd", c) == baseSig {
			t.Fatalf("changing a field must change the signature: %v", c)
		}
	}
	if sign("other-pwd", base) == baseSig {
		t.Fatalf("changing the password must change the signature")
	}
}

func TestVerifyNotification_ValidToken(t *testing.T) {
	password := "super-secret"
	params := map[string]string{
		"TerminalKey": "TestTerminalKey",
		"OrderId":     "11111111-1111-1111-1111-111111111111",
		"Success":     "true",
		"Status":      statusConfirmed,
		"PaymentId":   "123456789",
		"ErrorCode":   "0",
		"Amount":      "19200",
	}
	token := sign(password, params)

	raw, err := json.Marshal(map[string]any{
		"TerminalKey": params["TerminalKey"],
		"OrderId":     params["OrderId"],
		"Success":     true,
		"Status":      params["Status"],
		"PaymentId":   123456789,
		"ErrorCode":   params["ErrorCode"],
		"Amount":      19200,
		"Token":       token,
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	n, valid, err := verifyNotification(password, raw)
	if err != nil {
		t.Fatalf("verifyNotification: %v", err)
	}
	if !valid {
		t.Fatalf("expected a valid token")
	}
	if n.OrderId != params["OrderId"] {
		t.Fatalf("OrderId = %q, want %q", n.OrderId, params["OrderId"])
	}
	if n.PaymentId.String() != "123456789" {
		t.Fatalf("PaymentId = %q, want %q", n.PaymentId.String(), "123456789")
	}
	if n.Status != statusConfirmed {
		t.Fatalf("Status = %q, want %q", n.Status, statusConfirmed)
	}
}

func TestVerifyNotification_TamperedAmountIsRejected(t *testing.T) {
	password := "super-secret"
	params := map[string]string{
		"TerminalKey": "TestTerminalKey",
		"OrderId":     "11111111-1111-1111-1111-111111111111",
		"Amount":      "19200",
	}
	token := sign(password, params)

	// An attacker bumps Amount after the token was computed, without
	// recomputing it.
	raw, err := json.Marshal(map[string]any{
		"TerminalKey": params["TerminalKey"],
		"OrderId":     params["OrderId"],
		"Amount":      990000,
		"Token":       token,
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	_, valid, err := verifyNotification(password, raw)
	if err != nil {
		t.Fatalf("verifyNotification: %v", err)
	}
	if valid {
		t.Fatalf("a token computed over the original Amount must not validate a tampered one")
	}
}

func TestVerifyNotification_NestedFieldsExcludedFromSignature(t *testing.T) {
	password := "super-secret"
	params := map[string]string{
		"TerminalKey": "TestTerminalKey",
		"OrderId":     "o1",
	}
	token := sign(password, params)

	// DATA is a nested object T-Bank may echo back; it must not be part of
	// the signature, so the token computed without it still validates.
	raw, err := json.Marshal(map[string]any{
		"TerminalKey": params["TerminalKey"],
		"OrderId":     params["OrderId"],
		"DATA":        map[string]any{"Phone": "+70000000000"},
		"Token":       token,
	})
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	_, valid, err := verifyNotification(password, raw)
	if err != nil {
		t.Fatalf("verifyNotification: %v", err)
	}
	if !valid {
		t.Fatalf("nested fields must be excluded from the signature")
	}
}

func TestMapStatus(t *testing.T) {
	cases := map[string]string{
		statusConfirmed:       "confirmed",
		statusRejected:        "failed",
		statusDeadlineExpired: "failed",
		statusCanceled:        "failed",
		statusReversed:        "refunded",
		statusRefunded:        "refunded",
		"NEW":                 "pending",
		"AUTHORIZED":          "pending",
	}
	for in, want := range cases {
		if got := mapStatus(in).String(); got != want {
			t.Errorf("mapStatus(%q) = %q, want %q", in, got, want)
		}
	}
}
