package hotkey_test

import (
	"testing"

	"github.com/Djarvur/goswitch/internal/hotkey"
)

// TestParseBinding pins the closed key tables (D-31/D-33): single-key names
// resolve to their keyval with the key's own family bit in ModMask (a
// Shift_R press carries MaskShift on the wire).
func TestParseBinding(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		want hotkey.Binding
	}{
		{
			name: "default tap key",
			in:   "shift_r",
			want: hotkey.Binding{Keyval: hotkey.KeyvalShiftR, ModMask: hotkey.MaskShift},
		},
		{
			name: "left control",
			in:   "ctrl_l",
			want: hotkey.Binding{Keyval: hotkey.KeyvalCtrlL, ModMask: hotkey.MaskControl},
		},
		{
			name: "left alt",
			in:   "alt_l",
			want: hotkey.Binding{Keyval: hotkey.KeyvalAltL, ModMask: hotkey.MaskMod1},
		},
		{
			name: "right alt",
			in:   "alt_r",
			want: hotkey.Binding{Keyval: hotkey.KeyvalAltR, ModMask: hotkey.MaskMod1},
		},
		{
			name: "left super",
			in:   "super_l",
			want: hotkey.Binding{Keyval: hotkey.KeyvalSuperL, ModMask: hotkey.MaskMod4},
		},
		{
			name: "right super",
			in:   "super_r",
			want: hotkey.Binding{Keyval: hotkey.KeyvalSuperR, ModMask: hotkey.MaskMod4},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := hotkey.ParseBinding(tc.in)
			if err != nil {
				t.Fatalf("ParseBinding(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("ParseBinding(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

// TestParseBinding_Combos pins the combo grammar: "+"-joined tokens with
// the KEY last, every held modifier's bit OR the key's own family bit —
// the full modifier state of the bound key's event.
func TestParseBinding_Combos(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		want hotkey.Binding
	}{
		{
			name: "word+layout combo (D-36 default)",
			in:   "shift+ctrl_r",
			want: hotkey.Binding{Keyval: hotkey.KeyvalCtrlR, ModMask: hotkey.MaskShift | hotkey.MaskControl},
		},
		{
			name: "two explicit modifiers",
			in:   "shift+super+ctrl_r",
			want: hotkey.Binding{
				Keyval:  hotkey.KeyvalCtrlR,
				ModMask: hotkey.MaskShift | hotkey.MaskMod4 | hotkey.MaskControl,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := hotkey.ParseBinding(tc.in)
			if err != nil {
				t.Fatalf("ParseBinding(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("ParseBinding(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

// TestParseBinding_Rejected pins the closed-vocabulary refusals (D-33): an
// empty name or an unknown token anywhere in the binding is an error, never
// a silent fallback.
func TestParseBinding_Rejected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
	}{
		{name: "empty string", in: ""},
		{name: "unknown key", in: "shift_x"},
		{name: "unknown key in combo", in: "shift+ctrl_q"},
		{name: "unknown modifier", in: "shif+ctrl_r"},
		{name: "trailing separator", in: "shift+"},
		{name: "uppercase name", in: "Shift_R"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := hotkey.ParseBinding(tc.in); err == nil {
				t.Fatalf("ParseBinding(%q) succeeded, want a closed-vocabulary rejection", tc.in)
			}
		})
	}
}

// TestParseBinding_KeyvalProvenance pins the literal wire values against
// the installed header (acceptance: Control_R == 0xffe4, the values cited
// from ibuskeysyms.h — never trusted from memory).
func TestParseBinding_KeyvalProvenance(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		got   uint32
		want  uint32
		where string
	}{
		{name: "Shift_L", got: hotkey.KeyvalShiftL, want: 0xffe1, where: "ibuskeysyms.h:188"},
		{name: "Shift_R", got: hotkey.KeyvalShiftR, want: 0xffe2, where: "ibuskeysyms.h:189"},
		{name: "Control_L", got: hotkey.KeyvalCtrlL, want: 0xffe3, where: "ibuskeysyms.h:190"},
		{name: "Control_R", got: hotkey.KeyvalCtrlR, want: 0xffe4, where: "ibuskeysyms.h:191"},
		{name: "Alt_L", got: hotkey.KeyvalAltL, want: 0xffe9, where: "ibuskeysyms.h:196"},
		{name: "Alt_R", got: hotkey.KeyvalAltR, want: 0xffea, where: "ibuskeysyms.h:197"},
		{name: "Super_L", got: hotkey.KeyvalSuperL, want: 0xffeb, where: "ibuskeysyms.h:198"},
		{name: "Super_R", got: hotkey.KeyvalSuperR, want: 0xffec, where: "ibuskeysyms.h:199"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Errorf("%s = %#x, want %#x (%s)", tc.name, tc.got, tc.want, tc.where)
			}
		})
	}
}

// TestFamilyMask pins the press-side wire truth (live finding 2026-09-15,
// plan 03-04): a key press's state word carries only the modifiers held
// BEFORE the key — the key's own family bit appears on its release. Every
// bindable key maps to its family bit, everything else to none.
func TestFamilyMask(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		keyval uint32
		want   uint32
	}{
		{name: "Shift_L", keyval: hotkey.KeyvalShiftL, want: hotkey.MaskShift},
		{name: "right shift", keyval: hotkey.KeyvalShiftR, want: hotkey.MaskShift},
		{name: "Ctrl_L", keyval: hotkey.KeyvalCtrlL, want: hotkey.MaskControl},
		{name: "Ctrl_R", keyval: hotkey.KeyvalCtrlR, want: hotkey.MaskControl},
		{name: "Alt_L", keyval: hotkey.KeyvalAltL, want: hotkey.MaskMod1},
		{name: "Alt_R", keyval: hotkey.KeyvalAltR, want: hotkey.MaskMod1},
		{name: "Super_L", keyval: hotkey.KeyvalSuperL, want: hotkey.MaskMod4},
		{name: "Super_R", keyval: hotkey.KeyvalSuperR, want: hotkey.MaskMod4},
		{name: "letter", keyval: 0x061, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := hotkey.FamilyMask(tc.keyval); got != tc.want {
				t.Errorf("FamilyMask(%#x) = %#x, want %#x", tc.keyval, got, tc.want)
			}
		})
	}
}
