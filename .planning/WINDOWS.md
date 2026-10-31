---
schema_version: 1
open_count: 12
waived_count: 4
fixed_count: 7
total_count: 23
last_updated: 2026-10-31T21:18:56.408Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 02 | deviation | test/e2e/case_ladder.go |  | Plan oracle level:2 falsified live: google-chrome 153 reports caps 0x29 and applies DeleteSurroundingText — actual level is 1 (ibus#2354 obsolete on target); case pins actual level 1 + verify-after match; no live level-2 witness exists on this desktop (all surfaces report the bit), unit corpus carries level 2 | fixed |  | 2026-09-14T22:27:01.392Z | 2026-09-15T05:19:04.926Z |
| 2 | 3 | deviation | internal/correct/runs.go | 57 | D-24 nothing-to-convert example unsatisfiable under уточнение D-22 (homogeneous converts wholesale per matrix v1); changed=false surface implemented, reserved for selection path 03-03 — owner confirms at verify gate | fixed |  | 2026-09-15T08:57:29.492Z | 2026-09-16T14:53:15.673Z |
| 3 | 3 | deviation | internal/correct/runs.go | 57 | test | waived | junk entry appended by executor output-shape probe — not a real defect, removed intent | 2026-09-15T08:57:41.296Z | 2026-09-15T08:58:03.115Z |
| 4 | 3 | unmet-truth | test/e2e/case_macr.go | 595 | GTK4-Wayland (GNOME 46) does not apply IBus-forwarded key events: the relay reaches the focused InputContext (dbus-monitor pinned) but the widget never acts — MACR's Ctrl+letter remap, the ADR-003 level-2 Backspace replay and the D-28 Ctrl+V burst all stop at the client boundary (03-05 Deviation 3) | waived | owner accepted GTK4-Wayland forwarded-events platform limitation (UAT 2026-09-15, REQUIREMENTS MACR-01 caveat) | 2026-09-15T13:23:14.477Z | 2026-09-16T14:53:16.148Z |
| 5 | 04 | unmet-truth | test/e2e/perf.go | 1 | INST-03 latency budget verdict is FAIL under the prescribed ydotool→AT-SPI window (p95 190-205ms >= 50ms over two full runs): the window is dominated by ydotool 0.1.8 per-event injection overhead (~80-125ms/event), daemon reaction is 0.2-1.8ms; owner methodology decision required before the D-46 README table publishes latency numbers | waived | owner accepted the ydotool-injector-dominated stand numbers as-is (decision (в), 2026-09-16): publish v1.0.0 with the real ydotool→AT-SPI p50/p95/p99 and the budget-fail status stated honestly in the README perf table; daemon reaction is sub-millisecond, the window overhead is the stand's ydotool 0.1.8 injector | 2026-09-16T13:51:53.747Z | 2026-09-16T17:36:37.077Z |
| 6 | 04 | deviation | test/e2e/main.go | 201 | 04-04 (1b443eb) turned the matrix path's zero watchdog limit into a literal time.After(0): every matrix case (v1/v2/v3) failed instantly with 'did not complete within 0s' — live-caught by the first D-48 double-run dispatch (run 35107406144); fixed structurally by watchdogLimit resolving 0 to caseTimeout inside runCaseWatchdog (9f2fd75, RED->GREEN tests) | fixed |  | 2026-09-16T14:35:30.840Z | 2026-09-16T14:35:35.768Z |
| 7 | 04 | deviation | test/e2e/focus_helper.py |  | 04-08 live smoke: INPUT-hit-through-focused-frame not drivable from background context (mutter denies activation; FLAGGED focus-lag case) — witness answered frame-fallback arm in budget; in-situ proof rides the owner's D-48 dispatch | waived | Владелец закрыл как принятую особенность стенда (2026-10-08): запись нужна только автотестам (вывод тестового окна на передний план из фона mutter запрещает), продукт не затрагивает; прогоны сейчас ручные — при возврате автоматических ночных прогонов завести заново | 2026-09-17T11:53:10.594Z | 2026-10-08T10:49:36.953Z |
| 8 | quick-260927-sy8 | unrun-verify | .planning/quick/260927-sy8-engine-self-reactivation-re-activate-the |  | Task 3 live desktop restart-and-type human-check deferred to orchestrator (automated grep gates + README edits done) | fixed |  | 2026-09-27T18:28:09.875Z | 2026-09-27T18:57:43.384Z |
| 9 | 05 | unrun-verify | mise.toml | 153 | plan 05-02: mise run e2e-install-cycle full-case PASS deferred — live run proved the new two-source assertions (install wrapper machine-check, selfcheck ok input-source, verbatim restore) but the zenity correction step was focus-refused with the owner active; rerun consolidates in plan 05-05 | open |  | 2026-09-29T20:37:58.257Z |  |
| 10 | 05 | unmet-truth | docs/adr/ADR-006-two-engine-revision.md |  | plan 05-01: P3 XKB-truth unresolved — spike entry surface focus-refused both runs (witness gnome-shell:WINDOW:chars=-1); re-opens at 05-05 live proofs (ADR-006 unresolved row) | fixed |  | 2026-09-29T21:28:22.157Z | 2026-09-30T00:22:44.944Z |
| 11 | 05 | unmet-truth | docs/adr/ADR-006-two-engine-revision.md |  | plan 05-01: indicator verdict unknown (owner 'не разглядел') — routes to UAT/verify-work; criterion-2 semantics fixed by spec-delta regardless | fixed |  | 2026-09-29T21:28:22.438Z | 2026-10-08T10:39:18.770Z |
| 12 | 05 | unmet-truth | test/e2e/cases/matrix-v3.yaml |  | plan 05-05: word-mixed + phrase-mixed rows expect a correction ACROSS a flip boundary — under ADR-006 the two-source flip churns FocusOut/FocusIn and the designed context reset (05-CONTEXT: «переключение раскладки сбрасывает контекст») drops the pre-flip buffer; the daemon then corrects only the post-flip word (word-mixed: buffer привет → wholesale ru→en → gfbghbdtn; phrase-mixed → ghbdtn ghbdtn). Rows frozen (v3, CASES-FROZEN); the owner owns the verdict at the phase verify gate: either the rows re-shape to the reset semantics (new plan) or the reset semantics get revisited (architectural) | fixed |  | 2026-09-30T00:21:32.095Z | 2026-10-08T10:39:18.940Z |
| 13 | 06 | unrun-verify | test/e2e/cases/matrix-v4.yaml |  | plan 06-07: full mise run e2e-matrix-v4 green NOT achieved on this session — first-key-after-mode-flip session drift (each row typing right after a SetGlobalEngine flip loses exactly one leading char/mode record); control experiment: pre-phase-6 tree d96dcb2 fails BYTE-IDENTICALLY (/tmp/matrix-v3-oldtree.log, worktree /tmp/goswitch-0505) — environmental, phase-6 diff exonerated; autocorrect rows (fires/off/oov) green x2 and all three live cases green; formal fresh-session D-48 double-run rides plan 06-08 (checkpoint accept & defer 2026-10-02). 06-08 verdict (2026-10-02): drift CONFIRMED REAL and STICKY — the formal back-to-back double-run on the relogged-in session (ages 2h03m/2h08m) went 23/35 and 24/35 with the SAME row set (9 drift rows byte-similar to 06-07, +1 ibus-activation transient in run 1); autocorrect rows and word/phrase manual-contract carriers green x2 in the same runs (3rd independent exoneration of the phase diff); NEW diagnostic: the drift SURVIVES 'ibus restart' (24/35, identical set — /tmp/06-08-matrix-post-ibus-restart.log) — not the documented-cure class, lives in the login state; the formal <=30-min fresh_session=true gate (docs/ci-runner.md) remains the owner's machine-checked instrument at verify-work; ADR-007 Acceptance Evidence row 5 carries the full record | open |  | 2026-10-02T18:01:54.918Z |  |
| 14 | 10 | deviation | internal/session/actor_test.go | 5557 | Silence-matrix role-60 cell migrated to the arming expectation (terminal-veto, sync) — the D-55 semantic move the plan mandates; the confirm role switch stays as the second belt | open |  | 2026-10-12T15:31:24.063Z |  |
| 15 | 10 | deviation | internal/session/actor_test.go | 6945 | Plan 10-04 (OQ 1 unified): TestTerminalClass_VerifyDeadLearnedOnVerifyTimeout second half migrated — a LEARNED verify-dead context routes the next Double through the terminal write branch, so a second verify expiry is unreachable; marker discipline (no transition, no marker) preserved | open |  | 2026-10-12T18:02:30.083Z |  |
| 16 | 10 | deviation | internal/session/actor.go | 2970 | Plan 10-04 Task 2 acceptance partial: the D-24 instant-done changed=false INFO is unreachable by behavior test — ConvertRuns provably never returns changed=false (runs.go doctrine, WINDOWS-2 owner verdict), the branch stays defensive; the discriminator is planted in the same form as logCorrectionDone, changed=true is behaviorally pinned | open |  | 2026-10-12T18:02:30.777Z |  |
| 17 | 10 | deviation | test/e2e/case_autocorrect.go |  | terminal-silent re-pin gates moved to daemon focus_in (AT-SPI witness pinned to gnome-shell globally; 10-05 hand-over prescription) + noSurfacePreflight — closed the 10-05 blocked row | open |  | 2026-10-12T21:28:26.352Z |  |
| 18 | 10 | unrun-verify | .planning/debug/DEBUG-term-host-live-failures.md | 332 | Task 2 automated verify 'approved\|rejected' matches instruction/pending text only - owner re-acceptance verdict is PENDING (never synthesized); routing: UAT end-of-phase; on approved append gap-closed row, on rejected STOP without code changes | open |  | 2026-10-13T21:53:18.173Z |  |
| 19 | 10 | unrun-verify | .planning/debug/DEBUG-term-host-live-failures.md |  | 10-12 D4: вердикт re-acceptance round-2 (шесть пунктов, закрывающий оракул G-10-1) — ОЖИДАНИЕ ответа владельца; grep-гейт матчится с текстом инструкции, не с вердиктом | open |  | 2026-10-14T10:03:44.986Z |  |
| 20 | 10 | unrun-verify | .planning/debug/DEBUG-term-host-live-failures.md | 556 | Вердикт владельца round-4 (гейт: слово в wezterm ОСТАЁТСЯ на v1.1.1-180-gefac750) — human-check честно ОЖИДАЕТ; UAT Current Test №5 pending; не синтезируется | open |  | 2026-10-16T21:57:28.487Z |  |
| 21 | 9 | unrun-verify | .planning/phases/09-spayk-i-defekt-wezterm/09-02-SUMMARY.md | 72 | 09-02 (D3–D5): живые раунды R2–R6, полная лестница P1–P5 и two-round wezterm-flip RED-контракт не исполнены — владелец прекратил участие в живых тестах 2026-10-12 (T2-ядро: фоновые окна клавиатурно мертвы); ряды не синтезированы; capability-потребитель Фаз 10/12 обслужен стенд-рядами Фазы 10 (adr-008-spike-log.md) | open |  | 2026-10-20T14:00:23.408Z |  |
| 22 | 11 | unrun-verify | .planning/debug/orca-terminal-second-shift.md |  | Round 2 journal-verify: живой ряд «следующая коррекция с одного нажатия + lifecycle aftermath=true» не снят — владелец вышел из сессии 23:21; демон активен, ряд завершится на следующей сессии; by-ear зачёт — UAT end-of-phase (G-10-1) | open |  | 2026-10-31T20:39:21.173Z |  |
| 23 | 11 | unrun-verify | mise.toml |  | 11-03: literal 'mise run ci' unrun in the executor session (toolchain seam, 11-01 Rule-3 precedent): GOROOT(mise go1.23.12) vs go binary(system 1.27.3) mismatch in task shells; the SAME gates ran green directly — go build/vet/test -race ./... (22 packages) + golangci-lint 2.15.0 (mise install, GOROOT unset) 0 issues + go mod tidy git-diff clean, STACK.md-sanctioned | open |  | 2026-10-31T21:18:56.408Z |  |

````json
[
  {
    "id": 1,
    "kind": "deviation",
    "phase": "02",
    "file": "test/e2e/case_ladder.go",
    "line": null,
    "description": "Plan oracle level:2 falsified live: google-chrome 153 reports caps 0x29 and applies DeleteSurroundingText — actual level is 1 (ibus#2354 obsolete on target); case pins actual level 1 + verify-after match; no live level-2 witness exists on this desktop (all surfaces report the bit), unit corpus carries level 2",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-14T22:27:01.392Z",
    "resolved_at": "2026-09-15T05:19:04.926Z"
  },
  {
    "id": 2,
    "kind": "deviation",
    "phase": "3",
    "file": "internal/correct/runs.go",
    "line": 57,
    "description": "D-24 nothing-to-convert example unsatisfiable under уточнение D-22 (homogeneous converts wholesale per matrix v1); changed=false surface implemented, reserved for selection path 03-03 — owner confirms at verify gate",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-15T08:57:29.492Z",
    "resolved_at": "2026-09-16T14:53:15.673Z"
  },
  {
    "id": 3,
    "kind": "deviation",
    "phase": "3",
    "file": "internal/correct/runs.go",
    "line": 57,
    "description": "test",
    "status": "waived",
    "reason": "junk entry appended by executor output-shape probe — not a real defect, removed intent",
    "recorded_at": "2026-09-15T08:57:41.296Z",
    "resolved_at": "2026-09-15T08:58:03.115Z"
  },
  {
    "id": 4,
    "kind": "unmet-truth",
    "phase": "3",
    "file": "test/e2e/case_macr.go",
    "line": 595,
    "description": "GTK4-Wayland (GNOME 46) does not apply IBus-forwarded key events: the relay reaches the focused InputContext (dbus-monitor pinned) but the widget never acts — MACR's Ctrl+letter remap, the ADR-003 level-2 Backspace replay and the D-28 Ctrl+V burst all stop at the client boundary (03-05 Deviation 3)",
    "status": "waived",
    "reason": "owner accepted GTK4-Wayland forwarded-events platform limitation (UAT 2026-09-15, REQUIREMENTS MACR-01 caveat)",
    "recorded_at": "2026-09-15T13:23:14.477Z",
    "resolved_at": "2026-09-16T14:53:16.148Z"
  },
  {
    "id": 5,
    "kind": "unmet-truth",
    "phase": "04",
    "file": "test/e2e/perf.go",
    "line": 1,
    "description": "INST-03 latency budget verdict is FAIL under the prescribed ydotool→AT-SPI window (p95 190-205ms >= 50ms over two full runs): the window is dominated by ydotool 0.1.8 per-event injection overhead (~80-125ms/event), daemon reaction is 0.2-1.8ms; owner methodology decision required before the D-46 README table publishes latency numbers",
    "status": "waived",
    "reason": "owner accepted the ydotool-injector-dominated stand numbers as-is (decision (в), 2026-09-16): publish v1.0.0 with the real ydotool→AT-SPI p50/p95/p99 and the budget-fail status stated honestly in the README perf table; daemon reaction is sub-millisecond, the window overhead is the stand's ydotool 0.1.8 injector",
    "recorded_at": "2026-09-16T13:51:53.747Z",
    "resolved_at": "2026-09-16T17:36:37.077Z"
  },
  {
    "id": 6,
    "kind": "deviation",
    "phase": "04",
    "file": "test/e2e/main.go",
    "line": 201,
    "description": "04-04 (1b443eb) turned the matrix path's zero watchdog limit into a literal time.After(0): every matrix case (v1/v2/v3) failed instantly with 'did not complete within 0s' — live-caught by the first D-48 double-run dispatch (run 35107406144); fixed structurally by watchdogLimit resolving 0 to caseTimeout inside runCaseWatchdog (9f2fd75, RED->GREEN tests)",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-16T14:35:30.840Z",
    "resolved_at": "2026-09-16T14:35:35.768Z"
  },
  {
    "id": 7,
    "kind": "deviation",
    "phase": "04",
    "file": "test/e2e/focus_helper.py",
    "line": null,
    "description": "04-08 live smoke: INPUT-hit-through-focused-frame not drivable from background context (mutter denies activation; FLAGGED focus-lag case) — witness answered frame-fallback arm in budget; in-situ proof rides the owner's D-48 dispatch",
    "status": "waived",
    "reason": "Владелец закрыл как принятую особенность стенда (2026-10-08): запись нужна только автотестам (вывод тестового окна на передний план из фона mutter запрещает), продукт не затрагивает; прогоны сейчас ручные — при возврате автоматических ночных прогонов завести заново",
    "recorded_at": "2026-09-17T11:53:10.594Z",
    "resolved_at": "2026-10-08T10:49:36.953Z"
  },
  {
    "id": 8,
    "kind": "unrun-verify",
    "phase": "quick-260927-sy8",
    "file": ".planning/quick/260927-sy8-engine-self-reactivation-re-activate-the",
    "line": null,
    "description": "Task 3 live desktop restart-and-type human-check deferred to orchestrator (automated grep gates + README edits done)",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-27T18:28:09.875Z",
    "resolved_at": "2026-09-27T18:57:43.384Z"
  },
  {
    "id": 9,
    "kind": "unrun-verify",
    "phase": "05",
    "file": "mise.toml",
    "line": 153,
    "description": "plan 05-02: mise run e2e-install-cycle full-case PASS deferred — live run proved the new two-source assertions (install wrapper machine-check, selfcheck ok input-source, verbatim restore) but the zenity correction step was focus-refused with the owner active; rerun consolidates in plan 05-05",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-09-29T20:37:58.257Z",
    "resolved_at": null
  },
  {
    "id": 10,
    "kind": "unmet-truth",
    "phase": "05",
    "file": "docs/adr/ADR-006-two-engine-revision.md",
    "line": null,
    "description": "plan 05-01: P3 XKB-truth unresolved — spike entry surface focus-refused both runs (witness gnome-shell:WINDOW:chars=-1); re-opens at 05-05 live proofs (ADR-006 unresolved row)",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-29T21:28:22.157Z",
    "resolved_at": "2026-09-30T00:22:44.944Z"
  },
  {
    "id": 11,
    "kind": "unmet-truth",
    "phase": "05",
    "file": "docs/adr/ADR-006-two-engine-revision.md",
    "line": null,
    "description": "plan 05-01: indicator verdict unknown (owner 'не разглядел') — routes to UAT/verify-work; criterion-2 semantics fixed by spec-delta regardless",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-29T21:28:22.438Z",
    "resolved_at": "2026-10-08T10:39:18.770Z"
  },
  {
    "id": 12,
    "kind": "unmet-truth",
    "phase": "05",
    "file": "test/e2e/cases/matrix-v3.yaml",
    "line": null,
    "description": "plan 05-05: word-mixed + phrase-mixed rows expect a correction ACROSS a flip boundary — under ADR-006 the two-source flip churns FocusOut/FocusIn and the designed context reset (05-CONTEXT: «переключение раскладки сбрасывает контекст») drops the pre-flip buffer; the daemon then corrects only the post-flip word (word-mixed: buffer привет → wholesale ru→en → gfbghbdtn; phrase-mixed → ghbdtn ghbdtn). Rows frozen (v3, CASES-FROZEN); the owner owns the verdict at the phase verify gate: either the rows re-shape to the reset semantics (new plan) or the reset semantics get revisited (architectural)",
    "status": "fixed",
    "reason": "",
    "recorded_at": "2026-09-30T00:21:32.095Z",
    "resolved_at": "2026-10-08T10:39:18.940Z"
  },
  {
    "id": 13,
    "kind": "unrun-verify",
    "phase": "06",
    "file": "test/e2e/cases/matrix-v4.yaml",
    "line": null,
    "description": "plan 06-07: full mise run e2e-matrix-v4 green NOT achieved on this session — first-key-after-mode-flip session drift (each row typing right after a SetGlobalEngine flip loses exactly one leading char/mode record); control experiment: pre-phase-6 tree d96dcb2 fails BYTE-IDENTICALLY (/tmp/matrix-v3-oldtree.log, worktree /tmp/goswitch-0505) — environmental, phase-6 diff exonerated; autocorrect rows (fires/off/oov) green x2 and all three live cases green; formal fresh-session D-48 double-run rides plan 06-08 (checkpoint accept & defer 2026-10-02). 06-08 verdict (2026-10-02): drift CONFIRMED REAL and STICKY — the formal back-to-back double-run on the relogged-in session (ages 2h03m/2h08m) went 23/35 and 24/35 with the SAME row set (9 drift rows byte-similar to 06-07, +1 ibus-activation transient in run 1); autocorrect rows and word/phrase manual-contract carriers green x2 in the same runs (3rd independent exoneration of the phase diff); NEW diagnostic: the drift SURVIVES 'ibus restart' (24/35, identical set — /tmp/06-08-matrix-post-ibus-restart.log) — not the documented-cure class, lives in the login state; the formal <=30-min fresh_session=true gate (docs/ci-runner.md) remains the owner's machine-checked instrument at verify-work; ADR-007 Acceptance Evidence row 5 carries the full record",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-02T18:01:54.918Z",
    "resolved_at": null
  },
  {
    "id": 14,
    "kind": "deviation",
    "phase": "10",
    "file": "internal/session/actor_test.go",
    "line": 5557,
    "description": "Silence-matrix role-60 cell migrated to the arming expectation (terminal-veto, sync) — the D-55 semantic move the plan mandates; the confirm role switch stays as the second belt",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-12T15:31:24.063Z",
    "resolved_at": null
  },
  {
    "id": 15,
    "kind": "deviation",
    "phase": "10",
    "file": "internal/session/actor_test.go",
    "line": 6945,
    "description": "Plan 10-04 (OQ 1 unified): TestTerminalClass_VerifyDeadLearnedOnVerifyTimeout second half migrated — a LEARNED verify-dead context routes the next Double through the terminal write branch, so a second verify expiry is unreachable; marker discipline (no transition, no marker) preserved",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-12T18:02:30.083Z",
    "resolved_at": null
  },
  {
    "id": 16,
    "kind": "deviation",
    "phase": "10",
    "file": "internal/session/actor.go",
    "line": 2970,
    "description": "Plan 10-04 Task 2 acceptance partial: the D-24 instant-done changed=false INFO is unreachable by behavior test — ConvertRuns provably never returns changed=false (runs.go doctrine, WINDOWS-2 owner verdict), the branch stays defensive; the discriminator is planted in the same form as logCorrectionDone, changed=true is behaviorally pinned",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-12T18:02:30.777Z",
    "resolved_at": null
  },
  {
    "id": 17,
    "kind": "deviation",
    "phase": "10",
    "file": "test/e2e/case_autocorrect.go",
    "line": null,
    "description": "terminal-silent re-pin gates moved to daemon focus_in (AT-SPI witness pinned to gnome-shell globally; 10-05 hand-over prescription) + noSurfacePreflight — closed the 10-05 blocked row",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-12T21:28:26.352Z",
    "resolved_at": null
  },
  {
    "id": 18,
    "kind": "unrun-verify",
    "phase": "10",
    "file": ".planning/debug/DEBUG-term-host-live-failures.md",
    "line": 332,
    "description": "Task 2 automated verify 'approved|rejected' matches instruction/pending text only - owner re-acceptance verdict is PENDING (never synthesized); routing: UAT end-of-phase; on approved append gap-closed row, on rejected STOP without code changes",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-13T21:53:18.173Z",
    "resolved_at": null
  },
  {
    "id": 19,
    "kind": "unrun-verify",
    "phase": "10",
    "file": ".planning/debug/DEBUG-term-host-live-failures.md",
    "line": null,
    "description": "10-12 D4: вердикт re-acceptance round-2 (шесть пунктов, закрывающий оракул G-10-1) — ОЖИДАНИЕ ответа владельца; grep-гейт матчится с текстом инструкции, не с вердиктом",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-14T10:03:44.986Z",
    "resolved_at": null
  },
  {
    "id": 20,
    "kind": "unrun-verify",
    "phase": "10",
    "file": ".planning/debug/DEBUG-term-host-live-failures.md",
    "line": 556,
    "description": "Вердикт владельца round-4 (гейт: слово в wezterm ОСТАЁТСЯ на v1.1.1-180-gefac750) — human-check честно ОЖИДАЕТ; UAT Current Test №5 pending; не синтезируется",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-16T21:57:28.487Z",
    "resolved_at": null
  },
  {
    "id": 21,
    "kind": "unrun-verify",
    "phase": "9",
    "file": ".planning/phases/09-spayk-i-defekt-wezterm/09-02-SUMMARY.md",
    "line": 72,
    "description": "09-02 (D3–D5): живые раунды R2–R6, полная лестница P1–P5 и two-round wezterm-flip RED-контракт не исполнены — владелец прекратил участие в живых тестах 2026-10-12 (T2-ядро: фоновые окна клавиатурно мертвы); ряды не синтезированы; capability-потребитель Фаз 10/12 обслужен стенд-рядами Фазы 10 (adr-008-spike-log.md)",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-20T14:00:23.408Z",
    "resolved_at": null
  },
  {
    "id": 22,
    "kind": "unrun-verify",
    "phase": "11",
    "file": ".planning/debug/orca-terminal-second-shift.md",
    "line": null,
    "description": "Round 2 journal-verify: живой ряд «следующая коррекция с одного нажатия + lifecycle aftermath=true» не снят — владелец вышел из сессии 23:21; демон активен, ряд завершится на следующей сессии; by-ear зачёт — UAT end-of-phase (G-10-1)",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-31T20:39:21.173Z",
    "resolved_at": null
  },
  {
    "id": 23,
    "kind": "unrun-verify",
    "phase": "11",
    "file": "mise.toml",
    "line": null,
    "description": "11-03: literal 'mise run ci' unrun in the executor session (toolchain seam, 11-01 Rule-3 precedent): GOROOT(mise go1.23.12) vs go binary(system 1.27.3) mismatch in task shells; the SAME gates ran green directly — go build/vet/test -race ./... (22 packages) + golangci-lint 2.15.0 (mise install, GOROOT unset) 0 issues + go mod tidy git-diff clean, STACK.md-sanctioned",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-31T21:18:56.408Z",
    "resolved_at": null
  }
]
````
