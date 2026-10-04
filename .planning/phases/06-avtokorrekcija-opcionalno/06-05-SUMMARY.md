---
phase: 06-avtokorrekcija-opcionalno
plan: 05
subsystem: detect
tags: [wrong-layout-detection, dictionary-path, trigram-scoring, binary-search, golden-corpus, privacy, tdd]

# Dependency graph
requires:
  - phase: 06-avtokorrekcija-opcionalno (06-02)
    provides: DictRU 138 914 / DictEN 78 951 (sorted []string, ё→е folded) + TriRU/TriEN (log10, ≤4096/язык) — входы Data/Trigrams
  - phase: 06-avtokorrekcija-opcionalno (06-04)
    provides: конфиг-дефолты порогов 4/2.0/1.0 — пара мест с DefaultParams (кросс-комментарии обе стороны)
  - phase: 02 (internal/correct)
    provides: correct.Convert/Dir — единственный ремап-примитив (Don't Hand-Roll), отказ конвертации = abstain
provides:
  - internal/detect — чистый headless-пакет: Verdict{WrongLayout, Confident, Dir, Reason}, Data, Trigrams, Params, DefaultParams (4/2.0/1.0), Check(tok, mode, Data, Trigrams, Params) — детерминированная функция над рунами
  - Словарный путь D-52(а) с двухпробным lookup (точное написание + нормализованный запрос lower+ё→е) и двойным вето (cur-hit, both-hit) — валидное слово НИКОГДА не корректируется
  - Триграммный fallback D-52(б): scoreLang (Σ log10 P по 3-рунным окнам, neutralPenalty -6.0, нормировка длиной), уверенность только при margin+floor
  - Замкнутый словарь причин из 8 констант — единственная строка в Verdict, токен невосстановим (D-20/D-21, Pitfall 6)
  - Golden-корпус на реальных данных: 190 wrong-layout (оба направления, три регистра, Vfrcbv-класс), 404 легитимных (ноль ложных), 34 OOV (никогда не уверены), yo/mixed/boundary/приватность — критерий 1 фазы
affects: [06-06 session (хук границы слова зовёт Check), 06-07 ctlsvc (счётчики по причинам), 06-08 acceptance (полный регресс)]

# Actuals (#2632)
actuals:
  tokens: 12590        # chars/4 over the realized diff (50 358 bytes) — оценка плана 62 000 близка
  tasks: 3
  commits: 6           # MEASURED: git rev-list --count gsd-plan-head-before-06-05..HEAD

# Tech tracking
tech-stack:
  added: []            # ноль новых зависимостей — stdlib sort/strings/unicode + internal/correct + layouts (tidy-diff зелёный)
  patterns:
    - "Двухпробный словарный lookup: точное написание (капс-статьи подавляют) + нормализованный запрос (lower+ё→е — регистр- и ё-варианты набираемого находят е-статьи фолда 06-02)"
    - "Порогово-мёртвый fallback как санкционированный старт: floor 1.0 выше log10-шкалы — trigram-unsure всегда, пока корпус/конфиг не перевыпинят пару мест"
    - "Golden-корпус со строительством из данных: сэмпл фиксированным шагом + проверка определения класса (definition → expectation), ожидания выводимы, не захардкожены"

key-files:
  created:
    - internal/detect/detect.go
    - internal/detect/detect_test.go
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-05-task1-red.json
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-05-task2-red.json
    - .planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-05-task3-red.json
    - .planning/phases/06-avtokorrekcija-opcionalno/deferred-items.md
  modified: []

key-decisions:
  - "Двухпробный словарный lookup (точный + нормализованный lower/ё→е): план-корпус требует уверенную детекцию трёх регистров и ё-слов (Pitfall 3), а данные 06-02 хранят статьи в смешанном регистре с фолдом ё — запрос обязан фолдиться так же, как данные; точная проба сохраняет подавление капс-статей (АЗС, Москва)"
  - "Floor 1.0 при дефолтах лежит выше log10-шкалы таблиц (все значения ≤ 0): fallback честно отвечает trigram-unsure на каждом двустороннем miss — санкционированный планом консервативный старт (research A4 «fallback отключается порогами — санкционированный исход»); корпус подтвердил, config 06-04 не менялся"
  - "OOV-класс Pitfall 2 = двусторонний словарный miss: план-примеры Vfrcbv (ремап максим — статья DictRU) и vjcrdf (ремап москва — НЕ статья) лежат по разные стороны словарной границы; Vfrcbv перенесён в wrong-layout-корпус (уверенная коррекция = работа фичи по D-52а), OOV-набор из 34 кейсов требует unsure"
  - "Граница длины старше ё-фолда: ещё/АЗС (3 руны) из план-примеров корпуса абстейнятся до вопроса о фолде — пин порядка гейтов (len < MinWordLen → abstain-short всегда, empty/1-рунный вход тоже)"

patterns-established:
  - "Verdict без токена: 4 поля, Reason — замкнутый набор из 8 констант, структурный пин рефлексией в корпусе (T-06-05-02)"
  - "Класс корпуса строится из определения (обе стороны dictionary membership проверяются ассертами-предусловиями с громким Fatal) — дрейф данных громко валит construction, а не молча меняет семантику"

requirements-completed: ["CORR-04", "CORR-05", "CORR-06", "SPEC §10/§11 (spec-delta)"]  # провизорные метки плана: CORR-04/05/06 заблокированы шлагбаумом #2388 (разделяются с 06-06..06-08 без SUMMARY — чекбоксы REQUIREMENTS.md при последнем объявившем плане); свободная метка «SPEC §10/§11» не размечается verb'ом (пробел-лейбл, прецедент 06-02); формальная фиксация REQ-ID — при new-milestone v1.1.0 (D-51)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Словарный путь D-52(а): confident wrong-layout при miss-в-своём + hit-в-чужом через переиспользованный correct.Convert, двойное вето (cur-hit/both-hit — валидное слово никогда), abstain-границы (короткий/nil, без-букв, unknown-mode, mixed), SearchStrings-единственный поиск, slog-ноль"
    requirement: "CORR-04"
    verification:
      - kind: unit
        ref: "tests/internal/detect/detect_test.go#TestCheck_DictWrongLayout + TestCheck_DictCurHitVetoes + TestCheck_DictBothHitVetoes + TestCheck_AbstainShort + TestCheck_AbstainNoLetters + TestCheck_UnknownModeAbstains + TestCheck_AbstainConvertFail + TestCheck_DirMatchesCorrect"
        status: pass
      - kind: other
        ref: "греп-гейт DICT-PATH-OK: sort.SearchStrings + correct.Convert в detect.go, slog — 0 вхождений"
        status: pass
    human_judgment: false
  - id: D2
    description: "Триграммный fallback D-52(б): scoreLang (log10-окна по 3 руны, neutralPenalty -6.0, нормировка длиной), уверенность только при (score_other − score_cur) > margin И score_other > floor; dict-cur-hit не доходит до скоринга"
    requirement: "CORR-04"
    verification:
      - kind: unit
        ref: "tests/internal/detect/detect_test.go#TestTrigram_WrongLayout + TestTrigram_UnsureBelowMargin + TestTrigram_UnsureBelowFloor + TestTrigram_NeutralTrigramsUnknown + TestTrigram_NormalizedByLength + TestTrigram_DictHitSkipsScoring"
        status: pass
    human_judgment: false
  - id: D3
    description: "Golden-корпус wrong-layout: 190 кейсов (оба направления, три регистра, ≥40 реальных слов, флагманы ghbdtn/руддщ ×3 регистра, Vfrcbv-класс, йцукен→qwerty) — все уверенно детектированы на реальных данных"
    requirement: "CORR-05"
    verification:
      - kind: unit
        ref: "tests/internal/detect/detect_test.go#TestCorpus_WrongLayoutBothDirections (190 subtests green)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Критерий 1, сторона безопасности: 404 легитимных слова обоих алфавитов — ноль wrong-layout, каждое отвечает cur-hit-вето (бюджет ложных = РОВНО 0)"
    requirement: "CORR-05"
    verification:
      - kind: unit
        ref: "tests/internal/detect/detect_test.go#TestCorpus_LegitWordsZeroFalsePositives (404 subtests green)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Pitfall 3 (ё/е): ёлка/ёжик + регистры — cur-hit через ё→е-фолд запроса; латинские ремапы ёлка/ёжик — уверенная коррекция; ещё (3 руны) — пин порядка гейтов (abstain-short до вопроса о фолде)"
    requirement: "CORR-06"
    verification:
      - kind: unit
        ref: "tests/internal/detect/detect_test.go#TestCorpus_YoWords"
        status: pass
    human_judgment: false
  - id: D6
    description: "Pitfall 2 (OOV): 34 двусторонних словарных miss (имена/ники/аббревиатуры/клавиатурный мусор, оба направления) — ни одного Confident=true; construction-предусловия с громким Fatal ловят дрейф данных"
    requirement: "CORR-06"
    verification:
      - kind: unit
        ref: "tests/internal/detect/detect_test.go#TestCorpus_OOVNames (34 subtests green)"
        status: pass
    human_judgment: false
  - id: D7
    description: "CORR-06 (mixed — abstain-convert-fail, никогда частичная правка) + приватность: замкнутость 8 причин по всему корпусу, рефлексия 4 полей Verdict без второй строковой дороги, ни одна строка корпуса не входит в вердикт"
    requirement: "CORR-06"
    verification:
      - kind: unit
        ref: "tests/internal/detect/detect_test.go#TestCorpus_MixedTokens + TestCorpus_ReasonVocabularyClosed"
        status: pass
    human_judgment: false
  - id: D8
    description: "Зелёная итерация D-08: mise run ci (build+vet+lint strict+test -race) зелёный; tidy-diff без новых зависимостей; пороги подтверждены корпусом на дефолтах 4/2.0/1.0 — пара с config 06-04 не тронута"
    requirement: "SPEC §10/§11 (spec-delta)"
    verification:
      - kind: other
        ref: "mise run ci exit 0 (2026-10-02, /tmp/06-05-ci3.log); go mod tidy && git diff --exit-code -- go.mod go.sum → TIDY-CLEAN"
        status: pass
    human_judgment: false

# Metrics
duration: 42min
completed: 2026-10-01
status: complete
---

# Phase 6 Plan 5: internal/detect — гибридный детектор неверной раскладки Summary

**Чистый детектор D-52: словарный путь с двухпробным lookup (точный + нормализованный lower/ё→е) и двойным вето, триграммный fallback на запечённых таблицах (порогово-мёртвый при дефолтах — санкционированный консервативный старт), замкнутые причины без токена, и обязательный golden-корпус на реальных данных: 190 уверенных wrong-layout в три регистра, ноль ложных на 404 легитимных, 34 OOV никогда не уверены.**

## Performance

- **Duration:** 42 min (21:57–22:39 UTC)
- **Started:** 2026-10-01T21:56:51Z
- **Completed:** 2026-10-01T22:39:12Z
- **Tasks:** 3 (каждая — полный RED→GREEN по RED_EVIDENCE_OK)
- **Files modified:** 6 создано (2 кода + 3 red-evidence + deferred-items)

## Accomplishments
- internal/detect — чистый headless-пакет (дисциплина internal/correct: без D-Bus/часов/горутин, весь корпус под -race): Verdict{WrongLayout, Confident, Dir, Reason}, Data/Trigrams/Params, DefaultParams 4/2.0/1.0 с кросс-комментарием к паре config 06-04, детерминированный Check
- Словарный путь D-52(а): вето cur-hit и both-hit (каждый словарный hit подавляет коррекцию — полнота = безопасность), уверенный verdict при miss-в-своём + hit-в-чужом через переиспользованный correct.Convert; abstain-границы: короткий/nil, без-букв режима, unknown-mode, mixed (отказ Convert = молчание, никогда частичная правка)
- Двухпробный lookup: точная проба сохраняет подавление капс-статей (АЗС, Москва), нормализованная проба (lower + ё→е — тот же фолд, что запёк dictgen 06-02) находит регистр- и ё-варианты набираемого — трёхрегистровая уверенность и Pitfall 3 закрыты на данных, а не оговорками
- Триграммный fallback D-52(б): scoreLang по 3-рунным окнам с neutralPenalty -6.0 и нормировкой длиной, уверенность только при margin+floor; при дефолтах floor 1.0 выше log10-шкалы — fallback честно unsure (A4-санкционированный старт), словарный путь несёт всю уверенность
- Golden-корпус на реальных данных (критерий 1): 190 wrong-layout уверенно (оба направления, три регистра, Vfrcbv-класс, йцукен→qwerty), 404 легитимных с нулём ложных и поголовным cur-hit-вето, 34 OOV никогда не уверены (construction-предусловия с громким Fatal), yo/mixed/замкнутость причин + рефлексия приватности
- mise run ci зелёная на закрытии каждой задачи; ноль новых зависимостей (TIDY-CLEAN); греп-гейты DICT-PATH-OK

## Task Commits

Each task was committed atomically (TDD: RED → GREEN per task):

1. **Task 1: Словарный путь D-52(а)** — `359ac63` (test, RED) → `4a71d3d` (feat, GREEN)
2. **Task 2: Триграммный fallback D-52(б)** — `c4e3aaf` (test, RED) → `041fda3` (feat, GREEN)
3. **Task 3: Golden-корпус (критерий 1)** — `b696b03` (test, RED — естественный) → `8825908` (test, GREEN — калибровка классов корпуса)

**Plan metadata:** _этот коммит — docs(06-05)_

_Note: REFACTOR-коммиты не потребовались — линт-риффы (именованные константы DefaultParams, trigramLen, prealloc/goconst корпуса) входили в GREEN-коммиты (зелёная итерация D-08 до коммита, lint 0 issues на каждом)._

## TDD Gate Compliance

| Task | RED | GREEN | REFACTOR | Status |
|------|-----|-------|----------|--------|
| 06-05 Task 1 | ✓ `359ac63` (RED_EVIDENCE_OK — 8/8 TestCheck_* падают на утверждениях против RED-stub) | ✓ `4a71d3d` | — | Pass |
| 06-05 Task 2 | ✓ `c4e3aaf` (RED_EVIDENCE_OK — цель TestTrigram_WrongLayout; дерево RED несёт план-предписанную форму Task 1 «fallback пока unsure-only», скоринг возвращён в GREEN байт-в-байт) | ✓ `041fda3` | — | Pass |
| 06-05 Task 3 | ✓ `b696b03` (RED_EVIDENCE_OK — естественный RED на реальных данных: ещё/АЗС/to` из план-примеров ниже запиненного MinWordLen 4) | ✓ `8825908` (тип test — калибровка КЛАССОВ корпуса; детектор не требовал правок, «тюнинг порогами» свёлся к подтверждению дефолтов) | — | Pass (GREEN test-type задокументирован) |

Все три RED-корпуса прошли `check tdd-red-evidence` с вердиктом RED_EVIDENCE_OK (целевой тест падает на утверждении запланированного поведения, не на компиляции/фикстурах).

## Files Created/Modified
- `internal/detect/detect.go` — пакетная документация (headless + приватность D-20/D-21), 8 констант причин (замкнутый словарь), Verdict/Data/Trigrams/Params/DefaultParams, Check (гейты длина → скрипт/режим → словарные пробы → классификация вето/wrong-layout/abstain → fallback), inDict/has/normalizeLookup, scoreLang/neutralPenalty/trigramLen/trigramWrongLayout
- `internal/detect/detect_test.go` — фикстурный корпус Task 1-2 (мини-наборы + стрид-сэмпл реальных данных, синтетические триграммные таблицы) + golden-корпус Task 3 на реальных данных: TestCorpus_* (6 классов, 628+ subtests), TestCheck_* (8), TestTrigram_* (6)
- `.planning/phases/06-avtokorrekcija-opcionalno/red-evidence/06-05-task{1,2,3}-red.json` — RED-записи (все RED_EVIDENCE_OK)
- `.planning/phases/06-avtokorrekcija-opcionalno/deferred-items.md` — незапланированная находка (флак ctlsvc, вне скоупа)

## Decisions Made
См. key-decisions: двухпробный lookup (запрос фолдится как данные), порогово-мёртвый fallback при дефолтах (A4), OOV = двусторонний miss (Vfrcbv → wrong-layout-класс), порядок гейтов «длина старше фолда» (ещё/АЗС).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Порядок классификации словарных веток: действие плана противоречило собственному behavior-блоку**
- **Found during:** Task 1 (GREEN)
- **Issue:** action предписывает ранний возврат dict-cur-hit до конвертации, но behavior TestCheck_DictBothHitVetoes требует ReasonDictBothHit для слова, валидного в обоих словарях — при раннем возврате обе-hit-слово отвечало бы dict-cur-hit
- **Fix:** обе проверки членства вычисляются до классификации; both-hit проверяется первым, затем cur-hit, затем wrong-layout — каждое вето сохранено, все истины must_haves держатся
- **Files modified:** internal/detect/detect.go
- **Verification:** TestCheck_DictBothHitVetoes зелёный; все 8 TestCheck_* зелёные
- **Committed in:** 4a71d3d

**2. [Rule 2 - Missing critical] Нормализованная вторая словарная проба (lower + ё→е)**
- **Found during:** Task 3 (RED, честные падения корпуса)
- **Issue:** однопробный точный lookup из action-текста не может исполнить запиненные планом поведения: данные 06-02 хранят статьи в смешанном регистре с ё→е-фолдом — «GHBDTN»→«ПРИВЕТ», «Ghbdtn»→«Привет», «ёлка» не находят статей, трёхрегистровая уверенность и Pitfall 3 недостижимы
- **Fix:** inDict пробует точное написание, затем нормализованный запрос (lower + ё→е — тот же фолд, что запёк dictgen); точная проба сохраняет подавление капс-статей; поиск — по-прежнему только sort.SearchStrings (2 бинпоиска максимум)
- **Files modified:** internal/detect/detect.go
- **Verification:** TestCorpus_WrongLayoutBothDirections 190/190 (три регистра), TestCorpus_YoWords зелёный; легитимный корпус по-прежнему 0 ложных (вето раньше)
- **Committed in:** 4a71d3d (механизм), корпус-доказательство b696b03/8825908

**3. [Rule 1 - Bug] OOV-класс плана внутренне противоречив: Vfrcbv не OOV по определению D-52**
- **Found during:** Task 3 (конструирование корпуса)
- **Issue:** план требует «ни одного Confident=true» на OOV-наборе, называя Vfrcbv — но его ремап максим есть статья DictRU, что делает Vfrcbv словарным confident wrong-layout (тот же контракт, что флагман ghbdtn→привет); одновременно второй план-пример «Москва-в-латинице» (vjcrdf) — истинный двусторонний miss (москва НЕ статья), то есть примеры лежат по разные стороны словарной границы
- **Fix:** OOV-класс определён как двусторонний словарный miss ( literal «out of vocabulary»); Vfrcbv (×3 регистра) перенесён в wrong-layout-корпус с уверенным ожиданием; construction-предусловия OOV громко Fatal'ят любой дрейф данных; vjcrdf остался OOV → unsure
- **Files modified:** internal/detect/detect_test.go
- **Verification:** TestCorpus_OOVNames 34/34 unsure-at-most; name vfrcbv×3 confident в TestCorpus_WrongLayoutBothDirections
- **Committed in:** b696b03

**4. [Rule 3 - Blocking] Task 1 GREEN включил скоринг Task 2 (одноэтапная реализация) — честный RED Task 2 восстановлен мутацией дерева**
- **Found during:** Task 2 (RED)
- **Issue:** реализация легла одним куском, и TestTrigram_* были бы зелёными сразу — «unexpected GREEN in RED phase» (fail-fast tdd.md); план же предписывает Task-1-форме ветку «пока ReasonTrigramUnsure»
- **Fix:** для RED-коммита Task 2 дерево возвращено к план-предписанной форме (скоринг отсутствует, ветка unsure-only; lint 0), RED доказан честно (RED_EVIDENCE_OK), скоринг возвращён в GREEN байт-в-байт (прецедент mutation-RED 06-02 Task 2)
- **Files modified:** internal/detect/detect.go (RED-состояние в c4e3aaf, восстановление в 041fda3)
- **Verification:** c4e3aaf: TestTrigram_WrongLayout красный на утверждении; 041fda3: весь корпус зелёный
- **Committed in:** c4e3aaf, 041fda3

**5. [Rule 1 - Bug] План-примеры корпуса ещё/АЗС лежат ниже собственного запиненного порога MinWordLen 4**
- **Found during:** Task 3 (RED — естественные падения)
- **Issue:** поведение «трёхрегистровые ghbdtn-классы + ёлка/ещё» сталкивается с забиненным гейтом len < 4 → abstain-short: ещё/АЗС (3 руны) абстейнятся до всякого словарного вопроса; детектор прав по контракту
- **Fix:** классы корпуса выровнены с гейтом: ещё пинится как abstain-short-граница в TestCorpus_YoWords (порядок «длина старше фолда»), АЗС заменён 4+-рунными капс-статьями (Москва/Россия/Александр/Питер); порог НЕ тронут (пара с config 06-04)
- **Files modified:** internal/detect/detect_test.go
- **Verification:** TestCorpus_YoWords/LegitWords зелёные; honest-RED зафиксирован в red-evidence/06-05-task3-red.json
- **Committed in:** b696b03 (RED-состояние), 8825908 (калибровка)

---

**Total deviations:** 5 auto-fixed (3 Rule 1, 1 Rule 2, 1 Rule 3).
**Impact on plan:** все отклонения — сверки внутриплановых противоречий в пользу behavior-блоков и запиненных контрактов плана; расширения скоупа нет, ноль новых зависимостей, config 06-04 не тронут (пара мест согласована).

## Issues Encountered
- `internal/ctlsvc` TestRun_OnConnHookCalledOnce упал раз в полном `mise run ci` («OnConn called 0 times, want exactly 1») — предсуществующий флак локального unix-socket-беса чужого пакета (в логе того же прогона «another goswitchd instance already registered»); изолированно 4/4 зелёный, полный ci на том же дереве зелёный. Записан в deferred-items.md фазы (вне скоупа 06-05).
- `requirements.ready-ids` разбивает свободную метку «SPEC §10/§11 (spec-delta)» по пробелам на мусорные ID — маркирование пропущено по прецеденту 06-02 (формальная фиксация REQ-ID при new-milestone D-51); CORR-04/05/06 честно заблокированы шлагбаумом #2388 до SUMMARY планов 06-06..06-08.
- Уточнение масштаба порогов: значения таблиц 06-02 — log10 (все ≤ 0), поэтому дефолтный floor 1.0 делает fallback порогово-мёртвым; это зафиксировано в комментарии DefaultParams как санкционированный консервативный старт (research A4), а не дефект — уверенность при дефолтах несёт исключительно словарный путь.

## Known Stubs

None — реализация полная, заглушек нет (RED-stub Task 1 заменён реализацией в том же плане; все verify-гейты исполнены).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- 06-06 (актор) получает контракт Check: mode «en»/«ru» (строка режима актора D-34), Data/Trigrams из layouts.Dict*/Tri*, Params из config-дефолтов 06-04; уверенный вердикт = условие 4 конъюнкции D-53; счётчики причин — из замкнутого набора 8 констант (точно для ctl-статуса 06-07)
- Границы контракта, важные актору: пустой/1-рунный токен всегда abstain-short; mixed — abstain-convert-fail (никогда частичная правка); post-correction слово валидно → cur-hit → вето (дебаунс бесплатен)
- При будущем перевыпинивании порогов менять DefaultParams и config Defaults 06-04 ПАРОЙ (кросс-комментарии на обеих сторонах); fallback оживёт только при floor ≤ 0 (log10-шкала)

---
*Phase: 06-avtokorrekcija-opcionalno*
*Completed: 2026-10-01*

## Self-Check: PASSED

- Файлы: internal/detect/detect.go, internal/detect/detect_test.go, 3 red-evidence JSON, deferred-items.md — существуют на диске (6/6 FOUND)
- Коммиты: 359ac63, 4a71d3d, c4e3aaf, 041fda3, b696b03, 8825908 — присутствуют в git log; measured commits от ledger gsd-plan-head-before-06-05: 6
- План-level verification перезапущена на финальном дереве: `go test ./internal/detect/ -race -count=1` ok; `mise run ci` exit 0; греп-гейты DICT-PATH-OK (SearchStrings+Convert в, slog 0); `go mod tidy && git diff --exit-code -- go.mod go.sum` → TIDY-CLEAN
- TDD-гейт: test(06-05) `359ac63`/`c4e3aaf`/`b696b03` предшествуют feat-коммитам; все три RED-записи — RED_EVIDENCE_OK по check tdd-red-evidence
