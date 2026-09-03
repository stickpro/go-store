// Package mailer provides a small SMTP client for transactional email.
//
// It intentionally depends only on the standard library (net/smtp). A single
// message is delivered per call over a fresh connection, which is enough for the
// current volume of transactional mail (welcome, verification, …).
package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// dialTimeout bounds the whole send (connect + handshake + data).
const dialTimeout = 15 * time.Second

// Config holds SMTP connection and envelope settings.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string // envelope + From header address
	FromName string // optional display name for the From header
}

func (c Config) addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// Message is a single outgoing email. Body is treated as UTF-8 HTML.
type Message struct {
	To      []string
	Subject string
	HTML    string
}

// Mailer sends transactional email.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

// Noop discards every message. Used when email delivery is disabled in config.
type Noop struct{}

func (Noop) Send(context.Context, Message) error { return nil }

// SMTP delivers messages through an SMTP server.
type SMTP struct {
	cfg Config
}

// NewSMTP builds an SMTP mailer. It does not open a connection.
func NewSMTP(cfg Config) *SMTP {
	return &SMTP{cfg: cfg}
}

// Send delivers a single message. STARTTLS and AUTH are used only when the
// server advertises them, which keeps local servers such as maildev working
// without credentials.
func (s *SMTP) Send(ctx context.Context, msg Message) error {
	if len(msg.To) == 0 {
		return errors.New("mailer: no recipients")
	}

	ctx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", s.cfg.addr())
	if err != nil {
		return fmt.Errorf("mailer: dial %s: %w", s.cfg.addr(), err)
	}

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("mailer: smtp handshake: %w", err)
	}
	defer func() { _ = client.Close() }()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.cfg.Host}); err != nil {
			return fmt.Errorf("mailer: starttls: %w", err)
		}
	}

	if s.cfg.Username != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("mailer: auth: %w", err)
			}
		}
	}

	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("mailer: MAIL FROM: %w", err)
	}
	for _, rcpt := range msg.To {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("mailer: RCPT TO %s: %w", rcpt, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("mailer: DATA: %w", err)
	}
	if _, err := w.Write(s.buildMIME(msg)); err != nil {
		return fmt.Errorf("mailer: write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("mailer: close body: %w", err)
	}

	return client.Quit()
}

func (s *SMTP) buildMIME(msg Message) []byte {
	from := s.cfg.From
	if s.cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.From)
	}

	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + strings.Join(msg.To, ", ") + "\r\n")
	b.WriteString("Subject: " + encodeHeader(msg.Subject) + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(msg.HTML)

	return []byte(b.String())
}

// encodeHeader RFC 2047-encodes a header value when it contains non-ASCII bytes
// so subjects with Cyrillic text survive the transport.
func encodeHeader(v string) string {
	ascii := true
	for _, r := range v {
		if r > 127 {
			ascii = false
			break
		}
	}
	if ascii {
		return v
	}
	return mime.QEncoding.Encode("UTF-8", v)
}
