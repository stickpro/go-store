package opencartimport

import (
	"fmt"
	"strings"
)

var cyrillicToLatin = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch",
	'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

// Slugify transliterates Cyrillic to Latin, lowercases, and replaces every
// run of characters outside [a-z0-9] with a single hyphen.
func Slugify(s string) string {
	s = strings.ToLower(s)

	var transliterated strings.Builder
	for _, r := range s {
		if repl, ok := cyrillicToLatin[r]; ok {
			transliterated.WriteString(repl)
			continue
		}
		transliterated.WriteRune(r)
	}

	var slug strings.Builder
	prevHyphen := false
	for _, r := range transliterated.String() {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			slug.WriteRune(r)
			prevHyphen = false
		default:
			if !prevHyphen && slug.Len() > 0 {
				slug.WriteRune('-')
				prevHyphen = true
			}
		}
	}

	return strings.Trim(slug.String(), "-")
}

// SanitizeKeyword cleans an OpenCart SEO keyword (e.g. "catalog/podarki")
// into a plain slug segment.
func SanitizeKeyword(keyword string) string {
	keyword = strings.ReplaceAll(keyword, "/", "-")
	return Slugify(keyword)
}

// Uniquify appends "-2", "-3", ... to base until exists returns false.
func Uniquify(base string, exists func(string) bool) string {
	if base == "" {
		base = "item"
	}
	candidate := base
	for i := 2; exists(candidate); i++ {
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return candidate
}
