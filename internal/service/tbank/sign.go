package tbank

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-json"
)

// sign implements T-Bank's Token algorithm: take every flat request
// parameter plus the terminal Password, sort by key, concatenate the values
// (not the keys) in that order, SHA-256, hex. Nested objects/arrays (Receipt,
// DATA, …) and Token itself are never part of the signature.
func sign(password string, params map[string]string) string {
	keys := make([]string, 0, len(params)+1)
	for k := range params {
		keys = append(keys, k)
	}
	keys = append(keys, "Password")
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		if k == "Password" {
			sb.WriteString(password)
			continue
		}
		sb.WriteString(params[k])
	}

	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:])
}

// verifyNotification parses a raw webhook body, recomputes its Token over the
// flat fields actually present (future fields T-Bank adds are picked up
// automatically) and reports whether it matches.
func verifyNotification(password string, raw []byte) (notification, bool, error) {
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return notification{}, false, err
	}

	token, _ := generic["Token"].(string)
	params := make(map[string]string, len(generic))
	for k, v := range generic {
		if k == "Token" {
			continue
		}
		switch v.(type) {
		case map[string]any, []any:
			continue // nested objects/arrays never enter the signature
		}
		params[k] = stringifyJSONValue(v)
	}

	var n notification
	if err := json.Unmarshal(raw, &n); err != nil {
		return notification{}, false, err
	}

	return n, token != "" && strings.EqualFold(token, sign(password, params)), nil
}

// stringifyJSONValue renders a decoded JSON scalar the way T-Bank expects it
// in the signature: booleans as "true"/"false", whole numbers without a
// decimal point (encoding/json decodes all JSON numbers as float64).
func stringifyJSONValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case bool:
		if t {
			return "true"
		}
		return "false"
	case string:
		return t
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return ""
	}
}
