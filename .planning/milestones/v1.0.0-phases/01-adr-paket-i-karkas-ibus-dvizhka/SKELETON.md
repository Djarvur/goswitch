# Walking Skeleton — goswitch

**Phase:** 1
**Generated:** 2026-09-10

## Capability Proven End-to-End

goswitchd запускается от пользователя сессии, подключается к приватной шине IBus, программно регистрируется как input method engine (goswitch-en + goswitch-ru), будучи активным источником ввода получает каждое нажатие клавиши и пишет его в структурированный JSON-лог, пропуская набор транзитом (наблюдатель), переживает рестарт ibus-daemon циклом перерегистрации и завершается чисто по SIGTERM.

Это гейт M1 в миниатюре: «движок активен, событие видно в логе, ввод рабочего стола не пострадал».

## Architectural Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Язык/рантайм | Go, go.mod `go 1.23` (локальный тулчейн 1.27.x), stdlib-first | Конвейер поставки: статический бинарник, `go install`, аудит за вечер (AGENTS.md) |
| Транспорт IBus | Собственный адаптер `engine/` на `github.com/godbus/dbus/v5` v5.2.2, приватный сокет IBus | goibus без LICENSE — юридический блокер для MIT; wire-формат верифицирован исследованием (goibus source + libibus strings + заголовки) |
| Регистрация | Программная `org.freedesktop.IBus.RegisterComponent`, без XML и без root | Прототип владельца доказал путь на этой машине; XML-инсталляция — Фаза 4 |
| Инъекция текста | Только сигналы IBus-движка (CommitText и др.); uinput/evdev запрещён в продукте | Аксиома SPEC §3.1: не конфликтует с keyd/xremap, работает в любом поле GTK/Qt/Chromium |
| Логирование | stdlib `log/slog`, JSON handler на stderr, трассировка клавиш за opt-in `-debug` | INST-04; нулевые зависимости; e2e-стенд ассертит по JSON-строкам |
| Чистая логика | `internal/hotkey` (FSM) и `internal/session` (актор) без D-Bus; адаптер — единственный слой, знающий провод | Headless CI: FSM и таблицы тестируются без живой сессии |
| Таблицы раскладок | Генерация `go:generate` из /usr/share/X11/xkb/symbols (us basic + ru=winkeys), файл закоммичен, golden-тесты в CI | CORR-08: позиция клавиши, не транслитерация; CI не зависит от xkb |
| e2e-стенд | Go-раннер + ydotool (физический путь) + AT-SPI grabFocus через /usr/bin/python3-gi | SPEC §7.2-7.3; стенд симулирует человека, сам активирует окно |
| Надзор | dev systemd user unit: PartOf=graphical-session.target, Restart=on-failure | Как дистро запускает сам ibus-daemon; kill -9 покрывается перезапуском + цикл перерегистрации |
| Раскладка репо | cmd/ (бинарники), engine/ (адаптер), internal/ (чистая логика), layouts/ (данные+генератор), test/e2e/ (стенд), docs/adr/ | PATTERNS.md; без pkg/ (go-ultimate) |

## Stack Touched in Phase 1

- [x] Project scaffold (go.mod, .gitignore, CI vet+build+test-race)
- [x] Транспорт — реальное подключение и регистрация на приватной шине IBus
- [x] Приём событий — ProcessKeyEvent активного источника ввода → структурированный лог
- [x] Инъекция — эмиттер CommitText в адаптере (коррекция подключается в Фазе 2 без изменения архитектуры)
- [x] Запуск — документированная локальная команда + dev systemd user unit

## Out of Scope (Deferred to Later Slices)

- Коррекция слова (буфер, surrounding-text сверка, замена) — Фаза 2
- Фразы, выделение, смешанный текст, переключение раскладки, YAML-конфиг, hot reload, goswitchctl — Фаза 3
- Установка без root по инструкции, производительность, релиз — Фаза 4
- OSD/уведомление при переключении — deferred (CONTEXT)
- Рецепт Super+Буква через keyd в документации — кандидат Фазы 4, часть решения OPEN-01

## Subsequent Slice Plan

Каждая следующая фаза добавляет вертикальный срез поверх скелета, не меняя его архитектурных решений:

- Phase 2: буфер + коррекция слова EN↔RU (двойной Right Shift) на таблицах layouts/ и FSM internal/hotkey/ — e2e-матрица v1 (AT-SPI readback вместо лог-ассертов)
- Phase 3: жесты (фраза/выделение/смешанный текст), переключение раскладки механизмом ADR-001, YAML-конфиг + hot reload + goswitchctl — e2e-матрица v2
- Phase 4: установка (systemd user unit + IBUS_COMPONENT_PATH + go install/binary), производительность (<50 мс / <50 МБ), двойной зелёный прогон, ручная приёмка владельца
