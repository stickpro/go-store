package mail

// Message is one transactional email. Kind() identifies it on the queue and
// selects its templates ("<kind>/subject" and "<kind>/content"). A Message is a
// plain struct holding the template variables and must be JSON-serialisable —
// it travels through the queue as the Task payload.
//
// Adding a new email:
//  1. define a struct here with a Kind() method;
//  2. add one line to registry;
//  3. add templates/<kind>.html with "subject" and "content" defines.
//
// Nothing in Service, the worker, or the renderer changes.
type Message interface {
	Kind() string
}

const (
	KindWelcome           = "welcome"
	KindLoginCode         = "login_code"
	KindOrderConfirmation = "order_confirmation"
)

// registry maps a kind to a constructor for its empty payload so the worker can
// decode a queued Task back into a typed Message.
var registry = map[string]func() Message{
	KindWelcome:           func() Message { return &Welcome{} },
	KindLoginCode:         func() Message { return &LoginCode{} },
	KindOrderConfirmation: func() Message { return &OrderConfirmation{} },
}

// Kinds returns every registered kind (used for startup template validation).
func Kinds() []string {
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}

// Welcome is sent once, right after registration.
type Welcome struct {
	Email string `json:"email"`
}

func (Welcome) Kind() string { return KindWelcome }

// LoginCode carries the one-time code for passwordless login / registration.
type LoginCode struct {
	Code string `json:"code"`
	// TTLMinutes is shown in the email body ("valid for N minutes").
	TTLMinutes int `json:"ttl_minutes"`
}

func (LoginCode) Kind() string { return KindLoginCode }

// OrderConfirmation is sent once, right after a successful checkout. Line-level
// detail lives in the customer's order history; this is just the receipt header.
type OrderConfirmation struct {
	OrderNumber int64  `json:"order_number"`
	Currency    string `json:"currency"`
	// GrandTotal is pre-formatted (two decimals) so the template stays dumb.
	GrandTotal string `json:"grand_total"`
}

func (OrderConfirmation) Kind() string { return KindOrderConfirmation }
