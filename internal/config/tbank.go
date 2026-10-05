package config

import "time"

// TBankConfig configures the T-Bank (Т-Банк) acquiring integration (Init/
// Notification API, securepay.tbank.ru/v2). TerminalKey/Password are issued
// in the T-Bank merchant cabinet for the terminal.
type TBankConfig struct {
	Enabled     bool   `yaml:"enabled" default:"false" usage:"enable the T-Bank acquiring integration"`
	TerminalKey string `yaml:"terminal_key" usage:"T-Bank terminal key"`
	Password    string `yaml:"password" secret:"true" usage:"T-Bank terminal password, used to sign/verify requests"`
	BaseURL     string `yaml:"base_url" default:"https://securepay.tbank.ru/v2" usage:"T-Bank API base URL"`
	// NotificationURL is this store's public webhook endpoint
	// (/api/v1/payments/tbank/notification), sent on every Init call so T-Bank
	// knows where to POST payment status updates. Required unless a default
	// notification URL is configured on the terminal itself.
	NotificationURL string        `yaml:"notification_url" usage:"public URL T-Bank should POST payment notifications to"`
	SuccessURL      string        `yaml:"success_url" usage:"куда редиректить после успешной оплаты"`
	FailURL         string        `yaml:"fail_url" usage:"куда редиректить после неуспешной оплаты"`
	RequestTimeout  time.Duration `yaml:"request_timeout" default:"15s" usage:"HTTP timeout for a single T-Bank API request"`
}
