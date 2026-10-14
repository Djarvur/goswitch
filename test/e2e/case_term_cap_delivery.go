package main

// The phase-10 round-3 delivery-capability probe case (plan 10-13, G-10-1
// cycle 3): three probes × two terminal clients on the VM stand, dated rows
// to the journal section «Фаза 10 round-3 caps-проба доставки» — the
// machine-visible routing input plan 10-14 selects its delivery branch by.
//
// Probes per client (each on a FRESH focused fixture — an empty field is
// the known-state premise of the cursor arithmetic):
//
//	P1 (DST honored?) — type the 6-rune token, then the probe verb
//	                   probe-delete-surrounding 0 6 (DST + Require): does a
//	                   Require answer show the field shortened by exactly
//	                   nchars (cursor 6 → 0)?
//	P2 (receipt)      — type the token, probe-commit <привет>: does a
//	                   Require answer reflect the just-committed text
//	                   (cursor 6 → 12)?
//	P3 (drain-proof)  — type the token, 6× probe-forward BackSpace through
//	                   the daemon's own channel: does a Require answer
//	                   reflect the drained field (cursor 6 → 0)?
//
// The Require poll rides the DST(0,0) no-op ask (the plan's read_first
// decision: one verb — DeleteSurroundingText(0,0) deletes nothing, the
// Require forces the client's report). Every poll cycle is bounded by the
// case's internal constants (T-10-13-03) — a silent client yields an honest
// never-row, never a hang; the journal-derived window budgets for 10-14 are
// Task 3's derivation, never a config key. The rows carry outcomes, ask
// counts and latencies only — the field content lives in the daemon's
// -debug journal and never enters the spike-log (D-20/D-21); the CLIPBOARD
// sentinel rides through a full probe round untouched (D5).

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// The case's journal section (the header text is the contract; the section
// pins itself in spike_journal_test.go) and the frozen probe constants.
const (
	secCapsDel = "Фаза 10 round-3 caps-проба доставки (план 10-13): P1/P2/P3 — DST honored, receipt, drain-proof"

	// Row client tags (the verdict table's rows key on these).
	capsClientWez = "wezterm-nightly"
	capsClientVTE = "gnome-terminal-vte"

	// capsToken/capsWord are the probe corpus: the 6-rune synthetic token
	// and its conversion (the ghbdtn→привет class; synthetic only, Pitfall
	// 6). The deltas are the cursor arithmetic of the reflection predicates.
	capsToken      = corpusLatin
	capsWord       = "привет"
	capsTokenRunes = 6
	capsWordRunes  = 6

	// capsPollAsks bounds one probe's Require poll — the internal budget
	// constant (NOT a config key): a silent client exhausts it into an
	// honest never-row (T-10-13-03). capsPollGap paces the poll at the
	// level-2 burst quantum.
	capsPollAsks = 15
	capsPollGap  = 20 * time.Millisecond

	// capsPushWait bounds the wait for ONE ask's surrounding push to land
	// in the daemon journal (an ask is only consumed when its push lands
	// or the bound expires). 5 s: the nightly's stale-cache answer rides
	// its own cadence (the 10-08 drift rows) — the budget must not be the
	// reason an answer is missed; on a genuinely silent client the bound
	// is what makes the poll MEASURED silence instead of a hang.
	capsPushWait = 5 * time.Second

	// capsForwardCount is the P3 drain: 6 BackSpaces erase the 6-rune token.
	capsForwardCount = 6

	// capsKeyFloor is the key-trace floor of one typed token (6 keys ×
	// press+release).
	capsKeyFloor = 12

	// capsAskMark/capsCommitMark are the probe verbs' DEBUG emission
	// records — the latency anchors read them (probeForwardMark already
	// sits in wezterm.go). capsPushMark is the actor's DEBUG surrounding
	// record — the wire answer with the full text and cursor
	// (HandleSurroundingText); the reflection oracles read it.
	// (case_select.go's surroundingMark is the engine-side position trace
	// of the same event.)
	capsAskMark    = `"msg":"probe delete-surrounding"`
	capsCommitMark = `"msg":"probe commit"`
	capsPushMark   = `"msg":"surrounding push"`

	// The three probe names (row tags and predicate keys).
	capsP1 = "P1"
	capsP2 = "P2"
	capsP3 = "P3"
)

// capFixture is the minimal handle the probes need from either fixture
// family (both closes reap the window by its tmp markers).
type capFixture interface{ close() }

// capPush is the parsed form of one surrounding push record (the content
// stays inside the daemon journal; the ROWS carry cursor numbers and
// booleans only).
type capPush struct {
	Text   string `json:"text"`
	Cursor uint32 `json:"cursor"`
}

// capProbeResult is one probe's outcome — exactly what the row carries
// (content-free by construction).
type capProbeResult struct {
	probe        string // P1 | P2 | P3
	honored      bool
	asks         int // Require asks consumed until the reflecting answer (the action's own Require counts for P1)
	never        bool
	latencyMS    int64 // emission → reflecting push, journal timestamps (-1 unknown)
	baselineCur  int64
	reflectCur   int64 // the reflecting answer's cursor (-1 when never)
	baselineOK   bool  // the baseline ask was ANSWERED (a blind baseline anchors nothing)
	baselineSeen bool  // the baseline answer carried the typed token (the channel tracked the typing)
}

// capsRow appends one dated verdict row to the case's journal section (the
// appendSpikeRow form).
func capsRow(label, cmd, observed, verdict string) error {
	return appendSpikeRow(secCapsDel, label+" "+time.Now().Format("15:04:05"), cmd, observed, verdict)
}

// lastCapPush scans the journal from the end for the newest surrounding
// push record; ok reports whether one exists at all.
func (s *stand) lastCapPush() (idx int, push capPush, ok bool) {
	lines := s.logLines()
	for i := len(lines) - 1; i >= 0; i-- {
		if !strings.Contains(lines[i], capsPushMark) {
			continue
		}
		var p capPush
		if err := json.Unmarshal([]byte(lines[i]), &p); err != nil {
			return -1, capPush{}, false
		}

		return i, p, true
	}

	return -1, capPush{}, false
}

// waitCapPush polls the journal until a push lands ABOVE the given line
// index (a genuinely NEW record — never an older answer re-read), bounded
// by capsPushWait. The bound's expiry is MEASURED SILENCE (ok=false) — the
// honest never-input of the poll, never a case failure (T-10-13-03).
func (s *stand) waitCapPush(ctx context.Context, afterIdx int) (idx int, push capPush, answered bool, err error) {
	deadline := time.Now().Add(capsPushWait)
	for {
		idx, push, ok := s.lastCapPush()
		if ok && idx > afterIdx {
			return idx, push, true, nil
		}
		if time.Now().After(deadline) {
			return idx, push, false, nil
		}
		if err := sleepCtx(ctx, witnessPoll); err != nil {
			return idx, push, false, err
		}
	}
}

// capAsk drives ONE Require ask through the DST(0,0) no-op poll form and
// waits for its push (the plan's read_first decision: one verb, the
// zero-char delete is a no-op, the Require forces the client's report).
// ok reports whether the ask was ANSWERED within the bound; the returned
// idx feeds the next ask's freshness bound either way.
func (s *stand) capAsk(
	ctx context.Context, caseName, ctlBin string, afterIdx int,
) (idx int, push capPush, answered bool, err error) {
	if _, _, err := runCtl(ctx, ctlBin, "probe-delete-surrounding", "0", "0"); err != nil {
		return afterIdx, capPush{}, false, fmt.Errorf("%s: poll ask: %w", caseName, err)
	}

	return s.waitCapPush(ctx, afterIdx)
}

// capLineTime parses one journal record's timestamp (the latency anchor).
func capLineTime(s *stand, idx int) (time.Time, bool) {
	lines := s.logLines()
	if idx < 0 || idx >= len(lines) {
		return time.Time{}, false
	}
	ts, err := parseLogTime(lines[idx])

	return ts, err == nil
}

// capLatency resolves the journal latency between the action's emission
// record (the last one above emitLine carrying the mark) and the reflecting
// push at pushIdx — -1 when either record is unparseable (the honest
// unknown).
func capLatency(s *stand, emitLine int, emitMark string, pushIdx int) int64 {
	lines := s.logLines()
	emitIdx := -1
	for i := len(lines) - 1; i >= emitLine && i >= 0; i-- {
		if strings.Contains(lines[i], emitMark) {
			emitIdx = i

			break
		}
	}
	tEmit, okE := capLineTime(s, emitIdx)
	tPush, okP := capLineTime(s, pushIdx)
	if !okE || !okP {
		return -1
	}

	return tPush.Sub(tEmit).Milliseconds()
}

// capReflecting applies the probe's reflection predicate: cursor arithmetic
// against the baseline answer plus the token's presence class — content
// never leaves the journal, the predicate reads numbers and booleans.
func capReflecting(res *capProbeResult, base, push capPush) bool {
	if !res.baselineOK {
		return false // a blind baseline cannot anchor the arithmetic
	}
	switch res.probe {
	case capsP1: // the field shortened by exactly nchars
		return int64(push.Cursor) == int64(base.Cursor)-capsTokenRunes && !strings.Contains(push.Text, capsToken)
	case capsP2: // the just-committed text visible at the cursor
		return int64(push.Cursor) == int64(base.Cursor)+capsWordRunes && strings.Contains(push.Text, capsWord)
	case capsP3: // the drained field (the forwarded BackSpaces applied)
		return int64(push.Cursor) == int64(base.Cursor)-capsTokenRunes && !strings.Contains(push.Text, capsToken)
	}

	return false
}

// capTypeToken types the synthetic token into the focused fixture, gated by
// the daemon key-trace (never an AT-SPI witness — Pitfall 7).
func (s *stand) capTypeToken(ctx context.Context, caseName string) error {
	keyBase := s.countSub(keyTraceMark)
	if err := s.injectText(ctx, capsToken); err != nil {
		return fmt.Errorf("%s: type token: %w", caseName, err)
	}

	return s.waitForNew(ctx, keyTraceMark, keyBase+capsKeyFloor, keyWait)
}

// capAction drives the probe's action and answers whether it carries its
// own Require ask (only P1's verb does — DST + Require in one emission).
func (s *stand) capAction(ctx context.Context, caseName, ctlBin, probe string) (ownAsk bool, err error) {
	switch probe {
	case capsP1:
		_, _, err = runCtl(ctx, ctlBin, "probe-delete-surrounding", "0", "6")

		return true, err
	case capsP2:
		_, _, err = runCtl(ctx, ctlBin, "probe-commit", capsWord)

		return false, err
	case capsP3:
		for range capsForwardCount {
			if _, _, ferr := runCtl(ctx, ctlBin, "probe-forward", "BackSpace"); ferr != nil {
				return false, ferr
			}
		}

		return false, nil
	}

	return false, fmt.Errorf("%s: unknown probe %q", caseName, probe)
}

// capActionMark names the emission record the probe's latency anchors on.
func capActionMark(probe string) string {
	switch probe {
	case capsP1:
		return capsAskMark
	case capsP2:
		return capsCommitMark
	default:
		return probeForwardMark
	}
}

// capRunProbe drives one probe end to end on a fresh fixture and lands its
// dated row; the honest never-row replaces a hung poll (T-10-13-03). spawn
// opens one fresh focused fixture of the client under test.
func (s *stand) capRunProbe(
	ctx context.Context, caseName, ctlBin, client, probe, rowTag string,
	spawn func(tag string) (capFixture, error),
) error {
	fx, err := spawn("cap-" + strings.ToLower(probe))
	if err != nil {
		return err
	}
	defer fx.close()
	if err := s.capTypeToken(ctx, caseName); err != nil {
		return err
	}

	// Baseline ask: the channel's answer to the KNOWN typed state — the
	// arithmetic anchor of the reflection predicate. An unanswered
	// baseline is the measured blindness of the channel (the probes still
	// run — the rows carry baseline_cursor=-1 semantics).
	_, basePush, baseOK, err := s.capAsk(ctx, caseName, ctlBin, -1)
	if err != nil {
		return fmt.Errorf("%s: %s baseline ask: %w", caseName, probe, err)
	}
	res := capProbeResult{
		probe:        probe,
		asks:         1, // the baseline ask
		latencyMS:    -1,
		baselineCur:  int64(basePush.Cursor),
		reflectCur:   -1,
		baselineOK:   baseOK,
		baselineSeen: baseOK && strings.Contains(basePush.Text, capsToken),
	}

	// The action; its emission record (the last one after this mark) is the
	// latency anchor.
	emitLine := len(s.logLines())
	ownAsk, aerr := s.capAction(ctx, caseName, ctlBin, probe)
	if aerr != nil {
		return fmt.Errorf("%s: %s action: %w", caseName, probe, aerr)
	}
	if ownAsk {
		res.asks++ // P1's action IS the first poll ask
	}

	// The budgeted poll: P1's first answer is already on the wire above
	// emitLine; P2/P3 start asking here. The FIRST reflecting answer wins
	// with its journal latency; exhaustion is the honest never.
	refIdx, refPush, err := s.capPollForResult(ctx, caseName, ctlBin, &res, basePush, emitLine, ownAsk)
	if err != nil {
		return err
	}
	switch {
	case refIdx >= 0:
		res.honored = true
		res.reflectCur = int64(refPush.Cursor)
		res.latencyMS = capLatency(s, emitLine, capActionMark(probe), refIdx)
	default:
		res.never = true
	}

	return capsProbeRow(ctx, client, rowTag, res)
}

// capPollForResult runs the budgeted Require poll and returns the first
// reflecting answer (its journal line and record), or (-1, zero, nil) when
// the budget exhausted — the honest never. P1's action ask (ownAsk) is
// already on the wire above emitLine and counts as the first ask.
func (s *stand) capPollForResult(
	ctx context.Context, caseName, ctlBin string, res *capProbeResult,
	basePush capPush, emitLine int, ownAsk bool,
) (int, capPush, error) {
	if ownAsk {
		idx, push, ok, werr := s.waitCapPush(ctx, emitLine)
		if werr == nil && ok && capReflecting(res, basePush, push) {
			return idx, push, nil
		}
	}
	pollIdx, _, _ := s.lastCapPush()
	for res.asks-1 < capsPollAsks {
		var push capPush
		var ok bool
		var perr error
		pollIdx, push, ok, perr = s.capAsk(ctx, caseName, ctlBin, pollIdx)
		res.asks++
		if perr != nil {
			return -1, capPush{}, fmt.Errorf("%s: %s poll: %w", caseName, res.probe, perr)
		}
		if ok && capReflecting(res, basePush, push) {
			return pollIdx, push, nil
		}
		if gerr := sleepCtx(ctx, capsPollGap); gerr != nil {
			return -1, capPush{}, gerr
		}
	}

	return -1, capPush{}, nil
}

// rowCursor renders a cursor observable: the honest -1 for a blind
// (unanswered) baseline ask.
func rowCursor(cur int64, ok bool) int64 {
	if !ok {
		return -1
	}

	return cur
}

// capsProbeRow lands one probe's dated verdict row (content-free: cursors,
// booleans, ask counts, latencies).
func capsProbeRow(ctx context.Context, client, rowTag string, res capProbeResult) error {
	cmd := map[string]string{
		capsP1: "type ghbdtn → probe-delete-surrounding 0 6 → polls DST(0,0)+Require",
		capsP2: "type ghbdtn → probe-commit привет → polls DST(0,0)+Require",
		capsP3: "type ghbdtn → 6× probe-forward BackSpace → polls DST(0,0)+Require",
	}[res.probe]
	observed := fmt.Sprintf(
		"venue: vm-primary (%s); client=%s; probe=%s; baseline_cursor=%d; honored=%v; asks=%d;"+
			" latency_ms=%d; never=%v; reflect_cursor=%d; baseline_token_seen=%v",
		venueMarker(ctx), client, res.probe, rowCursor(res.baselineCur, res.baselineOK),
		res.honored, res.asks, res.latencyMS, res.never, res.reflectCur, res.baselineSeen)
	verdict := fmt.Sprintf("Verdict: %s=%s", res.probe, map[bool]string{true: "yes", false: "no"}[res.honored])
	switch {
	case res.never:
		verdict += " (never — бюджет опроса исчерпан без отражающего ответа)"
	case res.honored:
		verdict += fmt.Sprintf(" (отражение на опросе %d, латентность %d мс)", res.asks, res.latencyMS)
	}

	return capsRow(rowTag, cmd, observed, verdict)
}

// capClipRow is the frozen CLIPBOARD sentinel of the probe set (D5): the
// sentinel rides CLIPBOARD through one full probe round and wl-paste must
// read it back byte-identical — the probes never touch the buffers.
func (s *stand) capClipRow(
	ctx context.Context, caseName, ctlBin, client string, spawn func(tag string) (capFixture, error),
) error {
	const rowTag = "clip-frozen"
	if err := clipCopy(ctx); err != nil {
		return fmt.Errorf("%s: %w", caseName, err)
	}
	before := clipReadNoNewline(ctx)
	if before != primarySentinelMark {
		return fmt.Errorf("%s: CLIPBOARD pre-state %q, want sentinel %q", caseName, before, primarySentinelMark)
	}
	// One full probe round rides under the sentinel (a fresh fixture, the
	// typed token, one DST probe) — the buffer must survive it.
	fx, err := spawn("cap-clip")
	if err != nil {
		return err
	}
	defer fx.close()
	if err := s.capTypeToken(ctx, caseName); err != nil {
		return err
	}
	if _, _, err := runCtl(ctx, ctlBin, "probe-delete-surrounding", "0", "6"); err != nil {
		return fmt.Errorf("%s: clip probe action: %w", caseName, err)
	}
	if err := sleepCtx(ctx, capsPollGap); err != nil {
		return err
	}
	after := clipReadNoNewline(ctx)
	observed := fmt.Sprintf("venue: vm-primary (%s); client=%s; clip_before=%s; clip_after=%s"+
		" (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием)",
		venueMarker(ctx), client, before, after)
	cmd := "printf %s GS-SPIKE-SENTINEL | wl-copy → probe round → wl-paste --no-newline"
	if after == primarySentinelMark {
		return capsRow(rowTag, cmd, observed, "Verdict: pass (CLIPBOARD байт-нетронут пробами доставки)")
	}

	return capsRow(rowTag, cmd, observed, "Verdict: FAIL (клипборд изменён пробами)")
}

// runTermCapDelivery drives the phase-10 round-3 delivery-capability probes
// on the stand: the enabled-layer daemon, then the three probes × two
// clients + the CLIPBOARD sentinels. A failure lands the honest-abort row
// and fails the case (the exit contract) — never a retry loop.
func runTermCapDelivery(ctx context.Context, s *stand) error {
	const caseName = "term-cap-delivery"
	err := runTermCapProbes(ctx, s)
	if err == nil {
		fmt.Printf("%s: six probe rows (P1/P2/P3 × 2 clients) and clip sentinels landed (rows in %s)\n",
			caseName, spikeJournal)

		return nil
	}
	obs := "case aborted: " + boundRowText(err.Error())
	if rerr := capsRow(caseName, "term-cap-delivery probe round", obs, "Verdict: aborted"); rerr != nil {
		fmt.Fprintf(os.Stderr, "e2e: %s deferred journal row: %v\n", caseName, rerr)
	}

	return err
}

// runTermCapProbes is the row program of the case (extracted so the
// honest-abort wrapper stays under the complexity budget): wezterm first
// (the mandate's client), then VTE.
func runTermCapProbes(ctx context.Context, s *stand) error {
	const caseName = "term-cap-delivery"
	if err := startAutocorrectConfigDaemon(ctx, s, autocorrectConfigDoc); err != nil {
		return err
	}
	ctlBin, err := s.buildCtl(ctx)
	if err != nil {
		return err
	}

	clients := []struct {
		name  string
		spawn func(tag string) (capFixture, error)
	}{
		{capsClientWez, func(tag string) (capFixture, error) {
			return s.openWezFixture(ctx, caseName, tag)
		}},
		{capsClientVTE, func(tag string) (capFixture, error) {
			return s.openTermFixture(ctx, caseName, tag, capturePane("bash", "-c"))
		}},
	}
	for _, cl := range clients {
		for i, probe := range []string{capsP1, capsP2, capsP3} {
			rowTag := fmt.Sprintf("p%d-%s", i+1, strings.ToLower(cl.name))
			if err := s.capRunProbe(ctx, caseName, ctlBin, cl.name, probe, rowTag, cl.spawn); err != nil {
				return err
			}
		}
		if err := s.capClipRow(ctx, caseName, ctlBin, cl.name, cl.spawn); err != nil {
			return err
		}
	}

	return nil
}
