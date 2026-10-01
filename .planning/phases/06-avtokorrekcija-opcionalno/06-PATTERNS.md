# Phase 6: Система автокоррекции - Pattern Map

**Mapped:** 2026-09-28
**Files analyzed:** 21 (9 new, 12 modified)
**Analogs found:** 21 / 21 (a mature 4-phase repo: every touched file but the two genuinely-novel mechanisms has an in-repo analog — the planner needs the *in-file precedent* for each seam, named below with line anchors re-verified on branch `gsd/phase-06-avtokorrekcija-opcionalno`; several RESEARCH anchors had drifted and are corrected here)

All analog paths verified git-tracked (`git ls-files` non-empty for every path cited below; `.zcode/`, `.gsd/` excluded). All Go work follows the project skill `.zcode/skills/go-ultimate/SKILL.md` (no `pkg/`, interfaces at point of use, `errors.Is/As`, `fmt.Errorf("...: %w")`, `-race` always, `package xxx_test` tests named `TestF_suffixCamelCase`). Strict TDD: red → green → refactor; `mise run ci` green per commit.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `docs/SPEC.md` (modify: spec-delta §2/§10/§11, D-55 — FIRST plan) | documentation | n/a | itself — §2 item 3 (SPEC.md:31-32), §10 (185-190), §11 (192-196); owner-decision comment style §4.1 (67-71) | exact |
| `docs/adr/ADR-007-autocorrect-*.md` (new) | documentation (ADR) | n/a | `docs/adr/ADR-005-macr-01-owner-checkpoint.md` | exact (format) |
| `layouts/dictgen/main.go` (new, dev generator) | generator (dev tool) | batch + transform (file I/O → Go source) | `layouts/generator/main.go` | exact |
| `layouts/dict_ru.go` / `dict_en.go` / `trigrams.go` (new, generated, committed) | model (baked data) | lookup (binary search — NOT a runtime map, research Q3) | `layouts/tables.go` | exact (form) |
| `layouts/` dict golden tests (new file or tables_test.go extension) | test | lookup | `layouts/tables_test.go` | exact |
| `internal/detect/` (new package: detect.go + golden-corpus test) | service (pure logic) | transform | `internal/correct/direction.go` + `convert.go` (pure-package discipline); verdict shape from 06-RESEARCH Code Examples | role-match |
| `internal/appid/appid.go` (modify: store (sender,path), `Role()`) | service (a11y observer) | event-driven + request-response | itself — `run` state store (174-199), `Start` GetAddress call precedent (126-128), `FocusedApp` (160-168) | exact |
| `internal/appid/appid_test.go` (modify) | test | event-driven | itself — `focusSignal` builder (18-27), synthetic-feed corpus (47-101) | exact |
| `internal/config/config.go` (modify: `autocorrect` section) | config (schema) | load-time validation | itself — `MACR` struct (86-91), `MACR.validate` (218-237), `Defaults` MACR-off (124-129), `maxMACRApps` (25) | exact |
| `internal/config/config_test.go` / `load_test.go` (modify) | test | n/a | itself — MACR validation corpus (per existing macr cases) | exact |
| `internal/session/actor.go` (modify: word-boundary hook, D-53 conjunction, autocorrect counters) | service (event consumer / state machine) | event-driven | itself — `feedKey` (1315-1410), `AppidSource` seam (311-317), `applySnapshot` MACR fold (790-797), `MACRCounters` (359-367), `macrTargetActive`/`warnAppid` (1144-1181) | exact |
| `internal/session/actor_test.go` (modify) | test | event-driven | itself — `fakeSink` (115-227), `fakeAppid` (2937-2945), `TestActor_MACRPerAppMatch` (2947-2989) | exact |
| `internal/correct/buffer.go` (maybe modify: token-finished signal) | model (typed buffer) | transform | itself — `Push` token finish (36-49) | exact |
| `cmd/goswitchd/main.go` (modify: Options wiring) | config (composition root) | wiring | itself — `newActor` SetOptions (137-142) | exact |
| `internal/ctlsvc/ctlsvc.go` (modify: autocorrect status tokens) | service (D-Bus control) | request-response | itself — `renderStatus` (125-159) | exact |
| `cmd/goswitchctl/main.go` (likely zero-change; verify token grammar) | CLI client | request-response | itself — `parseStatusLine` (242-260) | exact |
| `docs/CONFIG.md` (modify: autocorrect section, D-54) | documentation | n/a | itself — macr schema rows (33-36), example (74-79), Privacy section (141-145) | exact |
| `test/e2e/case_autocorrect.go` (new: fires / password-silent / terminal-silent) | e2e test | batch (inject + witness/counters) | `test/e2e/case_macr.go` — `startMacrConfigDaemon` ladder (601-623), `macrConfigTmpl` (468-487) | exact |
| `test/e2e/fixtures/password_entry.py` (new, GTK4 fixture) | e2e fixture (python) | batch (window + stdout oracle) | `test/e2e/focus_helper.py` (witness print grammar) + research-verified prototype (06-RESEARCH Code Examples) | role-match |
| `test/e2e/fixtures/input.html` (modify: + `<input type="password">`) | e2e fixture | n/a | itself (1-10) | exact |
| `test/e2e/main.go` + `mise.toml` (modify: case registry + tasks) | e2e harness / task config | batch | itself — registry `pickCase` (232-276), `caseListUsage` (102-110); mise e2e task entries (mise.toml:39-155) | exact |

## Pattern Assignments

### `docs/SPEC.md` (documentation) — spec-delta §2/§10/§11, BEFORE detector code (D-55)

**Analog:** the file itself. Three edit sites:
- §2 item 3 currently says autocorrect is «явно ВНЕ объёма первой версии; см. §11» (SPEC.md:31-32) — rewrite to point at the new §11 verdict (opt-in, default off, v1.1.0 milestone D-51).
- §10 first bullet «Автокоррекция без горячей клавиши (см. §11)» (SPEC.md:187) — remove/replace.
- §11 (SPEC.md:192-196) currently records the rejection «отвергнута владельцем осознанно… Если когда-либо вернёмся — только отдельной опцией, default off, с white/black list по классам окон» — the delta APPENDS the owner decision 2026-09-27 (D-51): returned exactly on those prescribed terms (opt-in, default off, white-list), milestone v1.1.0.

**Owner-decision annotation style** — the §4.1 HTML comment is the repo's precedent for editing pinned spec text without losing the audit trail (SPEC.md:67-71):
```html
<!-- Правка закреплённой таблицы — решение владельца 2026-09-27
     (quick plan 260927-way): ... устаревшая оговорка «…» снята. -->
```
The §11 delta follows this shape: keep the old verdict visible, append the new dated decision.

---

### `docs/adr/ADR-007-autocorrect-*.md` (documentation, NEW)

**Analog:** `docs/adr/ADR-005-macr-01-owner-checkpoint.md` (103 lines). Format to copy: `## Status` (owner-gate state + dates + verbatim approval quote, ADR-005:3-17) → `## Context` (SPEC § requirement + verbatim D-52/D-53/D-54 blockquotes from 06-CONTEXT.md, ADR-005:19-43) → `## Decision` (candidate | verdict | rationale kill-table — ADR-005:49-57 is the table shape; ADR-007's rows: dictionary path / trigram own-mini-model vs whatlanggo / live GetRole vs cached role / fail-closed vs MACR fail-open) → `## Consequences` (bulleted scope + config/status obligations, ADR-005:85-96) → `## Reversibility` (verbatim D-quote, ADR-005:98-103). ADR-007 must record additionally: the licensing position (hunspell custom-BSD/SCOWL baked with notices; xneur/Easy Switcher GPL read-only reference, zero code copied — the goibus precedent; whatlanggo rejected on size/profile, not law), the GTK3-password live fact (role 61 indistinguishable → role is necessary-not-sufficient), and the honest correction that Easy Switcher contains no dictionary algorithm (06-RESEARCH Q1b).

---

### `layouts/dictgen/main.go` (generator, batch + transform) — NEW dev tool

**Analog:** `layouts/generator/main.go` (466 lines) — the golden-tables pattern this generator copies wholesale. Copy:

**Dev-side tool contract** (generator/main.go:1-10, the load-bearing comment):
```go
// Command generator rebuilds layouts/tables.go from the system XKB symbol
// files and the IBus keysym header.
//
// It is a dev-side tool (CORR-08): `go generate ./layouts` runs it, the
// generated file is committed, and CI only runs the golden tests over the
// committed tables — no xkb dependency in CI.
```
dictgen substitutes: inputs `/usr/share/hunspell/ru_RU.dic` + `en_US.dic` (installed on the dev machine, 06-RESEARCH Q3), outputs `dict_ru.go`/`dict_en.go`/`trigrams.go`; CI never needs hunspell.

**Structure to mirror:** system-input const block (generator/main.go:24-39: path consts + `filePerm = 0o644`); static error values wrapped with `%w` (41-51); `main` → `run` pipeline (79-116: parse → build → emit → `os.WriteFile`); hard errors on data anomalies, never silent skips (the `errConflictingEntry` discipline, 386-391).

**Hunspell parsing specifics (no in-repo analog — nearest is the xkb line parser):** the `.dic` format is `count\nword[/flags]...` per line — strip `/flags`, drop len<2, dedupe, ё→е normalization (Pitfall 3 — decided by the golden corpus), sort, emit. The regex-line-splitting style of `parseSymbols` (generator/main.go:228-269) is the nearest parsing precedent.

**Emission discipline** — deterministic, gofmt-clean by construction (generator/main.go:418-426, 438-466):
```go
const fileHeader = `// Code generated by goswitch/layouts/generator. DO NOT EDIT.

//go:generate go run ./generator

// Package layouts holds ...
package layouts

`
```
`emitTables` runs `format.Source` (444-447); `writeMap` sorts keys so regeneration is byte-for-byte deterministic (452-466). dictgen emits **sorted `[]string` slices** (research-measured: rodata, ~0 heap, `sort.SearchStrings` ≈ 17 comparisons over 146k) — NOT map literals like tables.go (a runtime map costs +10-12 МБ RSS). Each generated file needs its own `//go:generate` header naming `dictgen`, and the license notice (Lebedev custom-BSD for ru, SCOWL for en) goes in the generated file header comment or a sibling `LICENSE-data` doc — planner pins the location.

**mise task** (mise.toml:15-37 task style): a `[tasks.dictgen-regen]`-style entry running `go generate ./layouts`; NOT in `mise run ci` (CI runs only the committed goldens, per the generator contract above).

---

### `layouts/dict_*.go` / `trigrams.go` (generated, committed) + golden tests

**Analog:** `layouts/tables.go` (the committed golden's form: DO NOT EDIT header + go:generate directive + package doc + one documented var per table) and `layouts/tables_test.go` (the golden-test style). Copy from tables_test.go:
- `package layouts_test`, stdlib `testing` only, `t.Parallel()` throughout (tables_test.go:1-7, 27).
- Table-driven pairs with per-pair `t.Run` names (26-83: the SPEC examples corpus — `ghbdtn`→`привет` in three registers both directions).
- A size/invariants test pinning the table's self-consistency (127-144: `TestGolden_TableSize`) — the dict analog pins: both slices sorted (`slices.IsSorted`), mutually disjoint scripts (no Latin word in the ru dict and vice versa), ё-normalized (no `ё` in dict_ru), count within the measured envelope (146 261 ru / 78 951 en — a drifted regeneration fails loudly).
- Corpus MUST include the detector's mandatory golden cases (D-52 Discretion): ё-words (ёлка/ещё), OOV proper names, mixed tokens — these live in `internal/detect` tests, but the dict goldens pin the data side.

---

### `internal/detect/` (service, transform) — NEW pure package

**Analog:** `internal/correct` — the repo's pure-package discipline, stated in its package doc (buffer.go:1-8): «The package is headless: no D-Bus, no clock, no goroutines — the whole corpus runs under -race without a live session». detect copies this exactly: no D-Bus, no logging of the word, no goroutines.

**Classifier precedent** — `correct.Detect` is the shape of the entry point (direction.go:16-42): pure function over `[]rune`, `(value, ok)` returns, documented refusal cases. The research sketch (06-RESEARCH Code Examples) fixes the API surface:
```go
// Verdict — исход детекции; Reasons для счётчиков без слова (D-20/D-21).
type Verdict struct {
	WrongLayout bool
	Confident   bool // словарный путь = true; триграммный = по порогам
	Dir         correct.Dir
	Reason      string // "dict-cur-miss-other-hit" | "trigram" | "abstain-len" | ...
}
func Check(tok []rune, mode string, d Data, t Trigrams, p Params) Verdict
```
**Reuse, don't re-derive:** the remap half is `correct.Convert(tok, dir)` (convert.go:20-52 — register-preserving, bijection, letter-miss fails the whole conversion → reuse that refusal for mixed tokens); direction vocabulary is `correct.Dir`/`ENtoRU`/`RUtoEN` (direction.go:5-14). Dictionary hit = `sort.SearchStrings(dict, string(tok))` — stdlib, no trie/bloom (research "Don't Hand-Roll").

**Test corpus style:** table-driven golden corpus in `package detect_test`, mirroring `layouts/tables_test.go` + the privacy property test (Pitfall 6): the Verdict/reason output must be constructible WITHOUT the word ever reaching a log call — pin by asserting `Reason ∈` the closed reason vocabulary and that no log sink receives token content (`captureLogs` helper exists in actor_test.go for the actor side).

---

### `internal/appid/appid.go` (service, event-driven + request-response) — (sender,path) storage + `Role()`

**Analog:** the file itself. Four in-file precedents compose:

**The event that must be extended** — `run` currently stores ONLY the name on a focus gain (appid.go:192-196):
```go
			if name, gained := focusGain(sig); gained {
				o.mu.Lock()
				o.app = name
				o.mu.Unlock()
			}
```
The signal already carries `sig.Sender` and `sig.Path` (godbus `*dbus.Signal`); the extension stores `(focusSender, focusPath)` beside `o.app` under the same lock. `focusGain` (78-93) and `bridgeApp` (100-108) stay unchanged.

**The read seam** — `FocusedApp` (appid.go:160-168) is the exact shape of the new `Role(ctx)` read (mutex-guarded field snapshot, error return):
```go
func (o *Observer) FocusedApp() (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.app, o.err
}
```
`Role` follows the research-verified form (06-RESEARCH Code Examples — live busctl proof): snapshot `(sender, path)` under the mutex, empty path → `ErrBusClosed`-class «unknown» error (fail-closed, D-53), then one call `o.conn.Object(sender, path).CallWithContext(ctx, "org.a11y.atspi.Accessible.GetRole", 0).Store(&role)`. NOTE: `Observer` today holds no conn — `Start` (118-158) closes over the connection only in the cleanup; `Role` needs the live `*dbus.Conn` stored on the Observer (or a call-through closure field, the `startAppid` seam style). Role enum values are pinned: 40=password text, 60=terminal, 61=text, 79=entry, 94=document text — compare by NUMBER, never `GetRoleName` (localized strings diverge: "text box" vs gi-nick "text").

**The method-call precedent** — the GetAddress round trip (appid.go:126-128) is the only in-repo a11y Object() call; `Role` copies it verbatim in shape:
```go
	if err := sess.Object(a11yBusName, a11yBusPath).
		CallWithContext(ctx, a11yBusName+".GetAddress", 0).Store(&addr); err != nil {
```

**Error discipline** — `ErrBusClosed` sentinel + concrete type (appid.go:41-49, `errors.Is`-friendly, no mutable global) is the pattern for any new "role unknown" sentinel. Every failure is an error return, never a panic (package doc 1-17).

**Test extension** — `appid_test.go` `focusSignal` builder (18-27) gains a sender field; the synthetic-feed corpus (52-101) adds: gain stores the path, loss/non-bridge do NOT clear it (cache-keep semantics match `FocusedApp`), `Role` over a closed/stale path errors → policy silence. `waitApp` poll-until idiom (29-45), never fixed sleeps (except the documented 50 ms drain at 78).

---

### `internal/session/actor.go` (service, event-driven) — word-boundary hook + D-53 conjunction + counters

**Analog:** the file itself. Six precedents compose:

**The word-boundary site** — `feedKey` (actor.go:1315-1410). Two token-finish branches exist today:
1. CORR-09 reset keys (1374-1381): `isResetKeyval` (1759-1766) → `a.buf.HardReset()` — the last completed word is readable BEFORE the reset;
2. separator push (1402-1408 EN, 1392-1399 RU): `a.buf.Push(r)` — the Push itself finishes the token (buffer.go:44-48).
The hook needs the "token just finished" signal: EITHER compare `buf.Token()` before/after `Push` in the actor (zero Buffer change), OR a minimal Buffer extension (a returned flag / callback) — planner picks; the Buffer analog block below has the exact lines. Backspace honestly re-derives coordinates (buffer.go:55-61), so detector state follows the buffer for free. The boundary key TRANSITS as before («the reset is engine state, never consumption», actor.go:1377-1378) — the detector's decision never changes a consume verdict at the boundary itself; a correction fires through the existing pipeline asynchronously.

**The D-53 conjunction site** — every input the policy needs is already actor state, read under `a.mu`: token (`a.buf.Token()`), mode (`a.mode`, 99), caps (`a.caps`, 92 — verify-rung applicability, condition 3), engine (`a.eng`, 93), app list (`a.opts`, 100), appid (`a.appid.FocusedApp()`, 136). The new role query goes through a seam defined at the point of use, exactly like `AppidSource` (actor.go:311-317):
```go
// AppidSource is the per-app identity seam of the MACR layer (ADR-005 a):
// FocusedApp returns the a11y identity of the focused application — the
// live observer of internal/appid or a test double. Defined at the point
// of use; the interface travels with the consumer.
type AppidSource interface {
	FocusedApp() (string, error)
}
```
`RoleSource` (or an extension of AppidSource) follows this verbatim; installers `UseAppid`/`UseAppidStarter` (329-347) are the setter-seam templates.

**The degradation contrast — DIRECTION INVERTED** — `macrTargetActive` is the fail-OPEN precedent (actor.go:1144-1165):
```go
	if a.appid == nil {
		a.warnAppid(nil)

		return true // the degradation rung: the global rule stays in force
	}
```
Autocorrect inverts ONLY the return (any unknown → `false`/silence, fail-closed D-53) while REUSING the warn-once episode pattern `warnAppid` (1167-1181: one WARN per degradation episode, `appidWarned` flag reset on a healthy answer — never per-keystroke spam). `ensureAppid` (1183-1201) is the lazy-start precedent — autocorrect's app-white-list must start the observer on the same discipline (non-empty `apps` list ⇒ start once).

**Off-mutex D-Bus discipline** — `Role()` is a live D-Bus round trip in the hot path; the repo's precedent for off-mutex work is `handleSurroundingLocked` returning an armed payload (actor.go:548-552) + `runClipboardRung` running under `rungMu` with snapshots (1001-1016), and `rungMu` itself (85, WR-01). Either bound the GetRole ctx well under the budget (research: unix-socket RTT ≈ sub-ms; total budget SPEC §5 < 50 ms) inside the locked decision, or hand off like the rung — planner picks; Pitfall 5 requires a perf p95 comparison either way.

**The counters** — `MACRCounters`/`MACRStats` (actor.go:175-180, 359-367) is the exact template for `AutoCorrectCounters`/`AutoCorrectStats`:
```go
type MACRStats struct {
	SuperIntercepted int
	ConsumedUpstream int
}
```
Reason counters follow `skipCorrection` (707-717: `skipReasons map[string]int` + INFO record with the reason slug only). Status extension: `Status` struct (374-385) + `StatusSnapshot` (401-427) gain enabled/counters fields.

**The execution path — REUSE, never a second mechanism** (D-52/Pitfall 7): a detected word corrects through `startRangeCorrection` (1546-1617) / `executeLevel2` (1680-1704) with `correctionRange{token: …, tail: …, replace: a.buf.ReplaceToken}` (1443-1454) — the hook builds the range and calls the SAME entry the combo path uses (feedKey combo branch 1352-1363 is the in-actor caller example). Post-correction flip/settle (`settleCorrectionFlip` 667-680, `settleCombo` 637-648) applies unchanged if reached via the same pipeline.

**CRITICAL logging contrast (Pitfall 6 / D-20/D-21):** `logCorrectionDone` logs `source`/`result` at DEBUG (actor.go:1714-1719). The AUTOCORRECT branch must NOT copy that — the research binds the detector to counts+reasons only, never the word, at ANY level (an autocorrection fires without a user command; the word may be a password). This difference belongs in ADR-007.

**Config fold** — `applySnapshot` (726-798): the MACR fold block (790-797) is the template for the autocorrect fold (enabled/apps/thresholds), including the last-good-name parse-cache discipline (`macrLettersName`, 793-796). `Options` (163-173) gains the autocorrect fields with the same zero-value-off doc convention.

**Test doubles** — actor_test.go `fakeSink` (115-227: mutex-guarded records + unified `opLog` ordering + hooks fired OUTSIDE the lock) and `fakeAppid` (2937-2945: fixed answer/failure struct) are the templates; add `fakeRole` in the same style. `TestActor_MACRPerAppMatch` (2947-2989) is the corpus shape for the D-53 conjunction (in-list/out-of-list/unknown-silence); `TestActor_MACRAppidDegradation` (2991-3017) pins warn-once + the log mark — its autocorrect counterpart pins silence + the WARN mark instead of the global fallback.

---

### `internal/correct/buffer.go` (model, transform) — optional token-finished signal

**Analog:** itself. The token finish lives in `Push` (buffer.go:36-49):
```go
func (b *Buffer) Push(r rune) {
	if tokenCapable(r) {
		b.runes = append(b.runes, r)

		return
	}
	if cur := len(b.runes) - b.curStart; cur > 0 {
		b.lastStart, b.lastLen = b.curStart, cur
	}
	b.runes = append(b.runes, r)
	b.curStart = len(b.runes)
}
```
Lines 44-48 are the "separator finishes the current token" moment the detector hook needs. `Token()` (72-82) then returns the LAST finished token via `activeRange` (136-149) — after a separator, exactly the completed word. `ReplaceToken`'s toggle invariant (97-111) already makes a re-fire safe, and the detector's post-correction word is dictionary-valid → branch 2 → `no` (research debounce note). Keep any extension minimal: the package is pure/headless (package doc 1-8) — no callbacks with side effects, prefer a return value or the actor-side before/after `Token()` comparison.

---

### `internal/config/config.go` (config schema) — `autocorrect` section

**Analog:** itself — the MACR section end to end:
- Struct (config.go:86-91) + `Config` field (95-100; the doc comment «exactly the four sections» updates to five);
- Defaults OFF (124-129: `MACR: MACR{Enabled: false, Letters: "", Apps: nil, AltModifier: ""}`) — autocorrect defaults `enabled: false` + empty white-list (D-54), making the ZERO VALUE the off state (zero-value-Options convention, actor.go:146-162);
- Validation with ceilings and closed vocabularies (218-237): `maxMACRApps = 64` (25) is the app-list ceiling precedent (`errMACRAppsOverCeil`, 47); threshold ranges copy the `errTapWindowRange` shape (41-42, checked at 190-199) — min_word_len/θ ranges each get a sentinel + a `fmt.Errorf("autocorrect.… = %v: %w", …)` line naming the field (D-33);
- `Validate` chain gains one line (143-155);
- Strict decode + hot reload need NO new code — D-32/D-33 propagate to any new section automatically (watch.go/load.go untouched).
- Config-test corpus mirrors the existing macr cases: defaults-off, typo'd key rejects whole file, over-ceiling apps, out-of-range thresholds, empty-enables (if letters-style coupling is chosen for thresholds).

---

### `cmd/goswitchd/main.go` + `internal/ctlsvc/ctlsvc.go` + `cmd/goswitchctl/main.go` (wiring + status surface)

**Analog:** each file itself. `newActor`'s SetOptions call (cmd/goswitchd/main.go:137-142) is where the startup autocorrect options land IF enumerated explicitly — note MACR fields are deliberately ABSENT here today (they flow only via `applySnapshot`/Defaults), so the planner decides: same treatment (fold-only) or explicit fields; either way the succession contract (SetOptions = no-config surface, snapshot wins once a source exists, main.go:118-124) holds.

`renderStatus` (ctlsvc.go:125-159) is the token grammar the autocorrect status joins: counts-only tokens appended before the config block (144-147 appends `super_intercepted`/`super_upstream_consumed` — the exact insertion point), skip-reason flattening `strings.ReplaceAll(reason, "-", "_")` (140-143), `config_error` LAST (148-156). Field names in `session.Status` are the contract (374-385). `goswitchctl` needs NO change if tokens stay one-per-key space-free (`parseStatusLine` 242-260 handles new keys generically; `--json` types ints/bools automatically, 265-294) — verify only that no autocorrect value contains spaces.

---

### `docs/CONFIG.md` (documentation) — autocorrect section (D-54)

**Analog:** itself. The macr rows (CONFIG.md:33-36) are the table form (`| Key | Type | Default | Range / vocabulary | Meaning |`); the activated-MACR example (81-100) is the template for an enabled-autocorrect example; the "exactly four sections" intro (14-15) updates to five; the Privacy section (141-145) extends with the stricter autocorrect rule: the typed/corrected word never enters ANY log level or the status surface. Missing-key semantics note (18-21): pin whether absent autocorrect keys decode off (the macr.alt_modifier empty precedent, config.go:174-180).

---

### `test/e2e/case_autocorrect.go` (e2e test, batch) — NEW

**Analog:** `test/e2e/case_macr.go` — the config-driven case pattern complete:
- The complete-document template const (case_macr.go:468-487: `macrConfigTmpl` — full YAML, no defaults overlay, inline comment discipline) — autocorrect's version adds the `autocorrect:` section with the fixture app in `apps:`;
- The -config spawn ladder (601-623): `activateGoswitch` → `os.WriteFile(cfgPath, …, configFilePerm)` → `s.restartDaemonWithArgs("-config", cfgPath)` (main.go:409-416) → `s.waitForLog(ctx, `"msg":"config loaded"`, registrationWait)` → wait `componentRegisteredMark` → re-`activateGoswitch`;
- The journal-oracle loop (waitForNew/countSub, main.go:611-644) — autocorrect's INFO records (reason slugs, counts) become the marks; `injectText` (main.go:692-701) injects «ghbdtn » + separator for the fires case;
- The negative-oracle style for password-silent: witness the focused input's role+chars before/after (surface.go:426-470 `waitInputFocus` — refuses while a shell PASSWORD entry holds focus, 453-455) + the fixture's stdout oracle;
- Terminal-silent (wezterm, absent from a11y): counters-oracle — `goswitchctl status` shows `autocorrect_…=0` and the -debug log carries no correction records (the case_macr.go pinning-up-to-boundary precedent, 450-462);
- Registration: three entries in `pickCase` (main.go:234-264) + `caseListUsage` (102-110) + the error message registry list (267-272) + mise tasks (`[tasks.e2e-autocorrect-*]`, mise.toml:133-147 style — live cases never in `mise run ci`).

### `test/e2e/fixtures/password_entry.py` (e2e fixture, NEW) and `input.html` (modify)

**Analog (python):** `test/e2e/focus_helper.py` (660 lines, tracked) — the repo's python style: `/usr/bin/python3` + gi, one mode per argv, machine-readable stdout marks (`<app>:<role>:chars=<n>`, focus_helper.py:5, 380; `INPUT_ROLES` including `PASSWORD_TEXT` at 126). The fixture itself is the research-verified ~30-line prototype (06-RESEARCH Code Examples — GtkWindow + GtkEntry + GtkPasswordEntry, timeout argv, `print(f"FINAL:{secret.get_text()}", flush=True)` on exit — the stdout oracle in the zenity spirit). The stand spawns it like any surface (fresh window takes focus; `awaitPidInput`/`waitInputFocus` gates, surface.go:437-470). `input.html` gains `<input type="password">` beside the existing text input (fixtures/input.html:8) — but pin the Chromium password role LIVE before relying on it (research A2: assumed, not verified).

## Shared Patterns

### Interfaces at the point of use
**Source:** `internal/session/actor.go:311-317` (AppidSource doc: «Defined at the point of use; the interface travels with the consumer»), `internal/ctlsvc/ctlsvc.go:47-72` (StatusSnapshotProvider/Reloader/Corrector)
**Apply to:** `RoleSource` — define in session where consumed; test doubles in `_test` files; appid implements it implicitly.

### Fail-closed + warn-once (the inverted MACR ladder)
**Source:** actor.go:1144-1165 (fail-open precedent — DO NOT copy the direction), 1167-1181 (`warnAppid` episode pattern — DO copy), 1183-1201 (`ensureAppid` lazy start)
**Apply to:** every D-53 unknown: no tree / no focus event / role error / app not in list / caps absent / detector unsure → silence + at most one WARN per episode. ADR-007 records the contrast with ADR-005.

### Golden generator discipline (dev-side, committed, CI goldens only)
**Source:** `layouts/generator/main.go:1-10` (contract), 418-466 (deterministic emission); `layouts/tables_test.go` (committed-table tests); mise.toml:39-40 (live/e2e never in ci)
**Apply to:** dictgen + dict/trigram goldens + `[tasks.dictgen-regen]`.

### Logging privacy (D-20/D-21) — stricter for autocorrect
**Source:** actor.go:369-373 (Status doc: counts and states ONLY), 707-717 (reason-slug INFO), 1714-1719 (the DEBUG source/result detail — explicitly NOT for autocorrect)
**Apply to:** detector verdicts, autocorrect counters, e2e marks: reason slugs from a closed vocabulary; the word never appears at any level.

### Deadline-bounded bus calls; nothing slow under the actor mutex
**Source:** actor.go:85 (`rungMu`, WR-01), 548-552 (armed-payload handoff), 1001-1016 (snapshot-then-run); appid.go:126-128 (CallWithContext precedent)
**Apply to:** the live `GetRole` round trip — deadline ctx (sub-ms expected, budget < 50 ms) or off-mutex handoff; Pitfall 5 perf check after.

### Strict-decode sections with ceilings (D-33 + ASVS V5)
**Source:** config.go:25 (`maxMACRApps = 64`), 40-49 (sentinels), 218-237 (validate)
**Apply to:** the `autocorrect` section — app-list cap, threshold ranges, every error names its field; a bad section invalidates the WHOLE file.

### Status token grammar
**Source:** ctlsvc.go:125-159 (`renderStatus`: counts tokens, dash-flattened skip reasons, config_error LAST), cmd/goswitchctl/main.go:242-260
**Apply to:** `autocorrect_enabled=…`, `autocorrect_fired=…`, reason counters — space-free single tokens, appended before the config block.

### e2e config-case ladder
**Source:** case_macr.go:601-623 (`startMacrConfigDaemon`), main.go:409-416 (`restartDaemonWithArgs`)
**Apply to:** every autocorrect case needing a non-default config; wait `"msg":"config loaded"` then `componentRegisteredMark` before activating.

### TDD + green-iteration gates
**Source:** CONVENTIONS.md directives 1-3; go-ultimate SKILL.md
**Apply to:** every task — red → green → refactor; `mise run ci` (build + vet + golangci-lint + test -race) green per commit; zero new Go dependencies (tidy-diff gate; whatlanggo rejected).

## No Analog Found

| Mechanism | Novelty | Nearest precedent (partial) |
|-----------|---------|---------------------------|
| Trigram scoring model (own mini-model) | No statistical/n-gram code exists in the repo | `internal/correct/direction.go` (pure classifier discipline); algorithm + thresholds from 06-RESEARCH Q2/Detector Sketch; thresholds pinned by the mandatory golden corpus (Discretion) |
| Hunspell `.dic` parser (flag stripping, ё-normalization) | No dictionary-data parsing exists | `layouts/generator/main.go:228-269` (xkb line parser style: regex-per-line, hard errors); output-side fully analogged |
| Dictionary `sort.SearchStrings` lookup | No binary-search lookup exists (tables are maps) | Research-measured choice (Q3): sorted slice ≈ 17 comparisons, rodata; stdlib — nothing to copy |

All three are data/algorithm work with the research sketch as the spec — not copy-paste from any repo. Everything else in the phase (bus, focus, replacement, config, status, e2e) extends existing seams named above.

## Metadata

**Analog search scope:** entire module tree (`engine/`, `internal/`, `cmd/`, `layouts/`, `test/e2e/`, `docs/`, `mise.toml`) — 76 tracked .go files; `.zcode/skills/go-ultimate` consulted for conventions only (never an analog)
**Tracked-source gate:** every analog path verified via `git ls-files -- <path>` (all non-empty); branch `gsd/phase-06-avtokorrekcija-opcionalno`; no mirror/untracked paths cited
**Line-anchor verification:** all anchors re-read on the working tree 2026-09-28 — RESEARCH.md anchors that drifted on this branch and are corrected here: AppidSource 301→311-317, feedKey 1225→1315-1410, isResetKeyval 1597→1759-1766, MACRCounters 348→359-367, applySnapshot 277→726-798, appid wire-facts/Start 118-158 confirmed unchanged
**Files deep-read:** 15 full (appid.go, appid_test.go, config.go, buffer.go, convert.go, direction.go, actor.go, generator/main.go, tables.go, tables_test.go, ctlsvc.go, goswitchctl/main.go, test/e2e/main.go, SPEC.md, CONFIG.md, ADR-005, input.html, mise.toml) + targeted ranges in actor_test.go, case_macr.go, surface.go, case_m1.go (grep-verified), cmd/goswitchd/main.go, focus_helper.py (grep-verified)
**Pattern extraction date:** 2026-09-28
