// Package moneywords spells a ruble amount in Russian words, the way invoices
// and acts print it: "Одна тысяча двести рублей 50 копеек".
package moneywords

import (
	"strings"

	"github.com/shopspring/decimal"
)

var (
	onesMasc = []string{"", "один", "два", "три", "четыре", "пять", "шесть", "семь", "восемь", "девять"}
	onesFem  = []string{"", "одна", "две", "три", "четыре", "пять", "шесть", "семь", "восемь", "девять"}
	teens    = []string{
		"десять", "одиннадцать", "двенадцать", "тринадцать", "четырнадцать",
		"пятнадцать", "шестнадцать", "семнадцать", "восемнадцать", "девятнадцать",
	}
	tens = []string{
		"", "", "двадцать", "тридцать", "сорок", "пятьдесят",
		"шестьдесят", "семьдесят", "восемьдесят", "девяносто",
	}
	hundreds = []string{
		"", "сто", "двести", "триста", "четыреста", "пятьсот",
		"шестьсот", "семьсот", "восемьсот", "девятьсот",
	}
)

// scale is one group of three digits: its word forms (one / few / many) and
// grammatical gender.
type scale struct {
	forms    [3]string
	feminine bool
}

// scales from the lowest group up; the lowest group has no word of its own.
var scales = []scale{
	{},
	{forms: [3]string{"тысяча", "тысячи", "тысяч"}, feminine: true},
	{forms: [3]string{"миллион", "миллиона", "миллионов"}},
	{forms: [3]string{"миллиард", "миллиарда", "миллиардов"}},
	{forms: [3]string{"триллион", "триллиона", "триллионов"}},
}

// Rubles spells amount as "<rubles in words> <рубль/рубля/рублей> NN
// <копейка/копейки/копеек>", capitalised. The amount is rounded to kopecks;
// negative amounts are spelled by absolute value.
func Rubles(amount decimal.Decimal) string {
	amount = amount.Abs().Round(2)
	rub := amount.Truncate(0)
	kop := amount.Sub(rub).Shift(2).IntPart()
	r := rub.IntPart()

	words := Number(r, false)
	var b strings.Builder
	b.WriteString(capitalize(words))
	b.WriteString(" ")
	b.WriteString(Plural(r, "рубль", "рубля", "рублей"))
	b.WriteString(" ")
	if kop < 10 {
		b.WriteString("0")
	}
	b.WriteString(decimal.NewFromInt(kop).String())
	b.WriteString(" ")
	b.WriteString(Plural(kop, "копейка", "копейки", "копеек"))
	return b.String()
}

// Number spells a non-negative integer in words ("ноль" for 0). feminine
// selects "одна/две" for the last group (e.g. for feminine nouns).
func Number(n int64, feminine bool) string {
	if n <= 0 {
		return "ноль"
	}

	var groups []int64
	for n > 0 {
		groups = append(groups, n%1000)
		n /= 1000
	}

	parts := make([]string, 0, len(groups)*4)
	for i := len(groups) - 1; i >= 0; i-- {
		g := groups[i]
		if g == 0 || i >= len(scales) {
			continue
		}
		fem := scales[i].feminine || (i == 0 && feminine)
		parts = append(parts, triplet(g, fem)...)
		if i > 0 {
			f := scales[i].forms
			parts = append(parts, Plural(g, f[0], f[1], f[2]))
		}
	}
	return strings.Join(parts, " ")
}

func triplet(n int64, feminine bool) []string {
	var out []string
	if h := n / 100; h > 0 {
		out = append(out, hundreds[h])
	}
	rest := n % 100
	switch {
	case rest >= 10 && rest < 20:
		out = append(out, teens[rest-10])
	default:
		if t := rest / 10; t > 0 {
			out = append(out, tens[t])
		}
		if u := rest % 10; u > 0 {
			if feminine {
				out = append(out, onesFem[u])
			} else {
				out = append(out, onesMasc[u])
			}
		}
	}
	return out
}

// Plural picks the Russian plural form for n: one (1, 21), few (2-4, 22-24)
// or many (0, 5-20, 25-30).
func Plural(n int64, one, few, many string) string {
	if n < 0 {
		n = -n
	}
	n100 := n % 100
	n10 := n % 10
	switch {
	case n100 >= 11 && n100 <= 14:
		return many
	case n10 == 1:
		return one
	case n10 >= 2 && n10 <= 4:
		return few
	default:
		return many
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return strings.ToUpper(string(r[0])) + string(r[1:])
}
