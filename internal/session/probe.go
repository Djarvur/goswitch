package session

import (
	"errors"

	"github.com/Djarvur/goswitch/internal/engine"
)

// Probe keyspec refusals — the named errors of the strict keyspec grammar
// (plan 09-02, ASVS V5: every cause distinct and errors.Is-checkable).
// RED scaffold: the behavior lands in the GREEN commit; the declarations
// exist so the corpus compiles and fails on assertions, not on symbols.
var (
	ErrProbeEmptySpec   = errors.New("probe keyspec is empty")
	ErrProbeSpecFormat  = errors.New("probe keyspec must be [mod+]{Home|End|Left|Right} or one key")
	ErrProbeKeyUnknown  = errors.New("probe key unknown")
	ErrProbeModUnknown  = errors.New("probe modifier unknown")
	ErrProbeKeyUnmapped = errors.New("probe key has no physical keycode mapping")
)

// errProbeNotImplemented is the RED scaffold's refusal — every corpus
// assertion must fail against it until the GREEN implementation lands.
var errProbeNotImplemented = errors.New("probe forward not implemented")

// ParseProbeKey parses the probe verb's keyspec (plan 09-02). RED scaffold.
func ParseProbeKey(spec string) (keyval, keycode, mods uint32, err error) {
	_ = spec

	return 0, 0, 0, errProbeNotImplemented
}

// ProbeForward emits the synthetic press/release pair for spec through the
// live engine emitter (plan 09-02). RED scaffold: nothing is emitted.
func ProbeForward(eng engine.Emitter, spec string) error {
	_, _ = eng, spec

	return nil
}
