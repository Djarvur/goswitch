---
phase: "06"
slug: "avtokorrekcija-opcionalno"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-28"
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Source: 06-RESEARCH.md § Validation Architecture (Test Framework, Phase Requirements → Test Map,
> Sampling Rate, Wave 0 Gaps). Таблицы ниже заполняются планировщиком/validate-phase.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib testing + -race, mise tasks) |
| **Config file** | mise.toml |
| **Quick run command** | `mise exec -- go test ./... -race -count=1` |
| **Full suite command** | `mise run ci` |
| **Estimated runtime** | ~120 seconds |

---

## Sampling Rate

- **After every task commit:** Run `mise exec -- go test ./... -race -count=1`
- **After every plan wave:** Run `mise run ci`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | SPEC §10/§11 (spec-delta) | T-06-02a | audit-trail comments keep the §11 verdict visible | doc greps | `grep -q "v1.1.0" docs/SPEC.md && grep -q "2026-09-27" docs/SPEC.md && …` | ✅ docs/SPEC.md | ⬜ pending |
| 06-01-02 | 01 | 1 | SPEC §10/§11, MACR-ACL | T-06-01a | GPL read-only position, fail-closed recorded | doc greps | `test -f docs/adr/ADR-007-*.md && grep -q "## Status" …` | ❌ W0 | ⬜ pending |
| 06-02-01 | 02 | 2 | CORR-08 | T-06-02a/b | generator hard errors + golden invariants (sorted/disjoint/envelope) | unit+golden | `mise exec -- go test ./layouts/ -race -count=1` | ❌ W0 | ⬜ pending |
| 06-02-02 | 02 | 2 | CORR-08 | T-06-02b | table size caps 4096/lang, deterministic | unit+golden | `mise exec -- go test ./layouts/ -race -count=1 && mise run dictgen-regen && git diff --exit-code -- layouts/` | ❌ W0 | ⬜ pending |
| 06-03-01 | 03 | 2 | MACR-ACL | T-06-03a | (sender,path) cache-keep matches FocusedApp semantics | unit | `mise exec -- go test ./internal/appid/ -race -count=1` | ✅ appid_test.go | ⬜ pending |
| 06-03-02 | 03 | 2 | MACR-ACL | T-06-03a/b | live GetRole, ErrRoleUnknown sentinel, never GetRoleName | unit | `mise exec -- go test ./internal/appid/ -race -count=1` | ✅ appid_test.go | ⬜ pending |
| 06-04-01 | 04 | 2 | MACR-ACL | T-06-04a/b/c | default-off everywhere, ceilings, legacy compat | unit | `mise exec -- go test ./internal/config/ -race -count=1` | ✅ config_test.go | ⬜ pending |
| 06-04-02 | 04 | 2 | MACR-ACL | T-06-04a | hot-reload propagation + last-good | unit | `mise exec -- go test ./internal/config/ -race -count=1` + CONFIG-DOC greps | ✅ watch_test.go | ⬜ pending |
| 06-05-01 | 05 | 3 | CORR-04/05/06 | T-06-05a/b | dictionary-hit veto, closed reason vocabulary | unit (golden corpus) | `mise exec -- go test ./internal/detect/ -race -count=1` | ❌ W0 | ⬜ pending |
| 06-05-02 | 05 | 3 | CORR-04 | T-06-05b | OOV fires via trigram; legit-word budget exactly 0 | unit (golden corpus) | `mise exec -- go test ./internal/detect/ -race -count=1` | ❌ W0 | ⬜ pending |
| 06-05-03 | 05 | 3 | CORR-04/06 | T-06-05a/c | empty/tie/ordering boundary answers pinned | unit | `mise exec -- go test ./internal/detect/ -race -count=1 && mise run ci` | ❌ W0 | ⬜ pending |
| 06-06-01 | 06 | 4 | CORR-01/07/09, MACR-ACL | T-06-06a/c/d | D-53 conjunction fail-closed; silence matrix; existing pipeline only | unit (fakes) | `mise exec -- go test ./internal/correct/ ./internal/session/ -race -count=1` | ✅ actor_test.go | ⬜ pending |
| 06-06-02 | 06 | 4 | CORR-01, MACR-ACL | T-06-06b | word-free records at all levels; status counters | unit | `mise exec -- go test ./internal/session/ ./internal/ctlsvc/ -race -count=1` | ✅ ctlsvc_test.go | ⬜ pending |
| 06-06-03 | 06 | 4 | MACR-ACL | T-06-SC | wiring, zero new deps (tidy-diff) | gate | `mise run ci` | ✅ | ⬜ pending |
| 06-07-01 | 07 | 5 | MACR-ACL | T-06-07a | GTK4 fixture + live Chromium role pin | py_compile + greps | `python3 -m py_compile test/e2e/fixtures/password_entry.py && …` | ❌ W0 | ⬜ pending |
| 06-07-02 | 07 | 5 | MACR-ACL, CORR-01 | T-06-07a/b/c | password-silent + terminal-silent (stale-cache) live | e2e (live) | `mise run e2e-autocorrect-fires` / `-password-silent` / `-terminal-silent` | ❌ W0 | ⬜ pending |
| 06-07-03 | 07 | 5 | SPEC §10/§11 | T-06-07d | matrix v4 rows; v3 frozen | e2e (live) | `mise run e2e-matrix-v4` + v3 diff gate | ❌ W0 | ⬜ pending |
| 06-08-01 | 08 | 6 | SPEC §10/§11 | T-06-08b | perf p95 delta ≤ 5 ms (Pitfall 5) | e2e (live) | `mise run e2e-perf-autocorrect` + ADR greps | ❌ W0 | ⬜ pending |
| 06-08-02 | 08 | 6 | CORR-01..09 (base regression) | T-06-08a/b | double-run gate; README safety caveats | e2e (live) | `mise run e2e-matrix-v4 ×2 && mise run e2e-autocorrect ×2` + README greps | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

From 06-RESEARCH.md § Wave 0 Gaps, assigned to plans:

- [ ] `layouts/dictgen/` + golden dict/trigram data + layouts golden tests (06-02)
- [ ] `internal/detect/` package + mandatory golden corpus — ё-words, OOV names, mixed, legit-word budget, privacy, boundary semantics (06-05)
- [ ] `internal/appid` (sender,path) storage + Role + feed tests (06-03)
- [ ] `internal/config` autocorrect section + validation + hot-reload propagation tests (06-04)
- [ ] `internal/session` word-boundary hook + D-53 conjunction + counters (06-06)
- [ ] `test/e2e/fixtures/password_entry.py` (prototype live-verified by research) (06-07)
- [ ] `test/e2e/case_autocorrect.go` + registry names + mise tasks + matrix-v4 (06-07)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| TBD | TBD | TBD | TBD |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
