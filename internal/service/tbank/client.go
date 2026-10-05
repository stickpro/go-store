package tbank

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/goccy/go-json"

	"github.com/stickpro/go-store/internal/config"
)

// client talks to the T-Bank acquiring API (Init/Cancel). Signature
// generation lives in sign.go; this file only builds requests and performs
// the HTTP round-trip.
type client struct {
	cfg  config.TBankConfig
	http *http.Client
}

func newClient(cfg config.TBankConfig) *client {
	return &client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.RequestTimeout},
	}
}

type initCall struct {
	OrderID         string
	AmountKopecks   int64
	Description     string
	NotificationURL string
	SuccessURL      string
	FailURL         string
}

// init starts a one-stage payment. raw is always returned (even on a gateway
// rejection) so the caller can keep it for audit.
func (c *client) init(ctx context.Context, call initCall) (*initResponse, []byte, error) {
	params := map[string]string{
		"TerminalKey": c.cfg.TerminalKey,
		"Amount":      strconv.FormatInt(call.AmountKopecks, 10),
		"OrderId":     call.OrderID,
	}
	if call.Description != "" {
		params["Description"] = call.Description
	}
	if call.NotificationURL != "" {
		params["NotificationURL"] = call.NotificationURL
	}
	if call.SuccessURL != "" {
		params["SuccessURL"] = call.SuccessURL
	}
	if call.FailURL != "" {
		params["FailURL"] = call.FailURL
	}

	body := initRequest{
		TerminalKey:     c.cfg.TerminalKey,
		Amount:          call.AmountKopecks,
		OrderId:         call.OrderID,
		Description:     call.Description,
		NotificationURL: call.NotificationURL,
		SuccessURL:      call.SuccessURL,
		FailURL:         call.FailURL,
		Token:           sign(c.cfg.Password, params),
	}

	var resp initResponse
	raw, err := c.doPost(ctx, "/Init", body, &resp)
	if err != nil {
		return nil, raw, err
	}
	if err := resp.err(); err != nil {
		return &resp, raw, err
	}
	return &resp, raw, nil
}

// cancel reverses or refunds a payment. amountKopecks of 0 means "in full".
func (c *client) cancel(ctx context.Context, providerPaymentID string, amountKopecks int64) (*cancelResponse, []byte, error) {
	params := map[string]string{
		"TerminalKey": c.cfg.TerminalKey,
		"PaymentId":   providerPaymentID,
	}
	if amountKopecks > 0 {
		params["Amount"] = strconv.FormatInt(amountKopecks, 10)
	}

	body := cancelRequest{
		TerminalKey: c.cfg.TerminalKey,
		PaymentId:   providerPaymentID,
		Amount:      amountKopecks,
		Token:       sign(c.cfg.Password, params),
	}

	var resp cancelResponse
	raw, err := c.doPost(ctx, "/Cancel", body, &resp)
	if err != nil {
		return nil, raw, err
	}
	if err := resp.err(); err != nil {
		return &resp, raw, err
	}
	return &resp, raw, nil
}

func (c *client) doPost(ctx context.Context, path string, body, out any) ([]byte, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("tbank: marshal %s request: %w", path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("tbank: build %s request: %w", path, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tbank: request %s: %w", path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tbank: read %s response: %w", path, err)
	}
	if resp.StatusCode != http.StatusOK {
		return raw, fmt.Errorf("tbank: %s: unexpected status %d", path, resp.StatusCode)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return raw, fmt.Errorf("tbank: decode %s response: %w", path, err)
	}
	return raw, nil
}
