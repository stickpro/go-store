package invoice

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

//go:embed templates/invoice.html
var templatesFS embed.FS

var monthsGenitive = [...]string{
	"января", "февраля", "марта", "апреля", "мая", "июня",
	"июля", "августа", "сентября", "октября", "ноября", "декабря",
}

// templateFuncs are the helpers available to the invoice template.
func templateFuncs(loc *time.Location) template.FuncMap {
	return template.FuncMap{
		"money": formatMoney,
		"qty":   func(n int64) string { return groupThousands(fmt.Sprint(n)) },
		// dateLong: "30 октября 2026 г."
		"dateLong": func(t time.Time) string {
			t = t.In(loc)
			return fmt.Sprintf("%d %s %d г.", t.Day(), monthsGenitive[t.Month()-1], t.Year())
		},
		// date: "30.10.2026"
		"date": func(t time.Time) string { return t.In(loc).Format("02.01.2006") },
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
	}
}

// parseTemplate loads the built-in template, or the one at path when set.
func parseTemplate(path string, loc *time.Location) (*template.Template, error) {
	t := template.New("invoice.html").Funcs(templateFuncs(loc))
	if path == "" {
		return t.ParseFS(templatesFS, "templates/invoice.html")
	}
	src, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("invoice: read template %s: %w", path, err)
	}
	return t.Parse(string(src))
}

// renderHTML executes the template against a snapshot.
func renderHTML(t *template.Template, s Snapshot) ([]byte, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, s); err != nil {
		return nil, fmt.Errorf("invoice: render html: %w", err)
	}
	return buf.Bytes(), nil
}

// formatMoney prints 1234567.5 as "1 234 567,50" (non-breaking spaces).
func formatMoney(d decimal.Decimal) string {
	s := d.StringFixed(2)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	intPart, frac, _ := strings.Cut(s, ".")
	out := groupThousands(intPart) + "," + frac
	if neg {
		out = "-" + out
	}
	return out
}

func groupThousands(s string) string {
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
	}
	for i := pre; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
