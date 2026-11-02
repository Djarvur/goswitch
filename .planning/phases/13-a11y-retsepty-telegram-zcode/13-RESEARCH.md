# Фаза 13: A11Y — рецепты Telegram/ZCode — Research

**Researched:** 2026-11-03
**Domain:** включение доступности в самих приложениях (Electron-флаг ZCode, сборка Telegram) + внешний self-check-скрипт + документация рецептов; демон goswitchd НЕ меняется
**Confidence:** HIGH (стендовые факты и оракулы — живые пробы этой сессии + предыдущие живые вердикты; flatpak/snap a11y-механика для Telegram — честный probe-first, web-версии помечены)

## Summary

Фаза на три четверти — пробная и документационная, и почти весь её «стек» уже жив в репо и на
стенде. Демон не меняется (решение 2026-10-08): его наблюдение — `internal/appid` (identity из
bridge-namespace + живой `GetRole` под 25 мс) и двухразрешительная модель D-53/D-55 — уже
определяет, что именно надо измерять. Главный технический факт фазы: **«переключение работает» и
«автокоррекция работает» — разные слои**. Переключение раскладки — хоткей IBus-движка, ему a11y
не нужен вовсе; автокоррекции нужен входящий `org.a11y.atspi.Accessible.GetRole` в объекты
приложения. Поэтому сообщение владельца «сейчас у меня переключенив телеграме работает»
[VERIFIED: 13-CONTEXT.md D-01, дословно] **не противоречит** AppArmor-вердикту 2026-10-08
(snap зарегистрирован на a11y-шине, но входящие вызовы `Cache.GetItems`/`Accessible.GetRole`
policy-denied [CITED: .planning/research/STACK.md §(c)]) — и проба D-01 обязана разделить эти
слои по машинным оракулам, а не по формулировке.

Второй ключевой факт: **стенд-сеть асимметрична** (живые пробы этой сессии): snap store
доступен (`snap find telegram-desktop` → 7.3.3), `dl.flathub.org` отдаёт 200 (репо-CDN — то, что
нужно `flatpak install`), но `flathub.org` (сайт) и `telegram.org` недоступны. Flatpak на ВМ не
установлен (apt-кандидат 1.14.6-1ubuntu0.1 есть). Значит: snap-рунг ставится прямо с store,
flatpak-рунг — `apt install flatpak` + `flathub.flatpakrepo` с dl.flathub.org (жизнеспособен),
а tarball-рунг лестницы D-03 стейджится скачкой на хосте + scp/rsync (прямая скачка с
telegram.org на ВМ невозможна). Хост-реплика владельца: Telegram **snap 7.3.3 rev 7175**
сегодня (вердикт 2026-10-08 снимался на 7.2.9 rev 7150 — обновление snap само по себе повод
для перепроверки D-01); ZCode на хосте — самодостаточный Electron-каталог `/opt/ZCode`
(811 М), на ВМ переносится rsync-ом.

Третий факт — оракул «видимости поля» на ВМ имеет задокументированные качели: журнал Фазы 10
доказал живое дерево (GTK_MODULES-активация + RegisterEvent демона → 153 StateChanged), но
диагноз ci-runner 2026-10-27 фиксирует «AT-SPI-оракул на ВМ мёртв» для zenity-фикстур (нет
фокус-флагов/текста при реальном фокусе; софтверный рендер оболочки). Вывод для плана:
самопроверка оракула обязательна (калибровочная проба на известном GTK-приложении до любых
Telegram/ZCode-вердиктов), а главным машинным оракулом «автокоррекция работает» должен быть
**сам демон** — счётчики `goswitchctl status` (`autocorrect_fired`, `ac_skip_*`,
`ac_role_errors`) и журнал -debug, что Фаза 10 уже отработала на этом стенде байт-точно.

**Primary recommendation:** первый план — подъём стенда (recipe 10-01/10-08) + калибровка
a11y-оракула + живая проба D-01/D-02 (snap и flatpak одной сессией, таблица рядов в
adr-008-spike-log.md); второй план — self-check-скрипт `scripts/` (python3-gi, коды возврата
0/1/2 по прецеденту `test/e2e/focus_helper.py`) + рецепты в README.md/README.en.md +
детали/побочки в docs/CONFIG.md; третий — residual-risk заметка и фазовые grep-гейты доков.
Daemon-код — ноль строк.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Probe-first — первый план фазы живая ВМ-проба Telegram ДО любых решений о рецепте/сборке. Основание — сообщение владельца (дословно, без интерпретации): «сейчас у меня переключенив телеграме работает, проверь в виртуалке» (2026-11-02, про хост владельца). Проба устанавливает машинную истину и перепроверяет вердикт AppArmor из A11Y-02 (2026-10-08). Что именно работает (переключение раскладки / слово / строка) — решает проба, не формулировка сообщения.
- **D-02:** На ВМ ставятся ОБЕ установки одной сессией: snap (реплика хоста владельца — проверяет и его сообщение, и старый AppArmor-вердикт) и flatpak (кандидат-рунг для рецепта). Полная сравнительная таблица.
- **D-03:** Если unconfined-проба не подтвердит a11y-достижимость поля — исполнитель автоперебирает форматы (tarball → deb → flatpak), стоп на первом зелёном, все ряды в журнал.
- **D-04:** Рунг лестницы при выборе новой сборки — **Flatpak** (выбор владельца). — **Reversibility:** reversible — смена формата сборки не ломает контракт.
- **D-05:** Все проверки и зачёт — ТОЛЬКО на ВМ goswitch-stand (директива 6 + выбор владельца «Только ВМ» для ZCode): ZCode ставится на ВМ, стол владельца не трогаем вообще.
- Стенд сейчас запаркован (демон inactive, конфиг удалён): план обязан включать подъём — gap-сборка текущей ветки, конфиг стенда, свежая Wayland-сессия (recipe 10-08: `ssh stand@192.168.122.150`, `/tmp/stand-env.sh`, reboot-прецедент при мёртвой сессии).
- **D-06:** Готовый скрипт в репо (`scripts/`) с человекочитаемым вердиктом и кодом возврата; в документации — одна строка запуска. Демон не меняется: скрипт — внешний инструмент (AT-SPI-зонд видимости поля ввода).
- **D-07:** Краткие рецепты в `README.md` и `README.en.md`; детали, побочные эффекты и границы — `docs/CONFIG.md`; residual-risk заметка (ручная коррекция на скрытых полях копирует в PRIMARY) — в README.

### Claude's Discretion

- Точный текст рецептов и порядок шагов; внутренности self-check-скрипта; порядок snap→flatpak внутри пробной сессии; форма сравнительной таблицы пробы.

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| A11Y-01 | ZCode: автокоррекция работает — рецепт локального `.desktop` override с `--force-renderer-accessibility` (toolkit-accessibility через a11y-магию демона уже работает); приёмка по живому AT-SPI-дереву | Механизм флага документирован Electron/Chromium; шаблон override уже в docs/CONFIG.md:373+ (`Exec=env ACCESSIBILITY_ENABLED=1 /opt/ZCode/zcode --force-renderer-accessibility %U`); системный .desktop хоста прочитан живьём (`Exec=/opt/ZCode/zcode %U`, StartupWMClass=ZCode); путь установки ZCode на ВМ — rsync /opt/ZCode (811 М); оракул — живое дерево + счётчики демона |
| A11Y-02 | Telegram: автокоррекция достижима. Спайк подтверждает AppArmor-блокировку snap для входящих AT-SPI вызовов (env-рецепты не решают — доказано зондированием) и выносит решение владельца: unconfined-сборка (flatpak/deb) vs демон-fallback; при демон-fallback — отдельная приёмка с владельцем | Проба D-01/D-02 пере-исполнит вердикт на свежем snap 7.3.3 (хост-реплика обновилась с 7.2.9); различение исходящих/входящих соединений snap даёт машину для интерпретации; flatpak-рунг установлен по сетевой карте ВМ; лестница D-03 с host-стейджингом tarball (telegram.org с ВМ недоступен) |
| A11Y-03 | Рецепты документированы в README/CONFIG.md, включая побочный эффект Telegram (плашка «режим экранного диктора») и ограничение «флаги Electron применяются только на запуске» | Карта текущих разделов снята (README.md:239-255 + 307+, README.en.md:233-249 + 296+, CONFIG.md:255+/373+); побочка диктора — Qt-мост Telegram реагирует на ScreenReaderEnabled-семантику включённого a11y-стола [ASSUMED A5 — форму проверяет проба]; restart-семантика Electron закреплена SPEC §4.5 |
</phase_requirements>

## Project Constraints (from AGENTS.md — CLAUDE.md в репо нет)

- **Директива 6 (2026-11-01):** любые проверки продукта — ТОЛЬКО на ВМ goswitch-stand; a11y-пробы на хосте запрещены (инцидент 2026-10-31: UAF atk-bridge уронил сеанс владельца). Чтение журналов/конфигов хоста — пассивно, допустимо (живые факты этой research получены именно так: snap list, dpkg, чтение .desktop).
- **Строгий TDD** (red→green→refactor) — для поведенческих задач; живой AT-SPI headless-RED невозможен (прецедент focus_helper: «gi/AT-SPI needs a live session») — зелёный гейт скрипта = живая ВМ-проба, RED-форма = падающий прогон против заведомо невидимого приложения.
- **Зелёная итерация:** `go build ./...`, `go test -race ./...`, `golangci-lint run` — фаза не пишет Go-кода, но не должна ломать гейты (mise run ci обязан оставаться зелёным).
- **mise, не make**; **golangci-lint v2 strict** с Фазы 1; **CI-гейты go-ultimate** (build/vet/test -race/cover/tidy-diff/govulncheck).
- **SDD/spec-delta:** демон не меняется — spec-delta НЕ нужен; изменения только в docs/ + scripts/.
- **Git:** изменения через PR в защищённый main, conventional commits.
- **Ноль новых Go-зависимостей** (фаза их и не требует).

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Включение a11y в ZCode | Приложение (Electron-флаг запуска) | Демон (a11y-магия §4.5 — уже поставлена, default ON) | Флаг — свойство процесса приложения; демон лишь создаёт окружение (toolkit-accessibility + IsEnabled), принудительный рендер-флаг снимает зависимость от порядка запуска |
| Включение a11y в Telegram | Сборка/формат поставки приложения (snap-песочница vs flatpak) | — | Блокировка — свойство confinement формата, не приложения и не демона; демон не может её обойти (доN'T hand-roll ниже) |
| Наблюдение identity/роли | Демон (`internal/appid` + role-гейт D-53/D-55) | — | Уже реализовано; фаза только потребляет его как оракул |
| Пробный оракул «поле видно» | Внешний AT-SPI-зонд (python3-gi) + счётчики демона | Живое дерево (калибровка) | Демон — первичный потребитель a11y; его счётчики — самый честный «field visible» для автокоррекции; зонд — независимый свидетель |
| Self-check одной командой | scripts/ (репо) | docs/ (строка запуска) | D-06: скрипт — внешний инструмент, код возврата для пользователя |
| Рецепты/границы/risidual-risk | docs/ (README ×2 + CONFIG.md) | — | D-07 |

## Standard Stack

### Core (всё существующее — ноль новых модулей и пакетов)

| Инструмент | Версия | Назначение | Почему стандарт |
|------------|--------|------------|-----------------|
| `internal/appid` (in-repo) | текущее дерево | Оракул видимости: роли 60/61/79/94/40, bridge-identity | Уже живой потребитель a11y-шины; `RoleTerminal=60` и текст-класс определены кодом [VERIFIED: internal/appid/appid.go:90-96, цитата ниже] |
| `goswitchctl status` (in-repo) | текущее дерево | Машинный оракул пробы: `ac_role_errors`, `ac_skip_*`, `autocorrect_fired/abstained` | Токены статуса — wire-канон [VERIFIED: internal/ctlsvc/ctlsvc.go:252-296] |
| `test/e2e/focus_helper.py` (in-repo) | текущее дерево | Прецедент python3-gi AT-SPI-зонда: протокол `app:role:chars=N`, коды 0/1/2, `/usr/bin/python3` | Self-check-скрипт D-06 повторяет его форму; все ловушки уже закодированы в докстринге [VERIFIED: test/e2e/focus_helper.py:1-110, прочитан живьём] |
| `python3-gi` + `gir1.2-atspi-2.0` | 3.48.2 / 2.52.0 (dpkg на ВМ и хосте) | AT-SPI-обход self-check-скрипта | SPEC §7.3 предписывает; установлены на ВМ [VERIFIED: живая проба import на ВМ этой сессией] и на хосте [VERIFIED: dpkg этой сессии] |
| snapd (ВМ) | 2.77.1 | Snap-рунг D-02 | Store доступен: `snap find telegram-desktop` → 7.3.3 [VERIFIED: живая проба на ВМ] |
| flatpak (ВМ, установка фазы) | 1.14.6-1ubuntu0.1 (apt-кандидат) | Flatpak-рунг D-02/D-04 | apt-архив доступен; репо-CDN dl.flathub.org отдаёт 200 [VERIFIED: живые пробы] |
| rsync/scp хост→ВМ | существующий канал зеркала | Стейджинг tarball-рунга D-03 и ZCode (811 М) | telegram.org недоступен с ВМ [VERIFIED: CURL-FAIL]; github.com доступен [VERIFIED: 200] |

### Supporting

| Инструмент | Когда используется |
|------------|--------------------|
| `virsh send-key` / `virsh screenshot` | Единственный клавиатурный канал до IBus в любые окна ВМ (Electron тоже); визуальная верификация плашек |
| `busctl --user` (ВМ) | Проверка `org.a11y.Bus GetAddress` / `org.a11y.Status.IsEnabled` — жива ли шина (сделано этой сессией: alive, IsEnabled=true) |
| `gsettings` (ВМ) | sources-takeover стенда (recipe 10-01), чтение `toolkit-accessibility` (сейчас true на ВМ) |
| `snap connections` (ВМ) | Карта интерфейсов установленного snap — интерпретация вердикта пробы |

### Alternatives Considered

| Рекомендовано | Альтернатива | Tradeoff |
|---------------|--------------|----------|
| Snap-реплика D-02 (проба 7.3.3 на ВМ) | Доверять вердикту 2026-10-08 без перепробы | Снято D-01: snap обновился (7.2.9→7.3.3), слова владельца требуют машины; перепроба дешёвая |
| Flatpak как рунг D-04 | Tarball с telegram.org как первый рунг | Манifest flathub жив (gh api); tarball требует host-стейджинга (сеть ВМ); порядок лестницы уже зафиксирован владельцем D-03 (tarball→deb→flatpak) — D-04 выбирает flatpak при выборе новой сборки |
| Self-check на python3-gi | Чистый Go-зонд в goswitchctl | Запрещено границей фазы (демон не меняется); python3-gi предписан SPEC §7.3 и установлен всюду |
| Копия ZCode rsync-ом на ВМ | Скачивание дистрибутива ZCode с апстрима | Копия детерминирована (та же сборка, что у владельца), 811 М по локальной сети — минуты; апстрим-канал ZCode не исследован [ASSUMED A2] |

## Package Legitimacy Audit

Фаза не добавляет Go-модулей и npm/pypi-пакетов: `go.mod` не трогается, весь код — один
python3-gi-скрипт (stdlib-обвязка над системной gir) + docs. Ставимые на ВМ артефакты —
дистрибутивы конечных приложений, не библиотеки разработки:

| Артефакт | Реестр/источник | Статус | Disposition |
|----------|-----------------|--------|-------------|
| telegram-desktop snap | snapcraft store (`snap find` живой) | официальный publisher `telegram-desktop**` [VERIFIED: живой `snap find` на ВМ] | Approved |
| flatpak + org.telegram.desktop | Ubuntu archive + dl.flathub.org (manifest flathub/org.telegram.desktop прочитан через gh api) | официальные источники [VERIFIED: gh api этой сессии; runtime org.gnome.Platform 51] | Approved |
| ZCode (копия /opt/ZCode) | хост владельца | та же сборка, что daily-driver [VERIFIED: ls/du этой сессии] | Approved |

**Packages removed due to SLOP verdict:** none. **Packages flagged SUS:** none.

## Architecture Patterns

### System Architecture Diagram — путь пробы и рецептов

```
                        ┌── ВМ goswitch-stand (все проверки, D-05) ──────────────────────────┐
                        │                                                                    │
 Владелец (слова) ──D-01──▶ План A: подъём стенда (recipe 10-01/10-08)                       │
                        │   rsync-зеркало → gap-сборка → unit -debug → sources-takeover      │
                        │   калибровка a11y-оракула (известный GTK-апп: дерево+роли живы?)   │
                        │        │                                                           │
                        │        ▼                                                           │
                        │   План A: живая проба Telegram (D-01/D-02, одна сессия)            │
                        │   ┌─ рунг snap 7.3.3 (store) ──▶ слой 1: переключение (key-trace)  │
                        │   │                            слой 2: дерево/GetRole (зонд)      │
                        │   │                            слой 3: автокоррекция (счётчики)   │
                        │   └─ рунг flatpak (apt+dl.flathub.org) ──▶ те же 3 слоя            │
                        │        │  неудача unconfined → лестница D-03 (tarball→deb→flatpak,│
                        │        │  стейджинг через хост) — стоп на первом зелёном           │
                        │        ▼                                                           │
                        │   План B: ZCode на ВМ (rsync /opt/ZCode + .desktop override)      │
                        │   флаг --force-renderer-accessibility → рестарт → дерево+счётчики  │
                        │        │                                                           │
                        │        ▼                                                           │
                        │   План B: scripts/ self-check (python3-gi, exit 0/1/2)             │
                        │        │                                                           │
                        └────────┼────────────────────────────────────────────────────────────┘
                                 ▼
                    План C: README.md + README.en.md (кратко + residual-risk)
                            docs/CONFIG.md (детали, побочки, границы)
                            ряды пробы → docs/adr/adr-008-spike-log.md
```

Читается сверху вниз: слова владельца НЕ решение, а повод для пробы; каждый рунг сборки
измеряется тремя независимыми слоями (hotkey-канал IBus / AT-SPI-дерево / счётчики демона);
результаты оседают в журнал спайка и становятся текстами рецептов.

### Recommended Project Structure (дельта фазы)

```
scripts/
├── a11y-selfcheck.py (или .sh-обёртка)   # D-06: AT-SPI-зонд видимости поля, exit 0/1/2
README.md / README.en.md                   # краткие рецепты + residual-risk PRIMARY (D-07)
docs/CONFIG.md                             # детали рецептов, побочки, границы (D-07)
docs/adr/adr-008-spike-log.md              # датированные ряды пробы (форма venue-рядов)
```

### Pattern 1: Трёхслойная проба (различение «переключение vs автокоррекция»)

**Что:** каждый рунг сборки Telegram измеряется тремя независимыми оракулами — тем же набором,
которым Фаза 10 пинила терминалы.
**Когда использовать:** в каждом ряду таблицы пробы D-02.
**Слои:**

1. **Слой IBus (переключение):** virsh send-key / ydotool `shift_r` → журнал демона
   (`mode to ru/en`, `switch_engine`) — работает везде, где жив IBus-контекст; это, скорее
   всего, и есть «переключенив телеграме работает» владельца [ASSUMED A6 — проба решает].
2. **Слой AT-SPI (видимость):** python3-gi зонд — приложение в desktop-children? `GetRole`
   сфокусированного поля отвечает (не «Access denied»)? какого класса роль?
3. **Слой автокоррекции (результат):** синтетический корпус `ghbdtn`-класса в поле →
   `goswitchctl status` до/после: `autocorrect_fired` растёт ИЛИ растёт
   `ac_skip_role_unknown`/`role-timeout` (fail-closed молчание при невидимой роли) —
   `ac_role_errors` отдельно считает ошибки источника роли
   [VERIFIED: internal/ctlsvc/ctlsvc.go:284 — `"ac_role_errors="+strconv.Itoa(st.AutoCorrectRoleErrors)`].

### Pattern 2: Self-check по прецеденту focus_helper.py

**Что:** скрипт `scripts/` повторяет контракт `test/e2e/focus_helper.py` — протокол
`app:role:chars=N`, человекочитаемый вердикт, коды возврата 0/1/2
[VERIFIED: test/e2e/focus_helper.py:109 — «Exit codes: 0 success, 1 named failure, 2 usage
error»].
**Когда использовать:** D-06.
**Контракт (рекомендация):**

```
exit 0 — приложение в живом AT-SPI-дереве, текст-поле видимо (роль класса 61/79/94)
exit 1 — приложение запущено, но дерева/поля нет (a11y не включён / не перезапущен)
exit 2 — приложение не найдено/не запущено/ошибка использования
```

Скрипт печатает роли, не предполагая их (роль поля Telegram-Qt и ZCode-Electron устанавливает
проба, не документация). Обязателен `/usr/bin/python3` (PATH-python3 — linuxbrew без gi,
ловушка 01-PATTERNS в докстринге focus_helper).

### Pattern 3: К-place доки — расширение существующих разделов, не новые страницы

**Что:** рецепты ложатся в уже существующие точки (grep-карта снята живьём):
- README.md:239-255 «Где автокоррекция молчит» — уже называет ZCode/snap Telegram и оба
  механизма; сюда — краткие прикладные рецепты ZCode/Telegram + residual-risk PRIMARY-заметка;
- README.md:307+ «A11y-магия приложений» — сюда уточнение «что магия НЕ делает» (ссылка на
  рецепты);
- README.en.md:233-249 и 296+ — зеркально;
- docs/CONFIG.md:373+ «Ручной рецепт: per-app запуск с a11y-флагами» — сюда полный рецепт
  ZCode-override (живая валидация) + Telegram-сборка + побочки/границы.

### Anti-Patterns to Avoid

- **Интерпретировать слова владельца как «автокоррекция работает»** — D-01 запрещает явно;
  таблица пробы фиксирует, какой из трёх слоёв жив.
- **Принимать «приложение в дереве» за «поле видно»** — snap Telegram регистрируется на шине
  (исходящее соединение) при_policy-denied входящих вызовах; вердикт только по `GetRole`
  сфокусированного объекта.
- **Единичный зелёный ряд** — двух-рядовая дисциплина SPEC §7.2 применяется к пробным
  вердиктам площадки (прецедент venue-probe r1/r2).
- **Правка демона «помочь» пробе** — граница фазы; всё, что требует кода демона, — эскалация
  владельцу (демон-fallback ветка A11Y-02), не тихий фикс.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Обход snap AppArmor для входящих AT-SPI | snap connect / interface-твик / прокси | Смена формата сборки (лестница D-03) | Стена — сгенерированный профиль snapd; интерфейса, открывающего входящие вызовы в snap, не существует [CITED: forum.snapcraft.io/t/27274 — desktop-legacy даёт только исходящий доступ к сокету a11y; docs a11y-interface = 404] |
| Go AT-SPI-клиент для self-check | Новый пакет/код в демоне | python3-gi скрипт в scripts/ | Демон не меняется (граница фазы); gir установлен на всех Ubuntu 24.04 GNOME [VERIFIED: dpkg хост+ВМ] |
| Детекция «a11y включён» по косвенным признакам | Парсинг /proc environ как вердикт | Живое дерево (GetRole/GetItems) + счётчики демона | environ показывает переменные процесса, не доступность входящих вызовов (2026-10-08: QT_ACCESSIBILITY=1 стоял при полностью заблокированном GetRole) |
| GUI/CLI-фреймворк для self-check | Всё, кроме stdlib python3 + gi | Один файл, ноль зависимостей | Директива минимальных зависимостей; прецедент focus_helper.py |
| Своя семантика «текст-класса» ролей | Эвристики по именам ролей | Сравнение по НОМЕРУ роли | Имена ролей локализуемы и расходятся между тулкитами; номера закреплены кодом демона [VERIFIED: internal/appid/appid.go:85-96] |

**Key insight:** фаза — редкий случай, когда «не hand-roll» почти абсолютна: демон, оракулы,
протокол зонда и место в доках уже существуют; новое — только instalляция двух сборок на ВМ,
один скрипт и тексты.

## Runtime State Inventory

> Фаза СОЗДАЁТ runtime-state (установки на ВМ) и опирается на запаркованный стенд — инвентарь обязателен.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | Конфиг стенда удалён (`~/.config/goswitch/` пуст, живая проверка); Telegram/ZCode на ВМ отсутствуют (нет snap/flatpak/процесса — живые проверки + 13-CONTEXT) | Создать при подъёме: конфиг стенда (autocorrect on для пробы слоя 3), источники ввода (sources-takeover recipe 10-01 с откатом `[('xkb','us')]`) |
| Live service config | goswitchd user unit на ВМ — inactive (живая проверка); `/tmp/stand-env.sh` отсутствует (tmp зачищен ребутом) | Пересоздать stand-env.sh, unit -debug drop-in, `ibus engine goswitch-en` — рецепт 10-01 Task 1 / 10-08 Task 2 дословно |
| OS-registered state | НОВОЕ от фазы: snap telegram-desktop + flatpak + flathub remote + flatpak org.telegram.desktop + `/opt/ZCode`-копия + `~/.local/share/applications/zcode.desktop` на ВМ — всё VM-local, обратимо (snap remove / flatpak uninstall / rm) | Зафиксировать откаты каждого рунга в рядах пробы; стол хоста не трогаем (D-05) — на хосте ничего не создаём |
| Secrets/env vars | Логин Telegram на ВМ: вход по QR требует телефона владельца — сессия не создаётся без него; ZCode на ВМ может требовать авторизацию | Проба слоёв 1-2 возможна без логина (поля авторизации — тоже текст-поля); сообщение-поле — только после логина: чекпойнт владельца, не тихое обходное решение |
| Build artifacts | Зеркало `~/goswitch` на ВМ (существует, живая проверка) отстало от текущей ветки; `~/go/bin/{goswitchd,goswitchctl}` — старой сборки | rsync-рецепт + gap-сборка текущей ветки (прецедент: НОВЫЙ release-dir не нужен — стендовая сборка, не установка владельца) |

**Канонический вопрос:** после всех правок репо на ВМ останутся: snap/flatpak Telegram, копия
ZCode, override .desktop, конфиг стенда, изменённые sources ввода — каждое с documented
rollback; на хосте — ничего (D-05).

## Common Pitfalls

### Pitfall 1: «Переключение работает» принято за «автокоррекция работает»

**What goes wrong:** проба фиксирует зелёный хоткей-флип в Telegram и фаза закрывает A11Y-02
без a11y-дерева.
**Why:** переключение — слой IBus (не нужен a11y); автокоррекция — входящий GetRole. Слова
владельца почти наверняка про первый слой.
**How to avoid:** трёхслойная проба (Pattern 1); ряд таблицы без всех трёх слоёв — неполный.
**Warning signs:** в ряде пробы есть key-trace флипа, но нет GetRole-ответа.

### Pitfall 2: AT-SPI-оракул ВМ мёртв, а вердикт написан по его молчанию

**What goes wrong:** зонд не видит фокус-флаги/текст поля и проба объявляет сборку
недостижимой — хотя виноват оракул, не приложение.
**Why:** задокументированные качели: диагноз ci-runner 2026-10-27 «AT-SPI-оракул на ВМ мёртв»
(софтверный рендер, глухие zenity-фикстуры) против живых рядов Фазы 10 (дерево ожило после
GTK_MODULES-активации gnome-terminal и RegisterEvent демона, 153 StateChanged)
[CITED: docs/ci-runner.md:312-330; CITED: docs/adr/adr-008-spike-log.md term-word-vte-mechanics].
**How to avoid:** калибровочный ряд ПЕРВЫМ: известный GTK-апп (gnome-terminal) обязан
показывать дерево+роли; если глух — оживить (GTK_MODULES через activation-environment,
прецедент 10-05) до любых Telegram/ZCode рядов; вердикт по сборке — только при живом
контрольном ряде.
**Warning signs:** контрольный ряд глух при живой шине (busctl GetAddress отвечает — живая
проверка этой сессии: шина на ВМ жива, IsEnabled=true).

### Pitfall 3: Electron-флаг применён к уже запущенному приложению

**What goes wrong:** override `.desktop` создан, ZCode «перезапущен» крестиком (а процесс жил
в трее), дерево не появилось, рецепт объявлен неработающим.
**Why:** флаги Chromium/Electron читаются только при старте процесса (закреплено SPEC §4.5:
«проверяют ключ однократно при своём старте»); A11Y-03 требует явно задокументировать это.
**How to avoid:** в пробе — проверка pid-идентичности до/после рестарта (`pgrep -af zcode`) и
`/proc/<pid>/cmdline` с флагом; в рецепте — явный «полностью выйдите и запустите снова».
**Warning signs:** environ/cmdline нового процесса не содержит флага.

### Pitfall 4: Плашка «режим экранного диктора» в Telegram принята за поломку

**What goes wrong:** после включения a11y-стола Telegram показывает плашку диктора (Qt-мост
реагирует на ScreenReaderEnabled-семантику) — проба/рецепт тракует её как дефект.
**Why:** это документируемый побочный эффект включённого toolkit-accessibility, не регресс
goswitch.
**How to avoid:** снять плашку скриншотом в ряде пробы (virsh screenshot), включить в A11Y-03
как ожидаемый побочный эффект; форму/условия появления фиксирует живая проба
[ASSUMED A5 — не подтверждено документально в этой сессии].
**Warning signs:** плашка на скриншоте без строки в таблице пробы.

### Pitfall 5: Сеть ВМ тихо режет рунг лестницы

**What goes wrong:** `flatpak install` или скачка tarball падают по таймауту, ряд записан как
«сборка недостижима».
**Why:** асимметрия (живые проверки): flathub.org и telegram.org с ВМ недоступны (CURL-FAIL),
dl.flathub.org — 200; значит flatpak работает, но любые команды/браузинг против flathub.org —
нет; tarball-рунг напрямую невозможен.
**How to avoid:** карта каналов в Environment Availability — часть плана; tarball стейджится
через хост (скачка + scp); первый `flatpak install` — отдельный verify-шаг с выводом, не
молчаливый.
**Warning signs:** curl-таймауты в логе установки при живом apt.

### Pitfall 6: Затирание запаркованного состояния стенда без отката

**What goes wrong:** подъём стенда оставляет sources-takeover/idle-lock-off включёнными, ВМ
перестаёт быть «запаркованной» после фазы.
**Why:** recipe 10-01 меняет gsettings-состояние ВМ.
**How to avoid:** откаты — дословно из venue-verdict: sources `[('xkb','us')]`,
lock-enabled=true/idle-delay=300 [CITED: docs/adr/adr-008-spike-log.md venue-verdict 14:20];
итоговое состояние ВМ — строка SUMMARY.
**Warning signs:** в плане нет задачи-отката стенда (или осознанного «оставляем поднятым»).

### Pitfall 7: Единственный свидетель — самозахват (witness pinned to gnome-shell)

**What goes wrong:** зонд отвечает `gnome-shell:WINDOW:chars=-1` — фокус-свидетель прибит к
оболочке, поле приложения не читается, хотя ввод работает.
**Why:** известная ловушка стенда (Фаза 9: свидетель всегда gnome-shell; убран из маркеров).
**How to avoid:** фокус-свидетель — НЕ маркер; слои 1 и 3 (key-trace, счётчики) не зависят от
него; зонд ищет приложение в дереве по имени, а не по фокус-флагу.
**Warning signs:** все ряды дают gnome-shell.

### Pitfall 8: «Видимость» по имени приложения без проверки роли поля

**What goes wrong:** «telegram» появился в desktop-children → рецепт написан; но поле ввода
сообщает роль password/нестандартную — автокоррекция молчит по D-53.
**Why:** gate демона — роль текст-класса с известной identity, не факт присутствия в дереве.
**How to avoid:** self-check и ряды пробы печатают роль сфокусированного поля; зелёный
вердикт — роль класса 61/79/94 (список — verbatim ниже) И живой счётчик слоя 3.
**Warning signs:** в ряде нет числа роли.

## Code Examples

### Оракул демона — токены статуса (wire-канон, читается goswitchctl status)

```go
// Source: internal/ctlsvc/ctlsvc.go:252-296 (renderStatus) — прочитано живьём этой сессией
tokens := []string{
        "version=" + st.Version,
        "mode=" + st.Mode,
        "corrections_done=" + strconv.Itoa(st.CorrectionsDone),
        "corrections_skipped=" + strconv.Itoa(st.CorrectionsSkipped),
}
...
        "autocorrect_enabled="+strconv.FormatBool(st.AutoCorrectEnabled),
        "autocorrect_fired="+strconv.Itoa(st.AutoCorrectFired),
        "autocorrect_abstained="+strconv.Itoa(st.AutoCorrectAbstained),
        "ac_role_errors="+strconv.Itoa(st.AutoCorrectRoleErrors),
        "corr_dst_guarded="+strconv.Itoa(st.CorrDstGuarded),
        "sound_enabled="+strconv.FormatBool(st.SoundEnabled),
```

Дельты до/после синтетического корпуса — машинная истина слоя 3.

### Роли гейта демона — сравнение по номеру, дословно

```go
// Source: internal/appid/appid.go:85-96 — прочитано живьём этой сессией
const (
	RolePasswordText uint32 = 40 // password text — the autocorrect policy never corrects it
	RoleTerminal     uint32 = 60 // terminal — never corrected
	RoleText         uint32 = 61 // text box / text — correctable text input
	RoleEntry        uint32 = 79 // entry — correctable text input
	RoleDocumentText uint32 = 94 // document text — correctable text input
)
```

Self-check печатает эти номера; зелёный вердикт поля = роль ∈ {61, 79, 94}.

### Двухразрешительная модель — инвариант TERM-04 (контекст, почему a11y = разрешение)

```text
Source: docs/SPEC.md §11 (ревизия 2026-10-08, дословно):
«Автокоррекция никогда не армируется в контексте терминала (роль
AT-SPI 60 или сконфигурированный терминальный движок/identity). Вето
проверяется в точке армирования автокоррекции ДО запуска детектора;
оно не выражается через `autocorrect.apps_blocklist` и не зависит от
наблюдаемости identity.»
```

Для A11Y следствие: невидимая роль (нет дерева) = молчание fail-closed со слагом
`role-unknown`/`role-timeout` [VERIFIED: internal/session/actor.go:234-235 —
`acReasonRoleAmbiguous = "role-ambiguous"`, `acReasonRoleUnknown = "role-unknown"`];
неизвестная identity + роль 61 = `role-ambiguous` (CR-01) — то есть даже «роль есть, личности
нет» молчит: слой 3 зелёный только при полной паре.

### Скелет self-check (форма focus_helper, не финальный код)

```python
#!/usr/bin/python3
# Source: контракт по прецеденту test/e2e/focus_helper.py (выход 0/1/2, /usr/bin/python3)
import sys
import gi
gi.require_version("Atspi", "2.0")
from gi.repository import Atspi

TEXT_ROLES = {61, 79, 94}   # RoleText/RoleEntry/RoleDocumentText — номера, не имена

def find_app(desktop, needle):
    for i in range(desktop.get_child_count()):
        app = desktop.get_child_at_index(i)
        name = (app.get_name() or "")
        if needle.lower() in name.lower():
            return app
    return None

def has_text_field(node, depth=0):
    # присутствие текст-поля в поддереве; роли по номерам
    ...
```

Плановая задача доводит до контракта 0/1/2; скрипт запускается строкой из доков
(напр. `/usr/bin/python3 scripts/a11y-selfcheck.py zcode`).

### Установочные команды рунгов (сеть ВМ учтена)

```bash
# Рунг snap (store доступен — живая проверка):
ssh ... stand@192.168.122.150 'sudo snap install telegram-desktop'

# Рунг flatpak (apt + репо-CDN; flathub.org-сайт недоступен — не нужен):
sudo apt install -y flatpak
flatpak remote-add --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
flatpak install -y flathub org.telegram.desktop

# Рунг tarball (стейджинг через хост — telegram.org с ВМ недоступен):
# host: curl -L https://telegram.org/dl/desktop/linux -o /tmp/telegram.tar.xz
# host: scp /tmp/telegram.tar.xz stand@192.168.122.150:/tmp/
# vm:   tar -xJf /tmp/telegram.tar.xz -C ~/Applications

# ZCode на ВМ (копия сборки владельца):
# host: rsync -a /opt/ZCode/ stand@192.168.122.150:/opt/ZCode/   # 811 М
# vm:   ~/.local/share/applications/zcode.desktop (override ниже)
```

### ZCode-override (шаблон уже в CONFIG.md; системный .desktop хоста прочитан живьём)

```ini
# Системный (host, dpkg-owned): Exec=/opt/ZCode/zcode %U ; StartupWMClass=ZCode
# User-local override ~/.local/share/applications/zcode.desktop:
[Desktop Entry]
Type=Application
Name=ZCode
Exec=env ACCESSIBILITY_ENABLED=1 /opt/ZCode/zcode --force-renderer-accessibility %U
Icon=zcode
StartupWMClass=ZCode
```

Источник шаблона: docs/CONFIG.md:373+ (раздел «Ручной рецепт», уже в репо); флаг —
документированный Chromium-переключатель [CITED: electronjs.org/docs accessibility;
CITED: github.com/microsoft/vscode#84833]; `ACCESSIBILITY_ENABLED=1` — env-форма того же
[по прецеденту STACK.md; ELECTRON-доки].

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Вердиект AppArmor snap (7.2.9 rev 7150) как текущий факт | Пере-проба на 7.3.3 rev 7175 (хост обновился) | живые проверки этой сессии | D-01 не формальность: между версиями сменился и snap, и rev |
| AT-SPI-сокет в abstract namespace | `/run/user/<uid>/at-spi/bus` | at-spi2-core 2.39.1 (snapd 2.55 подстроился) | snap-приложения МОГУТ исходяще подключаться к шине — не путать с входящими вызовами [CITED: forum.snapcraft.io/t/27274] |
| `GTK_MODULES=gail:atk-bridge` в рецептах | GTK3+ автозагрузка моста от toolkit-accessibility | давно | В рецепты НЕ включать (STACK: «GTK2-era») |

**Deprecated/outdated:** lore «gsettings current read-only» (снят живой пробой Фазы 0);
lore «snap можно подключить интерфейсом» — не существует интерфейса входящих AT-SPI-вызовов
(docs a11y-interface = 404, живая проверка; forum 27274 — только исходящий сокет).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Flatpak Telegram a11y-достижим: runtime (org.gnome.Platform 51) несёт at-spi, песочница передаёт a11y-шину через прокси (не finish-arg — их в manifest нет) | Standard Stack / Pitfalls | Если прокси не пропускает входящие вызовы — flatpak-рунг красный, лестница D-03 идёт дальше; фаза имеет на это ответ |
| A2 | ZCode переносится на ВМ rsync-ом /opt/ZCode (самодостаточный Electron-каталог) и запускается без установки | Validation | Если требуются пост-установочные шаги/авторизация — чекпойнт; fields авторизации тоже текст-поля, слой 2 измерим |
| A3 | Telegram-проба слоёв 1-2 возможна на экране авторизации (без логина); поле сообщения — только после QR-логина владельцем | Runtime State | Логин-чекпойнт для финального «в поле сообщения»; видимость решается и без него |
| A4 | Роль поля Telegram (Qt) и ZCode (Electron) окажется в текст-классе 61/79/94 | Patterns/Examples | Если роль иная (напр. нестандартная) — автокоррекция молчит по D-53; это результат пробы, не провал исследования (скрипт печатает роли, не предполагая) |
| A5 | Плашка «режим экранного диктора» появляется в Telegram при включённом a11y-столе (форма/условия — пробой) | Pitfall 4 / A11Y-03 | Если плашки нет — пункт доки упрощается; текст A11Y-03 требует её задокументировать при наличии |
| A6 | «Переключение работает» у владельца = слой IBus (хоткей), согласуется с AppArmor-вердиктом | Summary / Pitfall 1 | Проба D-01 прямо измеряет оба слоя — неверность интерпретации не влияет на машины |
| A7 | ZCode на хосте не имеет user-local .desktop override (живая проверка: файл отсутствует) — рецепт для владельца будет первым override | Environment | Если override появится до фазы — рецепт становится апдейтом, не созданием |

## Open Questions (RESOLVED — решения вшиты в планы фазы, probe-first дизайн D-01 нетронут)

1. **Пропускает ли flatpak-песочница входящие GetRole к объектам Telegram?**
   - What we know: manifest без a11y-talk-имен (gh api, живое чтение); a11y-передача в flatpak не через finish-args; web-источники противоречивы (старые «bus can't be used», новые «проксируется»).
   - What's unclear: поведение конкретно org.gnome.Platform 51 + Qt6-мостом на Ubuntu 24.04.
   - Recommendation: не исследовать web дальше — D-02 проба даёт ответ машиной за одну сессию; это дешевле любого дополнительного источника.
   - RESOLVED (планирование 2026-11-03): web не исследуется — отвечает живая проба: план 13-01 Task 3, flatpak-рунг тремя слоями ×2 зелёных (при красном — лестница D-03); механизм различения исходящих/входящих вызовов закреплён в слое 2 (Pitfall «регистрация на шине ≠ входящие GetRole»).
2. **Нужен ли владельцу QR-логин Telegram на ВМ для приёмки «в поле сообщения»?**
   - What we know: слои 1-2 измеримы без логина; A11Y-02 говорит «входное поле реально видно … автокоррекция в нём работает».
   - Recommendation: плановый чекпойнт после пробы: если зелёный на авторизационном поле — предъявить владельцу решение (достаточно ли) или попросить QR-логин для сообщения-поля.
   - RESOLVED (планирование 2026-11-03): плановая развилка вместо тихого обхода — план 13-01 Task 3 шаг (6): слои 1-2 на экране авторизации без логина (A3); если слой 3 требовал поле СООБЩЕНИЯ — честная строка «сообщение-поле: без QR-логина» в итог-ряд + вопрос владельцу в checkpoint (б)/(в) или SUMMARY (a); молчаливый обход запрещён.
3. **Достаточно ли «зелёного по живому дереву» для зачёта ZCode (D-05: без стола владельца)?**
   - What we know: критерий 1 роадмапа требует живое дерево + работающую автокоррекцию на ВМ.
   - Recommendation: зачёт = живой ряд на ВМ (дерево + счётчики + readback если доступен); residual (субъективное «работает у меня») — вне фазы по D-05.
   - RESOLVED (планирование 2026-11-03): зачёт = живой ряд на ВМ — план 13-02 Task 1: RED→GREEN пара ×2 (exit 0 + номер роли + слой 3 дельтами счётчиков + pid/cmdline-доказательство); субъективное «работает у меня» — вне фазы по D-05, residual — UAT end-of-phase.


## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| ВМ goswitch-stand (ssh, BatchMode) | все проверки (D-05) | ✓ | живая проверка; ядро 6.8.0-147-generic | — |
| virsh send-key / screenshot (хост) | драйв клавиатуры, визуальные плашки | ✓ | goswitch-stand running (virsh list живой) | ydotool на ВМ |
| snapd + store (ВМ) | snap-рунг D-02 | ✓ | 2.77.1; `snap find telegram-desktop` → 7.3.3 (живые) | — |
| flatpak (ВМ) | flatpak-рунг D-02/D-04 | ✗ → ставится фазой | apt-кандидат 1.14.6-1ubuntu0.1 | — |
| dl.flathub.org (repo-CDN) | flatpak install | ✓ | HTTP 200 (живой) | — |
| flathub.org (сайт/API) | — | ✗ | CURL-FAIL (живой) | не нужен для install |
| telegram.org (ВМ) | tarball-рунг D-03 | ✗ | CURL-FAIL (живой) | стейджинг: скачка на хосте + scp |
| github.com (ВМ) | резервный канал дистрибутивов | ✓ | HTTP 200 (живой) | — |
| python3-gi + Atspi (ВМ) | зонд, self-check | ✓ | import Atspi OK (живой) | — |
| python3-gi (хост) | справочно (self-check юзера) | ✓ | 3.48.2 / gir 2.52.0 (dpkg живой) | — |
| a11y-шина ВМ (org.a11y.Bus) | проба слоя 2 | ✓ | GetAddress отвечает; IsEnabled=true; toolkit-accessibility=true (живые) | калибровка при глухом дереве (Pitfall 2) |
| Зеркало ~/goswitch + ~/go/bin (ВМ) | gap-сборка | ✓ (устарело) | rsync-рецепт 10-08 Task 2 | — |
| Место на ВМ | snap+flatpak+ZCode (~3-4 ГБ) | ✓ | 28 Г свободно, 6.1 Г RAM (живые) | — |
| goswitchd (ВМ) | оракул слоя 1/3 | ✗ inactive (запаркован) | подъём recipe 10-01/10-08 | — |
| Конфиг стенда ~/.config/goswitch | autocorrect on для пробы | ✗ удалён | пересоздать (для слоя 3 autocorrect.enabled обязателен) | — |

**Missing dependencies with no fallback:** none — всё недостающее ставится фазой на ВМ.
**Missing dependencies with fallback:** telegram.org с ВМ → host-стейджинг; flathub.org-сайт → не требуется.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `-race` — фаза Go-кода не пишет, `mise run ci` обязан оставаться зелёным; живые пробы — python3-gi зонд + счётчики демона на ВМ (mise-задачи NOT in ci по прецеденту e2e-*) |
| Config file | `mise.toml` (ci/lint/test/e2e-*), `.golangci.yml` v2 |
| Quick run command | `mise exec -- go build ./... && go test -race -count=1 ./internal/...` |
| Full suite command | `mise run ci` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| A11Y-02 | Подъём стенда + калибровка a11y-оракула (GTK-контроль жив) | e2e (стенд) | ssh-recipe 10-01/10-08 + `busctl GetAddress` + калибровочный ряд | ❌ Wave 0 (recipe в журнале) |
| A11Y-02 | Трёхслойная проба snap 7.3.3 (key-trace / GetRole / счётчики) ×2 зелёных | e2e (стенд) | ряд в adr-008-spike-log.md + `goswitchctl status` дельты | ❌ Wave 0 |
| A11Y-02 | Трёхслойная проба flatpak (или лестница D-03 до первого зелёного) | e2e (стенд) | то же + `flatpak install` verify-шаг | ❌ Wave 0 |
| A11Y-01 | ZCode на ВМ: override + рестарт (pid/cmdline с флагом) + дерево + слой 3 | e2e (стенд) | `/proc/<pid>/cmdline` grep + зонд + счётчики | ❌ Wave 0 |
| A11Y-03/A11Y-01 | Self-check: exit 0 при видимом поле, 1 при невидимом, 2 при отсутствии; человекочитаемый вердикт | e2e (стенд; headless-RED невозможен — прецедент focus_helper) | `/usr/bin/python3 scripts/a11y-selfcheck.py <app>` на ВМ (позитив+негатив) | ❌ Wave 0 (скрипт — предмет фазы) |
| A11Y-03 | Рецепты в доках: разделы несут все обязательные пункты (побочка-диктор, только-на-запуске, restart+self-check, residual-risk PRIMARY) | grep-гейты (прецедент ADR grep-gates) | grep-набор по README.md/README.en.md/docs/CONFIG.md | ❌ Wave 0 |
| A11Y-02 | Daemon-fallback-ветка (если все рунги красные) | manual | отдельная приёмка с владельцем (требование A11Y-02) | n/a (условная) |

### Sampling Rate

- **Per task commit:** `mise run ci` (зелёная итерация — кода почти нет, но гейт держим).
- **Per wave merge:** `mise run ci` + живые ряды волны на ВМ (два подряд зелёных для
  пробных вердиктов — SPEC §7.2 применённо к пробе).
- **Phase gate:** все ряды пробы датированы в adr-008-spike-log.md; self-check демонстрируется
  на ВМ (позитив и негатив); grep-гейты доков зелёные; `mise run ci` зелёный — до
  `/gsd:verify-work`.

### Machine-checkable vs owner acceptance (nyquist: разделение honestly)

- **Машинное на ВМ:** все три слоя пробы по каждому рунгу; калибровка оракула; поведение
  self-check (коды возврата); grep-гейты доков; откаты стенда.
- **Owner acceptance (human):** (1) Telegram: достаточность зелёного рунга ИЛИ выбор
  демон-fallback (A11Y-02 прямо требует решения владельца при fallback; QR-логин для
  сообщение-поля — по ситуации Open Question 2); (2) демон-fallback — отдельная приёмка с
  владельцем; (3) текст рецептов — владелец читает как пользователь (UAT end-of-phase).

### Wave 0 Gaps

- [ ] `scripts/a11y-selfcheck.py` (+ при необходимости .sh-обёртка) — D-06, весь предмет кода фазы
- [ ] конфиг стенда (autocorrect enabled) + stand-env.sh — подъём, recipe 10-01/10-08
- [ ] калибровочный ряд a11y-оракула — до любых вердиктов сборок
- [ ] таблица пробы (форма рядов) в adr-008-spike-log.md — секция Фазы 13
- [ ] grep-гейты доков (перечень обязательных пунктов A11Y-03)

*(Если не все: существующая инфраструктура покрывает оракулы демона и протокол зонда — новые только скрипт и ряды.)*

## Security Domain

security_enforcement включён (config: `security_enforcement: true`, ASVS L1). Фаза не пишет
daemona и не трогает конфиг-парсер; поверхность — скрипт + доки + установки на ВМ.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — |
| V3 Session Management | no | — |
| V4 Access Control | no | — |
| V5 Input Validation | yes (скрипт) | Self-check не передаёт пользовательский текст в argv команд (только фиксированные литералы + чтение AT-SPI); имена приложений — только для сопоставления в python, не для shell-интерполяции; argv-дисциплина SPEC §4.5 — прецедент |
| V6 Cryptography | no | — |

### Known Threat Patterns for {python3-gi скрипт + стенд}

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Подмена identity на a11y-шине (скрипт видит «не то» приложение) | Spoofing | Принятый риск 2026-10-31 (SPEC §11): сеанс локален/однопользователен; скрипт — диагностика, не решение о доступе; вердикт сопровождается pid-строкой |
| Инжект через имена приложений в вывод/лог | Tampering | Скрипт печатает имена как данные (никаких eval/формирование команд из имён); exit-коды — единственный машинный контракт |
| Стейджинг tarball с хоста (недоверенные данные) | Tampering | Архив распаковывается в новый пустой каталог, запуск — указанием полного пути бинарника; без запуска интерпретаторов из архива (правило среды) |

## Sources

### Primary (HIGH confidence)

- Живые проверки этой сессии (ВМ ssh + хост, пассивные чтения): состояние ВМ (snap/flatpak/telegram/zcode/goswitchd/конфиг/a11y-шина), сетевая карта (snap store/flathub CDN/flathub.org/telegram.org/github.com), dpkg python3-gi (хост+ВМ), хост-snap telegram 7.3.3 rev 7175, /opt/ZCode (811 М) + системный zcode.desktop, flathub-manifest через gh api, virsh list
- `internal/appid/appid.go:85-96` — роли гейта (прочитан, процитирован verbatim)
- `internal/ctlsvc/ctlsvc.go:252-296` — токены статуса (прочитан)
- `internal/session/actor.go:234-235, 3685-3786, 4017-4075` — слаги/граница/autoConfirm (прочитаны)
- `test/e2e/focus_helper.py:1-110` — контракт зонда (прочитан)
- `docs/SPEC.md` §4.5/§11 — a11y-магия, двухразрешительная модель, TERM-04 (прочитаны)
- `13-CONTEXT.md` — D-01..D-07 verbatim; слова владельца verbatim

### Secondary (MEDIUM confidence)

- `.planning/research/STACK.md` §(c), Alternatives, What NOT to Use — AppArmor-вердикт живой пробы 2026-10-08 (7.2.9 rev 7150), QT_ACCESSIBILITY в environ, план лестницы
- `docs/ci-runner.md:312-330` — диагноз «AT-SPI-оракул на ВМ мёртв» (2026-10-27)
- `docs/adr/adr-008-spike-log.md` — venue-ряды/recipe стенд-механики (10-01/10-05/10-08), откаты
- [forum.snapcraft.io/t/27274](https://forum.snapcraft.io/t/accessibility-support-failing-for-desktop-snaps-on-recent-distro-releases/27274) — desktop-legacy даёт snap исходящий a11y-сокет (snapd 2.55+); входящие вызовы не рассматриваются
- [flathub/org.telegram.desktop.yml](https://github.com/flathub/org.telegram.desktop) (gh api) — runtime org.gnome.Platform 51, finish-args без a11y-имён
- [docs.flatpak.org sandbox-permissions](https://docs.flatpak.org/en/latest/sandbox-permissions.html) — a11y-сокета в документированном наборе разрешений НЕТ
- [Electron accessibility docs](https://electronjs.org/docs/latest/tutorial/accessibility), [VS Code #84833](https://github.com/microsoft/vscode/issues/84833) — флаг/ACCESSIBILITY_ENABLED, старт-семантика

### Tertiary (LOW confidence — сознательно не углублялись, проба решает)

- Web-механика flatpak a11y-прокси ([flatpak#4394](https://github.com/flatpak/flatpak/issues/4394), [easyeffects#178](https://github.com/wwmm/easyeffects/issues/178) — противоречивые эпохи) → [ASSUMED A1]
- Синтезированный поиском «snapd a11y interface auto-connected» — НЕ подтверждён (docs-страница 404); единственное подтверждённое — исходящий сокет desktop-legacy
- Плашка диктора Telegram — проектное ожидание, форма устанавливается пробой [ASSUMED A5]

## Metadata

**Confidence breakdown:**
- Оракулы и seams демона: **HIGH** — прочитаны в коде этой сессии, процитированы verbatim
- Стенд и сеть: **HIGH** — живые проверки этой сессии (ssh/virsh/curl/snap find/dpkg)
- AppArmor-контекст snap: **MEDIUM-HIGH** — живой вердикт 2026-10-08 + форумное различение исходящих/входящих; обновление snap 7.2.9→7.3.3 честно требует пере-пробы (D-01)
- Flatpak a11y-механика для Telegram: **LOW-MEDIUM (ASSUMED A1)** — противоречивые web-источники; фаза по дизайну probe-first, ответ даёт машина
- Docs-карта: **HIGH** — grep/чтение файлов этой сессии

**Research date:** 2026-11-03
**Valid until:** ~2026-11-17 (стабильно: локальный стек; главный риск дрейфа — обновление snap telegram-desktop на хосте/ВМ и состояние запаркованного стенда)
