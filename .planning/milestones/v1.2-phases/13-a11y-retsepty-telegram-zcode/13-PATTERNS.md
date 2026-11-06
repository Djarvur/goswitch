# Фаза 13: A11Y — рецепты Telegram/ZCode — Карта паттернов

**Составлено:** 2026-11-03
**Файлов проанализировано:** 8 (1 новый скрипт, 4 правки доков, 3 условных e2e-файла — Go-case-носитель пробных рядов)
**Аналоги найдены:** 8 / 8
**База кода:** main 00ddfa9 (post-Phase-12); номера строк пинены по фактическому состоянию файлов 2026-11-03
**Граница фазы:** демон goswitchd — НОЛЬ строк правок (решение владельца 2026-10-08): `internal/*`, `cmd/*` только читаются как оракулы; весь новый код — один python3-gi-скрипт + тексты + (условно) e2e-кейсы стенда

## Классификация файлов

| Файл (новый/правимый) | Роль | Поток данных | Ближайший аналог | Качество |
|------------------------|------|--------------|------------------|----------|
| `scripts/a11y-selfcheck.py` (новый, D-06) | utility (внешний AT-SPI-зонд) | request-response (D-Bus GetRole-раунды) | `test/e2e/focus_helper.py` | exact |
| `README.md` (правка, D-07) | docs | — | сам: «Где автокоррекция молчит» :234–260 + «A11y-магия приложений» :307–341 | exact |
| `README.en.md` (правка, D-07) | docs | — | зеркало README.md: :228–254 + :296–331 | exact |
| `docs/CONFIG.md` (правка, D-07) | docs | — | сам: «Ручной рецепт: per-app запуск с a11y-флагами» :373–409 | exact |
| `docs/adr/adr-008-spike-log.md` (правка) | docs (журнал пробы) | — | сам: venue-ряды :253–263, фазовые секции :265+, :678+ | exact |
| `test/e2e/case_a11y_*.go` (новый, условно — см. «Развилка носителя») | test (e2e live-ряды) | event-driven (probe) | `test/e2e/case_spike_wezterm.go` (appendSpikeRow :246–251) + `case_term_word_vte.go` | exact |
| `test/e2e/main.go` (правка, условно) | config (диспетчер кейсов) | — | сам: реестр кейсов ~:395–435 | exact |
| `mise.toml` (правка, условно) | config (задачи) | — | сам: секция e2e-term-* :175–204 | exact |

**Неприкасаемые по построению:** весь `internal/` (наблюдатель `internal/appid`, статус `internal/ctlsvc`, актор `internal/session`) и `cmd/goswitchd`/`cmd/goswitchctl` — только потребляются как машинные оракулы пробы. Любая задача, требующая правки демона, — эскалация владельцу (демон-fallback ветка A11Y-02), не тихий фикс.

## Назначения паттернов

### `scripts/a11y-selfcheck.py` (utility, request-response) — новый

**Аналог:** `test/e2e/focus_helper.py` целиком (661 строка, прочитана живьём). Это тот же класс артефакта: одиночный python3-gi-файл над системной gir, ноль зависимостей, контракт в докстринге, коды возврата как единственный машинный интерфейс.

**Шебанг + докстринг-контракт** (:1–2, :106–110 — форма первого экрана скрипта):
```python
#!/usr/bin/python3
"""AT-SPI helper for the goswitch e2e stand (TEST-03.
...
Every caller must use /usr/bin/python3: PATH python3 is linuxbrew without
gi (01-PATTERNS § test/e2e).

Exit codes: 0 success, 1 named failure, 2 usage error.
"""
```
`/usr/bin/python3` — НЕ факультативно: PATH-python3 на машинах владельца — linuxbrew без gi (ловушка закреплена докстрингом и каждым Go-вызывателем). Дословная строка запуска в доках (D-06) — `/usr/bin/python3 scripts/a11y-selfcheck.py <app>`.

**Диспетчер + коды возврата** (:619–656, форма main — проблема как строка, не исключение):
```python
def main(argv):
    if len(argv) == 2 and argv[1] == "witness":
        problem = cmd_witness()
    elif len(argv) == 3 and argv[1] == "text":
        problem = cmd_text(argv[2])
    ...
    else:
        print("usage: focus_helper.py witness | text <app> | ...", file=sys.stderr)
        return 2
    if problem is not None:
        print(f"focus_helper: {problem}", file=sys.stderr)
        return 1
    return 0
```
Каждая `cmd_*` возвращает `None` при успехе (ответ уже напечатан в stdout) либо строку-проблему (→ stderr + exit 1). Контракт D-06 из RESEARCH Pattern 2: **0** — приложение в живом дереве, текст-поле видно (роль ∈ {61, 79, 94}); **1** — приложение запущено, но дерева/поля нет (a11y не включён / не перезапущен); **2** — приложение не найдено/не запущено/ошибка использования. Человекочитаемый вердикт — в stdout, код — машине.

**Перечисление desktop-приложений + обход дерева** (:135–141, :144–162 — копировать как есть, ловушки уже закодированы):
```python
def applications():
    """Yield every AT-SPI application node on the desktop."""
    desktop = Atspi.get_desktop(0)
    for i in range(desktop.get_child_count()):
        app = desktop.get_child_at_index(i)
        if app is not None:
            yield app


def walk(app):
    """Breadth-first yield every node of one application tree.
    ...
    Every step is individually best-effort — a
    transient node error skips that node instead of killing the witness.
    """
    stack = [app]
    while stack:
        node = stack.pop(0)
        if node is None:
            continue
        yield node
        try:
            children = [node.get_child_at_index(k) for k in range(node.get_child_count())]
        except Exception:
            continue
        stack.extend(children)
```
Шина a11y отстаёт композитора: ноды исчезают на ходу, всякий доступ к мёртвой ноде кидает — поштучный try/except, никогда общий.

**КЛЮЧЕВОЕ РАСХОЖДЕНИЕ с аналогом — роли по НОМЕРУ, не по имени.** focus_helper сравнивает роли по именам (:126, :129–132):
```python
INPUT_ROLES = ("TEXT", "PASSWORD_TEXT", "ENTRY")

def role_name(node):
    """Return the stable role label (TEXT, PASSWORD_TEXT, ...)."""
    match = re.match(r".*ATSPI_ROLE_(\w+)", str(node.get_role()))
    return match.group(1) if match else str(node.get_role())
```
Гейт демона сравнивает по номеру — `internal/appid/appid.go:85–96` (прочитан живьём, комментарий дословно запрещает имена):
```go
// The AT-SPI role enum values the autocorrect policy gates on (gi Atspi
// 2.52.0, live-verified on the target GNOME 46 desktop 2026-09-27 —
// 06-RESEARCH Q4). Roles compare by NUMBER only: role-name strings are
// localized and diverge between toolkits for the same enum value (61 is
// "text box" over the wire and "text" in the GI registry).
const (
	RolePasswordText uint32 = 40 // password text — the autocorrect policy never corrects it
	RoleTerminal     uint32 = 60 // terminal — never corrected
	RoleText         uint32 = 61 // text box / text — correctable text input
	RoleEntry        uint32 = 79 // entry — correctable text input
	RoleDocumentText uint32 = 94 // document text — correctable text input
)
```
Self-check выносит вердикт по `{61, 79, 94}` (номер из `node.get_role()`), роль 40 печатает как «видимо, но неисправимо», все прочие номера печатает БЕЗ предположения (какую роль несёт поле Qt-Telegram и Electron-ZCode — решает проба A4, не скрипт). Имена — только для человекочитаемого вывода. Отдельная дыра аналога: `INPUT_ROLES` focus_helper не содержит DOCUMENT_TEXT (94) — текст-класс демона шире; переносить список имён в self-check нельзя.

**Readback текста (если зонд печатает chars=N) — три серверных причуды** (:261–287, копировать порядок дословно): GTK4-мост отвечает только современным `get_string_at_offset`, Chromium — только депрекированным `get_text_at_offset`, сигнатура `get_text` сломана в atspi 2.52 (последний резерв):
```python
    try:
        span = node.get_string_at_offset(0, Atspi.TextGranularity.LINE)
        return span.content
    except Exception:
        pass
    try:
        span = node.get_text_at_offset(0, Atspi.TextBoundaryType.LINE_START)
        return span.content
    except Exception:
        pass
```

**Коллизия имени:** у `goswitchctl` УЖЕ есть субкоманда `selfcheck` (валидация конфига: `cmd/goswitchctl/main.go:56` usage-строка `... | install | uninstall [--purge] | selfcheck`, диспетч `:269`). Имя скрипта `a11y-selfcheck.py` не конфликтует, но строка запуска в доках обязана нести полный путь `/usr/bin/python3 scripts/a11y-selfcheck.py` — никогда не «goswitchctl selfcheck» и не голый `a11y-selfcheck` из PATH.

**Если нужна .sh-обёртка** (RESEARCH допускает): единственный житель `scripts/` — `d48-nightly-dispatch.sh`; форма: `#!/usr/bin/env bash` + шапка-комментарий «что/зачем/чьё решение» (:1–10), `set -euo pipefail` (:12), именованные ошибки в stderr вместо трассировок (:21–24), `exit 1` как отказ.

### `README.md` (docs) — правка, краткие рецепты + residual-risk

**Аналог:** сам файл, раздел «Где автокоррекция молчит — и как это починить» :234–260 — целевая точка K-place УЖЕ существует и УЖЕ называет оба приложения (:243):
```markdown
Известные примеры: ZCode, snap Telegram. Чтобы
автокоррекция заработала в таком приложении, включите toolkit
accessibility глобально и перезапустите приложение:
```
Сюда ложатся краткие прикладные рецепты (D-07): ZCode — user-local `.desktop` override (шаблон целиком в docs/CONFIG.md:385–392, в README — одна строка сути + ссылка); Telegram — выбор сборки по вердикту пробы (snap vs flatpak, результат D-01/D-02); residual-risk заметка «ручная коррекция на скрытых полях копирует в PRIMARY» (D-07). Вторая точка — «A11y-магия приложений» :307–341: сюда уточнение «что магия НЕ делает» (глобальный ключ ≠ флаг конкретного Electron-приложения; ссылка на рецепты и self-check). Семантика «только при старте» уже описана :329–331 — рецепт обязан ссылаться, не дублировать.

### `README.en.md` (docs) — зеркальная правка

**Аналог:** собственное зеркало README.md: «Where autocorrect stays silent — and how to fix it» :228–254 (те же ZCode/the Telegram snap :237, тот же gsettings-рецепт :241–243) и «App accessibility magic (a11y)» :296–331. Дисциплина: оба README правятся в одном PR байт-параллельно; grep-гейты A11Y-03 проверяют ОБА (Validation Map, RESEARCH). Замечание-предостережение: зеркало ЧУТЬ ОТСТАЁТ от RU (например, :261–262 ещё описывает canberra-gtk-play, тогда как RU :269–270 — прямое соединение со звуковым сервером) — новые абзацы вносить синхронно, не трогая застарелое чужое расхождение (отдельная уборка вне фазы).

### `docs/CONFIG.md` (docs) — детали, побочки, границы

**Аналог:** сам файл, раздел «Ручной рецепт: per-app запуск с a11y-флагами» :373–409. Шаблон ZCode-override уже в репо ДОСЛОВНО (:383–392) — тот самый, что RESEARCH цитирует как основу:
```ini
Фиксированный пример шаблона (флаги — литералы):

```ini
[Desktop Entry]
Type=Application
Name=ZCode
Exec=env ACCESSIBILITY_ENABLED=1 /opt/ZCode/zcode --force-renderer-accessibility %U
Icon=zcode
StartupWMClass=ZCode
```
```
Расширение секции: (1) живая валидация шаблона после пробы ZCode на ВМ (A11Y-01); (2) рецепт Telegram-сборки по вердикту пробы; (3) побочка «плашка режима экранного диктора» (Pitfall 4, A5 — форму фиксирует проба); (4) граница «флаги Electron применяются только на запуске» — существующая клауза :280–285 уже несёт семантику для gsettings-ключа, рецепт флага ссылается на неё + требование полного выхода из процесса (трей жив — Pitfall 3); (5) строка запуска self-check (D-06). Форма предупреждений — нумерованный список :394–405 (замороженный override маскирует обновления; «snap-приложения могут требовать иных путей и форм override» :405 — прямо расширяется рядом про flatpak по итогам пробы). Смежные точки: «Секция a11y: применение и откат» :255–295 и клауза «Где слой молчит» :356–371 (роль 40/61-семантика) — согласовать формулировки, не дублировать.

### `docs/adr/adr-008-spike-log.md` (docs, журнал пробы) — новая секция Фазы 13

**Аналог:** сам журнал. Форма фазовой секции + venue-таблицы (:253–263):
```markdown
## 2026-10-12 ~13:42-14:20 UTC — Wave-0 площадка Фазы 10 (план 10-01): ...

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| venue-ready 13:49 | `stand-env (/tmp/stand-env.sh: VM_ADDR=192.168.122.150, ...) + rsync + unit -debug` | ... | Verdict: стенд готов ... |
| venue-probe r1 2026-10-12 14:12 | `armed: BackSpace x3 ... → ydotool type --key-delay 40 $'ghbdtn\n'` | ... | Verdict: keyboard-live (round 1 из 2 подряд) |
| venue-verdict 2026-10-12 14:20 | `вердикт площадки по venue-probe r1 + r2 (два подряд зелёных раунда — дисциплина ... SPEC §7.2 ...)` | ... | Verdict: vm-primary |
```
Обязательные ряды фазы (по RESEARCH Pattern 1/Pitfalls): (а) КАЛИБРОВОЧНЫЙ ряд первым — известный GTK-апп (gnome-terminal) обязан показать живое дерево+роли; глухой контрольный = вердиктов по сборкам НЕТ (Pitfall 2); (б) по каждому рунгу (snap 7.3.3, flatpak, ZCode) — ТРИ слоя в одном ряду: key-trace флипа / `GetRole`-номер роли сфокусированного поля / дельты счётчиков `goswitchctl status`; (в) ×2 зелёных подряд на вердиктный ряд (r1/r2-форма :261–262); (г) механика-ряд находок установки (форма venue-mechanics :260 — сети ВМ, стейджинг tarball, плашка диктора на `virsh screenshot`); (д) откаты стенда в вердикт-ряду дословно (форма :263: «Откаты стенда задокументированы: sources [("xkb","us")], lock-enabled=true/idle-delay=300» + откаты рунгов: snap remove / flatpak uninstall / rm -r /opt/ZCode).

### `test/e2e/case_a11y_*.go` + `test/e2e/main.go` + `mise.toml` (условная тройка)

**Развилка носителя пробных рядов** (решение планировщика): (i) живые ряды через Go e2e-кейс на зеркале ВМ — прецедент Фазы 10 (повторяемость, арминг-гейты, watchdog); (ii) ssh-ручные ряды — прецедент venue-строк :253–263 (план 10-01 исполнял их руками оркестратора). Для ×2-дисциплины и grep-гейтов повторяемость кейса предпочтительна; установка snap/flatpak — всё равно внешние ssh-команды стенда (кейс начинает работу после готовности рунга).

**Аналог кейса:** `test/e2e/case_spike_wezterm.go` — шов записи в журнал (:246–251):
```go
func appendSpikeRow(section, probe, cmd, observed, verdict string) error {
	return appendSpikeRowTo(spikeJournal, section, probe, cmd, observed, verdict)
}
```
плюс датированный ряд-хелпер recipeChordRow того же файла (цитирован 12-PATTERNS :774–796: сентинел → действие → settle → оракул → `appendSpikeRow(secRecipe, ...)`) и счётчиковый оракул `case_term_word_vte.go` (`goswitchctl status` дельты fired/abstained). Вызов python-зонда из Go — каноническая форма `runCmd(ctx, "/usr/bin/python3", s.cfg.helper, "focused-input-pid", pidStr)` (`case_autocorrect.go:259`, также `case_word.go:303`, `matrix.go:802`); self-check-скрипт вызывается той же формой с полным путём `scripts/a11y-selfcheck.py`.

**Аналог регистрации:** реестр `test/e2e/main.go` ~:395–435, форма стендовых кейсов (проверено живьём):
```go
		"term-word-vte":      {fn: runTermWordVTE, noSurfacePreflight: true, watchdog: recipeWatchdog},
		"term-word-wezterm":  {fn: runTermWordWezterm, noSurfacePreflight: true, watchdog: recipeWatchdog},
```
Новые кейсы — `noSurfacePreflight: true` (chromium/gte/gedit-префлайт не нужен: кейсы водят Telegram/ZCode, не стендовые фикстуры — философия 10-05/11-05, комментарий в реестре) + `watchdog: recipeWatchdog`; плюс usage-строки.

**Аналог mise-задачи:** `mise.toml` :175–204, форма (:178–180):
```toml
# Phase-10 live VTE rows (10-05, TERM-02): they drive the goswitch-stand VM
# desktop (vm-primary venue) and append dated verdict rows to
# docs/adr/adr-008-spike-log.md — deliberately NOT part of `mise run ci`.
[tasks.e2e-term-word-vte]
description = "e2e: 10-05 term-word-vte — live VTE word rows on the VM stand: ... rows to adr-008-spike-log.md (live GNOME session; NOT in ci)"
run = "go run ./test/e2e -case term-word-vte"
```
Новые задачи — НЕ in ci (шапка-комментарий секции повторить дословно), исполнение НА ВМ через зеркало (директива 6).

## Общие паттерны

### `/usr/bin/python3` — жёсткий интерпретатор
**Источник:** `test/e2e/focus_helper.py:106–107` (докстринг) + каждый Go-вызыватель (`case_autocorrect.go:259`, `case_word.go:303`, `case_m1.go:302`, `matrix.go:802`)
**Применять ко:** скрипту (шебанг), строке запуска в доках, любому вызову из Go-кейса — PATH-python3 это linuxbrew без gi.

### Коды возврата 0/1/2 + проблема-строка
**Источник:** `test/e2e/focus_helper.py:109` (контракт), :619–656 (main-диспетчер)
**Применять ко:** self-check — контракт D-06: 0 видно / 1 не видно / 2 нет приложения или usage; человекочитаемый вердикт в stdout, никогда exit-код вместо текста.

### Роли по НОМЕРУ, текст-класс {61, 79, 94}
**Источник:** `internal/appid/appid.go:85–96` (гейт демона; имена локализуемы и расходятся)
**Применять ко:** вердикту self-check и строкам таблицы пробы — номер роли всегда в ряду (Pitfall 8: «в ряде нет числа роли» = warning sign); focus_helper-форма `role_name` — только для печати.

### Счётчики — единственный оракул слоя 3; текст пользователя — никогда
**Источник:** `internal/ctlsvc/ctlsvc.go:255–296` (renderStatus: `autocorrect_fired`, `autocorrect_abstained`, `ac_role_errors`, `ac_skip_<slug>`; комментарий :250–254 «Counts and states only — never user text»)
**Применять ко:** пробе слоя 3 и выводу self-check — дельты до/после синтетического корпуса `ghbdtn`-класса; имена приложений печатать как данные, не интерполировать в команды (ASVS V5, RESEARCH Security Domain).

### Probe-first: калибровка → три слоя → ×2 зелёных → датированные ряды → откаты
**Источник:** `docs/adr/adr-008-spike-log.md:253–263` (venue-протокол Фазы 10 на этом же стенде)
**Применять ко:** каждому рунгу D-02/D-03 и ZCode-ряду; слова владельца — повод для пробы, не вердикт (D-01); единичный зелёный ряд — не вердикт (SPEC §7.2).

### K-place доков: расширять существующие разделы, не создавать страницы
**Источник:** README.md :239–255 + :307+, README.en.md :233–249 + :296+, docs/CONFIG.md :373+ (карта снята RESEARCH живьём)
**Применять ко:** всем док-правкам; оба README синхронно в одном PR; grep-гейты A11Y-03 (побочка-диктор, только-на-запуске, restart+self-check, residual-risk PRIMARY) прогоняются по трём файлам.

### Проверки — только на ВМ goswitch-stand
**Источник:** AGENTS.md директива 6 (2026-11-01); D-05; прецедент-инцидент UAF atk-bridge 2026-10-31
**Применять ко:** ВСЕМ живым прогонам (зонд, self-check-демонстрация, установка рунгов); host — только пассивные чтения и стейджинг скачки; RED-форма self-check — падающий прогон против заведомо невидимого приложения НА ВМ (headless-RED невозможен, прецедент focus_helper :102–104).

## Нет аналога

| Файл | Роль | Поток данных | Причина |
|------|------|--------------|---------|
| — | — | — | Нет: единственный новый код (self-check) имеет точного предшественника (focus_helper), доки-правки ложатся в существующие разделы, установка snap/flatpak-рунгов — процессный аналог venue-mechanics-ряда (adr-008 :260), не код. |

## Метаданные

**Область поиска аналогов:** `scripts/`, `test/e2e` (focus_helper.py, case_*, main.go, matrix.go), `mise.toml`, `README.md`, `README.en.md`, `docs/CONFIG.md`, `docs/adr/adr-008-spike-log.md`, `internal/appid`, `internal/ctlsvc`, `cmd/goswitchctl` (оракулы, read-only)
**Просмотрено файлов:** ~25 (из них прочитано дословно: 11 — focus_helper.py целиком, d48-nightly-dispatch.sh, mise.toml, README.md :225–344, README.en.md :225–339, docs/CONFIG.md :245–409, adr-008 :253–339, appid.go :60–104, ctlsvc.go :250–299, main.go :395–435, case_spike_wezterm.go :246–251)
**Tracked-source gate:** все названные аналоги проверены `git ls-files` — tracked, зеркал нет
**Координация:** фаза не пересекается ни с одним кодовым файлом других фаз (правки — scripts/, docs/, условно test/e2e/mise.toml); демон не трогается — конфликтов слияния нет
**Дата извлечения:** 2026-11-03
