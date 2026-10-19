package main

// The phase-10 round-6 DST-apply experiment case (the 10-13 scope
// continuation): the machine-visible answer to THE round-6 question — does
// the CLIENT APPLY the engine-emitted DeleteSurroundingText, i.e. do the
// deleted characters actually leave the terminal's input line? The 10-16
// bisection proved warm-context DST is crash-SAFE (R1b/R3/R4 survive; only
// the ≲1.5 s virgin window kills mutter 46.2) but never measured
// application; the host defect's round-6 fix candidate (commit a196101)
// routes the terminal write through DST+commit — same text-input channel
// as the commit, order guaranteed by the protocol — and that routing is
// shippable only if the clients honor the request.
//
// The matrix (one case run; dated rows to the round-6 journal section):
//
//	wez-apply  — warm wezterm-nightly pane, the typed token materializes the
//	             client's state, then probe-delete -6 6 covering exactly the
//	             token; the RAW pty capture (stty raw — the drain-case form)
//	             answers: erase bytes (0x7f/0x08) after the token = APPLIED,
//	             silence = IGNORED, some = PARTIAL.
//	vte-apply  — the same probe on gnome-terminal/VTE, the differential
//	             comparator separating «client ignores» from «stack drops».
//	wez-choreo-a/b — the full round-6 construction twice (the two-greens
//	             discipline): token → DST(6) → a 1 s artificial drain-delay →
//	             probe-commit привет; the pty oracle is the byte ORDER (every
//	             erase strictly before the word's first byte) plus the
//	             final-line net (6 erases remove exactly the token) — the
//	             construction works end to end only on that shape.
//
// The warm-discipline (the safety gate, honored by EVERY cell), in the
// bisection's own terms: the trigger boundary is TIME from the context's
// fresh focus_in (crash at +0.49 s — R2/R6b/R7; survive at +1.5 s — R3;
// survive at +23 s with zero pushes — R4), so the gate holds every DST
// until the journal's last focus_in is ≥ 3 s old, with the TYPED TOKEN
// (real keys through the context — the mandate's materialization shape)
// as the precondition. The Require-only ask is attempted too (R5b-safe in
// ANY window) and its answer RECORDED — but it never gates: the 0x9-class
// clients (the wezterm 10-06 signature, live in run 1) answer nothing, and
// a witness they can never produce must not make the matrix unmeasurable.
// No gate ⇒ the cell lands an honest aborted-warm row and the case fails —
// DST is NEVER emitted inside an unproven window.
// gnome-shell survival is recorded per row (the crash, if it ever comes, is
// a data point — the session dies loudly and the case fails with it).
// The rows carry outcomes, counters and booleans — never field content
// (D-20/D-21); the corpus is the synthetic ghbdtn class only.

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// The case's journal section (the header text is the contract — the section
// exists in the journal BEFORE the first run; TestSpikeJournalSectionsExist
// pins it) and the frozen experiment constants.
const (
	secTermDstApply = "Фаза 10 round-6 DST-применение (план 10-13 продолжение): " +
		"apply-пробы и хореография на тёплых контекстах"

	// dstDelMark/dstCommitMark are the probe verbs' DEBUG emission records —
	// each emission waits for its own record (the fail-closed proof the
	// emission left the daemon; the probe family's key-trace discipline).
	dstDelMark    = `"msg":"probe delete"`
	dstCommitMark = `"msg":"probe commit"`

	// The erase byte classes a raw pty stream carries for a deleted
	// character: DEL (0x7f — the backspace key's terminal byte) and BS
	// (0x08) — a client synthesizing its own deletion may emit either.
	dstEraseDel byte = 0x7f
	dstEraseBS  byte = 0x08

	// dstTokenErases is the exact erase count a fully-applied DST over the
	// 6-rune token must produce (BuildPlan's Offset=-n NChars=n convention,
	// internal/correct/plan.go — the probe mirrors it verbatim).
	dstTokenErases = 6

	// dstWarmMinAge is the warm gate's time floor on the context's last
	// focus_in — the bisection's trigger boundary IS time (crash at
	// +0.49 s — rows R2/R6b/R7; survive at +1.5 s — row R3, zero pushes;
	// survive at +23 s — row R4): 3 s is a 2× margin over the pinned
	// survive point, and the typed token (real keys through the context)
	// rides before the floor check. dstWarmAgeWait bounds the wait for the
	// floor (a focus-churn storm delays the mint age; expiry = refusal).
	dstWarmMinAge  = 3 * time.Second
	dstWarmAgeWait = 15 * time.Second

	// dstSettleWait is the post-DST settle before the capture read: long
	// enough for the client to apply (or not), and — honestly — long enough
	// for the mutter crash to MANIFEST (the bisection's ~2.5 s), so the
	// survival witness below means something.
	dstSettleWait = 3 * time.Second

	// dstDrainDelay is the choreography's artificial drain-delay between
	// the DST and the commit (the cheap stand-in for the round-4
	// drain-settle window): the order oracle must hold across it.
	dstDrainDelay = time.Second

	// dstCaptureWait bounds the post-emission capture-growth wait (the
	// token already rode the typing; the commit always lands — the proven
	// channel — so growth is the poll).
	dstCaptureWait = 5 * time.Second

	// The two cells' cmd texts (the journal row's command cell).
	dstApplyCmd = "armGate + type ghbdtn → probe-require witness → probe-delete -6 6 → raw pty delta"
	dstChorCmd  = "armGate + type ghbdtn → witness → probe-delete -6 6 → drain-delay 1s →" +
		" probe-commit привет → raw pty порядок"
)

// The verdict classes (content-free row verdicts; the unit corpus pins the
// classification BEFORE the stand runs).
const (
	dstVerdictNoToken = "no-token"
	dstVerdictApplied = "applied"
	dstVerdictIgnored = "ignored"
	dstVerdictPartial = "partial"
	dstVerdictNoWord  = "no-word"
	dstChoreoPass     = "pass"
	dstChoreoConcat   = "concat"
	dstChoreoOrder    = "order-broken"
)

// dstRow appends one dated verdict row to the case's journal section (the
// appendSpikeRow form).
func dstRow(label, cmd, observed, verdict string) error {
	return appendSpikeRow(secTermDstApply, label+" "+time.Now().Format("15:04:05"), cmd, observed, verdict)
}

// dstInspect is the content-free analysis of one raw-capture delta: the
// typed token anchors it (LAST occurrence — the arming residue sits before
// the offset), the erase bytes count strictly AFTER the token, and the
// order predicate reads the replacement word's position against them.
type dstInspect struct {
	tokenSeen bool
	delAfter  int // 0x7f bytes after the token's last byte
	bsAfter   int // 0x08 bytes after the token's last byte
	wordSeen  bool
	orderOK   bool // word present AND every erase byte strictly before the word's first byte
}

// inspectDstDelta computes the case's oracle from one raw delta. The
// erases-before-token residue (the arming pair in a raw capture) never
// counts — the token's LAST occurrence is the anchor.
func inspectDstDelta(delta []byte) dstInspect {
	ins := dstInspect{}
	tokenLast := bytes.LastIndex(delta, []byte(corpusLatin))
	if tokenLast < 0 {
		return ins
	}
	ins.tokenSeen = true
	tail := delta[tokenLast+len(corpusLatin):]
	for _, b := range tail {
		switch b {
		case dstEraseDel:
			ins.delAfter++
		case dstEraseBS:
			ins.bsAfter++
		}
	}
	wordAt := bytes.Index(tail, []byte(drainWordRU))
	if wordAt < 0 {
		return ins
	}
	ins.wordSeen = true
	// The order predicate: every erase byte strictly BEFORE the word's
	// first byte — so the scan flags an erase AFTER the word start (the
	// late-deletion defect shape), never the required leading erases.
	ins.orderOK = true
	for _, b := range tail[wordAt:] {
		if b == dstEraseDel || b == dstEraseBS {
			ins.orderOK = false

			break
		}
	}

	return ins
}

// dstEraseTotal is the applied-erase count of one inspection.
func dstEraseTotal(ins dstInspect) int { return ins.delAfter + ins.bsAfter }

// dstApplyVerdictClass classifies one apply cell: the token must have been
// delivered (the in-run control), then the erase count answers the
// client's honor directly.
func dstApplyVerdictClass(ins dstInspect) string {
	switch {
	case !ins.tokenSeen:
		return dstVerdictNoToken
	case dstEraseTotal(ins) == dstTokenErases:
		return dstVerdictApplied
	case dstEraseTotal(ins) == 0:
		return dstVerdictIgnored
	default:
		return dstVerdictPartial
	}
}

// dstChoreoVerdictClass classifies one choreography cell: the full
// construction's pass needs the applied deletion (all erases) AND the word
// AND the byte order; every deviation gets its own class.
func dstChoreoVerdictClass(ins dstInspect) string {
	switch {
	case !ins.tokenSeen:
		return dstVerdictNoToken
	case !ins.wordSeen:
		return dstVerdictNoWord
	case dstEraseTotal(ins) == 0:
		return dstChoreoConcat
	case dstEraseTotal(ins) != dstTokenErases:
		return dstVerdictPartial
	case !ins.orderOK:
		return dstChoreoOrder
	default:
		return dstChoreoPass
	}
}

// dstApplyVerdictText renders one apply cell's content-free verdict line.
func dstApplyVerdictText(ins dstInspect) string {
	switch class := dstApplyVerdictClass(ins); class {
	case dstVerdictApplied:
		return "Verdict: DST=" + class +
			fmt.Sprintf(" (%d erase-байта после токена в raw pty — удаление ПРИМЕНЕНО клиентом)", dstEraseTotal(ins))
	case dstVerdictIgnored:
		return "Verdict: DST=" + class +
			" (ни одного erase-байта после токена — клиент/стек DST проигнорировал)"
	case dstVerdictPartial:
		return "Verdict: DST=" + class +
			fmt.Sprintf(" (лишь %d из %d erase-байтов)", dstEraseTotal(ins), dstTokenErases)
	default:
		return "Verdict: DST=" + class + " (токен не доехал до pty — контроль доставки)"
	}
}

// dstChoreoVerdictText renders one choreography cell's verdict line.
func dstChoreoVerdictText(ins dstInspect) string {
	switch class := dstChoreoVerdictClass(ins); class {
	case dstChoreoPass:
		return "Verdict: choreography=" + class +
			" (финальная строка = привет: удаление применено, замена села, каждый erase строго прежде слова)"
	case dstChoreoConcat:
		return "Verdict: choreography=" + class +
			" (слово село, но удаление НЕ применилось — в поле ghbdtnпривет)"
	case dstChoreoOrder:
		return "Verdict: choreography=" + class +
			" (erase-байты ПОСЛЕ начала слова — порядок нарушен, класс дефекта владельца)"
	case dstVerdictPartial:
		return "Verdict: choreography=" + class +
			fmt.Sprintf(" (частичное удаление: %d из %d)", dstEraseTotal(ins), dstTokenErases)
	case dstVerdictNoWord:
		return "Verdict: choreography=" + class + " (слово не доехало до pty)"
	default:
		return "Verdict: choreography=" + class + " (токен не доехал до pty — контроль доставки)"
	}
}

// dstFixture is the minimal handle the cells need from either fixture
// family: the raw capture path and the teardown.
type dstFixture interface {
	close()
	pty() string
}

// pty exposes the wezterm round's raw capture path to the case interface.
func (wf *wezFixtureRound) pty() string { return wf.ptyPath }

// pty exposes the VTE fixture's raw capture path to the case interface.
func (tf *termFixture) pty() string { return tf.ptyPath }

// openWezRawFixture is openWezFixture with the RAW pane program (the
// drain-case form, stty raw -echo): the cooked cat of the word case
// line-buffers at the newline and erases the very bytes this case
// measures — the capture must be the byte stream itself.
func (s *stand) openWezRawFixture(ctx context.Context, caseName, tag string) (*wezFixtureRound, error) {
	readyPath := filepath.Join(s.tmpDir, tag+"-ready")
	ptyPath := filepath.Join(s.tmpDir, tag+"-in")
	focusBase := s.focusBaseline()
	wcap, err := startWeztermCapture(ctx, weztermFixturePath(wezFixture), "gs-e2e term-dst-apply",
		drainPane(readyPath, ptyPath)...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", caseName, err)
	}
	wf := &wezFixtureRound{wcap: wcap, readyPath: readyPath, ptyPath: ptyPath}
	if werr := waitPaneReady(ctx, caseName, readyPath); werr != nil {
		wf.close()

		return nil, fmt.Errorf("%s: %w", caseName, werr)
	}
	fmt.Printf("%s: %s raw pane ready — awaiting daemon focus_in (map focus usually self-fires)\n", caseName, tag)
	if werr := s.waitFocusInFrom(ctx, caseName, tag+" wezterm raw pane", focusBase); werr != nil {
		wf.close()

		return nil, fmt.Errorf("%s: %w", caseName, werr)
	}

	return wf, nil
}

// shellAlive is the per-row mutter-survival witness: a live gnome-shell
// process group. The bisection's crash tears the whole session down, so a
// dead pgrep means the emission killed the shell — the loud data point.
func shellAlive(ctx context.Context) bool {
	out, err := runCmd(ctx, "pgrep", "-c", "gnome-shell")
	if err != nil {
		return false
	}
	n, perr := strconv.Atoi(out)

	return perr == nil && n > 0
}

// dstFocusInAge reads the context's freshness off the daemon journal: the
// age of the LAST parseable focus_in record (the INFO engine records — the
// bridge's focus deliveries). Unparseable lines are SKIPPED, never counted
// as fresh: an overestimated age is the unsafe direction. ok=false means
// no parseable record — the caller refuses.
func dstFocusInAge(lines []string, now time.Time) (age time.Duration, ok bool) {
	for i := len(lines) - 1; i >= 0; i-- {
		if !strings.Contains(lines[i], `"msg":"focus_in"`) {
			continue
		}
		ts, err := parseLogTime(lines[i])
		if err != nil {
			continue
		}

		return now.Sub(ts), true
	}

	return 0, false
}

// dstWaitWarmAge waits until the journal's last focus_in is at least
// dstWarmMinAge old (the typed token already flowed by the caller). The
// expiry is the honest refusal — the window discipline never guesses.
func (s *stand) dstWaitWarmAge(ctx context.Context, caseName string) error {
	deadline := time.Now().Add(dstWarmAgeWait)
	for {
		if age, ok := dstFocusInAge(s.logLines(), time.Now()); ok && age >= dstWarmMinAge {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s: warm age floor %v unmet within %v (focus churn?)",
				caseName, dstWarmMinAge, dstWarmAgeWait)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return err
		}
	}
}

// dstWarmGate is the safety gate of EVERY DST emission, in the bisection's
// own terms: the trigger boundary is TIME from the context's fresh
// focus_in (the floor above), and the materialization precondition is the
// TYPED TOKEN (real keys through the context — the mandate's warm shape;
// rows R3/R4 survived DST with ZERO surrounding pushes). The Require-only
// ask is attempted too (R5b-safe in ANY window) and its answer RECORDED —
// a stale cache counts (the 10-08 finding) — but it never gates: the
// 0x9-class clients (the wezterm 10-06 signature) answer nothing, and
// refusing them would make the matrix unmeasurable while the time+keys
// discipline is exactly the pinned survive-shape.
func (s *stand) dstWarmGate(ctx context.Context, caseName, ctlBin string) (push capPush, answered bool, err error) {
	if werr := s.dstWaitWarmAge(ctx, caseName); werr != nil {
		return capPush{}, false, werr
	}
	if _, _, aerr := runCtl(ctx, ctlBin, "probe-require"); aerr != nil {
		return capPush{}, false, fmt.Errorf("%s: warm-gate ask: %w", caseName, aerr)
	}
	baseIdx := -1
	if idx, _, baseOK := s.lastCapPush(); baseOK {
		baseIdx = idx
	}
	_, push, answered, werr := s.waitCapPush(ctx, baseIdx)

	return push, answered, werr
}

// dstTypeToken types the synthetic token into the focused fixture, gated by
// the daemon key-trace (the probe family's delivery proof).
func (s *stand) dstTypeToken(ctx context.Context, caseName string) error {
	keyBase := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, corpusLatin); err != nil {
		return fmt.Errorf("%s: type token: %w", caseName, err)
	}

	return s.waitForNew(ctx, keyTraceMark, keyBase+capsKeyFloor, keyWait)
}

// dstEmitDelete emits the probe-delete covering exactly the typed token
// (offset -6, nchars 6 — the BuildPlan level-1 convention) and waits for
// the daemon's own emission record (the fail-closed proof on the wire).
func (s *stand) dstEmitDelete(ctx context.Context, caseName, ctlBin string) error {
	base := s.countSub(dstDelMark)
	if _, _, err := runCtl(ctx, ctlBin, "probe-delete", "-6", "6"); err != nil {
		return fmt.Errorf("%s: probe-delete: %w", caseName, err)
	}

	return s.waitForNew(ctx, dstDelMark, base+1, keyWait)
}

// dstWaitCaptureGrowth polls the raw capture until it grows past offset and
// returns the growth — the delta one inspection reads.
func dstWaitCaptureGrowth(ctx context.Context, path string, offset int64, timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)
	for {
		if data, err := os.ReadFile(path); err == nil && int64(len(data)) > offset {
			return data[offset:], nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("capture %s never grew past offset %d within %v", path, offset, timeout)
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return nil, err
		}
	}
}

// dstWarmOrRefuse is the cells' shared preamble: the arming gate, the typed
// token, the warm gate (the time floor + the recorded Require answer). On a
// gate failure it lands the honest refusal row and returns the error — the
// caller stops WITHOUT emitting (the warm-discipline refusal is a row and a
// case failure, never a guess).
func (s *stand) dstWarmOrRefuse(
	ctx context.Context, caseName, ctlBin, client, rowTag string,
) (push capPush, answered bool, err error) {
	if aerr := s.termArmGate(ctx, caseName); aerr != nil {
		return capPush{}, false, aerr
	}
	if terr := s.dstTypeToken(ctx, caseName); terr != nil {
		return capPush{}, false, terr
	}
	push, answered, werr := s.dstWarmGate(ctx, caseName, ctlBin)
	if werr == nil {
		return push, answered, nil
	}
	obs := fmt.Sprintf("venue: %s; client=%s; warm_gate=unmet (%v)"+
		" — эмиссия УДЕРЖАНА (дисциплина тёплых контекстов), shell_alive=%v",
		venueMarker(ctx), client, boundRowText(werr.Error()), shellAlive(ctx))
	if rerr := dstRow(rowTag, "armGate + type ghbdtn → warm gate → [эмиссия WITHHELD]",
		obs, "Verdict: aborted-warm (тёплость не доказана — DST запрещён)"); rerr != nil {
		return capPush{}, false, rerr
	}

	return capPush{}, false, fmt.Errorf("%s: %s warm gate unmet — DST withheld: %w", caseName, rowTag, werr)
}

// dstObserved renders the content-free observable cell shared by both cell
// kinds (counters, booleans, the warm ask's answer class — never the field
// content).
func dstObserved(
	ctx context.Context, client string, ins dstInspect, push capPush, answered, alive bool,
) string {
	return fmt.Sprintf("venue: %s; client=%s; token_seen=%v; erase_del=%d; erase_bs=%d;"+
		" erases_total=%d (want %d); word_seen=%v; order_ok=%v; shell_alive=%v;"+
		" push_answered=%v; push_cursor=%d",
		venueMarker(ctx), client, ins.tokenSeen, ins.delAfter, ins.bsAfter,
		dstEraseTotal(ins), dstTokenErases, ins.wordSeen, ins.orderOK, alive,
		answered, rowCursor(int64(push.Cursor), answered))
}

// dstApplyCell drives one apply cell end to end on a warm fixture and lands
// its dated row: arm, type, warm gate, DST(-6,6), settle, the raw delta
// inspection, the survival witness, the verdict.
func (s *stand) dstApplyCell(ctx context.Context, caseName, ctlBin, client, rowTag string, fx dstFixture) error {
	push, answered, err := s.dstWarmOrRefuse(ctx, caseName, ctlBin, client, rowTag)
	if err != nil {
		return err
	}
	offset := captureSize(fx.pty())
	if derr := s.dstEmitDelete(ctx, caseName, ctlBin); derr != nil {
		return derr
	}
	if serr := sleepCtx(ctx, dstSettleWait); serr != nil {
		return serr
	}
	data, gerr := dstWaitCaptureGrowth(ctx, fx.pty(), offset, dstCaptureWait)
	var ins dstInspect
	if gerr == nil {
		ins = inspectDstDelta(data)
	}
	observed := dstObserved(ctx, client, ins, push, answered, shellAlive(ctx))
	if gerr != nil {
		if rerr := dstRow(rowTag, dstApplyCmd, observed, "Verdict: FAIL (pty-дельта не дождалась)"); rerr != nil {
			return rerr
		}

		return gerr
	}
	if !shellAlive(ctx) {
		if rerr := dstRow(rowTag, dstApplyCmd, observed,
			"Verdict: FAIL (gnome-shell НЕ ПЕРЕЖИЛ эмиссию — крах-дата-пойнт)"); rerr != nil {
			return rerr
		}

		return fmt.Errorf("%s: %s gnome-shell died after the warm DST", caseName, rowTag)
	}

	return dstRow(rowTag, dstApplyCmd, observed, dstApplyVerdictText(ins))
}

// dstChoreoCell drives one full round-6 construction: token → warm gate →
// DST(6) → the artificial drain-delay → commit → the byte-order and
// final-line oracle on the raw delta.
func (s *stand) dstChoreoCell(ctx context.Context, caseName, ctlBin, rowTag string, fx dstFixture) error {
	push, answered, err := s.dstWarmOrRefuse(ctx, caseName, ctlBin, capsClientWez, rowTag)
	if err != nil {
		return err
	}
	offset := captureSize(fx.pty())
	if derr := s.dstEmitDelete(ctx, caseName, ctlBin); derr != nil {
		return derr
	}
	if serr := sleepCtx(ctx, dstDrainDelay); serr != nil {
		return serr
	}
	commitBase := s.countSub(dstCommitMark)
	if _, _, cerr := runCtl(ctx, ctlBin, "probe-commit", drainWordRU); cerr != nil {
		return fmt.Errorf("%s: probe-commit: %w", caseName, cerr)
	}
	if werr := s.waitForNew(ctx, dstCommitMark, commitBase+1, keyWait); werr != nil {
		return fmt.Errorf("%s: commit emission record: %w", caseName, werr)
	}
	data, gerr := dstWaitCaptureGrowth(ctx, fx.pty(), offset, dstCaptureWait)
	var ins dstInspect
	if gerr == nil {
		ins = inspectDstDelta(data)
	}
	observed := dstObserved(ctx, capsClientWez, ins, push, answered, shellAlive(ctx))
	if gerr != nil {
		if rerr := dstRow(rowTag, dstChorCmd, observed, "Verdict: FAIL (pty-дельта не выросла)"); rerr != nil {
			return rerr
		}

		return gerr
	}
	if !shellAlive(ctx) {
		if rerr := dstRow(rowTag, dstChorCmd, observed,
			"Verdict: FAIL (gnome-shell НЕ ПЕРЕЖИЛ хореографию)"); rerr != nil {
			return rerr
		}

		return fmt.Errorf("%s: %s gnome-shell died during the choreography", caseName, rowTag)
	}

	return dstRow(rowTag, dstChorCmd, observed, dstChoreoVerdictText(ins))
}

// runTermDstApply drives the round-6 matrix: the enabled-layer daemon, then
// wezterm-apply, VTE-apply and the twice-repeated wezterm choreography —
// each cell warm-gated, each row carrying the survival witness. A failure
// lands the honest-abort row and fails the case (the exit contract).
func runTermDstApply(ctx context.Context, s *stand) error {
	const caseName = "term-dst-apply"
	err := runTermDstCells(ctx, s)
	if err == nil {
		fmt.Printf("%s: apply rows (wez, vte) and both choreography rows landed (rows in %s)\n",
			caseName, spikeJournal)

		return nil
	}
	obs := "case aborted: " + boundRowText(err.Error())
	if rerr := dstRow(caseName, "term-dst-apply round", obs, "Verdict: aborted"); rerr != nil {
		fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, rerr)
	}

	return err
}

// runTermDstCells is the row program of the case (extracted so the
// honest-abort wrapper stays under the complexity budget): one daemon for
// the whole matrix (every cell re-proves its own warm gate).
func runTermDstCells(ctx context.Context, s *stand) error {
	const caseName = "term-dst-apply"
	if err := startAutocorrectConfigDaemon(ctx, s, autocorrectConfigDoc); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}

	// Cell 1: wezterm apply.
	wf, err := s.openWezRawFixture(ctx, caseName, "wez-apply")
	if err != nil {
		return err
	}
	if err := s.dstApplyCell(ctx, caseName, ctlBin, capsClientWez, "wez-apply", wf); err != nil {
		wf.close()

		return err
	}
	wf.close()

	// Cell 2: VTE apply (the differential comparator).
	vf, err := s.openTermFixture(ctx, caseName, "vte-apply", drainPane)
	if err != nil {
		return err
	}
	if err := s.dstApplyCell(ctx, caseName, ctlBin, capsClientVTE, "vte-apply", vf); err != nil {
		vf.close()

		return err
	}
	vf.close()

	// Cells 3-4: the full choreography twice on fresh warm wezterm panes
	// (the two-greens discipline, Pitfall 1).
	for _, tag := range []string{"wez-choreo-a", "wez-choreo-b"} {
		cf, err := s.openWezRawFixture(ctx, caseName, tag)
		if err != nil {
			return err
		}
		if err := s.dstChoreoCell(ctx, caseName, ctlBin, tag, cf); err != nil {
			cf.close()

			return err
		}
		cf.close()
	}

	return nil
}
