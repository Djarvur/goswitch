# goswitch

## What This Is

`goswitch` — корректор раскладки и введённого текста для GNOME Wayland (Ubuntu 24.04, T2 MacBook Pro): исправление текста, набранного не в той раскладке (EN↔RU), и переключение раскладки по горячим клавишам. Реализуется как собственный демон на Go, работающий в роли **IBus input method engine** через D-Bus. Полное техническое задание: `docs/SPEC.md` (SDD фаза 1, владелец Daniel Podolsky).

## Core Value

По горячей клавише исправить текст, набранный не в той раскладке (EN↔RU), в любом поле ввода GNOME Wayland — через IBus engine, без root и без конфликтов с ремапперами (keyd/xremap).

## Business Context

- **Customer**: Сам владелец (dogfooding); публичный open-source инструмент `Djarvur/goswitch`
- **Revenue model**: Нет (MIT, публичный репозиторий)
- **Success metric**: Зелёная e2e-матрица автотестов полного цикла без участия человека + ручная приёмка владельца
- **Strategy notes**: `docs/SPEC.md` §9 (вехи M0–M4)

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Коррекция по горячей клавише: перепечатать последнее слово / всю фразу / выделенный текст в другой раскладке
- [ ] Переключение раскладки по горячей клавише (проксируется в GNOME Super+Space), включая комбо «переключить и перепечатать»
- [ ] Направления EN→RU и RU→EN по позиции клавиш (ЙЦУКЕН↔QWERTY, включая знаки), с сохранением регистра (`GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет`, `ghbdtn`→`привет`)
- [ ] Автоопределение направления по составу буфера; смешанный текст — корректируется только часть не в той раскладке
- [ ] Замена без визуального «прыжка»: `commit_text` заменяет ровно диапазон неверного текста
- [ ] Буфер очищается по Enter, Tab, клику мыши, смене фокуса окна, Escape
- [ ] Чтение/установка текущей раскладки через GNOME `org.gnome.desktop.input-sources`
- [ ] Совместимость с keyd/xremap на уровне клавиатуры (разные слои, без EVIOCGRAB-монополии)
- [ ] Замена Super+Буква → Ctrl+Буква на лету в определённых приложениях (клавиатурные макросы)
- [ ] YAML-конфиг: все горячие клавиши переназначаются без перезапуска (hot reload)
- [ ] Автоматизированное тестирование полного цикла: unit/golden в CI (headless) + e2e-стенд (ydotool + AT-SPI, матрица кейсов YAML, отчёт PASS/FAIL)
- [ ] Установка без root: systemd user unit, `go install`, регистрация engine в IBus, бинарник из GitHub releases

### Out of Scope

- Автокоррекция без горячей клавиши — осознанно отвергнута владельцем (пароли, код, невозможность исключений по приложениям на Wayland); если вернёмся — отдельной опцией default off с white/black list по классам окон
- Другие пары раскладок кроме EN↔RU — интерфейс таблиц обобщённый, наполнение позже
- Wayland-композиторы кроме GNOME (KDE/Sway) — архитектурно допускаются, не тестируются
- GUI настроек — YAML-конфиг достаточен
- Эвристики «не трогать URL/e-mail/пути/hex» — v2

## Context

- Проблема: на GNOME Wayland нет полноценного аналога Punto Switcher
- WaylandSwitcher (Nim): работает, но форк чужого кода, EVIOCGRAB монополен (несовместим с keyd/xremap), нет автотестов
- keyd/xremap: только ремап, нет буферизации текста и коррекции, монополизируют клавиатуру
- IBus-engine (Python): правильный уровень (не конфликтует по построению), но прототип не доведён
- Терминалы (wezterm): IBus-предкоммит может не применяться — fallback через wl-clipboard или отказ в v1 (открытый вопрос Q3)
- В идеале использовать клавиатурные биндинги свитчера Caramba
- Открытые вопросы SDD фазы 2: Q1 (жизнеспособность IBus engine на Go), Q2 (замена диапазона текста через IBus API), Q3 (терминалы), Q4 (перехват горячих клавиш: IBus vs XKB-слой), Q5 (не-латинские/не-кириллические раскладки), Q6 (автозапуск: systemd user unit + ibus-daemon restart)

## Constraints

- **Tech stack**: Go ≥ 1.23, только stdlib + `godbus/dbus` + минимальные зависимости — простота поставки и аудита
- **Architecture**: Инъекция текста через IBus engine (`commit_text`), НЕ через uinput — не конфликтует с evdev-ремапперами, работает в любом поле ввода GTK/Qt/Chromium/Electron
- **Architecture**: Один демон `goswitchd` — IBus engine + D-Bus сервис + наблюдение за раскладкой
- **Permissions**: Демон без root — членство в группе доступа IBus-сокета; systemd user unit
- **Performance**: Реакция на горячую клавишу < 50 мс; память < 50 МБ
- **Process**: SDD (spec-driven development), трёхфазная модель; изменения требований через spec-delta
- **Quality**: Автоматизированное тестирование полного цикла без участия человека; e2e требует сам активировать целевое окно перед кейсом
- **Licensing**: MIT, репозиторий `Djarvur/goswitch` public; conventional commits, основные ветки защищены, изменения через PR

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| IBus engine вместо uinput для инъекции текста | Видит клавиатурные события до приложений на уровне метода ввода; коммитит текст напрямую; не требует grab клавиатуры; совместим с keyd/xremap | — Pending |
| Один демон `goswitchd` (IBus engine + D-Bus + наблюдение раскладки) | Простота эксплуатации и отладки единого процесса | — Pending |
| Язык Go (stdlib + godbus) | Владелец хочет свой код на Go; статический бинарник; минимальные зависимости | — Pending |
| Автокоррекция отвергнута | Пароли, код, невозможность исключений по приложениям на Wayland | — Pending |
| Горячие клавиши: double/triple right shift, right shift, Shift+RightCtrl | Привычные Punto-подобные жесты; все переназначаются в YAML; желательно биндинги Caramba | — Pending |
| Эталон приёмки — зелёная e2e-матрица (ydotool + AT-SPI) | Полный цикл без участия человека; тот же путь, что живая клавиатура | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-09-10 after initialization*
