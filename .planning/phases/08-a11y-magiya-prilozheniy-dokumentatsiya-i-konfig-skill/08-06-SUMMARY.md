---
phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill
plan: 06
subsystem: docs
tags: [docs, config-reference, russian-first, a11y, spec-audit, spec-delta-form, restart-semantics, tdd-free]

# Dependency graph
requires:
  - phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill (планы 08-01..08-05)
    provides: "SPEC §4 ревизия 2026-10-05 (форма датированных ревизий) + фактическое поведение: схема A11y (08-02), Reconciler (08-03), снапшот/откат (08-04), проводка актора (08-05)"
provides:
  - "docs/CONFIG.md — русскоязычный справочник конфигурации (D-8-9): весь повествовательный текст по-русски, структура и технические литералы сохранены"
  - "Таблица ключей, сверенная построчно с internal/config/config.go (D-8-8): имена == yaml-теги, дефолты == Defaults(), диапазоны == валидации; формат байт-строгий (5 колонок, одна строка на ключ) — источник генератора 08-08"
  - "Секция a11y задокументирована (D-8-6/D-8-1): «ровно семь секций», строки a11y.enabled/a11y.apps, пример, абзац поведения (старт+reload, restart-семантика, снятие запрещено, uninstall-revert), усиленная Privacy-оговорка, ручной рецепт per-app оверрайда"
  - "docs/SPEC.md — сверка-ревизия 2026-10-05 в форме 08-01: полный обход поведенческих утверждений (меню/тумблеры, звуки, blocklist, adopt+watch, «где молчит», числа §4) против дерева 08-02..08-05; одна ин-плейс правка заведомо ложного факта (install не пишет конфиг), отражённая в сверка-блоке; аудит-трейл цел"
affects: [08-07 (README-реструктура сверяется с финальным CONFIG.md), 08-08 (skillgen парсит таблицу CONFIG.md)]

# Actuals (#2632) — pairs with the plan's estimate to calibrate future estimates.
actuals:
  tokens: 11316    # chars/4 over the realized diff (45265 chars across docs/CONFIG.md + docs/SPEC.md)
  tasks: 3
  commits: 3      # MEASURED: git rev-list --count gsd-plan-head-before-08-06..HEAD
plan_head_before: c35da7bf0d9eef46e3fc3d53483af7b8dddafd4a

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Русский-первый перевод пользовательского дока: технические литералы/имена ключей/заголовок таблицы не переводятся (заголовок | Key | Type | Default | Range / vocabulary | Meaning | пиннут должен_haves) — проза вокруг русская"
    - "Сверка дока с кодом: каждая строка таблицы == yaml-тег + Defaults() + validate(); расхождение = правка документа, дефект поведения = todo (в этом плане дефектов не найдено)"

key-files:
  created: []
  modified:
    - docs/CONFIG.md
    - docs/SPEC.md
    - .planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/deferred-items.md

key-decisions:
  - "Пример-литералы плана для a11y.apps («zcode» находит «org.zdev.Zed») заменены регистр- и подстрок-корректными («chromium» находит «org.chromium.Chromium») — пример противоречил чувствительности к регистру, пиннутой тем же планом в этой же строке (Rule 1)"
  - "SPEC-сверка: единственное расхождение — «install лишь поставляет секцию в конфиг»: фактически установщик конфиг не пишет (решение 04-02, adopt+watch); исправлено ин-плейс + сверка-блок; дефектов поведения нет — новые todo не заводились"
  - "§4.4 (flip через input-sources) оставлен как есть: вне перечня сверки todo-первоисточника; канонический контракт флипа записан в ADR-006 (решение владельца, не оспаривается)"
  - "Полные YAML-примеры дополнены выключенной секцией a11y (зеркало Defaults()) — «полный файл» обязан нести все семь секций"

patterns-established:
  - "Сверка-ревизия SPEC по форме 08-01: HTML-комментарий-пометка + датированный абзац «Сверка (дата, фаза)» с перечнем подтверждённого и каждой ин-плейс правкой (раздел, что утверждалось, что фактически, источник истины)"

requirements-completed: ["D-8-8", "D-8-9", "D-8-6"]

coverage:
  - id: D1
    description: "docs/CONFIG.md русскоязычен (D-8-9): повествовательный текст по-русски, 19 прежних ключей таблицы на месте и сверены с config.go (плюс две сверка-правки: tap_key — только голая клавиша; autocorrect_event — пустое значение читается как дефолт); операционные доки не тронуты"
    requirement: "D-8-9"
    verification:
      - kind: other
        ref: "плановый гейт: head -30 docs/CONFIG.md | grep -qP «[А-Яа-яЁё]{3,}» → RU-FIRST"
        status: pass
      - kind: other
        ref: "плановый гейт: 19 ключей grep по docs/CONFIG.md → ALL-KEYS-PRESENT; 27 строк «| `» ≥ 15"
        status: pass
      - kind: other
        ref: "плановый гейт: git status по docs/ACCEPTANCE.md docs/ci-runner.md SECURITY.md — пусто (операционные доки чисты)"
        status: pass
    human_judgment: true
    rationale: "Качество русской прозы и полнота перевода — суждение владельца на verify-гейте; автоматика доказывает только структуру (гейты, ключи, формат)"
  - id: D2
    description: "Секция a11y в CONFIG.md (D-8-6/D-8-1): «ровно семь секций», две строки a11y.enabled/a11y.apps (false/[]/потолок 64 — байт-согласовано со схемой 08-02), пример полного документа с активной секцией, абзац поведения (две ручки, старт+reload идемпотентно, restart-семантика, снятие запрещено, uninstall-revert only-if-present), усиленная Privacy-оговорка, подраздел «Ручной рецепт: per-app запуск с a11y-флагами» с фиксированным шаблоном и предупреждениями"
    requirement: "D-8-6"
    verification:
      - kind: other
        ref: "плановые гейты: A11Y-DOC-OK (семь секций/a11y.enabled/a11y.apps/toolkit-accessibility); SEMANTICS-OK (следующего|следующем + uninstall + force-renderer-accessibility); ROWS-SEVEN-SECTIONS (29 ≥ 21); TWO-ROWS (ровно 2 a11y-строки)"
        status: pass
      - kind: other
        ref: "сверка значений со схемой 08-02: A11y{Enabled false, Apps nil}, maxA11yApps=64, Active() = enabled && непустой список — internal/config/config.go"
        status: pass
    human_judgment: true
    rationale: "Честность restart-семантики и рецепта (не обещает лишнего, не скрывает ограничений) — суждение владельца; гейты доказывают наличие и структуру"
  - id: D3
    description: "SPEC-сверка (D-8-8): сверка-ревизия 2026-10-05 в форме 08-01 с полным перечнем обхода (меню/тумблеры, звуки, blocklist D-53, adopt+watch, «где молчит», числа/дефолты/потолки §4); единственная ин-плейс правка (install не пишет конфиг) отражена в сверка-блоке; аудит-трейл цел; дефектов поведения нет"
    requirement: "D-8-8"
    verification:
      - kind: other
        ref: "плановый гейт: grep 2026-09-27/2026-10-04/D-53 docs/SPEC.md → AUDIT-TRAIL-OK"
        status: pass
      - kind: other
        ref: "плановый гейт: git diff HEAD --name-only | grep .go$ → пусто (NO-GO-IN-DIFF)"
        status: pass
      - kind: other
        ref: "плановый гейт: a11y.enabled/toolkit-accessibility в SPEC и CONFIG.md → S4-CONFIG-SYNC"
        status: pass
    human_judgment: true
    rationale: "Полнота обхода поведенческих утверждений и корректность ин-плейс правки против записанных решений владельца — суждение владельца на verify-гейте"
  - id: D4
    description: "Зелёная итерация mise run ci на выходе плана (build + vet + lint + test -race, exit 0); документы не ломают CI-гейты (включая будущий golden-парсер 08-08: таблица 5 колонок, ровно 2 a11y-строки)"
    verification:
      - kind: other
        ref: "mise run ci → MISE_EXIT=0 (финальный прогон задачи 3)"
        status: pass
    human_judgment: false

# Metrics
duration: 18min
completed: 2026-10-05
status: complete
---

# Phase 8 Plan 06: CONFIG.md русский-первый + a11y-секция + SPEC-сверка Summary

**docs/CONFIG.md переведён на русский и построчно сверён со схемой config.go, дополнен секцией a11y (семь секций, честная restart-семантика, рецепт per-app оверрайда); docs/SPEC.md прошёл полную сверку против фактического поведения v1.1.0 в форме 08-01 — единственное расхождение (install не пишет конфиг) исправлено с целым аудит-трейлом.**

## Performance

- **Duration:** 18 min
- **Started:** 2026-10-05T18:30:58Z
- **Completed:** 2026-10-05T18:48:35Z
- **Tasks:** 3
- **Files modified:** 2 (docs/CONFIG.md +307/−145, docs/SPEC.md +49/−2) + 1 planning-артефакт (deferred-items.md)

## Accomplishments

- docs/CONFIG.md — русскоязычный справочник (D-8-9): заголовки, вводные, абзацы-поведения, миграционные заметки, Privacy; технические литералы, имена ключей и заголовок таблицы (`| Key | Type | Default | Range / vocabulary | Meaning |`) не переводились — формат таблицы байт-строгий (5 колонок, одна строка на ключ) как источник генератора 08-08
- Таблица сверена построчно с internal/config/config.go (D-8-8): все 19 прежних ключей подтверждены (имена == yaml-теги, дефолты == Defaults(), диапазоны == валидации); две сверка-правки документа — `hotkeys.tap_key` дополнен правилом «только голая клавиша без модификаторов» (errTapKeyBare), `sound.autocorrect_event` дополнен «пустое значение читается как встроенный дефолт» (EffectiveAutocorrectEvent)
- Секция a11y задокументирована (D-8-6/D-8-1): «ровно семь секций», строки `a11y.enabled` (bool, `false`) и `a11y.apps` (list, `[]`, потолок 64 RE2, подстрока/анкеровка/регистр/порядок), пример полного документа с активной секцией, абзац поведения (демон применяет при старте и на hot reload идемпотентно; приложения подхватывают при следующем запуске — одноразовая стартовая проверка Chromium/Electron; снятие ключа не предусмотрено; откат — только `goswitchctl uninstall` из снапшота, restore only-if-present), дефолтный пример дополнен выключенной секцией (зеркало Defaults())
- Privacy усилен (одноразовая стартовая проверка — работающее приложение ключ не подхватит) + подраздел «Ручной рецепт: per-app запуск с a11y-флагами»: фиксированный шаблон override в ~/.local/share/applications, предупреждения о сталости Exec/перекрытии системного файла, путь удаления, snap-оговорка; никакой автоматизации
- docs/SPEC.md — сверка-ревизия 2026-10-05 в форме 08-01: обход подтвердил меню/тумблеры («Автокоррекция»→autocorrect.enabled, «Звук»→sound.enabled, adopt+watch через EnsureDocument/DefaultPath), звуки (bell фиксирован, message дефолт, canberra-gtk-play/paplay, WARN-only), blocklist D-53, «где молчит» (role-гейт fail-closed, GTK4-Wayland, терминалы) и все числа §4 против Defaults()/валидации; единственная ин-плейс правка — «install лишь поставляет секцию в конфиг» заменено фактическим поведением (конфиг не пишется, решение 04-02); аудит-трейл цел (13 маркеров прежних ревизий на месте); дефектов поведения нет — новые todo не заводились

## Task Commits

Each task was committed atomically:

1. **Task 1: Русский-первый + сверка таблицы/примеров с кодом (D-8-9 + D-8-8)** - `27b2da2` (docs)
2. **Task 2: Секция a11y — семь секций, restart-семантика, рецепт оверрайда** - `0cf7006` (docs)
3. **Task 3: SPEC-сверка в форме 08-01 (D-8-8)** - `ca30cbe` (docs)

**Plan metadata:** отдельный metadata-коммит (SUMMARY + STATE + ROADMAP + REQUIREMENTS) следует за этим файлом.

_Note: план type: execute, docs-only — TDD-гейты не применялись._

## Files Created/Modified

- `docs/CONFIG.md` — русский-первый перевод + построчная сверка + секция a11y (+307/−145)
- `docs/SPEC.md` — сверка-ревизия 2026-10-05 + ин-плейс правка install-факта (+49/−2)
- `.planning/phases/08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill/deferred-items.md` — журнал повторных средовых флаков (ctlsvc ×2, appid ×1)

## Decisions Made

- Пример-литералы a11y.apps из текста плана («zcode» находит «org.zdev.Zed») заменены корректными («chromium» находит «org.chromium.Chromium»): исходный пример противоречил чувствительности к регистру И подстрочности, пиннутым той же строкой того же плана (Rule 1 — bug в плане, исправлен документально)
- Полные YAML-примеры (дефолтный и a11y-активный) несут все семь секций — «полный файл» обязан зеркалить Defaults(), куда после 08-02 входит a11y
- §4.4 SPEC (flip через org.gnome.desktop.input-sources) оставлен как есть: вне перечня сверки todo-первоисточника; канон флипа записан в ADR-006 (решение владельца, не оспаривается)
- Секция `a11y` в поведенческом абзаце описана через Active()-семантику (enabled И непустой список) — единственное определение из кода (08-02/08-05)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Пример-литералы a11y.apps противоречили собственной семантике матчинга**
- **Found during:** Task 2 (a11y-строки таблицы и пример)
- **Issue:** текст плана предлагал «совпадение подстрокой (zcode находит org.zdev.Zed-класс имён)» и пример `apps: ["^org\\.chromium\\.", "zcode"]` — но «zcode» не является подстрокой «org.zdev.Zed» даже без учёта регистра, а регистр-чувствительность пиннута той же строкой плана
- **Fix:** примеры заменены на внутренне-согласованные и фактические: «`chromium` находит `org.chromium.Chromium`» (форма строки — как в канонической строке blocklist D-53) и `apps: ["^org\\.chromium\\.", "Chromium"]` (форма «анкерованный + голый substring» сохранена по плану)
- **Files modified:** docs/CONFIG.md
- **Verification:** TWO-ROWS/A11Y-DOC-OK гейты; пример согласован с config.go doc-комментарием A11y
- **Committed in:** 0cf7006 (Task 2 commit)

**2. [Rule 2 - Missing Critical] В таблице отсутствовали два документированных поведения валидации**
- **Found during:** Task 1 (построчная сверка)
- **Issue:** док не нёс (а) правило «голой клавиши» hotkeys.tap_key (errTapKeyBare: модификаторный tap_key прошёл бы грамматику и «применился» бы с молча проигнорированными модификаторами) и (б) поведение пустого sound.autocorrect_event (EffectiveAutocorrectEvent: пустое читается как дефолт)
- **Fix:** обе строки таблицы дополнены (Range/Meaning), в «Именах привязок» добавлен абзац о «голой» клавише
- **Files modified:** docs/CONFIG.md
- **Verification:** сверка с config.go (:310-311 bare-key check; :165-171 EffectiveAutocorrectEvent)
- **Committed in:** 27b2da2 (Task 1 commit)

---

**Total deviations:** 2 auto-fixed (1 bug плана, 1 недостающая критичная документация валидации)
**Impact on plan:** обе правки — в docs/CONFIG.md, суть D-8-8 (док == поведение); формат таблицы и гейты не пострадали.

## Issues Encountered

- **Средовые флаки тестов при `mise run ci` (out of scope, не чинились).** Из шести прогонов плана три падали и все три проходили зелёно на немедленном повторе: `internal/ctlsvc` ×2 (известный с 08-01 флак — гонка владения `org.djarvur.goswitch` с ЖИВЫМ демоном владельца на реальной session bus) и `internal/appid` ×1 (новый transient: `appid_test.go:117 FocusedApp error after close = <nil>, want ErrBusClosed` — тайминг-окно вердикта закрытия шины под -race). Диф плана — только markdown, ноль .go. Задокументировано в deferred-items.md фазы.
- SPEC-ин-плейс правка потребовала сверки с решением 04-02 (adopt+watch): внутренний/установочный контракты подтверждены чтением internal/install/install.go (ни одной записи конфига) и internal/config/writer.go.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Таблица CONFIG.md — строгий источник для генератора 08-08: 21 строка ключей в сплошном блоке, ровно 5 колонок, ровно 2 a11y-строки (гейты прошли)
- 08-07 (README-реструктура) получает сверенный CONFIG.md как эталон русской терминологии («сочетание клавиш», «режим скрипта», «секция»)
- SPEC-контракт и доки согласованы (S4-CONFIG-SYNC); блокеров нет; средовые флаки тестов — в deferred-items.md

---
*Phase: 08-a11y-magiya-prilozheniy-dokumentatsiya-i-konfig-skill*
*Completed: 2026-10-05*

## Self-Check: PASSED

- docs/CONFIG.md — FOUND (коммиты 27b2da2, 0cf7006 в git log)
- docs/SPEC.md — FOUND (коммит ca30cbe)
- 08-06-SUMMARY.md — FOUND в каталоге фазы
- commits с ledger gsd-plan-head-before-08-06: 3 (27b2da2, 0cf7006, ca30cbe)
- плановые гейты: RU-FIRST / TABLE-ROWS-OK / ALL-KEYS-PRESENT / A11Y-DOC-OK / SEMANTICS-OK / ROWS-SEVEN-SECTIONS / TWO-ROWS / AUDIT-TRAIL-OK / NO-GO-IN-DIFF / S4-CONFIG-SYNC — все PASS; mise run ci exit 0
