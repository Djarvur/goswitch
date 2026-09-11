package layouts_test

import (
	"testing"

	"github.com/Djarvur/goswitch/layouts"
)

// translateStrict maps every rune of s through table; a rune missing from the
// table maps to NUL, so an absent golden entry fails the comparison visibly
// instead of passing the character through unnoticed.
func translateStrict(s string, table map[rune]rune) string {
	mapped := make([]rune, 0, len(s))
	for _, r := range s {
		mapped = append(mapped, table[r])
	}

	return string(mapped)
}

// TestGolden_SpecExamples pins the SPEC correction examples against the
// committed tables: ghbdtn→привет in all three registers, both directions.
// The per-character pairs are the key-position join of us(basic) and the
// effective ru layout (winkeys): g and п share AC05, h and р share AC06,
// b and и share AB05, d and в share AC03, t and е share AD05, n and т AB06.
func TestGolden_SpecExamples(t *testing.T) {
	t.Parallel()

	pairs := []struct {
		en rune
		ru rune
	}{
		{en: 'g', ru: 'п'},
		{en: 'h', ru: 'р'},
		{en: 'b', ru: 'и'},
		{en: 'd', ru: 'в'},
		{en: 't', ru: 'е'},
		{en: 'n', ru: 'т'},
		{en: 'G', ru: 'П'},
		{en: 'H', ru: 'Р'},
		{en: 'B', ru: 'И'},
		{en: 'D', ru: 'В'},
		{en: 'T', ru: 'Е'},
		{en: 'N', ru: 'Т'},
	}
	for _, tt := range pairs {
		t.Run(string(tt.en)+"-"+string(tt.ru), func(t *testing.T) {
			t.Parallel()
			if got := layouts.ENToRU[tt.en]; got != tt.ru {
				t.Errorf("ENToRU[%q] = %q, want %q", tt.en, got, tt.ru)
			}
			if got := layouts.RUToEN[tt.ru]; got != tt.en {
				t.Errorf("RUToEN[%q] = %q, want %q", tt.ru, got, tt.en)
			}
		})
	}

	words := []struct {
		name string
		in   string
		want string
		en   bool // true: ENToRU, false: RUToEN
	}{
		{name: "lower", in: "ghbdtn", want: "привет", en: true},
		{name: "upper", in: "GHBDTN", want: "ПРИВЕТ", en: true},
		{name: "mixed", in: "Ghbdtn", want: "Привет", en: true},
		{name: "lower-back", in: "привет", want: "ghbdtn", en: false},
		{name: "upper-back", in: "ПРИВЕТ", want: "GHBDTN", en: false},
		{name: "mixed-back", in: "Привет", want: "Ghbdtn", en: false},
	}
	for _, tt := range words {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			table := layouts.RUToEN
			if tt.en {
				table = layouts.ENToRU
			}
			if got := translateStrict(tt.in, table); got != tt.want {
				t.Errorf("translate %q = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestGolden_Punctuation pins every punctuation pair of both shift levels in
// both directions. The asymmetric pairs are the point: the join is by key
// position, not by character — '&' lives on AE07-shift in us while '?' lives
// on AE07-shift in ru, so ENToRU['&'] = '?' and RUToEN['?'] = '&' while
// ENToRU['?'] = ',' (AB10-shift in us maps to comma on AB10-shift in ru).
func TestGolden_Punctuation(t *testing.T) {
	t.Parallel()

	pairs := []struct {
		en rune
		ru rune
	}{
		{en: '[', ru: 'х'},
		{en: ']', ru: 'ъ'},
		{en: ';', ru: 'ж'},
		{en: '\'', ru: 'э'},
		{en: ',', ru: 'б'},
		{en: '.', ru: 'ю'},
		{en: '/', ru: '.'},
		{en: '?', ru: ','},
		{en: '`', ru: 'ё'},
		{en: '~', ru: 'Ё'},
		{en: '@', ru: '"'},
		{en: '#', ru: '№'},
		{en: '$', ru: ';'},
		{en: '^', ru: ':'},
		{en: '&', ru: '?'},
		{en: '|', ru: '/'},
	}
	for _, tt := range pairs {
		t.Run(string(tt.en)+"-"+string(tt.ru), func(t *testing.T) {
			t.Parallel()
			if got := layouts.ENToRU[tt.en]; got != tt.ru {
				t.Errorf("ENToRU[%q] = %q, want %q", tt.en, got, tt.ru)
			}
			if got := layouts.RUToEN[tt.ru]; got != tt.en {
				t.Errorf("RUToEN[%q] = %q, want %q", tt.ru, got, tt.en)
			}
		})
	}
}

// TestGolden_TableSize pins that both tables cover every key position of
// us(basic) at both shift levels (47 positions, ≥47 entries after merging the
// two levels into one map) and stay mutually inverse in size.
func TestGolden_TableSize(t *testing.T) {
	t.Parallel()

	enLen := len(layouts.ENToRU)
	ruLen := len(layouts.RUToEN)
	if enLen != ruLen {
		t.Errorf("table sizes differ: len(ENToRU) = %d, len(RUToEN) = %d", enLen, ruLen)
	}
	if enLen < 47 {
		t.Errorf("len(ENToRU) = %d, want at least 47 (all us(basic) key positions)", enLen)
	}
	if ruLen < 47 {
		t.Errorf("len(RUToEN) = %d, want at least 47 (all us(basic) key positions)", ruLen)
	}
}
