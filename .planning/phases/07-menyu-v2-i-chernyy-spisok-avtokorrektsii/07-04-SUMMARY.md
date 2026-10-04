---
phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii
plan: 04
subsystem: session
tags: [autocorrect, blocklist, actor, gate-semantics, tdd, ibus]

# Dependency graph
requires:
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii (07-02)
    provides: blocklist schema + compiled pattern cache (acBlocklist), acReasonAppBlocked slug, conservative actor adaptation (unknown identity still silent — inverted here)
  - phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii (07-01)
    provides: SPEC §11 revision + ADR-007 amendment — the locked semantics this plan implements
  - phase: 06-avtokorrekcija-opcionalno
    provides: the D-53 conjunction gates (autoCorrectBoundary/autoConfirm/acConfirmRefusals), silence-matrix corpus, payload-stale/CR-01 review fixes
provides:
  - Final gate semantics (locked 07-CONTEXT): fire ⇔ enabled ∧ role-gate ∧ NOT blocklist; unknown app identity is NOT a prohibition (pass-through with an app-less payload) on BOTH the arm and confirm gates
  - Confirm gate blocklist-only on the CURRENT identity; the WR-01 arming/confirm equality removed with no remainder (constant + branch + test + doc echo; negative grep clean)
  - ensureAppid starts the a11y observer when the macr list is non-empty OR autocorrect is enabled; enabled:false stays byte-as-today (D-54)
  - Warn discipline: app-identity WARN only on a source ERROR (installed but broken, once per episode); a missing identity is the quiet norm (debug)
  - Slug vocabulary reduced to the closed set (app-blocked, no-caps, role-unknown, role-forbidden, role-timeout, recheck-disabled, payload-stale); app-unknown/app-changed gone
  - Inverted silence-matrix corpus + direction pin (IdentityUnknownNotProhibition) + confirm blocklist-only cells; red-evidence records task1/2/3 (RED_EVIDENCE_OK)
affects: [07-05-switchmode, 07-06-e2e-oracles (app-blocked oбразцы), 07-07, verify-work phase 07, ADR-007 amendment consumers]

# Actuals (#2632)
actuals:
  tokens: 9339    # chars/4 over the realized diff (37355 chars, +366/-119 lines incl. red-evidence)
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Inverted-unknown gate semantics: exactly ONE conjunction segment (app identity) flips to pass-through; every other segment keeps fail-closed — pinned by a dedicated direction test"
    - "Red-evidence TAP transcription for Go corpora (06-03 precedent): go test -v output transcribed to node-test TAP for check tdd-red-evidence"

key-files:
  created:
    - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/red-evidence/07-04-task1-red.json
    - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/red-evidence/07-04-task2-red.json
    - .planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/red-evidence/07-04-task3-red.json
  modified:
    - internal/session/actor.go
    - internal/session/actor_test.go

key-decisions:
  - "Arm inversion touches ONLY the identity segment: unknown passes with an app-less payload and the detector still runs; role/caps/detector stay fail-closed (TestAutoCorrect_IdentityUnknownNotProhibition pins the direction, T-07-04-02)"
  - "Confirm gate is blocklist-only on the CURRENT identity (owner variant 1): known+match refuses app-blocked, unknown passes; the WR-01 equality removed code+constant+test+comment with no remainder — a focus move inside the role-RTT window is payload-stale's business (FocusOut → HardReset; the replacement pipeline re-verifies before any Delete), and keeping the equality conditionally would break the locked «unknown at arm → learned by confirm» scenario with a phantom mismatch"
  - "acFocusedApp treats an empty identity as unknown (no focus event yet): no WARN, debug only — the app-identity WARN is reserved for source ERRORS (installed but broken), once per episode"
  - "ensureAppid condition: macr list non-empty OR AutoCorrectEnabled — the blocklist negatives and the confirm's live role query need the observer whenever the layer is on"
  - "Transitional confirm pass-through (unknown) landed with Task 2 because the acReasonAppUnknown slug had to leave the vocabulary with its emitters (no dead constants — lint); fractional GREEN sanctioned by the plan"

patterns-established:
  - "Direction-pin test pattern: when one segment of a fail-closed conjunction inverts, a dedicated table test re-asserts every other segment's refusal under the same unknown"
  - "Corpus slug constants (acReasonNoCaps/acReasonTrigramUnsure/acUnsureWord) — goconst keeps the matrix tables constant-driven"

requirements-completed: ["MACR-ACL (blocklist-ревизия)", "CORR-01..09 (регресс)"]  # copied verbatim; MARKING deferred: MACR-ACL is shared by 07-02/03/04/06/07 and 07-06/07-07 lack summaries (shared-ID gate #2388; 07-02/07-03 precedent) — the LAST declaring plan marks it. CORR-01..09 are already Complete (phases 2–3; this plan is their regression pin, nothing to mark)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Arm gate inversion: an UNKNOWN app identity passes (payload without app, detector runs, word fires with a healthy role); a KNOWN identity matching a blocklist pattern refuses app-blocked"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestAutoCorrect_UnknownIdentityFires"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestAutoCorrect_SilenceMatrix/unknown_identity_is_not_a_prohibition:_no_source_fires"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestAutoCorrect_SilenceMatrix/blocklist_match_forbids_a_known_app"
        status: pass
    human_judgment: false
  - id: D2
    description: "Confirm gate blocklist-only on the CURRENT identity; WR-01 equality removed with no remainder (constant+branch+test+doc echo)"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestAutoCorrect_ConfirmBlocklistRefuses"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestAutoCorrect_ConfirmUnknownPasses"
        status: pass
      - kind: other
        ref: "grep -E 'AppChanged|app-changed' internal/session/actor.go internal/session/actor_test.go → zero matches (WR01-GONE)"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestAutoConfirm_RevalidatesArmedPayload (payload-stale byte-preserved, CR-01)"
        status: pass
    human_judgment: false
  - id: D3
    description: "ensureAppid starts the a11y observer when the macr list is non-empty OR autocorrect is enabled; enabled:false with no lists starts nothing (D-54 zero behavior)"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_AppidStartsForAutocorrect"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestActor_AutocorrectDisabledSilent (off = zero counters/records)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Direction pin: with an unknown identity, role error/timeout, missing caps and unsure detector keep their fail-closed refusals (the inversion never creeps into safety segments)"
    requirement: "CORR-01..09 (регресс)"
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go#TestAutoCorrect_IdentityUnknownNotProhibition"
        status: pass
      - kind: unit
        ref: "internal/session/actor_test.go#TestAutoCorrect_SilenceMatrix (role/caps/detector cells byte-unchanged vs 07-02 at Task 1; values identical after the Task 3 lint constant-ification)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Warn discipline: app-identity WARN only on a source ERROR (once per episode); a missing identity (no source / no focus event) never warns — quiet debug norm"
    requirement: "MACR-ACL (blocklist-ревизия)"
    verification: []
    human_judgment: true
    rationale: "No unit test asserts the ABSENCE of the app-identity WARN on the missing-identity path (the role-warn episode pin survives unchanged); the journal shape is observable on a live desktop only — deferred to phase UAT/matrix (07-06)"
  - id: D6
    description: "Slug vocabulary reduced to the closed set; manual correction paths (Double/phrase/selection) untouched by the blocklist — existing corpora green without edits"
    requirement: "CORR-01..09 (регресс)"
    verification:
      - kind: other
        ref: "grep -cE 'acReasonNoApps|acReasonAppNotListed|acReasonAppUnknown' internal/session/actor.go → 0"
        status: pass
      - kind: unit
        ref: "mise run ci (build+vet+lint+test -race ./...) — full package green incl. TestActor_Correction*/manual corpora"
        status: pass
    human_judgment: false

# Metrics
duration: 22min
completed: 2026-10-04
status: complete
---

# Phase 7 Plan 04: Гейт-реворк blocklist-семантики Summary

**Финальная семантика гейтов автокоррекции (locked 07-CONTEXT): unknown-идентичность приложения — не запрет (проходит и арм, и confirm с payload без приложения), confirm сверяет ТОЛЬКО blocklist текущей идентичности, равенство WR-01 удалено без остатка; роль/капсы/детектор остались fail-closed.**

## Performance

- **Duration:** 22 min
- **Started:** 2026-10-04T22:11:20Z
- **Completed:** 2026-10-04T22:29:38Z
- **Tasks:** 3
- **Files modified:** 2 code files + 3 red-evidence records

## Accomplishments

- Арм-гейт инвертирован РОВНО одним сегментом: неизвестная идентичность (нет источника / нет фокус-события) больше не молчит — идентичность опускается (payload без приложения), детектор выполняется, слово исправляется; известная идентичность с матчащим паттерном отказывает `app-blocked` (полярность 07-02 сохранена)
- Confirm-гейт переведён на blocklist-only по ТЕКУЩЕЙ идентичности (вариант 1 владельца): известная+матч → отказ `app-blocked`, неизвестная → пропуск; payload-stale (CR-01) байт-сохранён; равенство приложения WR-01 удалено вместе с константой, веткой, тестом и коммент-эхом — негативный греп `AppChanged|app-changed` по коду и корпусу чист (T-07-04-03)
- ensureAppid стартует a11y-наблюдателя при `AutoCorrectEnabled` даже с пустым списком (негативы blocklist и role-источник confirm'а требуют наблюдателя); enabled:false — ноль стартов/счётчиков/записей (D-54)
- Warn-дисциплина: WARN идентичности только на ОШИБКУ источника (warn-once на эпизод), отсутствие идентичности — тихая норма (debug)
- Словарь слагов сокращён до закрытого набора (app-unknown/app-changed/no-apps/app-not-listed удалены; app-blocked — единственный идентичностный отказ); ручные пути (Double/фраза/выделение) не читают blocklist — существующий корпус зелёный без правок

## TDD Gate Compliance

- RED: `d5b9ab0` `test(07-04)` — 4 из 5 именованных кейсов красные против кода 07-02 на запланированных утверждениях (KnownNotBlockedFires — green-by-design continuity pin, прецедент 06-03 ErrRoleUnknown); red-evidence `07-04-task{1,2}-red.json` → RED_EVIDENCE_OK
- RED (Task 3): после Task 2 ровно ConfirmBlocklistRefuses красная (`app-changed` вместо `app-blocked` — живой WR-01-остаток); `07-04-task3-red.json` → RED_EVIDENCE_OK
- GREEN: `25bdc5d` (арм), `e01ae82` (confirm) — оба `feat(07-04)`
- REFACTOR: не отдельным коммитом — lint-принуждённая констант-фикация двух прежних ячеек матрицы (goconst) вошла в GREEN Task 3 с нулевым семантическим диффом

## Task Commits

Each task was committed atomically:

1. **Task 1: RED — инверсия корпуса** - `d5b9ab0` (test)
2. **Task 2: GREEN арм — unknown ⇒ пропуск, ensureAppid по enabled, warn-дисциплина** - `25bdc5d` (feat)
3. **Task 3: GREEN confirm — blocklist-only, удаление равенства (WR-01)** - `e01ae82` (feat)

## Files Created/Modified

- `internal/session/actor.go` — новая семантика autoCorrectBoundary (unknown⇒пропуск), acConfirmRefusals (blocklist-only на текущей идентичности), acFocusedApp (пустая идентичность = unknown, WARN только на ошибку источника), ensureAppid по enabled; реестр слагов без app-unknown/app-changed; doc-комментарии переписаны (обоснование удаления WR-01, ссылка на ADR-007 amendment)
- `internal/session/actor_test.go` — инвертированный silence-matrix (fire-ячейка unknown), пять новых кейсов (UnknownIdentityFires, KnownNotBlockedFires, ConfirmBlocklistRefuses, ConfirmUnknownPasses, IdentityUnknownNotProhibition), ensureAppid-пин; тесты равенства и FailClosedNoFailOpen удалены (заменены направлением)
- `.planning/phases/07-.../red-evidence/07-04-task{1,2,3}-red.json` — TAP-транскрипции RED-прогонов (RED_EVIDENCE_OK все три)

## Decisions Made

- Равенство WR-01 удалено целиком без условных форм — консервативная альтернатива отвергнута владельцем (07-CONTEXT, вариант 1); doc-комментарий несёт обоснование (payload-stale ловит перенос фокуса, условное равенство ломало бы сценарий «unknown на арме → выучили к confirm» мнимым mismatch)
- Пустая идентичность от источника (нет фокус-события ещё) трактуется как unknown в acFocusedApp — оба гейта видят одинаковую семантику «не видно = не запретить»
- Переходный confirm-пропуск вошёл в Task 2: слаг app-unknown обязан был покинуть реестр вместе с эмиттерами (verify-греп Task 2), дробный GREEN санкционирован планом

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] ConfirmUnknownPasses: assert ladder без settles верифай-раунда**
- **Found during:** Task 2 (GREEN-прогон)
- **Issue:** мой новый тест Task 1 требовал deletes==1 сразу после fired — а конвейер запускает pre-correction verify, который надо settle'ить (танец fireAutocorrect); RED-утверждение было честным, GREEN-проверка — нет
- **Fix:** тест доводит верифай-раунд полем post-boundary состояния (eventually requireCount>=1 → HandleSurroundingText) перед assert deletes
- **Files modified:** internal/session/actor_test.go
- **Verification:** TestAutoCorrect_ConfirmUnknownPasses зелёный, полный пакет зелёный
- **Committed in:** 25bdc5d (Task 2 commit)

**2. [Rule 3 - Blocking] goconst/lll: строгий lint против нового корпуса**
- **Found during:** Task 3 (mise run ci)
- **Issue:** литералы "no-caps"/"trigram-unsure"/"vjcrdf" достигли порога goconst (мои новые таблицы добавили третьи/четвёртые вхождения); комментарий SetOptions превысил lll 120
- **Fix:** константы корпуса acReasonNoCaps/acReasonTrigramUnsure/acUnsureWord (две прежние ячейки матрицы механически переведены на константы — значения идентичны), комментарий перенесён на свою строку
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go
- **Verification:** mise run ci зелёный (exit 0)
- **Committed in:** e01ae82 (Task 3 commit)

---

**Total deviations:** 2 auto-fixed (1 bug in own new test, 1 lint-blocking). **Impact on plan:** оба фикса необходимы для зелёной итерации (директива 2); семантика ячеек роли/капсов/детектора не изменилась.

## Issues Encountered

- None — оба известных флака (TestRun_OnConnHookCalledOnce, TestWatch_BrokenBlocklistPatternKeepsLastGood) прошли с первого раза, re-run не понадобился.

## TDD RED Evidence

- Записи: `.planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/red-evidence/07-04-task{1,2,3}-red.json` — TAP-транскрипция вывода `go test -v` per repo precedent (verb парсит только node-TAP), все три проверены `check tdd-red-evidence` → `RED_EVIDENCE_OK`; сырые выводы в `/tmp/07-04-t{1,2,3}-red-raw.txt`.

## Threat Model Mitigations (T-07-04-01..06)

- T-07-04-01 (half-переворот): RED-пин «unknown ⇒ fired» + IdentityUnknownNotProhibition; слаги/ветки удалены, корпус инвертирован — mitigate выполнено
- T-07-04-02 (расползание): направление пинено — роль/капсы/детектор отказывают при неизвестной идентичности; сохранённые ячейки зелёные
- T-07-04-03 (тихий остаток равенства): негативный греп код+корпус чист; doc-комментарий переписан
- T-07-04-05 (утечка слова): новые записи — слаги/класс ошибки, слово не появляется (D-20/D-21, captureLogs-пины целы)
- T-07-04-06 (DoS): кэш компиляции 07-02 переиспользуется обоими гейтами — одна семантика матчинга

## Threat Flags

None — новых поверхностей план не создаёт (гейт потребляет существующую a11y-идентичность; новые refusal-слаги входят в закрытый словарь ctlsvc renderStatus автоматически).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Гейт-семантика финальная и запинена корпусом — план 07-06 (e2e-оракулы app-blocked, упрощение хореографий) и 07-05 (SwitchMode поверх того же актора) строят на стабильном контракте
- MACR-ACL в REQUIREMENTS.md остаётся открытой до 07-06/07-07 (shared-ID gate, прецедент 07-02/03)
- WINDOWS-леджер не пополнялся: стабов, skipped-тестов и невыполненных verify нет; оба известных флака зелёные с первого прогона

---
*Phase: 07-menyu-v2-i-chernyy-spisok-avtokorrektsii*
*Completed: 2026-10-04*

## Self-Check: PASSED

- Created files verified on disk: actor.go, actor_test.go, red-evidence/07-04-task{1,2,3}-red.json — all FOUND
- Commits verified in git log: d5b9ab0 (test), 25bdc5d (feat), e01ae82 (feat) — all FOUND; measured commits from ledger 8ed1bd8..HEAD = 3 (matches frontmatter `commits: 3`)
- Plan-level verification re-run at close: full `internal/session` green under -race, negative greps (app-unknown/app-changed/no-apps/app-not-listed = 0; AppChanged|app-changed = 0), `mise run ci` exit 0
