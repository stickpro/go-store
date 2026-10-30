// Package requisites validates Russian company identifiers (INN, KPP, OGRN)
// by format and check digit.
package requisites

import "regexp"

var kppRe = regexp.MustCompile(`^\d{4}[\dA-Z]{2}\d{3}$`)

// IsLegalINN reports whether s is a valid 10-digit INN of a legal entity.
func IsLegalINN(s string) bool {
	d, ok := digits(s, 10)
	if !ok {
		return false
	}
	return checkDigit(d[:9], []int{2, 4, 10, 3, 5, 9, 4, 6, 8}) == d[9]
}

// IsPersonINN reports whether s is a valid 12-digit INN of a person (used by
// sole proprietors).
func IsPersonINN(s string) bool {
	d, ok := digits(s, 12)
	if !ok {
		return false
	}
	return checkDigit(d[:10], []int{7, 2, 4, 10, 3, 5, 9, 4, 6, 8}) == d[10] &&
		checkDigit(d[:11], []int{3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8}) == d[11]
}

// IsINN reports whether s is a valid INN of either kind.
func IsINN(s string) bool {
	return IsLegalINN(s) || IsPersonINN(s)
}

// IsKPP reports whether s has the KPP format: 4 digits, 2 digits or Latin
// capitals, 3 digits. KPP has no check digit.
func IsKPP(s string) bool {
	return kppRe.MatchString(s)
}

// IsOGRN reports whether s is a valid 13-digit OGRN (legal entity) or a
// 15-digit OGRNIP (sole proprietor).
func IsOGRN(s string) bool {
	switch len(s) {
	case 13:
		return ogrnCheck(s, 11)
	case 15:
		return ogrnCheck(s, 13)
	default:
		return false
	}
}

// ogrnCheck: the last digit is (number formed by the other digits mod m) mod 10.
func ogrnCheck(s string, m int64) bool {
	d, ok := digits(s, len(s))
	if !ok {
		return false
	}
	var n int64
	for _, v := range d[:len(d)-1] {
		n = n*10 + int64(v)
	}
	return int(n%m%10) == d[len(d)-1]
}

func checkDigit(d, weights []int) int {
	sum := 0
	for i, w := range weights {
		sum += d[i] * w
	}
	return sum % 11 % 10
}

func digits(s string, n int) ([]int, bool) {
	if len(s) != n {
		return nil, false
	}
	out := make([]int, n)
	for i := range n {
		c := s[i]
		if c < '0' || c > '9' {
			return nil, false
		}
		out[i] = int(c - '0')
	}
	return out, true
}
