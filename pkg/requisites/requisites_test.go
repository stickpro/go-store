package requisites

import "testing"

func TestINN(t *testing.T) {
	cases := []struct {
		in            string
		legal, person bool
	}{
		{"7707083893", true, false},   // Сбербанк
		{"7736207543", true, false},   // Яндекс
		{"7707083894", false, false},  // bad check digit
		{"500100732259", false, true}, // valid person INN
		{"500100732258", false, false},
		{"77070838", false, false},
		{"770708389a", false, false},
		{"", false, false},
	}
	for _, c := range cases {
		if got := IsLegalINN(c.in); got != c.legal {
			t.Errorf("IsLegalINN(%q) = %v, want %v", c.in, got, c.legal)
		}
		if got := IsPersonINN(c.in); got != c.person {
			t.Errorf("IsPersonINN(%q) = %v, want %v", c.in, got, c.person)
		}
		if got := IsINN(c.in); got != (c.legal || c.person) {
			t.Errorf("IsINN(%q) = %v", c.in, got)
		}
	}
}

func TestKPP(t *testing.T) {
	for in, want := range map[string]bool{
		"773601001": true,
		"7736AB001": true,
		"77360100":  false,
		"7736ab001": false,
		"77360100x": false,
	} {
		if got := IsKPP(in); got != want {
			t.Errorf("IsKPP(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestOGRN(t *testing.T) {
	for in, want := range map[string]bool{
		"1027700132195":   true, // Сбербанк
		"1027700229193":   true, // Яндекс
		"1027700132196":   false,
		"304500116000157": true, // OGRNIP
		"304500116000158": false,
		"102770013219":    false,
	} {
		if got := IsOGRN(in); got != want {
			t.Errorf("IsOGRN(%q) = %v, want %v", in, got, want)
		}
	}
}
