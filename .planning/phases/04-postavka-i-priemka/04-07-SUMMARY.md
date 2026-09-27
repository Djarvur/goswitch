---
phase: 04-postavka-i-priemka
plan: 07
subsystem: docs-release
tags: [readme, bilingual, acceptance-uat, windows-ledger, perf-budget, release-gate, inst-01, v1.0.0]

# Dependency graph
requires:
  - phase: 04-postavka-i-priemka
    provides: "04-01: goswitchctl install/uninstall facts; 04-02: selfcheck six steps; 04-03: release channels (.goreleaser.yaml + release.yml); 04-04: perf-report.txt live numbers + the latency escalation; 04-05: matrix v3; 04-06: D-48 double-run mechanics + ci-runner owner procedure"
provides:
  - "Двуязычная пара README.md (EN, первичен) + README.ru.md (RU): установка одной командой goswitchctl install (канал релиз-архива + go install), selfcheck, жесты/CLI, конфиг (дефолты до первого YAML), perf-таблица из живого прогона С честным статусом бюджета, расширенное -debug-предупреждение, troubleshooting, uninstall --purge"
  - "docs/ACCEPTANCE.md — письменный чек-лист приёмки владельца D-49 (формат 03-UAT): 6 пунктов expected/result/note + Summary-блок"
  - "WINDOWS-ledger свёрнут В НОЛЬ глаголами: #2 fixed, #4 waived, #5 waived решением владельца (в) — open_count 0, ship-гейт свободен"
  - "ОПУБЛИКОВАН РЕЛИЗ v1.0.0: тег → release-workflow run 35130309174 (success) → GitHub Release v1.0.0 с goswitch_1.0.0_linux_amd64.tar.gz (оба бинарника + LICENSE + README-пара) + checksums.txt; канал go install виден module proxy (v1.0.0.info, Hash == коммит тега)"
  - "release.yml зарегистрирован на default branch (bootstrap-прецедент 02-07, PR #5) — последующие теги публикуются без ручных шагов"
affects: [verify-work фазы 4 (UAT по docs/ACCEPTANCE.md — следующий гейт), milestone v1.0.0 close]

# Actuals (#2632) — pairs with the plan's `estimate` to calibrate future estimates.
actuals:
  tokens: 9100    # chars/4 over the realized diff (доки-пара + чек-лист + ledger + релизные вердикты)
  tasks: 3        # of 3 — Task 3 завершён полностью: гейты → тег → публикация → все проверки
  commits: 6      # MEASURED: git rev-list --count 657fbc4..2bc8c6a (b694e64, 8ac3451, dd645f6, b1f0aaa, b8e969d, 2bc8c6a + финальные доки)

# Tech tracking
tech-stack:
  added: []       # только доки, ledger-глаголы и публикация; ноль зависимостей
  patterns:
    - "README-таблица D-46: числа из последнего perf-report.txt + строка методики (окно ydotool→AT-SPI, оверхед инжектора назван явно) + ЧЕСТНЫЙ статус бюджета (end-to-end окно ПРЕВЫШАЕТ 50 мс — превышение даёт инжектор стенда, демон < 2 мс) + примечание об обновлении каждым релизом"
    - "ledger-сверка только глаголами gsd-tools (fixed/waive с причиной из решений владельца) — никаких ручных правок статусов"
    - "Bootstrap-регистрация workflow на default branch ЕДИНИСТВЕННЫМ файлом в диффе (02-07-прецедент) — до мержа фазы; затем тег публикует без ручных шагов"

key-files:
  created:
    - README.ru.md
    - docs/ACCEPTANCE.md
  modified:
    - README.md
    - .planning/WINDOWS.md
    - .planning/REQUIREMENTS.md (verb: INST-01 Complete)
    - Публичные артефакты: тег v1.0.0, GitHub Release, main d815e72→80d8534 (bootstrap PR #5)

key-decisions:
  - "ВЛАДЕЛЕЦ решил WINDOWS #5 вариантом (в) 2026-09-16: принять числа ydotool→AT-SPI как есть; #5 закрыт глаголом waive (open_count 0), e2e-perf budget-fail — САНКЦИОНИРОВАННЫЙ результат, не блокирующий гейт; README публикует реальные числа и статус бюджета честно"
  - "release.yml зарегистрирован bootstrap-прецедентом 02-07 (PR #5, squash 80d8534): в диффе ТОЛЬКО файл workflow, byte-идентичный фаза-ветке; мерж фазы в main НЕ выполнялся — остаётся решением владельца на verify-гейте"
  - "STAMPED-OK исполнен по честному эквиваленту: goreleaser по умолчанию штампует .Version БЕЗ 'v' (пин 04-03-SUMMARY: snapshot печатал 0.0.0-SNAPSHOT-<sha>) — проверены 'goswitchd 1.0.0' + полный sha == коммит тега 2bc8c6ade5a0…; grep плана 'v1.0.0' писался до этого пина"
  - "Первый локальный прогон матрицы v3 (22/31) ОБНУЛЁН самим исполнителем: переключение рабочего дерева на bootstrap-ветку посреди живого прогона подсунуло стенду старый focus_helper.py с main (без subcommand focused-input-pid); чистый повтор на фаза-ветке — 31/31"

patterns-established:
  - "Чек-лист приёмки фазы живёт в docs/ACCEPTANCE.md (discretion плана), формат 03-UAT, порядок: свежесессионный D-48 → install → жесты → конфиг → status → uninstall"
  - "Релизная последовательность воспроизводима: гейты локально → bootstrap workflow (если не зарегистрирован) → аннотированный тег → gh run watch → ассеты/checksum/-version/прокси"

requirements-completed: [INST-01]  # verb requirements mark-complete INST-01: оба канала наполнены реальным релизом v1.0.0 (архив + go install через proxy). INST-03 сознательно НЕ тронут: его disposition (waive #5, решение (в)) зафиксирован в WINDOWS/STATE; галку REQUIREMENTS ставит verify-work фазы — прецедент MACR-01 Фазы 3

coverage:
  - id: D1
    description: "Двуязычная пара README: установка одной командой (оба канала), selfcheck/uninstall, perf-таблица из живого прогона С статусом бюджета, приватность -debug, troubleshooting; команды/числа синхронны; голого write-cache нет"
    requirement: INST-01
    verification:
      - kind: other
        ref: "grep-гейт README-COMPLETE (goswitchctl install/selfcheck/uninstall/p95 в обоих языках) — pass"
        status: pass
      - kind: other
        ref: "grep-гейт NO-BARE-CACHE (ни одного 'ibus write-cache' вне #/> строк в обоих файлах) — pass"
        status: pass
      - kind: other
        ref: "grep-гейт синхрона: счётчик 'goswitchctl install' en=4 ru=4 — pass"
        status: pass
      - kind: other
        ref: "числа таблицы == perf-report.txt (прогон 2026-09-16T14:58:14Z: p50 166.9 / p95 184.4 / p99 186.5 мс, VmHWM 12368 кБ = 12.4 МБ); статус бюджета добавлен решением (в): 'превышает бюджет 50 мс — превышение даёт инжектор стенда, не демон (реакция < 2 мс)' в ОБОИХ языках"
        status: pass
    human_judgment: true
    rationale: "Фактическая достаточность и полная синхронность пары (D-50) — flagged verification:judgment плана: грей-зоны формулировок и полнота для читателя проверяются владельцем/верификатором на UAT"
  - id: D2
    description: "docs/ACCEPTANCE.md — чек-лист приёмки владельца D-49 в формате 03-UAT (6 пунктов с expected/result/note + Summary-блок)"
    requirement: INST-01
    verification:
      - kind: other
        ref: "grep-гейт CHECKLIST-OK (expected: / goswitchctl install / uninstall / Summary) — pass"
        status: pass
    human_judgment: true
    rationale: "Чек-лист — артефакт прохода: смысл гейта — его ПРОХОЖДЕНИЕ владельцем в одной сессии (SPEC §7.2) на verify-work фазы; автоматизация здесь по определению неполна"
  - id: D3
    description: "WINDOWS-ledger свёрнут В НОЛЬ глаголами: #2 fixed, #4 waived, #5 waived решением владельца (в); фронматтер: open 0 / waived 3 / fixed 3 / total 6"
    verification:
      - kind: other
        ref: "gsd-tools windows waive 5 '<решение (в) дословно>' → ok; open_count: 0 — гейт LEDGER-CLEAN плана достигнут честно после решения владельца"
        status: pass
    human_judgment: false
  - id: D4
    description: "Релиз v1.0.0 опубликован конвейером 04-03: тег → run 35130309174 (success) → ассеты + checksums + прокси + живой go install"
    requirement: INST-01
    verification:
      - kind: e2e
        ref: "RELEASE-ASSETS-OK: gh release view v1.0.0 → checksums.txt + goswitch_1.0.0_linux_amd64.tar.gz (3.67 МБ, оба бинарника + LICENSE + README-пара внутри)"
        status: pass
      - kind: e2e
        ref: "STAMPED-OK (честный эквивалент, v-стриппинг goreleaser — пин 04-03): скачанный goswitchd -version → 'goswitchd 1.0.0 2bc8c6ade5a092b36e471ec677cceccb848c7f55 2026-09-16T17:47:13Z'; sha == коммит тега v1.0.0"
        status: pass
      - kind: e2e
        ref: "CHECKSUM-OK: sha256sum -c checksums.txt → goswitch_1.0.0_linux_amd64.tar.gz: OK (013f48058a068f1d322c06ab0555eed6b3a2959495e6bf839511b4230c9bf10f)"
        status: pass
      - kind: e2e
        ref: "PROXY-OK (первая попытка, без лага): proxy.golang.org …/@v/v1.0.0.info → {\"Version\":\"v1.0.0\",\"Hash\":\"2bc8c6ade5a0…\",\"Ref\":\"refs/tags/v1.0.0\"}; живой go install cmd/goswitchd@v1.0.0 + cmd/goswitchctl@v1.0.0 успешен, -version честно печатает dev (D-37-оговорка README дословно подтвердилась на реальном канале)"
        status: pass
      - kind: e2e
        ref: "mise run e2e-perf (2026-09-16T14:58:14Z, до чекпоинта): budget FAIL p95 184.4 >= 50 — САНКЦИОНИРОВАНО решением владельца (в), WINDOWS #5 waived; перегон НЕ выполнялся по прямому указанию владельца"
        status: pass
    human_judgment: true
    rationale: "Disposition латентного вердикта — решение владельца (в) 2026-09-16, записано в STATE decisions и причине waive #5"
  - id: D5
    description: "Гейты тега зелёные на HEAD 2bc8c6a (коммит тега): ci, install-cycle, матрица v3 локально"
    verification:
      - kind: e2e
        ref: "mise run ci — GREEN (build+vet+lint+test -race, 2026-09-16 ~17:38 UTC); mise run e2e-install-cycle — PASS (preflight 4/4 + полный жизненный цикл); go run ./test/e2e -matrix matrix-v3.yaml — 31/31 PASS, exit 0 (чистый повтор после инцидента №2)"
        status: pass
    human_judgment: false

# Metrics
duration: 33min   # 16 мин (до чекпоинта) + ~17 мин продолжения после решения владельца; пауза на решение владельца (2.5 ч) не считается
completed: 2026-09-16
status: done
---

# Phase 4 Plan 7: Публикация и приёмка Summary

**Двуязычный README с одно-командной установкой и честной perf-таблицей, письменный чек-лист приёмки, WINDOWS-ledger свёрнут в ноль решением владельца (в) — и ОПУБЛИКОВАН РЕЛИЗ v1.0.0: run 35130309174 success, ассеты + checksums + прокси + go install проверены; INST-01 закрыт**

## Performance

- **Duration:** 33 min active (14:46–15:02 UTC до чекпоинта; 17:36–17:53 UTC продолжение; пауза на решение владельца в счёт не идёт)
- **Started:** 2026-09-16T14:46:40Z
- **Completed:** 2026-09-16T17:53:00Z (публикация + все проверки; финализация доков — сразу после)
- **Tasks:** 3 of 3
- **Files modified:** 4 + публичные артефакты (тег, релиз, main через bootstrap PR #5)

## Accomplishments

- **Двуязычная пара README (D-50)**: README.md (EN, первичен) переписан, README.ru.md создан зеркально — что это и как работает (IBus engine, без root, сосуществование с keyd/xremap), требования, установка ОДНОЙ командой `goswitchctl install` (канал А: релизный архив; канал Б: `go install ...@v1.0.0` с честной оговоркой про `dev`), verify через `goswitchctl selfcheck` (шесть проверок), жесты и CLI, конфиг (вшитые дефолты до первого YAML; схема и hot reload — ссылка на docs/CONFIG.md), perf-таблица с ЧЕСТНЫМ статусом бюджета, приватность, troubleshooting, uninstall `--purge`, MIT. Ни одного ручного шага ibus/systemctl и ни одного упоминания голого `ibus write-cache`; счётчики команд установки синхронны (en=4 ru=4).
- **Статус бюджета в README (после решения владельца, 2bc8c6a)**: оба языка дословно stating «это окно end-to-end **превышает** бюджет реакции 50 мс — превышение даёт собственная инжекция измерительного стенда, а не демон (реакция < 2 мс; бюджет памяти соблюдён). v1.0.0 публикуется с этими числами, как измерены» — решение (в) исполнено в публичном фронте.
- **docs/ACCEPTANCE.md (D-49)** — чек-лист приёмки владельца в формате 03-UAT: 6 нумерованных пунктов с expected/result/note (формальный D-48 свежесессионный двойной прогон по docs/ci-runner.md; установка с нуля по README; живые жесты; hot-reload конфига; счётчики status; uninstall) + Summary-блок; проход — гейт владельца на verify-work фазы.
- **WINDOWS-ledger свёрнут В НОЛЬ глаголами** (Pitfall 11): #2 fixed, #4 waived (решения владельца 2026-09-15), **#5 waived решением владельца (в) 2026-09-16** — причина waive несёт решение дословно. open_count: 0, waived 3 / fixed 3 / total 6 — гейт LEDGER-CLEAN плана достигнут честно (до решения он был честно недостижим — см. Deviations предыдущей редакции).
- **Релиз v1.0.0 опубликован и проверен** (Task 3, полный цикл):
  - Гейты до тега на HEAD 2bc8c6a (коммит тега): `mise run ci` **GREEN**; `mise run e2e-install-cycle` **PASS** (preflight 4/4); матрица v3 локально **31/31 PASS**; `mise run e2e-perf` — budget-fail **санкционирован** решением (в), не гейт.
  - Workflow-регистрация: release.yml НЕ был зарегистрирован (origin/main = d815e72 его не содержал) → bootstrap-прецедент 02-07: ветка от main с ЕДИНСТВЕННЫМ файлом workflow (byte-идентичен фаза-ветке) → PR #5 → squash 80d8534. Мерж фазы НЕ выполнялся.
  - Тег: аннотированный `v1.0.0` → 2bc8c6ade5a092b36e471ec677cceccb848c7f55; push запустил **release run 35130309174** — conclusion **success** (все шаги зелёные; goreleaser 2.18.1 из mise-пина).
  - RELEASE-ASSETS-OK: `checksums.txt` + `goswitch_1.0.0_linux_amd64.tar.gz` (3 665 732 Б; внутри goswitchd + goswitchctl парой, LICENSE, README-пара).
  - STAMPED-OK: скачанный архив распакован; `goswitchd -version` → **`goswitchd 1.0.0 2bc8c6ade5a092b36e471ec677cceccb848c7f55 2026-09-16T17:47:13Z`** — версия = тег (без 'v': дефолтный штамп goreleaser, пин 04-03), sha = точный коммит тега.
  - CHECKSUM-OK: `sha256sum -c checksums.txt` → `goswitch_1.0.0_linux_amd64.tar.gz: OK` (sha256 013f48058a068f1d322c06ab0555eed6b3a2959495e6bf839511b4230c9bf10f).
  - PROXY-OK: `proxy.golang.org/github.com/!djarvur/goswitch/@v/v1.0.0.info` → `{"Version":"v1.0.0","Hash":"2bc8c6ade5a0…","Ref":"refs/tags/v1.0.0"}` — виден с ПЕРВОЙ попытки, без лага.
  - Живой канал go install: `go install …/cmd/goswitchd@v1.0.0` + `…/cmd/goswitchctl@v1.0.0` прошли; бинарник печатает `goswitchd dev` — D-37-оговорка README (dev-версия в этом канале) подтвердилась дословно.
- **INST-01 закрыт** verb'ом requirements mark-complete: оба канала наполнены реальным релизом с проверяемой версией и checksums. INST-03 сознательно не тронут исполнителем — его disposition решён владельцем в #5, галку ставит verify-work (прецедент MACR-01 Фазы 3).

## Task Commits

1. **Task 1: Двуязычный README (D-50/D-46/INST-01)** - `b694e64` (docs); refresh таблицы до свежайшего прогона - `dd645f6`
2. **Task 2: docs/ACCEPTANCE.md (D-49) + WINDOWS-ledger сверка** - `8ac3451` (docs)
3. **Task 3: решение владельца (в) → ledger #5 waive + STATE-решение + статус бюджета в README** - `2bc8c6a` (docs); **тег v1.0.0 → run 35130309174 → релиз опубликован и проверен** (публичные артефакты, вне git-истории фаза-ветки); bootstrap PR #5 (80d8534 на main)

## Files Created/Modified

- `README.md` — публичный фронт EN: установка/selfcheck/жесты/конфиг/perf+статус бюджета/приватность/troubleshooting/uninstall
- `README.ru.md` — русское зеркало, синхрон команд/чисел/версий дословно
- `docs/ACCEPTANCE.md` — чек-лист UAT-гейта фазы (D-49, формат 03-UAT)
- `.planning/WINDOWS.md` — ledger: #2 fixed, #4 waived, #5 waived; open_count 0
- `.planning/REQUIREMENTS.md` — INST-01 → Complete (verb)
- Публичное: тег `v1.0.0`, GitHub Release v1.0.0 (2 ассета), main d815e72 → 80d8534 (только release.yml)

## Decisions Made

- **Решение владельца (в) исполнено дословно**: #5 закрыт глаголом waive (не fixed — дефект не устранён, принят), причина = решение дословно; e2e-perf НЕ перегонялся (прямое указание «do NOT re-run or re-litigate»); README дополнен статусом бюджета в обоих языках (проверка владельца «adjust only if not stated» — статус отсутствовал, добавлен).
- **STAMPED-OK по честному эквиваленту**: план-греп `grep "v1.0.0"` против `-version` не соответствовал зафиксированному в 04-03-SUMMARY поведению goreleaser (`.Version` = тег без 'v'; snapshot-пруф 04-03 печатал `0.0.0-SNAPSHOT-<sha>`). Вместо ослабления грепа до `"1.0.0"` проверено СТРОЖЕ: точный префикс `goswitchd 1.0.0` + полный sha строки == sha коммита тега. Отклонение задокументировано здесь.
- **Bootstrap workflow-регистрации до тега** (прецедент 02-07, санкционирован ответом владельца «используй bootstrap-прецедент, если он нужен»): эмпирическая проверка «заведётся ли workflow с тега без регистрации» не выполнялась тегом-бросайкой — это был бы публичный мусорный релиз; безопасный порядок «сначала регистрация, потом единственный тег» исключает сценарий неудачной публикации и re-tag вовсе.
- **INST-01 closed, INST-03 не тронут**: публикация = последнее отсутствующее условие INST-01 (закрыт verb'ом сразу по проверке прокси); INST-03 (латентность) — disposition зафиксирован в #5/STATE, статусная галка REQUIREMENTS — компетенция verify-work фазы (как MACR-01 в Фазе 3).

## Deviations from Plan

### Auto-fixed Issues / documented departures

**1. [Rule 3 - Documented] Греп STAMPED-OK плана не соответствует фактической прошивке релиза**
- **Found during:** Task 3 (проверка публикации)
- **Issue:** verify плана: `goswitchd -version | grep -q "v1.0.0"`; goreleaser по умолчанию штампует `.Version` БЕЗ 'v' — пин 04-03-SUMMARY (snapshot `0.0.0-SNAPSHOT-<sha>`, «version stamping rides goreleaser DEFAULT ldflags… deliberately NO custom ldflags»). Тег v1.0.0 честно печатает `1.0.0`.
- **Fix:** НЕ менять конфиг goreleaser (ломало бы discretion 04-03 и переопределяло бы принятый пин); проверен эквивалент повышенной строгости: `goswitchd 1.0.0` + sha строки == sha коммита тега. README-канал `@v1.0.0` не затронут (там тег с 'v').
- **Files modified:** нет (конфигурация не менялась)
- **Verification:** STAMPED-OK вывод в Accomplishments; коммит совпадает с тегом
- **Committed in:** этот SUMMARY

**2. [Executor-caused, обнулён повтором] Первый локальный прогон матрицы v3 — 22/31 по вине исполнителя**
- **Found during:** Task 3 (гейты)
- **Issue:** фоновой прогон матрицы был запущен, после чего исполнитель переключил рабочее дерево на bootstrap-ветку (origin/main) для регистрации workflow — стенд продолжил вызывать `test/e2e/focus_helper.py` ИЗ РАБОЧЕГО ДЕРЕВА, а копия main не знает subcommand `focused-input-pid <pid>` (добавлена на фаза-ветках) → 9 кейсов FAIL по гейту focused-input-pid (22/31, exit 1). Регрессии кода НЕТ.
- **Fix:** возврат на фаза-ветку (2bc8c6a), чистый повтор без переключений дерева: **31/31 PASS, exit 0**. Урок: живой e2e-прогон монопольно владеет рабочим деревом.
- **Files modified:** нет
- **Verification:** /tmp/matrix-v3-0407-run2.log: `matrix: 31/31 PASS`, EXIT=0
- **Committed in:** вердикт в этом SUMMARY (артефакт прогона e2e-report.txt — gitignored)

**3. [Sanctioned bootstrap] release.yml добавлен на main PR'ом #5 (squash 80d8534) до тега**
- **Found during:** Task 3 (workflow-регистрация)
- **Issue:** план: «убедиться, что workflow release зарегистрирован от default branch (если фаза ещё не смержена — bootstrap)»; реестр workflows не содержал release.yml (проверено gh api), а заголовок release.yml сам фиксирует «до мержа фазы тег не публикуется».
- **Fix:** ветка от origin/main, дифф = ТОЛЬКО `.github/workflows/release.yml` (byte-идентичен фаза-ветке), PR #5 → squash. Никакого кода и planning-доков фазы на main не попало; мерж фазы остаётся решением владельца.
- **Files modified:** .github/workflows/release.yml на main (через PR #5)
- **Verification:** реестр workflows показывает release.yml; tag-push v1.0.0 запустил run 35130309174
- **Committed in:** 80d8534 (main)

---

**Total deviations:** 3 (1 задокументированная адаптация грепа; 1 исполнительная ошибка, обнулённая чистым повтором; 1 санкционированный планом bootstrap). **Impact on plan:** нет — все verify-гейты Task 3 исполнены (RELEASE-ASSETS-OK / STAMPED-OK / PROXY-OK), публичных мусорных артефактов нет, re-tag не понадобился (v1.0.0 опубликован ровно один раз).

## Issues Encountered

- Красный perf-гейт (p95 184.4 мс ≥ 50 мс; третий подряд) — **решён владельцем вариантом (в)**: принять как есть; #5 waived, таблица README с честным статусом. Полная доказательная база — в предыдущей редакции SUMMARY и WINDOWS #5.
- Известный средовой флейк combo-word-layout (deferred-items) не воспроизвёлся ни в одном из прогонов.
- Node.js 20 deprecation annotation от actions/checkout@v4 в run 35130309174 — предупреждение GitHub, не сбой; кандидат в рутинное обновление dependabot.

## Known Stubs

None.

## User Setup Required

None — публикация завершена; проход docs/ACCEPTANCE.md — шаг приёмки владельца на verify-work фазы, не настройка.

## Next Phase Readiness

- План завершён полностью: доки-пара, чек-лист, ledger open_count 0, релиз v1.0.0 опубликован и проверен по всем трём verify-гейтам плана.
- Осталось на фазу 4 (вне этого плана): verify-work — владелец проходит docs/ACCEPTANCE.md (одна сессия §7.2, включая формальный D-48 свежесессионный прогон по docs/ci-runner.md), verifier закрывает REQUIREMENTS (INST-03 — с оговоркой решения (в)), мерж фазы в main и закрытие milestone v1.0.0 — решения владельца.

---
*Phase: 04-postavka-i-priemka*
*Completed: 2026-09-16*

## Self-Check: PASSED

- Files on disk: README.md, README.ru.md, docs/ACCEPTANCE.md, .planning/WINDOWS.md (open_count: 0), .planning/REQUIREMENTS.md (INST-01 [x]) — all FOUND
- Commits: b694e64, 8ac3451, dd645f6, b1f0aaa, b8e969d, 2bc8c6a — all present in git log on gsd/phase-04-postavka-i-priemka; 80d8534 on main (PR #5)
- Release: tag v1.0.0 == 2bc8c6ade5a092b36e471ec677cceccb848c7f55; run 35130309174 success; assets checksums.txt + goswitch_1.0.0_linux_amd64.tar.gz (sha256 013f48058a068f1d322c06ab0555eed6b3a2959495e6bf839511b4230c9bf10f); -version: `goswitchd 1.0.0 2bc8c6ade5a0… 2026-09-16T17:47:13Z`; PROXY v1.0.0.info OK
- Gates at tag (HEAD 2bc8c6a): ci GREEN, e2e-install-cycle PASS, matrix v3 31/31 PASS, e2e-perf budget-exceeded verdict sanctioned by owner decision (в)
