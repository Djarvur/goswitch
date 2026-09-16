# Requirements: goswitch

**Defined:** 2026-09-10
**Core Value:** По горячей клавише исправить текст, набранный не в той раскладке (EN↔RU), в любом поле ввода GNOME Wayland — через IBus engine, без root и без конфликтов с keyd/xremap.

## v1 Requirements

Требования первоначального релиза. Каждое отображается на фазу роадмапа. Источник: `docs/SPEC.md` (авторитет) + `.planning/research/` (проверенные механизмы).

### CORR — Коррекция текста

- [x] **CORR-01**: По двойному нажатию Right Shift исправить последнее слово в другой раскладке (EN↔RU по позиции клавиш)
- [x] **CORR-02**: По тройному нажатию Right Shift исправить всю фразу, набранную в буфере
- [x] **CORR-03**: При активном выделении двойной Right Shift исправляет выделенный текст
- [x] **CORR-04**: Направление EN→RU / RU→EN определяется автоматически по составу буфера (букв какой раскладки больше)
- [x] **CORR-05**: Регистр сохраняется посимвольно: `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет`, `ghbdtn`→`привет`
- [x] **CORR-06**: Смешанный текст: корректируется только часть, набранная не той раскладкой (точная семантика — ADR фазы решения)
- [x] **CORR-07**: Замена применяется ровно к диапазону неверного текста без визуального «прыжка» (surrounding-text; fallback Backspace×N / буфер обмена — по capability ladder)
- [x] **CORR-08**: Таблицы ЙЦУКЕН↔QWERTY включают знаки `[ ] ; ' , . /` и генерируются `go:generate` из xkb symbols
- [x] **CORR-09**: Буфер очищается по Enter, Tab, Escape, смене фокуса окна; клик мыши — по результатам ADR (на уровне IME ненаблюдаем)

### SWCH — Переключение раскладки

- [x] **SWCH-01**: Right Shift переключает раскладку EN↔RU; механизм (two-engine регистрация или внутренний флип режима движка) фиксируется ADR фазы решения — прямая запись `gsettings current` рантаймом GNOME игнорируется
- [x] **SWCH-02**: Комбо «исправить слово и переключить раскладку» одной комбинацией (дефолт Shift+RightCtrl, переназначается)
- [x] **SWCH-03**: Родное GNOME-переключение Super+Space и MRU продолжают работать (goswitch зарегистрирован как input source, индикатор GNOME актуален)
- [x] **SWCH-04**: Конфликт тайминга single/double/triple Right Shift разрешён ADR фазы решения (Caramba-подобная схема «свитч на первый тап» — кандидат)

### CONF — Конфигурация

- [x] **CONF-01**: Все горячие клавиши, таймауты тапов и параметры задаются в YAML-конфиге
- [x] **CONF-02**: Изменения конфига применяются без перезапуска демона (hot reload)
- [x] **CONF-03**: Схема биндингов совместима со свитчером Caramba [wish-уровень: уточняется при обсуждении фазы]

### INTEG — Интеграция с окружением

- [x] **INTEG-01**: `goswitchd` работает как IBus input method engine (D-Bus `org.freedesktop.IBus`, регистрация `_RegisterComponent`), текст инжектируется через `commit_text`
- [x] **INTEG-02**: Совместимость с keyd/xremap: обычный набор идёт транзитом, никакого EVIOCGRAB / захвата клавиатуры
- [x] **INTEG-03**: Работает поверх дефолтного IM-стека Ubuntu 24.04 GNOME Wayland без root (членство в группе IBus-сокета)
- [x] **INTEG-04**: Движок перерегистрируется после рестарта ibus-daemon (известная потеря регистрации — ibus#2910)
- [x] **INTEG-05**: Паника обработчика не роняет движок и не убивает ввод сессии (recover-шим на точке диспетчеризации; падение движка = смерть ввода всего рабочего стола)

### MACR — Клавиатурные макросы

- [x] **MACR-01**: Замена Super+Буква → Ctrl+Буква на лету в определённых приложениях — реализуется внутри goswitch (решение владельца D-12: keyd отвергнут — сложен и ненадёжен, goswitch сам упрощает конфигурацию); механизм идентификации приложений выбирает ADR-005 фазы 1 — Complete с оговоркой (владелец, UAT 2026-09-15): механизм реализован по ADR-005 и проводка доказана (burst движка + релей ibus-daemon, dbus-monitor); конечный эффект применяется Chromium-семейством, GTK4-виджеты GNOME 46 forwarded-события игнорируют — документированное платформенное ограничение (WINDOWS ledger #4 закрыт)

### TEST — Автоматизированное тестирование

- [x] **TEST-01**: Unit/golden-тесты в headless CI: таблицы раскладок, детектор направления, регистры, логика буфера (клавиша→буфер→решение), парсер конфига и hot reload
- [x] **TEST-02**: e2e-стенд: инжекция нажатий физическим путём (ydotool/uinput — тот же путь, что живая клавиатура), чтение результата из поля редактора через AT-SPI
- [x] **TEST-03**: e2e-стенд сам активирует целевое окно перед каждым кейсом (фокус непредсказуем)
- [x] **TEST-04**: Матрица кейсов в YAML (ввод → ожидание): `ghbdtn`+хоткей→`привет`, регистры, фразы, выделение, смешанный текст, разные приложения (gnome-text-editor, Chrome); отчёт PASS/FAIL, код выхода ≠ 0 при падении (закрыта матрицей v1 Фазы 2 — 16 кейсов, локально и на CI-раннере; широта наращивается в Фазе 3)

### INST — Поставка и эксплуатация

- [x] **INST-01**: Установка без root: systemd user unit, регистрация engine в IBus, `go install` + бинарник из GitHub releases
- [x] **INST-02**: CLI `goswitchctl`: статус, перечитать конфиг, принудительно скорректировать
- [ ] **INST-03**: Реакция на горячую клавишу < 50 мс; потребление памяти < 50 МБ
- [x] **INST-04**: Структурные логи с уровнями; debug-режим с трассировкой клавиш

## v2 Requirements

Отложено (не в текущем роадмапе).

- **CORR-10**: Эвристики «не трогать»: URL, e-mail, пути, hex-строки (настраиваемые)
- **LAYOUT-01**: Другие пары раскладок кроме EN↔RU (интерфейс таблиц обобщённый)
- **TERM-01**: Коррекция в терминалах через fallback wl-clipboard + эмуляция вставки (если ADR фазы решения отклонит IBus-путь для терминалов, срок v1)
- **DESK-01**: Поддержка KDE/Sway (архитектурно допускается, не тестируется)

## Out of Scope

Явно исключено. Зафиксировано против scope creep.

| Feature | Reason |
|---------|--------|
| Автокоррекция без горячей клавиши | Осознанно отвергнута владельцем: пароли, код, невозможность исключений по приложениям на Wayland. Если вернёмся — отдельной опцией, default off, white/black list по классам окон |
| GUI настроек | YAML-конфиг достаточен |
| Wayland-композиторы кроме GNOME | Архитектурно допускаются, не тестируются в v1 |
| Форк/реюз WaylandSwitcher или goibus | Владелец требует свой код; goibus без LICENSE — юридический блокер для MIT-проекта |

## Traceability

Заполнено при создании роадмапа (2026-09-10). Структура фаз следует вехам `docs/SPEC.md` §9: M0+M1 → Фаза 1, M2 → Фаза 2, M3 → Фаза 3, M4 → Фаза 4.

| Requirement | Phase | Status |
|-------------|-------|--------|
| CORR-01 | Phase 2 | Complete |
| CORR-02 | Phase 3 | Complete |
| CORR-03 | Phase 3 | Complete |
| CORR-04 | Phase 2 | Complete |
| CORR-05 | Phase 2 | Complete |
| CORR-06 | Phase 3 | Complete |
| CORR-07 | Phase 2 | Complete |
| CORR-08 | Phase 1 | Complete |
| CORR-09 | Phase 2 | Complete |
| SWCH-01 | Phase 3 | Complete |
| SWCH-02 | Phase 3 | Complete |
| SWCH-03 | Phase 3 | Complete |
| SWCH-04 | Phase 3 | Complete |
| CONF-01 | Phase 3 | Complete |
| CONF-02 | Phase 3 | Complete |
| CONF-03 | Phase 3 | Complete |
| INTEG-01 | Phase 1 | Complete |
| INTEG-02 | Phase 1 | Complete |
| INTEG-03 | Phase 1 | Complete |
| INTEG-04 | Phase 1 | Complete |
| INTEG-05 | Phase 1 | Complete |
| MACR-01 | Phase 3 | Complete (проводка доказана; эффект — Chromium-семейство, GTK4-Wayland GNOME 46 не применяет forwarded-события — документированное платформенное ограничение, WINDOWS #4 закрыт владельцем 2026-09-15) |
| TEST-01 | Phase 1 | Complete |
| TEST-02 | Phase 1 | Complete |
| TEST-03 | Phase 1 | Complete |
| TEST-04 | Phase 2 | Complete |
| INST-01 | Phase 4 | Complete |
| INST-02 | Phase 3 | Complete |
| INST-03 | Phase 4 | Pending |
| INST-04 | Phase 1 | Complete |

**Coverage:**

- v1 requirements: 30 total (по фактическому числу REQ-ID; ранее указанное «29» было ошибкой счёта)
- Mapped to phases: 30 (Phase 1: 10, Phase 2: 6, Phase 3: 12, Phase 4: 2)
- Unmapped: 0

---
*Requirements defined: 2026-09-10*
*Last updated: 2026-09-15 — MACR-01 flipped Complete with GTK4-Wayland forwarded-events platform caveat (owner decision, Phase 3 UAT item 3)*
