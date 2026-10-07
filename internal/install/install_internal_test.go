package install

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestVerifyWritten pins the mismatch branch of the ASVS V14 read-back
// guard (T-04-01-01): an exact read-back passes, a divergent file fails
// with the named tampering error.
func TestVerifyWritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "goswitchd.service")
	want := []byte("[Service]\nExecStart=/opt/goswitch/goswitchd\n")
	if err := os.WriteFile(path, want, permPublic); err != nil {
		t.Fatalf("write the file: %v", err)
	}
	if err := verifyWritten(path, want); err != nil {
		t.Errorf("verifyWritten(exact) = %v, want nil", err)
	}

	tampered := []byte("[Service]\nExecStart=/usr/bin/evil\n")
	if err := os.WriteFile(path, tampered, permPublic); err != nil {
		t.Fatalf("tamper the file: %v", err)
	}
	err := verifyWritten(path, want)
	if err == nil {
		t.Fatal("verifyWritten(divergent) = nil, want the named tampering error")
	}
	if !errors.Is(err, errWriteVerify) {
		t.Errorf("verifyWritten(divergent) err = %v, want an errWriteVerify wrap", err)
	}
	if !strings.Contains(err.Error(), "tampering") {
		t.Errorf("err %q does not name the tampering suspicion", err)
	}
}

// The wrap corpus's raw gsettings shapes (goconst: named once).
const (
	wrapPairENRU        = `[('xkb', 'us'), ('xkb', 'ru')]`
	wrapPairRUEN        = `[('xkb', 'ru'), ('xkb', 'us')]`
	wrapWrappedENRU     = `[('ibus', 'goswitch-en'), ('ibus', 'goswitch-ru')]`
	wrapWrappedRUEN     = `[('ibus', 'goswitch-ru'), ('ibus', 'goswitch-en')]`
	wrapTriple          = `[('xkb', 'us'), ('xkb', 'ru'), ('xkb', 'fr')]`
	wrapUSFR            = `[('xkb', 'us'), ('xkb', 'fr')]`
	wrapSingleUS        = `[('xkb', 'us')]`
	wrapSingleRU        = `[('xkb', 'ru')]`
	wrapSingleFR        = `[('xkb', 'fr')]`
	wrapSingleUSWrapped = `[('ibus', 'goswitch-en')]`
	wrapSingleRUWrapped = `[('ibus', 'goswitch-ru')]`
	wrapForeignKind     = `[('xkb', 'us'), ('wayland', 'ru')]`
	// The half-wrapped forms of the G-5-5 corpus (owner decision
	// 2026-10-07): one goswitch tuple beside ONE foreign tuple — the ru
	// form completes to the owned pair, the us form dedupes to the
	// canonical single source, the fr form is the residue refusal.
	wrapHalfWrapped    = `[('ibus', 'goswitch-en'), ('xkb', 'ru')]`
	wrapHalfWrappedEN  = `[('ibus', 'goswitch-en'), ('xkb', 'us')]`
	wrapHalfWrappedFR  = `[('ibus', 'goswitch-en'), ('xkb', 'fr')]`
	wrapOwnedSingle    = `[('ibus', 'goswitch-en')]`
)

// TestWrapSourcesPairPreserved pins the D-54 wrap of the user's own pair:
// every xkb us/ru entry becomes the goswitch engine of its layout and the
// POSITIONS survive — a ru-first desktop gets goswitch-ru first.
func TestWrapSourcesPairPreserved(t *testing.T) {
	t.Run("us ru order", func(t *testing.T) {
		got, err := wrapSources(wrapPairENRU)
		if err != nil {
			t.Fatalf("wrapSources() err = %v, want nil", err)
		}
		if got != wrapWrappedENRU {
			t.Errorf("wrapSources() = %q, want %q", got, wrapWrappedENRU)
		}
	})

	t.Run("ru us order keeps positions", func(t *testing.T) {
		got, err := wrapSources(wrapPairRUEN)
		if err != nil {
			t.Fatalf("wrapSources() err = %v, want nil", err)
		}
		if got != wrapWrappedRUEN {
			t.Errorf("wrapSources() = %q, want %q (the ru slot must wrap first)", got, wrapWrappedRUEN)
		}
	})
}

// TestWrapSourcesSingleSource pins the owner decision of 2026-09-30 (one
// source in the GNOME switcher): a desktop carrying ONE wrappable xkb
// source wraps into exactly its own goswitch engine — single 'us' →
// goswitch-en, single 'ru' → goswitch-ru.
func TestWrapSourcesSingleSource(t *testing.T) {
	t.Run("single us wraps to goswitch-en", func(t *testing.T) {
		got, err := wrapSources(wrapSingleUS)
		if err != nil {
			t.Fatalf("wrapSources() err = %v, want nil", err)
		}
		if got != wrapSingleUSWrapped {
			t.Errorf("wrapSources() = %q, want %q", got, wrapSingleUSWrapped)
		}
	})

	t.Run("single ru wraps to goswitch-ru", func(t *testing.T) {
		got, err := wrapSources(wrapSingleRU)
		if err != nil {
			t.Fatalf("wrapSources() err = %v, want nil", err)
		}
		if got != wrapSingleRUWrapped {
			t.Errorf("wrapSources() = %q, want %q", got, wrapSingleRUWrapped)
		}
	})
}

// TestWrapSourcesForeignResidueRefused pins the ≥1-model residue refusal:
// a foreign source beside wrappable xkb us/ru entries leaves a list
// goswitch cannot own ENTIRELY — refused, never wrapped around (D-53: the
// daemon sees no keys through a foreign source, so install must never
// write a list its own audit rejects).
func TestWrapSourcesForeignResidueRefused(t *testing.T) {
	_, err := wrapSources(wrapTriple)
	if !errors.Is(err, errMixedSources) {
		t.Fatalf("wrapSources(%s) err = %v, want an errMixedSources wrap", wrapTriple, err)
	}
	if !strings.Contains(err.Error(), "fix") {
		t.Errorf("verdict %q carries no fix hint", err)
	}
}

// TestWrapSourcesHalfWrappedCompletes pins the G-5-5 completion contract
// (owner decision 2026-10-07, verbatim: «автоматически приводить
// конфигурацию к правильной, а если не получилось - отказ»): a
// hand-edited half-wrapped list — one goswitch tuple beside one foreign
// tuple — is no longer a terminal refusal. Every wrappable foreign xkb
// us/ru tuple completes to its goswitch engine (positions preserved) and
// repeated engine tuples dedupe keeping the first position, so the result
// is the canonical goswitch-owned configuration.
func TestWrapSourcesHalfWrappedCompletes(t *testing.T) {
	t.Run("goswitch-en beside xkb ru completes to the owned pair", func(t *testing.T) {
		got, err := wrapSources(wrapHalfWrapped)
		if err != nil {
			t.Fatalf("wrapSources(%s) err = %v, want the completion to succeed", wrapHalfWrapped, err)
		}
		if got != wrapWrappedENRU {
			t.Errorf("wrapSources(%s) = %q, want %q (completion, positions preserved)", wrapHalfWrapped, got, wrapWrappedENRU)
		}
	})

	t.Run("goswitch-en beside xkb us dedupes to the canonical single source", func(t *testing.T) {
		got, err := wrapSources(wrapHalfWrappedEN)
		if err != nil {
			t.Fatalf("wrapSources(%s) err = %v, want the completion to succeed", wrapHalfWrappedEN, err)
		}
		if got != wrapSingleUSWrapped {
			t.Errorf("wrapSources(%s) = %q, want %q (the wrapped duplicate engine tuple dedupes, first position kept)",
				wrapHalfWrappedEN, got, wrapSingleUSWrapped)
		}
	})
}

// TestWrapSourcesRefusalTable pins the refusal table of the G-5-5 revision
// (owner decision 2026-10-07): a layout outside us/ru, an entry of an
// unsupported kind, an already-owned list with nothing wrappable, and the
// FOREIGN RESIDUE forms — a foreign layout beside wrappable xkb entries or
// beside a goswitch engine — are named refusals, never a forced pair. The
// superseded contract of the 2026-09-30 single-source model — an IMMEDIATE
// errMixedSources refusal of every half-wrapped list — is replaced by the
// completion attempt (TestWrapSourcesHalfWrappedCompletes); only the
// un-completable residue refuses. Every verdict names the failing tuple
// TYPE and the fix, never the raw user line (D-20-safe).
func TestWrapSourcesRefusalTable(t *testing.T) {
	cases := map[string]struct {
		raw  string
		want error
	}{
		"us+fr leaves a foreign residue":                     {raw: wrapUSFR, want: errMixedSources},
		"goswitch-en beside foreign residue":                 {raw: wrapHalfWrappedFR, want: errMixedSources},
		"triple leaves a foreign residue":                    {raw: wrapTriple, want: errMixedSources},
		"single foreign layout":                              {raw: wrapSingleFR, want: errUnsupportedPair},
		"unsupported kind in list":                           {raw: wrapForeignKind, want: errUnsupportedPair},
		"already-owned single (nothing wrappable)":           {raw: wrapOwnedSingle, want: errUnsupportedPair},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := wrapSources(tc.raw)
			if !errors.Is(err, tc.want) {
				t.Fatalf("wrapSources(%s) err = %v, want an %v wrap", tc.raw, err, tc.want)
			}
			if got != "" {
				t.Errorf("wrapSources(%s) = %q alongside the refusal, want an empty value", tc.raw, got)
			}
			if strings.Contains(err.Error(), tc.raw) {
				t.Errorf("verdict %q carries the raw user line — it must name the tuple type, not the content", err)
			}
			if !strings.Contains(err.Error(), "fix") {
				t.Errorf("verdict %q carries no fix hint", err)
			}
		})
	}
}

// TestWrapSourcesAlreadyOwnedUpgrade pins the Pitfall-7 upgrade path: the
// live desktop is already goswitch-owned (the phase-4 single source), so
// the two-source wrapper is computed from the SAVED ORIGINAL in the state
// file — never from the live value — and the state file stays byte-intact
// (the first backup is sacred).
func TestWrapSourcesAlreadyOwnedUpgrade(t *testing.T) {
	home := t.TempDir()
	statePath := filepath.Join(home, stateDirRel, stateFile)
	original := wrapPairENRU
	plant := `{"sources":"` + original + `","switch_input_source":"[]","switch_input_source_backward":"[]"}`
	if err := os.MkdirAll(filepath.Dir(statePath), permDir); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	if err := os.WriteFile(statePath, []byte(plant), permPrivate); err != nil {
		t.Fatalf("plant the state file: %v", err)
	}

	var sets [][]string
	run := func(_ context.Context, name string, args []string, _ []string, _ []byte) ([]byte, error) {
		if name == binGSettings && len(args) == 4 && args[0] == "set" {
			sets = append(sets, slices.Clone(args))
		}
		if name == binGSettings && len(args) == 3 && args[0] == "get" && args[2] == gsettingsKey {
			return []byte(wrapOwnedSingle), nil
		}

		return nil, nil
	}
	i := New(WithRunner(run), WithHome(home))
	if err := i.takeoverSources(context.Background()); err != nil {
		t.Fatalf("takeoverSources() err = %v, want the upgrade from the saved original to succeed", err)
	}

	wantSet := []string{"set", gsettingsSchema, gsettingsKey, wrapWrappedENRU}
	if len(sets) != 1 || !slices.Equal(sets[0], wantSet) {
		t.Errorf("set calls = %v, want exactly %v (the wrapper computed from the SAVED original)", sets, wantSet)
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("re-read state: %v", err)
	}
	if string(data) != plant {
		t.Errorf("state rewritten to %q, want %q byte-intact (the first backup is sacred)", data, plant)
	}
}
