package mail

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed templates/*.html
var templatesFS embed.FS

// renderLang is the lang attribute emitted in the layout. i18n is out of scope
// for now; when it lands this becomes a parameter and templates move under
// per-language directories.
const renderLang = "ru"

// view is the single data shape passed to every mail template. Message fields
// live under .Data; brand-wide values sit at the root.
type view struct {
	Data      Message
	BrandName string
	SiteURL   string

	// filled in progressively while rendering
	Title   string
	Content template.HTML
	Lang    string
}

type renderer struct {
	tmpl      *template.Template
	brandName string
	siteURL   string
}

func newRenderer(brandName, siteURL string) (*renderer, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("mail: parse templates: %w", err)
	}

	r := &renderer{tmpl: tmpl, brandName: brandName, siteURL: siteURL}

	if err := r.validate(); err != nil {
		return nil, err
	}
	return r, nil
}

// validate makes sure every registered kind has both template blocks and that
// the shared layout exists. Fails startup instead of the first send.
func (r *renderer) validate() error {
	if r.tmpl.Lookup("layout") == nil {
		return fmt.Errorf("mail: template block %q is missing", "layout")
	}
	for _, kind := range Kinds() {
		for _, block := range []string{kind + "/subject", kind + "/content"} {
			if r.tmpl.Lookup(block) == nil {
				return fmt.Errorf("mail: template block %q is missing", block)
			}
		}
	}
	return nil
}

// render produces the subject line and the full HTML body for a message.
func (r *renderer) render(m Message) (subject, body string, err error) {
	v := view{
		Data:      m,
		BrandName: r.brandName,
		SiteURL:   r.siteURL,
		Lang:      renderLang,
	}

	subject, err = r.exec(m.Kind()+"/subject", v)
	if err != nil {
		return "", "", err
	}
	subject = strings.TrimSpace(subject)
	v.Title = subject

	content, err := r.exec(m.Kind()+"/content", v)
	if err != nil {
		return "", "", err
	}
	v.Content = template.HTML(content) //nolint:gosec // content comes from a trusted embedded template

	body, err = r.exec("layout", v)
	if err != nil {
		return "", "", err
	}

	return subject, body, nil
}

func (r *renderer) exec(name string, v view) (string, error) {
	var b bytes.Buffer
	if err := r.tmpl.ExecuteTemplate(&b, name, v); err != nil {
		return "", fmt.Errorf("mail: render %q: %w", name, err)
	}
	return b.String(), nil
}
