package tbank

import (
	"fmt"
	"strings"
)

// flexString decodes a JSON field that T-Bank sometimes sends as a number and
// sometimes as a string (PaymentId, in particular) into a plain string.
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	*f = flexString(strings.Trim(string(b), `"`))
	return nil
}

func (f flexString) String() string { return string(f) }

type initRequest struct {
	TerminalKey     string `json:"TerminalKey"`
	Amount          int64  `json:"Amount"`
	OrderId         string `json:"OrderId"`
	Description     string `json:"Description,omitempty"`
	NotificationURL string `json:"NotificationURL,omitempty"`
	SuccessURL      string `json:"SuccessURL,omitempty"`
	FailURL         string `json:"FailURL,omitempty"`
	Token           string `json:"Token"`
}

type initResponse struct {
	Success     bool       `json:"Success"`
	ErrorCode   string     `json:"ErrorCode"`
	Message     string     `json:"Message"`
	Details     string     `json:"Details"`
	TerminalKey string     `json:"TerminalKey"`
	Status      string     `json:"Status"`
	PaymentId   flexString `json:"PaymentId"`
	OrderId     string     `json:"OrderId"`
	Amount      int64      `json:"Amount"`
	PaymentURL  string     `json:"PaymentURL"`
}

func (r initResponse) err() error {
	if r.Success {
		return nil
	}
	return fmt.Errorf("tbank Init failed: code %s: %s (%s)", r.ErrorCode, r.Message, r.Details)
}

type cancelRequest struct {
	TerminalKey string `json:"TerminalKey"`
	PaymentId   string `json:"PaymentId"`
	Amount      int64  `json:"Amount,omitempty"`
	Token       string `json:"Token"`
}

type cancelResponse struct {
	Success     bool       `json:"Success"`
	ErrorCode   string     `json:"ErrorCode"`
	Message     string     `json:"Message"`
	Details     string     `json:"Details"`
	TerminalKey string     `json:"TerminalKey"`
	Status      string     `json:"Status"`
	PaymentId   flexString `json:"PaymentId"`
	OrigAmount  int64      `json:"OrigAmount"`
	NewAmount   int64      `json:"NewAmount"`
}

func (r cancelResponse) err() error {
	if r.Success {
		return nil
	}
	return fmt.Errorf("tbank Cancel failed: code %s: %s (%s)", r.ErrorCode, r.Message, r.Details)
}

// notification is the webhook body T-Bank POSTs on every payment status
// change. OrderId echoes back whatever Init was called with — here, our own
// payments.id — which is how it's matched back to a row without depending on
// PaymentId having been persisted first.
type notification struct {
	TerminalKey string     `json:"TerminalKey"`
	OrderId     string     `json:"OrderId"`
	Success     bool       `json:"Success"`
	Status      string     `json:"Status"`
	PaymentId   flexString `json:"PaymentId"`
	ErrorCode   string     `json:"ErrorCode"`
	Amount      int64      `json:"Amount"`
	Token       string     `json:"Token"`
}

// T-Bank payment statuses relevant to a one-stage (PayType=O) payment. See
// https://www.tbank.ru/kassa/dev/payments/#section/Statusy-platezha
const (
	statusConfirmed       = "CONFIRMED"
	statusRejected        = "REJECTED"
	statusDeadlineExpired = "DEADLINE_EXPIRED"
	statusCanceled        = "CANCELED"
	statusReversed        = "REVERSED"
	statusPartialReversed = "PARTIAL_REVERSED"
	statusRefunded        = "REFUNDED"
	statusPartialRefunded = "PARTIAL_REFUNDED"
)
