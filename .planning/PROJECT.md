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

- ✓ Совместимость с keyd/xremap на уровне клавиатуры (разные слои, без EVIOCGRAB-монополии) — Phase 1 (INTEG-02: живая сессия с keyd, владелец подтвердил; архитектура IBus engine не захватывает evdev)
- ✓ Направления EN→RU и RU→EN по позиции клавиш (ЙЦУКЕН↔QWERTY, включая знаки), с сохранением регистра (`GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет`, `ghbdtn`→`привет`) — Phase 2 (CORR-05: три регистра в e2e-матрице на gnome-text-editor + юнит-корпус)
- ✓ Замена без визуального «прыжка»: `commit_text` заменяет ровно диапазон неверного текста — Phase 2 (владелец подтвердил визуально 2026-09-15: ghbdtn→привет in place; лог демона: action n:2 → correction done → verify match)
- ✓ Автоматизированное тестирование полного цикла: unit/golden в CI (headless) + e2e-стенд (ydotool + AT-SPI, матрица кейсов YAML, отчёт PASS/FAIL) — Phase 1+2 (матрица v1: 16/16 на трёх поверхностях, exit≠0 контракт; CI-контур e2e на self-hosted GNOME runner)
- ✓ Полный словарь жестов коррекции: слово (двойной Right Shift), фраза (тройной), выделенный текст (двойной при выделении, обе геометрии) — Phase 3 (CORR-02/03: матрица v2 21/21 PASS на HEAD 54ea460, живой прогон 2026-09-15)
- ✓ Смешанный текст: конвертируются только чужие раны, якорь — последняя буква; однородный — целиком (чтение D-24 утверждено владельцем) — Phase 3 (CORR-06, D-22/D-23: TestConvertRuns_GoldenCorpus + строка word-mixed)
- ✓ Переключение раскладки одиночным Right Shift и комбо Shift+RightCtrl «исправить и переключить» — Phase 3 (SWCH-01/02: внутренний флип за `layoutset` по ADR-001, матрица layout-single/combo-word-layout зелёные; oracle — mode-лог, не gsettings)
- ✓ Буфер очищается по Enter/Tab/клику/смене фокуса/Escape (ADR-004) — Phase 2+3 (матрицы v1/v2: reset-триггеры прогоняются в кейсах с фокусом и перезапуском серий)
- ✓ Замена Super+Буква → Ctrl+Буква на лету в определённых приложениях (MACR-01) — Phase 3 (проводка доказана: burst движка + релей ibus-daemon; эффект применяется Chromium-семейством; GTK4-Wayland GNOME 46 forwarded-события игнорирует — документированное платформенное ограничение, владелец принял 2026-09-15)
- ✓ YAML-конфиг: все хоткеи/таймауты/капы перенастраиваются без перезапуска (hot reload, last-good) + `goswitchctl status/reload/correct` — Phase 3 (CONF-01/02, INST-02: строгий KnownFields-декод, потолки значений, reload через ctl в матрице)

### Active

- [ ] Установка без root: systemd user unit, `go install`, регистрация engine в IBus, бинарник из GitHub releases; бюджет производительности (<50 мс / <50 МБ); полная матрица дважды зелёная на нетронутой сессии; ручная приёмка владельца — релиз v1 (Phase 4)

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
| IBus engine вместо uinput для инъекции текста | Видит клавиатурные события до приложений на уровне метода ввода; коммитит текст напрямую; не требует grab клавиатуры; совместим с keyd/xremap | ✓ Phase 1 — каркас работает как активный input source, переживает рестарт ibus-daemon и паники |
| Один демон `goswitchd` (IBus engine + D-Bus + наблюдение раскладки) | Простота эксплуатации и отладки единого процесса | ✓ Phase 1 — единый процесс: регистрация, логирование, рестарты |
| Язык Go (stdlib + godbus) | Владелец хочет свой код на Go; статический бинарник; минимальные зависимости | ✓ Phase 1 — godbus v5.2.2 + yaml.v3 + fsnotify, ничего сверх таблицы стека |
| Автокоррекция отвергнута | Пароли, код, невозможность исключений по приложениям на Wayland | — Pending |
| Горячие клавиши: double/triple right shift, right shift, Shift+RightCtrl | Привычные Punto-подобные жесты; все переназначаются в YAML; желательно биндинги Caramba | ✓ Phase 3 — полный словарь жестов живой, tap_key/window/verify_wait перенастраиваются hot reload'ом (CR-01/02), таблица соответствия Caramba в docs/CONFIG.md |
| Эталон приёмки — зелёная e2e-матрица (ydotool + AT-SPI) | Полный цикл без участия человека; тот же путь, что живая клавиатура | ✓ Phase 1 — стенд работает: e2e-m1, ibus-restart, kill9-survive, d01 зелёные |
| ADR-001..005 — пакет архитектурных решений Фазы 1 | Закрыты развилки: механизм переключения раскладки, тайминги тапов, capability ladder замены, триггеры очистки буфера, судьба макросов (MACR-01 остаётся в v1) | Accepted 2026-09-10 (владелец, M0-гейт; D-01 журнал: 15 вердиктов) |
| Отклонение ADR-003 от level-2 лора ibus#2354: фактический уровень лестницы = 1 на всех поверхностях стола (chromium применяет DeleteSurroundingText) | Живая находка 02-05; владелец принял отклонение без правки ADR-003/ROADMAP (WINDOWS #1 resolved); контракт уровня 2 несёт юнит-корпус | Accepted 2026-09-15 (verify-work UAT) |
| Корпусное чтение D-24: однородный текст конвертируется целиком (уточнение D-22); `changed=false` зарезервирована для внешне-заякоренных диапазонов (выделение); D-16→D-23 — смешанное слово конвертирует чужие раны | Буквальный пример плана («привет → out==input») разрешён в пользу must-haves; матрица v1 word-ru-en остаётся зелёной | Accepted 2026-09-15 (владелец, UAT item 2 — WINDOWS #2 закрыт) |
| GTK4-Wayland (GNOME 46) не применяет IBus-forwarded события — документированное платформенное ограничение | Демон-сторона MACR/level-2/Ctrl+V доказана на проводке (dbus-monitor); эффект применяется Chromium-семейством; REQUIREMENTS.md MACR-01 выровнен | Accepted 2026-09-15 (владелец, UAT item 3 — WINDOWS #4 закрыт) |
| Вердикт спайка 03-03 по выделению zenity развёрнут: выжившее ctrl+a выделение поглощает коммит (замена выделенного residue), поле остаётся с конвертированным словом | Спайк был загрязнён probe-кандидатами, разрушавшими выделение; исправленный факт пинируется строкой select-all-zenity | Accepted 2026-09-15 (владелец, UAT item 4) |
| Переключение раскладки — внутренний флип режима за `layoutset` (ADR-001 Option B), oracle = mode-лог; gsettings sources/current записью игнорируется shell-рантаймом GNOME 46 | Живая верификация Фазы 1 (01-03); поверхностей-свидетелей обратного нет; D-34 рабочая интерпретация super-space-alive | ✓ Phase 3 — layout-single оба направления зелёные в матрице v2 |
| Программная регистрация IBus-компонента (runtime, без XML) | В Фазе 1 нет файла поставки, который можно подменить; XML-инсталляция и права каталогов — Фаза 4 | Accepted (T-01-04) |

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
*Last updated: 2026-09-15 after Phase 3*
