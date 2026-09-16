# Phase 04 — UI Review

**Audited:** 2026-09-16
**Baseline:** abstract 6-pillar standards (no UI-SPEC.md exists)
**Screenshots:** not captured — no dev server (headless Go daemon project; ports 3000/5173 closed, 8080 is an unrelated local HTTP proxy). Code- and docs-only audit.

---

## Scope Adaptation (mandatory reading for this review)

goswitch is a headless IBus daemon + CLI + e2e harness. There is no frontend, no web UI, and no UI-SPEC.md. The user-visible surfaces created in Phase 04 are:

1. `README.md` (EN, rewritten) and `README.ru.md` (RU, created) — the bilingual doc pair (D-50);
2. `docs/ACCEPTANCE.md` — the owner's UAT checklist (D-49);
3. `goswitchctl` console output: `install` / `uninstall [--purge]` report lines (D-39/D-42) and `selfcheck` verdict lines (D-41), plus the pre-existing `status` / `reload` / `correct` one-line canon.

Pillars are graded against what is actually user-visible: docs readability/structure and CLI output clarity/consistency. Pillar 3 (Color) has no surface at all and is recorded N/A with evidence. Nothing was invented to fill a web-shaped template.

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 2/4 | Corrupted example `ГHBDTN` (Cyrillic + Latin) in the UAT checklist; untranslated `Troubleshooting` heading breaks the D-50 mirror; uninstall fallback verdict misnames its cause |
| 2. Visuals | 3/4 | Docs hierarchy strong; ACCEPTANCE item 3 packs six gesture checks into one run-on paragraph — the least scannable line of the UAT gate |
| 3. Color | N/A | No color surface exists (zero ANSI escapes in output; plain markdown); textual-emphasis analog audited and honest |
| 4. Typography | 3/4 | Heading levels, fences and the one-line canon consistent; install report reuses keys (`unit:`, `engine:`) with heterogeneous value shapes |
| 5. Spacing | 3/4 | README hanging indents disciplined; ACCEPTANCE `expected:/result:/note:` fields wrap flush-left with no hanging indent |
| 6. Experience Design | 2/4 | `install` is silent for its entire (up to 120 s) runtime — inconsistent with selfcheck's streaming; `--purge` destroys data with no pre-echo; error/empty states otherwise excellent |

**Overall: 13/20** (5 pillars scored; Color not applicable — no normalization performed)

---

## Top 3 Priority Fixes

1. **`docs/ACCEPTANCE.md:50` — the "three cases" example is corrupted: `ГHBDTN` (Cyrillic Г + Latin `HBDTN`) where the uppercase-Latin example `GHBDTN` belongs** — the owner executes this checklist literally as the release UAT gate; the mixed-script string is not a valid "uppercase EN layout" case, so the intended lower-EN / UPPER-EN / lower-RU triple cannot be verified as written — fix the byte (`ГHBDTN` → `GHBDTN`).
2. **`goswitchctl install` prints nothing until the whole lifecycle finishes** (`cmd/goswitchctl/main.go:120-131` buffers `install.Install(ctx)`'s `[]string`; `internal/install/install.go:273-310` runs up to two 20 s bounded waits plus an `ibus restart` inside a 120 s budget) — a silent terminal for 40+ seconds invites Ctrl-C mid-lifecycle, which can strand a half-installed desktop with sources already taken over; emit each report line as its step completes, matching the streaming discipline `selfcheck` already has (`internal/install/selfcheck.go:65-91`).
3. **`internal/install/install.go:626-631` — the uninstall fallback verdict reads "saved state unreadable or malformed — safe default applied" even when the state file is simply absent** (never installed / already uninstalled) — the visible line misleads the user about what just happened to their desktop sources during a destructive operation (it is the user-facing face of CR-01 in 04-REVIEW.md); distinguish the three cases in the copy ("no install state — sources untouched" vs "malformed — fallback applied").

---

## Detailed Findings

### Pillar 1: Copywriting (2/4)

The D-41 red-verdict contract is genuinely excellent — every static error in `internal/install/selfcheck.go:29-45` names the failing step AND the fix (`errUnitInactive` → "fix: systemctl --user enable --now goswitchd", etc.), and the friendly no-daemon verdict (`cmd/goswitchctl/main.go:49`) is a model of terminal copy. No generic labels anywhere ("Submit"/"OK"/"Cancel" grep: zero hits). The honest-`dev` version fallback wording in both READMEs matches observed `go install` behavior (verified in 04-07-SUMMARY STAMPED/PROXY evidence). But the phase's core deliverable was precisely these documents, and two of the three doc surfaces plus one CLI line carry real copy defects:

- **`docs/ACCEPTANCE.md:50` — mixed-script typo in the UAT gate.** Byte-verified: `M-PM-^SHBDTN` = U+0413 (Cyrillic Г) followed by Latin `HBDTN`. Context: "(и так же с `ГHBDTN`/`сщьввте` — три регистра)" — the intended triple is lower-EN `ghbdtn` / UPPER-EN `GHBDTN` / lower-RU `сщьввте`. An IME slip (ironic, given the product) corrupted the middle case; the checklist instruction is not executable as written. (Warning)
- **`README.ru.md:157` — `## Troubleshooting` left untranslated.** All nine sibling H2 headings are Russian (Установка, Проверка установки, …, Лицензия); this one heading breaks the D-50 mirror. The phase's own sync greps (04-07-SUMMARY D1: "counter `goswitchctl install` en=4 ru=4 — pass", verified: indeed 4/4, fences 10/10) only checked command mentions and never headings — the miss slipped through. (Warning)
- **`internal/install/install.go:626-631` — uninstall fallback verdict misattributes its cause.** `restoreSources` emits "sources: fallback [('xkb', 'us')] (saved state unreadable or malformed — safe default applied)" whenever `savedSources` returns untrusted — including the state-file-absent case (never installed / second uninstall). "Unreadable or malformed" is false when the file does not exist; the user reads a wrong diagnosis during a sources-rewriting operation. See also CR-01 in 04-REVIEW.md for the underlying behavior. (Warning)

Positive: empty-state copy is right — a missing config file is a green `config (defaults: no config file)` (`selfcheck.go:163`), worded exactly as the README promises.

### Pillar 2: Visuals (3/4)

Visual hierarchy analogs for docs + terminal:

- README (both languages) has a clear focal point — the `ghbdtn` → `привет` transformation in the lede — followed by scannable H2 sections, two well-formed tables (gestures, performance), fenced `sh` blocks, and a blockquote warning for `-debug` that stands out appropriately. Emphasis (bold) is load-bearing, never decorative.
- CLI output is strictly scannable: one verdict per line, consistent prefixes (`ok NAME` / `FAIL NAME: <fix>` / `key: value`), no banners or ASCII art.
- **`docs/ACCEPTANCE.md:48-55` (item 3, "Живой набор жестов") — the single most important checklist item is the least scannable.** Six distinct gesture checks (single tap, double tap ×3 cases, triple tap, selection, chord, mixed word) are packed into one semicolon-chained `expected:` paragraph. The owner reads this mid-session, on a live desktop, with hands on the keyboard. Sub-bullets (one per gesture) would make the gate executable at a glance. (Warning)
- Minor: `expected:/result:/note:` fields blur into prose because they carry no visual field separation (see Pillar 5).

### Pillar 3: Color — N/A

No color surface exists: grep for ANSI escapes (`\033`, `\x1b`) and color helpers across `cmd/` and `internal/` user-facing output found zero uses; docs are plain markdown. The nearest analog — textual emphasis — was audited: bold is used sparingly and honestly, including bolding "**exceeds**" in the perf budget-status paragraph (`README.md:136-139`, `README.ru.md:138-141`), which is the correct, owner-sanctioned (decision «в») disclosure rather than spin. Verdict words ("red"/"green" in docs) map to textual `ok`/`FAIL` prefixes, which survive pipes, ssh and screen readers. Pillar not scored; nothing to fix.

### Pillar 4: Typography (3/4)

Analogs: heading discipline, code fencing, output canon.

- All three documents use strict H1 → H2 → H3 without skipped levels; both READMEs are structurally identical (heading map 1:1 apart from the Pillar-1 translation miss; code fences 10/10 each, all tagged `sh`).
- The CLI's one-line canon is enforced and documented: `status` prints the daemon's verbatim key=value line, `--json` types tokens best-effort (numbers/booleans/string), and the `config_error=` last-field grammar (spaces allowed) is implemented and tested (`cmd/goswitchctl/main.go:246-279`).
- Nit: the install report reuses keys with heterogeneous value shapes — `unit:` appears as both "`unit: <path> content verified (read-back)`" and "`unit: daemon-reload + enable --now done`"; same for `engine:` (a registration claim, then an activation claim) (`internal/install/install.go:586-598`). Anyone scripting or skimming the report cannot rely on a key's shape. (Minor)

### Pillar 5: Spacing (3/4)

Analogs: whitespace and indent discipline in docs and output.

- Both READMEs use consistent hanging indents (3 spaces under ordered items, 2 under bullets) across every wrapped list entry — verified section by section.
- **`docs/ACCEPTANCE.md` — `expected:` field continuations wrap flush-left with no hanging indent** (e.g. lines 27-31, 38-43, 48-55), so a multi-line expectation visually merges with the following `result:`/`note:` fields. Contrast with the README pair's careful indents; a 3-space hanging indent would keep each field a visual unit. (Minor)
- CLI output spacing is right: one line per verdict, no blank-line noise, no alignment padding that would break the key=value canon.

### Pillar 6: Experience Design (2/4)

Error and empty states are the strongest part of this phase's UX; loading and destructive-action states are the weakest:

- **Loading states: FAIL for `install`.** The whole lifecycle (up to 120 s; internally: an `ibus restart` plus two 20 s bounded waits, `install.go:38-42, 293-301`) produces zero output until completion — `Install()` returns the report as a `[]string` and the CLI prints it afterwards (`main.go:120-131`). `selfcheck`, by contrast, streams per-step verdicts to stdout (`selfcheck.go:77-88`). The inconsistency is internal evidence that streaming was available; the doc comment on `runInstallCmd` ("prints the step-by-step report") overstates what the user experiences. During a silent 40+ second window the user cannot distinguish a healthy install from a wedged `systemctl`/`ibus`, and Ctrl-C mid-takeover strands a half-installed desktop. (Warning — priority fix #2)
- **Confirmation for destructive actions: absent.** `uninstall --purge` removes `~/.config/goswitch/` (`install.go:654-663`) with no pre-echo of what is about to be deleted and no prompt. Explicit-flag-as-consent is a defensible CLI convention, but even a one-line "removing <dir> (user config)" before `RemoveAll` would make the destruction visible in the transcript. (Warning)
- Error states: excellent. Every failure names step + fix (Pillar 1); fail-fast selfcheck exits non-zero on the first red; multi-line YAML errors are whitespace-flattened to keep the one-verdict-per-line contract (`selfcheck.go:79-85`); every external call is deadline-bounded (5 s control / 60 s selfcheck / 120 s lifecycle / 10 s per subprocess) so nothing hangs the terminal.
- Empty/no-op states: well handled — missing config is green with an explanatory value; uninstall tolerates an already-removed unit (`unitGone`, though see WR-01 in 04-REVIEW.md for the error-swallowing around it); malformed restore state falls back visibly, never silently.
- Note (accepted trade-off, no action required): selfcheck's fail-fast means a user with multiple problems discovers them one run at a time — deliberate per D-41's e2e-preflight discipline, and each red is actionable, so this is recorded, not penalized beyond the streaming finding above.

---

## Registry Safety

Skipped — `components.json` does not exist (no shadcn, no third-party UI registries; the project's dependency policy is stdlib + godbus + minimal, per AGENTS.md).

---

## Files Audited

- `/home/nil/DiskD/W/Djarvur/goswitch/README.md`
- `/home/nil/DiskD/W/Djarvur/goswitch/README.ru.md`
- `/home/nil/DiskD/W/Djarvur/goswitch/docs/ACCEPTANCE.md`
- `/home/nil/DiskD/W/Djarvur/goswitch/cmd/goswitchctl/main.go`
- `/home/nil/DiskD/W/Djarvur/goswitch/internal/install/install.go`
- `/home/nil/DiskD/W/Djarvur/goswitch/internal/install/selfcheck.go`
- `/home/nil/DiskD/W/Djarvur/goswitch/cmd/goswitchd/version.go`
- `/home/nil/DiskD/W/Djarvur/goswitch/perf-report.txt` (cross-check of README perf table — numbers match: p50 166.9 / p95 184.4 / p99 186.5 ms, VmHWM 12368 kB ≈ 12.4 MB)
- `.planning/phases/04-postavka-i-priemka/04-CONTEXT.md`, `04-01-SUMMARY.md`, `04-07-SUMMARY.md`, `04-REVIEW.md` (context and cross-reference)

---

*Phase: 04-Поставка и приёмка*
*Reviewed: 2026-09-16*
