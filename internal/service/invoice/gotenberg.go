package invoice

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// maxPDFSize caps the rendered PDF read into memory.
const maxPDFSize = 20 << 20

// PDFRenderer turns a self-contained HTML document into a PDF.
type PDFRenderer interface {
	RenderPDF(ctx context.Context, html []byte) ([]byte, error)
}

// gotenberg renders HTML through a Gotenberg service
// (https://gotenberg.dev, route /forms/chromium/convert/html).
type gotenberg struct {
	baseURL string
	client  *http.Client
}

func newGotenberg(baseURL string, timeout time.Duration) *gotenberg {
	return &gotenberg{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (g *gotenberg) RenderPDF(ctx context.Context, html []byte) ([]byte, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	fw, err := w.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, fmt.Errorf("invoice: gotenberg form: %w", err)
	}
	if _, err := fw.Write(html); err != nil {
		return nil, fmt.Errorf("invoice: gotenberg form: %w", err)
	}
	// A4 portrait; margins come from the template's own CSS.
	for k, v := range map[string]string{
		"paperWidth":      "8.27",
		"paperHeight":     "11.7",
		"marginTop":       "0",
		"marginBottom":    "0",
		"marginLeft":      "0",
		"marginRight":     "0",
		"printBackground": "true",
	} {
		if err := w.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("invoice: gotenberg form: %w", err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("invoice: gotenberg form: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/forms/chromium/convert/html", &body)
	if err != nil {
		return nil, fmt.Errorf("invoice: gotenberg request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRenderUnavailable, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxPDFSize))
	if err != nil {
		return nil, fmt.Errorf("invoice: read gotenberg response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: gotenberg status %d: %s", ErrRenderUnavailable, resp.StatusCode, truncate(string(data), 300))
	}
	return data, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
