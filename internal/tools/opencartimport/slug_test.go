package opencartimport

import "testing"

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Подарки":               "podarki",
		"Ноутбуки и планшеты":   "noutbuki-i-planshety",
		"Iphone 15 Pro Max":     "iphone-15-pro-max",
		"  --Trim me--  ":       "trim-me",
		"Товар №1 (100% нужен)": "tovar-1-100-nuzhen",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeKeyword(t *testing.T) {
	if got, want := SanitizeKeyword("catalog/podarki"), "catalog-podarki"; got != want {
		t.Errorf("SanitizeKeyword() = %q, want %q", got, want)
	}
}

func TestUniquify(t *testing.T) {
	taken := map[string]bool{"foo": true, "foo-2": true}
	got := Uniquify("foo", func(s string) bool { return taken[s] })
	if got != "foo-3" {
		t.Errorf("Uniquify() = %q, want %q", got, "foo-3")
	}

	if got := Uniquify("bar", func(string) bool { return false }); got != "bar" {
		t.Errorf("Uniquify() = %q, want %q", got, "bar")
	}
}
