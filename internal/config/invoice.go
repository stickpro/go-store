package config

import "time"

// InvoiceConfig configures the "pay by invoice" method for legal entities: the
// seller requisites printed on every invoice, VAT, the payment term and the
// Gotenberg service the PDF is rendered with. The invoice method is offered at
// checkout only once the seller requisites are filled in (see Configured).
type InvoiceConfig struct {
	// TTL is the payment term: invoice due date = issue date + TTL.
	TTL          time.Duration `yaml:"ttl" default:"120h" usage:"payment term of an invoice (due date = issue date + ttl)"`
	NumberPrefix string        `yaml:"number_prefix" usage:"prefix printed before the invoice number, e.g. \"СЧ-\""`
	// VATRate is the VAT rate included in prices, in percent ("20", "10",
	// "5", "7", "0"); empty means the seller is not a VAT payer ("Без НДС").
	VATRate string `yaml:"vat_rate" usage:"VAT rate in percent included in prices (20, 10, 7, 5, 0); empty = \"Без НДС\""`
	// TemplatePath overrides the built-in HTML template (Go html/template).
	TemplatePath string `yaml:"template_path" usage:"path to a custom HTML invoice template; empty = built-in standard form"`

	GotenbergURL  string        `yaml:"gotenberg_url" default:"http://localhost:3000" usage:"Gotenberg base URL used to render invoice PDFs"`
	RenderTimeout time.Duration `yaml:"render_timeout" default:"20s" usage:"timeout for rendering one invoice PDF"`

	Seller InvoiceSellerConfig `yaml:"seller"`
}

// InvoiceSellerConfig is the seller block of the invoice.
type InvoiceSellerConfig struct {
	Name        string `yaml:"name" usage:"seller legal name, e.g. ООО \"Ромашка\""`
	INN         string `yaml:"inn" usage:"seller INN"`
	KPP         string `yaml:"kpp" usage:"seller KPP (empty for a sole proprietor)"`
	OGRN        string `yaml:"ogrn" usage:"seller OGRN / OGRNIP"`
	Address     string `yaml:"address" usage:"seller legal address"`
	Phone       string `yaml:"phone" usage:"seller phone printed on the invoice"`
	BankName    string `yaml:"bank_name" usage:"seller bank name"`
	BIK         string `yaml:"bik" usage:"seller bank BIK"`
	Account     string `yaml:"account" usage:"seller settlement account (р/с)"`
	CorrAccount string `yaml:"corr_account" usage:"seller bank correspondent account (к/с)"`
	Director    string `yaml:"director" usage:"director (or sole proprietor) name for the signature line"`
	Accountant  string `yaml:"accountant" usage:"chief accountant name for the signature line; empty = same as director"`
}

// Configured reports whether the seller requisites an invoice cannot be
// issued without are set.
func (c InvoiceConfig) Configured() bool {
	s := c.Seller
	return s.Name != "" && s.INN != "" && s.Address != "" &&
		s.BankName != "" && s.BIK != "" && s.Account != "" && s.CorrAccount != ""
}
