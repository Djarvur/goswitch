# Phase 4: Поставка и приёмка - Context

**Gathered:** 2026-09-15
**Status:** Ready for planning

<domain>
## Phase Boundary

Продукт доезжает до пользователя и принимается владельцем — релиз v1 (веха M4 SPEC §9). Установка с нуля без root на чистой Ubuntu 24.04 через `goswitchctl install` (component XML под `$HOME` через `IBUS_COMPONENT_PATH` + `ibus write-cache`, systemd user unit, автоматический «один хозяин»), рабочий сквозной self-check и полный uninstall; сборка и публикация релиза goreleaser'ом с прошитой версией (`go install` + GitHub Releases — оба канала INST-01); подтверждённый измерением бюджет производительности (p95 < 50 мс выделенным perf-прогоном ydotool→AT-SPI, память < 50 МБ по /proc VmHWM, числа — таблицей в README); полная e2e-матрица v3 (v2 + gedit + оконные режимы Chromium) зелёная дважды подряд на нетронутой сессии раннера; ручная приёмка владельца по письменному чек-листу как UAT-гейт; двуязычный README (EN+RU) с инструкцией установки. Новой функциональности поведения нет — весь жестовый словарь закрыт Фазами 2–3; терминальный класс и OSD остаются во v2.

</domain>

<decisions>
## Implementation Decisions

### Релиз v1: артефакты и публикация
- **D-37:** Версия прошивается в бинарник на release-сборках (`-ldflags -X`: тег + commit); `goswitchd --version` и статус в `goswitchctl` показывают точную сборку — база разбора баг-репортов публичного инструмента. Сборки через `go install` идут без прошивки и показывают фолбэк (dev/unknown). — **Reversibility:** reversible — механика stamping'а
- **D-38:** Сборку и публикацию релиза ведёт goreleaser (`.goreleaser.yaml`; CI-инструмент — НЕ зависимость демона, рантайм-принцип «stdlib + godbus + минимум» не затронут). Каналы: GitHub Releases-артефакты + `go install github.com/Djarvur/goswitch/cmd/goswitchd@vX.Y.Z`. — **Reversibility:** reversible — конфиг инструмента

### Установщик и uninstall (INST-01)
- **D-39:** Установку выполняют подкоманды `goswitchctl install / uninstall / selfcheck` — не bash-скрипт и не README-шаги. install идемпотентен: пишет component XML в `~/.config/ibus/component/` (регистрация через `IBUS_COMPONENT_PATH`), systemd user unit (`~/.config/systemd/user/`), гоняет `ibus write-cache`, `systemctl --user daemon-reload` + enable --now. Тестируется Go-корпусом и e2e (принцип «полный цикл без человека»). — **Reversibility:** costly — публичный CLI-контракт инструмента
- **D-40:** install автоматически приводит стол к «одному хозяину» (D-03/ADR-001): gsettings input-sources = только goswitch, прежние источники сохраняются для восстановления при uninstall. После установки всё работает сразу, без ручных шагов. — **Reversibility:** costly — UX-контракт установки и модель ADR-001
- **D-41:** selfcheck — сквозной живой: версия бинарника → компонент виден ibus (`ibus list-engine`/registry) → user unit active → движок реально зарегистрирован (ListActiveEngines) → конфиг валиден → источник ввода goswitch. Каждый пункт зелёный/красный с подсказкой фикса — fail-fast стиль e2e-preflight. — **Reversibility:** reversible
- **D-42:** uninstall — полный откат: stop/disable юнита, удаление юнита + component XML + ibus write-cache, восстановление сохранённых источников ввода. Пользовательский `~/.config/goswitch/` НЕ трогается по умолчанию (наработки пользователя); флаг `--purge` удаляет и его. — **Reversibility:** reversible

### Perf-приёмка (INST-03)
- **D-43:** Статистика латентности — выделенный perf-прогон (mise-задача/e2e-режим): N повторов горячего кейса (коррекция слова + переключение раскладки), отчёт p50/p95/p99. Выборка однородная, не размазана по разнородным кейсам матрицы. — **Reversibility:** reversible
- **D-44:** «Реакция на горячую клавишу» = только e2e-окно ydotool→AT-SPI (дословно критерий №3 роадмапа). Интервал опроса AT-SPI-свидетеля фиксируется методикой — число честной верхней оценкой, воспроизводимо. Внутренний tap→commit-лог в приёмку НЕ входит. — **Reversibility:** reversible
- **D-45:** Память: оракул — `/proc/<pid>/status` VmHWM (пик RSS) против бюджета < 50 МБ; VmRSS снимается в контрольных точках (после старта, конец прогона). Ноль средовых зависимостей (cgroup-учёт user-сессий не нужен). — **Reversibility:** reversible
- **D-46:** Числа perf-измерений публикуются таблицей в README (p50/p95/p99 + RSS peak); актуализация таблицы — часть релизного цикла. — **Reversibility:** reversible

### Финальная матрица и приёмка (критерии №2/№4)
- **D-47:** Расширение покрытия — matrix v3 (новый YAML по образцу v1→v2): v2 остаётся замороженным принятым эталоном Фазы 3; v3 добавляет поверхность gedit и оконные режимы Chromium. Состав кейсов — планировщик по результатам исследования поверхностей. — **Reversibility:** reversible
- **D-48:** «Матрица зелёная дважды подряд на нетронутой сессии» = свежеподнятая GNOME-сессия раннера → прогон v3 №1 → прогон v3 №2 сразу, без пересоздания сессии и ручного вмешательства между прогонами (проверяет, что первый прогон не портит состояние для второго); любое падение — exit ≠ 0. — **Reversibility:** reversible
- **D-49:** Ручная приёмка владельца — письменный чек-лист в repo (установка с нуля по README на живом столе → живой набор всех жестов → uninstall), проходится как UAT-гейт verify-work фазы — та же механика, что приняла Фазы 1–3 (одна сессия по §7.2 спеки). — **Reversibility:** reversible
- **D-50:** README и инструкция установки двуязычные: `README.md` (EN) + `README.ru.md` (RU), синхронизация обоих — часть релизного цикла. — **Reversibility:** costly — опубликованный контент; рассинхрон = врущие доки

### Claude's Discretion
- Формат релизных ассетов (tar.gz vs голые бинарники + checksums), список архитектур (amd64-only vs +arm64), схема тегов и release notes — планировщик в связке с goreleaser-конфигом (решение владельца: «на усмотрение»)
- Точный N повторов и состав горячего кейса perf-прогона — пинируется методикой при реализации D-43
- Число и состав кейсов matrix v3 — после исследования gedit-поверхности (её IM-путь/класс виджета) и оконных режимов Chromium
- Дефолтный YAML-конфиг при первой установке (генерировать ли `~/.config/goswitch/config.yaml` при install, или демон работает на вшитых дефолтах до первого конфига) — в рамках принципов D-31..D-33
- Расположение чек-листа приёмки (отдельный docs/файл vs раздел README) и его точная структура
- Порядок задач, разбиение на планы, детали goreleaser/workflow-интеграции

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Спецификация и требования
- `docs/SPEC.md` §5 — NFR: waitless-действия < 50 мс, действия Right Shift ≤ окно различения; §7.2 — ручная приёмка (сессия владельца); §7 — тестирование; §9 — веха M4 (релиз v1)
- `.planning/REQUIREMENTS.md` — требования Фазы 4: INST-01 (установка без root: systemd user unit, IBus-регистрация, `go install` + GitHub Releases), INST-03 (< 50 мс / < 50 МБ)
- `.planning/ROADMAP.md` §Phase 4 — цель и 4 критерия успеха (установка с self-check/uninstall; матрица дважды; perf-измерение; приёмка + опубликованные доки)

### ADR и архитектура установки
- `docs/adr/ADR-001-layout-switching-mechanism.md` — «один хозяин»: install обязан ставить goswitch единственным источником (D-40), uninstall — восстанавливать прежние
- `.planning/research/STACK.md` — systemd user unit (цепочка `graphical-session.target`, `Restart=on-abnormal`), `IBUS_COMPONENT_PATH`, `ibus write-cache`, требования к e2e-стенду; раздел Installation — образец user unit
- `.planning/phases/01-adr-paket-i-karkas-ibus-dvizhka/01-CONTEXT.md` — D-07..D-11 (TDD/mise/lint/CI — BIND), T-01-04 (runtime-регистрация Фазы 1; XML-инсталляция — эта фаза)

### Контексты предыдущих фаз
- `.planning/phases/02-korrektsiya-slova-en-ru/02-CONTEXT.md` — D-17..D-19 (матрица v1, self-hosted раннер), D-20/D-21 (приватность логов)
- `.planning/phases/03-frazy-vydelenie-pereklyuchenie-i-konfiguratsiya/03-CONTEXT.md` — D-31..D-33 (схема конфига — публичный контракт, который install не должен ломать), D-34/D-35 (BIND-решения переключения), e2e-стенд v2

### Существующие доки и инфраструктура
- `docs/CONFIG.md` — публичный конфиг-контракт (таблица соответствия Caramba) — источник для README-инструкции
- `docs/ci-runner.md` — self-hosted GNOME-раннер green106 (токены, dispatch) — двойной прогон v3 гоняется там
- `README.md` — текущий минимальный (2 КБ) — будет переписан двуязычно (D-50)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `cmd/goswitchctl/main.go` — CLI уже жив (status/reload/correct, канон one-line key=value 03-06); подкоманды install/uninstall/selfcheck добавляются сюда
- `internal/ctlsvc` — D-Bus сервис демона с семантическим single-instance guard (RequestName ReplaceExisting без AllowReplacement) — selfcheck может переиспользовать диалог со сервисом
- `engine/engine.go` — ListActiveEngines-подобная live-проверка регистрации уже доказана Фазой 1 (live-gate) — база для selfcheck-шага «движок реально зарегистрирован»
- `test/e2e/` — стенд: `matrix.go`/`matrix_test.go` (YAML-кейсы v1/v2), `surface.go` (zenity/GTE/Chromium драйверы), `preflight.go` (fail-fast диагностика — образец для selfcheck), snapshot/restore gsettings (Фаза 1) — механика для install/uninstall-тестов и матрицы v3
- `mise.toml` — полный набор e2e-задач; perf-прогон (D-43) и matrix-v3 добавляются по образцу
- `.github/workflows/e2e-matrix.yml` — CI-контур e2e на green106; двойной прогон (D-48) надстраивается здесь
- `go.mod` (`go 1.23`, три прямых зависимости) — `go install`-канал уже совместим, ничего менять не нужно

### Established Patterns
- Строгий TDD + гейты итерации `go build` / `go test -race` / `golangci-lint` (D-07/D-08); mise — единственный интерфейс сценариев (D-09)
- Приватность логов: INFO — счётчики/причины, детали — `-debug` (D-20/D-21) — spread на install/selfcheck-вывод
- Fail-fast диагностика с подсказкой фикса (e2e preflight) — паттерн для selfcheck (D-41)
- Snapshot/restore gsettings вокруг живых прогонов — паттерн для install/uninstall e2e и для восстановления источников (D-40/D-42)

### Integration Points
- `goswitchctl install` → `~/.config/ibus/component/goswitch.xml` + `~/.config/systemd/user/goswitchd.service` + `ibus write-cache` + `systemctl --user` + gsettings input-sources (сохранение прежних)
- goreleaser → GitHub Releases (ассеты + checksums + прошивка версии D-37); тег v* триггерит
- perf-прогон → README-таблица (D-46); e2e-workflow green106 → двойной прогон v3 (D-48)
- UAT-гейт verify-work фазы → чек-лист приёмки (D-49)

</code_context>

<specifics>
## Specific Ideas

- Прежние источники ввода: install сохраняет, uninstall восстанавливает (дословная пара из обсуждения)
- README-таблица perf: p50 / p95 / p99 латентности + RSS peak — обновляется каждым релизом
- Двуязычные доки: `README.md` (EN) первичен + `README.ru.md` (RU) в синхроне
- Владелец просил пояснять незнакомые термины без жаргона (из 01-CONTEXT, действует)

</specifics>

<deferred>
## Deferred Ideas

- Терминальный класс (clipboard-уровень для терминалов, SPEC §6 Q3 / TERM-01) — остаётся во v2 (перенос из Фаз 2–3, во Фазу 4 не входит)
- OSD-уведомление при переключении раскладки (компенсация индикатора ADR-001) — вне объёма, остаётся во v2

</deferred>

---

*Phase: 4-Поставка и приёмка*
*Context gathered: 2026-09-15*
