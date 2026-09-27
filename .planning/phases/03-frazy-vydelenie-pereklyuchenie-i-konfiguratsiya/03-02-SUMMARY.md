---
phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya
plan: 02
subsystem: input-correction
tags: [config, yaml, fsnotify, hot-reload, hotkey-bindings, tdd, docs]

# Dependency graph
requires:
  - phase: 01-adr-paket-i-karkas-ibus-dvizhka
    provides: mise ci gate (D-08), strict-lint discipline (D-10), pure-package dependency direction
  - phase: 03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya plan 01
    provides: DefaultBackspaceCap=50 awaiting YAML wiring, actor window seam (NewActor(window))
provides:
  - internal/config: Config/Hotkeys/Timeouts/Correction/MACR types, Defaults(), Validate() (named-field errors), Load(path) strict KnownFields decode (D-33)
  - config.NewWatcher(ctx, path) + Watcher.Snapshot() config.Config (value-copy; the pinned 03-04 consumer contract interface{ Snapshot() config.Config }) + Watcher.LastError() (status surface 03-06); WARN "config reload rejected"; last-good (D-32); 200 ms debounce; directory-watch
  - internal/hotkey.ParseBinding(string) (Binding, error), Binding{Keyval, ModMask uint32}, closed name tables (shift_l..super_r; modifiers shift/ctrl/alt/super) with ibuskeysyms.h provenance
  - goswitchd -config flag: startup Load (refusal = visible exit 1), FSM window from tap_window_ms, watcher on signal-ctx
  - docs/CONFIG.md schema reference + Caramba correspondence table (CONF-03, no name cloning D-31)
affects: [03-03-selection, 03-04-config, 03-06-ctl, 03-07-matrix-v2]

actuals:
  tokens: 17662   # chars/4 over the realized diff (git diff 57a3a67..HEAD = 70649 chars)
  tasks: 3
  commits: 5      # measured: git rev-list --count 57a3a67..HEAD
  plan_head_before: 57a3a67fb6335e59f566cef09362081c7e650589

tech-stack:
  added:
    - "github.com/fsnotify/fsnotify v1.10.1 (the phase's single approved dependency — proxy.golang.org verified, STACK/RESEARCH audit Approved; tidy-clean)"
  patterns:
    - "directory-watch + debounce (STACK): fsnotify on filepath.Dir, base-name filter, Write|Create|Remove|Rename, AfterFunc re-arm (armTimer discipline)"
    - "EventSource interface seam: the watcher corpus runs headless on synthetic events, no real inotify"
    - "atomic.Pointer[Config] snapshot publication; value-copy Snapshot() pinned as the consumer contract"

key-files:
  created:
    - internal/config/config.go
    - internal/config/config_test.go
    - internal/config/load.go
    - internal/config/load_test.go
    - internal/config/watch.go
    - internal/config/watch_test.go
    - internal/hotkey/names.go
    - internal/hotkey/names_test.go
    - docs/CONFIG.md
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-02-task1-red.json
    - .planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/red-evidence/03-02-task2-red.json
  modified:
    - cmd/goswitchd/main.go
    - go.mod
    - go.sum

key-decisions:
  - "Binding.ModMask = held modifiers OR the bound key's own family bit — the full wire state of the bound key's event (a Shift_R press carries MaskShift; held Shift + pressed Ctrl_R carries Shift|Control). Matches the plan's shift_r literal; refines the combo literal, which no uniform rule could satisfy as written (see Deviations 1)"
  - "A config document must be COMPLETE: no overlay on Defaults() (the matrix.go strict precedent) — Defaults() is the no-flag path only; an explicit -config that omits keys is a validation error, and an empty/missing file is a visible start refusal"
  - "fsnotify sits behind an exported EventSource interface; options WithSource/WithLoader/WithDebounce are the test seams — the corpus is headless (synthetic events), the daemon gets real inotify"
  - "reload() WARNs BEFORE storing LastError so the log record always precedes the observable state (test-race ordering)"

requirements-completed: [CONF-01, CONF-02, CONF-03, SWCH-04]

coverage:
  - id: D1
    description: "YAML schema of all daemon parameters (hotkeys/timeouts/correction/macr, D-31) with strict KnownFields decode — a typo'd key rejects the whole document with the field named (D-33); ranges with DoS ceilings (T-03-02-02)"
    requirement: CONF-01
    verification:
      - kind: unit
        ref: internal/config/load_test.go#TestLoad_UnknownKeyRejected
        status: pass
      - kind: unit
        ref: internal/config/config_test.go#TestValidate_TimeoutRanges
        status: pass
      - kind: unit
        ref: internal/config/config_test.go#TestValidate_CorrectionRanges
        status: pass
      - kind: unit
        ref: internal/config/config_test.go#TestDefaults
        status: pass
    human_judgment: false
  - id: D2
    description: "Closed binding-name tables in the pure hotkey package: ParseBinding resolves shift_l..super_r (keyvals verbatim from ibuskeysyms.h:188-199) and shift/ctrl/alt/super masks; unknown token rejects the config (D-33)"
    requirement: CONF-01
    verification:
      - kind: unit
        ref: internal/hotkey/names_test.go#TestParseBinding
        status: pass
      - kind: unit
        ref: internal/hotkey/names_test.go#TestParseBinding_KeyvalProvenance
        status: pass
      - kind: unit
        ref: internal/config/config_test.go#TestValidate_Bindings
        status: pass
    human_judgment: false
  - id: D3
    description: "Hot-reload watcher (CONF-02 infrastructure half): directory-watch + 200 ms debounce + atomic.Pointer publication; invalid rewrite WARNs config reload rejected, last-good keeps serving, LastError exposed (D-32); snapshots await the 03-04 actor consumer"
    requirement: CONF-02
    verification:
      - kind: unit
        ref: internal/config/watch_test.go#TestWatch_ValidRewritePublishes
        status: pass
      - kind: unit
        ref: internal/config/watch_test.go#TestWatch_InvalidRewriteKeepsLastGood
        status: pass
      - kind: unit
        ref: internal/config/watch_test.go#TestWatch_DebounceCoalesces
        status: pass
      - kind: unit
        ref: internal/config/watch_test.go#TestWatch_RemoveAndRenameTrigger
        status: pass
      - kind: unit
        ref: internal/config/watch_test.go#TestWatch_ContextCancelClosesSource
        status: pass
    human_judgment: false
  - id: D4
    description: "docs/CONFIG.md: full schema reference (all keys, defaults, ranges, examples) + Caramba correspondence table without cloning their names, with the model-difference caveat (D-31)"
    requirement: CONF-03
    verification:
      - kind: command
        ref: "grep Caramba/tap_window_ms/clipboard_rung docs/CONFIG.md → DOCS-OK"
        status: pass
    human_judgment: true
    rationale: "The correspondence table's adequacy as the CONF-03 (wish-level) implementation and the schema as a public contract are owner confirmations scheduled at the verify gate (the plan's FLAGGED assumptions name exactly these)."
  - id: D5
    description: "Daemon -config flag (SWCH-04/D-35): window from tap_window_ms at startup, no flag = built-in defaults equal to hotkey.DefaultWindow; broken/missing config = visible exit 1"
    requirement: SWCH-04
    verification:
      - kind: command
        ref: "grep -config/TapWindowMs/config.NewWatcher cmd/goswitchd/main.go → WIRED"
        status: pass
      - kind: command
        ref: "live run: tap_window_ms 450 config → INFO config loaded path+450, component registered, zero ERROR; typo'd key → exit 1 'field verify_wait_mss not found'"
        status: pass
      - kind: unit
        ref: internal/config/config_test.go#TestDefaults (DefaultWindow continuity pin)
        status: pass
    human_judgment: false

duration: 30 min
completed: 2026-09-15
status: complete
---

# Phase 3 Plan 2: Конфигурация — схема, strict-декод, hot reload Summary

**internal/config: полная YAML-схема демона со strict-декодом KnownFields (опечатка = невалидный конфиг целиком, D-33), watcher на fsnotify с directory-watch + debounce 200 мс + last-good (D-32), закрытые таблицы имён биндингов в чистом hotkey, флаг -config у демона со стартовым окном из YAML и docs/CONFIG.md с таблицей соответствия Caramba.**

## Performance

- **Duration:** 30 min
- **Started:** 2026-09-15T09:02:01Z
- **Completed:** 2026-09-15T09:31:36Z
- **Tasks:** 3 (2 tdd: RED→GREEN with RED_EVIDENCE_OK ×2, 1 auto)
- **Files modified:** 12 (9 code/docs + go.mod/go.sum + 2 red-evidence records)

## Accomplishments

- Schema + strict decode (CONF-01/D-31/D-33): `internal/config` with the four pinned sections and snake_case tags; `Validate()` names the field in every error (ranges `(0,2000]` windows, `[1,500]` cap, 64-app ceiling — T-03-02-02); `Load()` decodes with `dec.KnownFields(true)` (matrix.go:113-116 pattern verbatim) and refuses empty/missing documents — an explicit `-config` must yield a working config.
- Binding tables (D-33): `internal/hotkey/names.go` — `ParseBinding` over closed key/modifier tables, all keyvals cited from `/usr/include/ibus-1.0/ibuskeysyms.h:188-199` (verified this session, never from memory), masks duplicated into the pure package with the KeyvalShiftR precedent; the FSM corpus (11 tests) stayed green untouched (D-35).
- Hot reload (CONF-02/D-32): `config.NewWatcher` watches the config's **directory** (atomic-rename editors swap the inode), filters by base name, re-arms a 200 ms debounce AfterFunc (armTimer discipline), re-parses fully on expiry; valid → `atomic.Pointer[Config]` store, invalid → WARN `config reload rejected` + last-good serving + `LastError()` for the 03-06 status surface. `Snapshot() config.Config` is the value-copy producer contract 03-04 consumes verbatim.
- fsnotify v1.10.1 joined go.mod — the phase's single approved dependency (proxy.golang.org-verified, RESEARCH legitimacy audit Approved); `go mod tidy` clean.
- Daemon wiring (SWCH-04): `goswitchd -config PATH` loads at startup (refusal = stderr + exit 1, verified live with a typo'd key naming the field), the FSM window comes from `tap_window_ms`, no flag = built-in defaults numerically equal to `hotkey.DefaultWindow` (pinned by `TestDefaults`), INFO `config loaded` carries path+window only (T-03-02-04), the watcher rides run()'s signal ctx.
- docs/CONFIG.md (CONF-03): every key/default/range, binding-name table, complete examples, hot-reload semantics (incl. the stale-timer note), the Caramba correspondence table WITHOUT cloning names — with the explicit model-difference caveat (their first-tap switch vs our ADR-002 window expiry) — and the privacy note.

## Task Commits

Each task committed atomically (TDD: RED before GREEN):

1. **Task 1: Схема Config + strict-декод + таблица имён биндингов** — `a5f0ccb` (test, RED) + `36be082` (feat, GREEN)
2. **Task 2: Watcher — fsnotify dir-watch + debounce + last-good + atomic.Pointer** — `ff06c01` (test, RED) + `7695d5e` (feat, GREEN)
3. **Task 3: Флаг -config + стартовая проводка окна + docs/CONFIG.md** — `22f6d4a` (feat)

**Plan metadata:** this commit (docs: complete plan)

## Files Created/Modified

- `internal/config/config.go` (new) — schema types, Defaults(), per-section Validate() with sentinels
- `internal/config/load.go` (new) — strict Load: open/decode(KnownFields)/validate chain
- `internal/config/watch.go` (new) — Watcher: EventSource seam, options, dir-watch loop, debounce, reload/last-good
- `internal/hotkey/names.go` (new) — Binding, keyval/mask constants (provenance), closed tables as functions, ParseBinding
- `internal/config/{config,load,watch}_test.go`, `internal/hotkey/names_test.go` (new) — the corpora (TDD)
- `cmd/goswitchd/main.go` — `-config` flag, loadConfig/watchConfig split, window from config
- `docs/CONFIG.md` (new) — schema reference + Caramba table
- `go.mod` / `go.sum` — fsnotify v1.10.1

## TDD Gate Compliance

Both tdd tasks followed RED → GREEN with machine-validated red evidence:

| Task | RED commit | Evidence record | Verdict | GREEN commit |
|------|-----------|-----------------|---------|--------------|
| 1 | `a5f0ccb` test(03-02) | red-evidence/03-02-task1-red.json | RED_EVIDENCE_OK (11 new tests failing on assertions, 47 with subtests; 11 pre-existing FSM tests green) | `36be082` feat(03-02) |
| 2 | `ff06c01` test(03-02) | red-evidence/03-02-task2-red.json | RED_EVIDENCE_OK (5 watcher tests failing on assertions) | `7695d5e` feat(03-02) |

No REFACTOR commit — GREEN implementations landed lint-clean under the strict v2 config (all lint findings fixed inside the GREEN iterations per D-08).

## Decisions Made

- ModMask semantics (see Deviations 1): the full wire state of the bound key's event — the one rule consistent with both the IBus state word and the plan's `shift_r` literal.
- Complete-document schema (matrix.go strict precedent): no defaults overlay inside Load; Defaults() exists for the no-flag path only. A partial YAML is a validation error, which keeps "the config that loaded is the config in force" auditable.
- The watcher's WARN precedes the LastError store, so the log record can never lag the observable state (test-race ordering; the corpus polls both).
- gomodguard_v2 stays disabled with its existing comment — the plan asserted the disable rationale remains valid (the module's allow-list is not hardcoded in the lint config); no .golangci.yml change needed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Plan-literal ambiguity] Combo ModMask literal is unsatisfiable as written**
- **Found during:** Task 1 (RED authoring)
- **Issue:** The behavior line pins BOTH `shift_r → Keyval 0xffe2, ModMask Shift` AND `shift+ctrl_r → Keyval 0xffe4, ModMask Shift`. No uniform rule produces both: explicit-modifiers-only gives `shift_r → mask 0`; a key-family rule without the explicit bits loses `shift` for the combo.
- **Fix:** Resolved to the wire-truthful rule — ModMask = held modifiers OR the bound key's own family bit (a Shift_R press carries MaskShift in the IBus state word; held Shift + a pressed Control_R carries Shift|Control). Matches the `shift_r` literal exactly; the combo pins Shift|Control (a superset refinement enabling exact-match against the state word). The consumer contract (`ParseBinding`, `Binding{Keyval, ModMask}`) is untouched; 03-04 decides the matching predicate.
- **Files modified:** internal/hotkey/names.go, internal/hotkey/names_test.go
- **Verification:** TestParseBinding + TestParseBinding_Combos green under -race; mise run ci green
- **Committed in:** `a5f0ccb` / `36be082`

**2. [Rule 3 - Blocking] Strict-lint findings inside the GREEN iterations (D-08)**
- **Found during:** Task 1 and Task 2 GREEN (mise run ci)
- **Issue:** err113 (dynamic errors), mnd (default literals), goconst/funlen (test corpora), funcorder/wrapcheck/errcheck in watch.go.
- **Fix:** Package sentinels + %w wrapping (conn.go idiom); named default constants; corpus constants + test splits (TestValidate_Ranges → TestValidate_TimeoutRanges/TestValidate_CorrectionRanges — acceptance's named test split in two, both green); method reorder; adapter error wrapping. All inside the same iterations.
- **Verification:** mise run ci green after each fix (lint 0 issues)
- **Committed in:** `36be082`, `7695d5e`

**3. [Rule 3 - Blocking] RED stub shape for genuine assertion failures**
- **Found during:** Task 1 RED (first run)
- **Issue:** The initial Load stub returned `(nil, nil)` — the corpus nil-dereferenced (a fixture crash, INVALID_RED material) instead of failing on assertions.
- **Fix:** Stub returns `(&Config{}, nil)` so failures land on field assertions; recorded in the red-evidence actual field.
- **Files modified:** internal/config/load.go (stub form)
- **Verification:** RED_EVIDENCE_OK (11 failing on assertions)
- **Committed in:** `a5f0ccb`

---

**Total deviations:** 3 auto-fixed (1 plan-literal resolution, 2 blocking lint/tooling)
**Impact on plan:** All resolved toward the locked must-haves; no scope creep. The schema-as-contract and the Caramba table route to the owner at the verify gate (coverage D4), exactly as the plan's FLAGGED assumptions schedule.

## Issues Encountered

None beyond the deviations above. Live gates on the target desktop: daemon with a 450 ms window config starts, logs `config loaded` (path+window), registers both engines, zero ERROR records; a typo'd key refuses startup with `exit 1` naming the field.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `Watcher.Snapshot() config.Config` is the pinned seam for 03-04 (window on new series, Backspace cap, clipboard flag, combo binding consumption; stale-timer unit pin lives there too, per the plan's FLAGGED assumption).
- `LastError()` is ready for the `goswitchctl status` surface (03-06).
- matrix v2 config-reload cases (03-07) can drive the real watcher through `-config` on temp files.
- The D-35 guard held: fsm_test.go untouched across the whole plan (11/11 green in every gate).

## Self-Check: PASSED

All key-files exist on disk; all five task commits found in history; plan ledger measured 5 commits from plan_head_before 57a3a67 (matches `commits:` in frontmatter); `mise run ci` green at the final tree.
