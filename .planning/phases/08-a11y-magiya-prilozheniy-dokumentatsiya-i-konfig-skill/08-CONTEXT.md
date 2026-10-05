---
status: draft
phase: 08-a11y-магия-приложений-документация-и-конфиг-skill
created: 2026-10-05
source: беклог владельца (3 todo от 2026-10-05, дословные решения) + анализ зависимостей
  /gsd-manager --analyze-deps. ВАЖНО: интерактивное обсуждение (AskUserQuestion по 4 серым
  областям) осталось без ответа — владелец отошёл. Решения ниже помечены источником:
  «владелец (todo)» = дословно его слова; «рекомендация агента» = выбор по прецедентам
  проекта, НЕ подтверждён владельцем — подтвердить на verify-гейте или правкой файла.
---

# Phase 8: a11y-магия приложений, документация и конфиг-skill - Context

**Gathered:** 2026-10-05
**Status:** Ready for planning (решения-рекомендации ждут подтверждения владельца — см. frontmatter)

<domain>
## Phase Boundary

Три потока разобранного беклога, порядок = порядок зависимостей (a11y-код → документация → skill):

1. **a11y-магия** — конфиг-секция со списком приложений, которым goswitch автоматически
   включает accessibility-магию; механизм применения (демон/установщик), без root.
2. **Документация** — сверка README.ru/README/CONFIG.md/SPEC с фактическим поведением
   v1.1.0 + реструктура «русский-первый».
3. **Skill «Конфигурация goswitch»** — SKILL.md для AI-ассистентов: что можно настроить
   и как; синхрон с docs/CONFIG.md.

Новые возможности коррекции/переключения — вне фазы. GUI, терминалы, OSD — v2 (как раньше).
</domain>

<decisions>
## Implementation Decisions

### A11y-магия: что решено владельцем дословно

- **D-8-1 (владелец, todo a11y-magic-apps-list, 2026-10-05):** нужен НЕ перечень в
  документах, а **секция в конфиге** со списком приложений, которым goswitch автоматически
  включает accessibility-магию «как сделали вручную для ZCode» (прецедент подтверждён).
  Форма — «секция типа `a11y.apps` / расширение `autocorrect`» (точное имя — discretion).
- **D-8-2 (владелец, todo):** механику применения (gsettings toolkit-accessibility,
  snap-оверрайды, флаги запуска) **исследовать при планировании** — research определяет
  фактические механизмы per-app на GNOME 46 до кода.

### A11y-магия: рекомендации агента (не подтверждены владельцем)

- **D-8-3:** Момент применения — **демон-reconcile**: при старте и на каждом hot reload
  списка демон идемпотентно приводит систему к состоянию из конфига. Обоснование: hot
  reload — канон проекта (CONF-02), смена списка не должна требовать переустановки;
  install лишь создаёт секцию в конфиге. Альтернативы (отвергнуты): install-time
  однократно; ручная команда `goswitchctl a11y apply`.
- **D-8-4:** Откат — **revert при uninstall** (ASVS-прецедент: чорды wm.keybindings
  восстанавливаются, снапшот-дисциплина sources). Удаление приложения из списка НЕ
  откатывает применённое (невозможно отличить от включённого владельцем вручную до
  goswitch). — **Reversibility:** reversible — снапшот + revert-шаг, тот же паттерн, что у sources/чордов в install.
- **D-8-5:** Матчинг списка приложений — **regex-подстрока** (RE2, анкеровка явная) —
  консистентность с `autocorrect.apps_blocklist` (ревизия D-53 фазы 7); строгий decode,
  валидация компиляции на Load.
- **D-8-6:** Набор магии — минимально достаточный: глобальный
  `gsettings org.gnome.desktop.interface toolkit-accessibility true` (ровно то, что
  включалось вручную для ZCode) + опциональные per-app user-level правки запуска
  (`~/.local/share/applications` override с флагами типа
  `--force-renderer-accessibility`). Всё без root, всё в $HOME. Snap-оверрайды —
  research решает по необходимости; глобальный ключ НЕ снимается динамически
  (нельзя выключать «для одного приложения» — ключ общий).

### Документация: русский-первый

- **D-8-7 (владелец, todo docs-accuracy-and-russian-first, дословно):** «README.ru как
  главная, README как перевод» → **README.md = русский**, английский перевод —
  README.en.md; синхронность поддерживается.
- **D-8-8:** Проверка соответствия — README/CONFIG.md/SPEC сверяются с фактическим
  поведением v1.1.0 (меню, тумблеры, звуки, blocklist, adopt+watch, «где молчит» —
  список из todo); расхождения в ДОКАХ правятся; если расхождение = дефект поведения —
  новая todo, код в этой фазе не правится. (Рекомендация агента.)
- **D-8-9:** CONFIG.md — русский-первый (пользовательский док); ACCEPTANCE.md /
  ci-runner.md / SECURITY.md остаются EN (операционные/CI). SPEC.md уже русский.
  (Рекомендация агента.)

### Skill «Конфигурация goswitch»

- **D-8-10:** Синхрон с доками — **single source**: конфиг-раздел SKILL.md генерируется
  из docs/CONFIG.md mise-задачей (рассинхрон невозможен); разделы процедур и диагностики
  — ручной каркас, CI lint-гейт сверяет ключи/дефолты. Todo даёт выбор «генерация из
  CONFIG.md или lint-гейт» — взято первое с гейтом для ручной части. (Рекомендация
  агента; точную механику уточняет планировщик.)
- **D-8-11:** Расположение — `skills/goswitch-config/SKILL.md` (project skill) — ОТСЛЕЖИВАЕМОЕ место в репо, НЕ машинная .zcode/ (решение владельца 2026-10-05: скилл нужен в репо, но не в машинной обвязке).

### Claude's Discretion

- Имя секции конфига (`a11y.apps` vs иное) и схема ключей — по канонам схемы
  (ACTIVE-условия, потолки, strict decode, complete-document — прецедент 06-04/07-02).
- Разбиение на планы/волны (естественный порядок потоков 1→2→3; 07-06/07-08 — прецедент
  параллельных планов с непересекающимися файлами).
- Объём регресса: `mise run ci` обязателен; e2e-матрица не расширяется (a11y-магия вне
  пути коррекции); юнит/golden на новый код по директивам TDD.
- Состав документации-сверки: полный обход README/CONFIG/SPEC против кода.

### Folded Todos

- **a11y-magic-apps-list** — «Конфиг-секция: приложения с автoвключением a11y-магии»:
  поток 1 фазы; дословные решения D-8-1/D-8-2.
- **docs-accuracy-and-russian-first** — «Документация: проверка соответствия,
  русский-первый»: поток 2; решение D-8-7.
- **config-skill-for-ai** — «Skill: конфигурация сервиса как документация для AI»:
  поток 3; решение D-8-10.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Контракт конфигурации и спецификация
- `docs/CONFIG.md` — канон всех секций config.yaml; источник генерации skill (поток 3) и объект правок (потоки 1-2)
- `docs/SPEC.md` §4 (конфигурация), §10–§11 (автокоррекция/политика) — правка spec-delta'ом, если a11y-секция касается контракта (D-55-дисциплина: спека до кода)
- `docs/adr/ADR-007*` — политика автокоррекции (a11y-секция её не меняет, но соседствует)

### Решения предыдущих фаз
- `.planning/phases/07-menyu-v2-i-chernyy-spisok-avtokorrektsii/07-CONTEXT.md` — канонический макет меню, adopt+watch (`EnsureDocument`/`DefaultPath`), blocklist-семантика D-53
- `.planning/phases/06-avtokorrekcija-opcionalno/06-CONTEXT.md` — D-51..D-55; role-гейт и 25 мс контур не трогаются
- `.planning/phases/05-integratsiya-s-gnome-indikatsiya/05-CONTEXT.md` — D-52..D-54 (двухисточниковая схема, снапшот-дисциплина install)

### Todo-первоисточники (дословные решения владельца)
- `.planning/todos/pending/a11y-magic-apps-list.md`
- `.planning/todos/pending/docs-accuracy-and-russian-first.md`
- `.planning/todos/pending/config-skill-for-ai.md`

### Документация как объект работы
- `README.md`, `README.ru.md` — текущая двуязычная структура (реструктурируется D-8-7)
- `docs/ACCEPTANCE.md`, `docs/ci-runner.md`, `SECURITY.md` — остаются EN (D-8-9)

### Конвенции
- `.planning/codebase/CONVENTIONS.md` — TDD, зелёная итерация, mise, golangci-lint

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/config` — паттерн секций с ACTIVE-условиями, strict decode, потолки (06-04/07-02) — образец для a11y-секции
- `internal/config` Watcher + `applySnapshot` — hot reload, точка reconcile-хука a11y-магии
- Меню-persist (07-03): `SetAutocorrectEnabled`/`EnsureDocument`/`DefaultPath`, yaml.v3 Node round-trip, temp+rename, 0600 — механика записи конфига
- `internal/sound` (07-08) + `clipboard.go` — no-pipes fork-shaped subprocess (gsettings-вызовы той же формы)
- `goswitchctl install/uninstall/selfcheck` (04-01/04-02) — снапшот/restore-дисциплина, точка для revert-шага D-8-4
- `internal/ctlsvc` — точка расширения `goswitchctl` (если понадобится a11y-команда статуса)

### Established Patterns
- Спека до кода (D-55): секция меняет контракт — spec-delta в плане №1 волны
- TDD red → green → refactor; golden-корпуса для генерируемых данных
- Русская проза: «сочетание клавиш» (не «чорда»), без жаргона в владельческих текстах

### Integration Points
- Старт демона — рядом с `activate.IfOwned` (самореактивация): reconcile a11y-состояния
- `watch.go` — reload → reconcile списка a11y-магии
- Uninstall — revert-шаг рядом с восстановлением чордов wm.keybindings
- CI (`.github/workflows/`) — lint-гейт синхрона skill↔CONFIG.md рядом с pr-sanity

</code_context>

<specifics>
## Specific Ideas

- ZCode-прецедент (владелец, 2026-10-05): a11y-магия уже включалась вручную
  (toolkit-accessibility; Electron-рецепты `--force-renderer-accessibility` /
  toolkit-accessibility из док фазы 7 UAT) — фаза автоматизирует ровно этот путь.
- Автокоррекция в ZCode/Telegram (snap/Electron) молчала по design (role-unknown
  fail-closed); a11y-магия из списка должна делать такие приложения наблюдаемыми —
  но role-гейт остаётся фильтром безопасности (не трогаем контракт фазы 6/7).

</specifics>

<deferred>
## Deferred Ideas

- Snap-оверрайды как механизм магии — если research одобрит (D-8-6 оставил на его усмотрение).
- Показ наблюдаемой a11y-идентичности в `goswitchctl status` — backlog 06-UAT Deferred, сюда не входит.

### Reviewed Todos (not folded)

Нет — все 3 совпавшие todo свёрнуты в фазу (она из них и собрана).

</deferred>

---

*Phase: 08-a11y-магия-приложений-документация-и-конфиг-skill*
*Context gathered: 2026-10-05*

## Owner Revision — 2026-10-06 (после исполнения, до закрытия фазы; вербатим)

- **D-8-1/REV (владелец, 2026-10-06):** «раз настройка глобальная, то список не нужен, а нужен
  bool параметр, по дефолту настройка включен» → секция `a11y` теряет `apps` (RE2-список),
  остаётся `a11y.enabled` (bool), **default ON**. Вся семантика матчинга/декларации приложений
  уходит; reconcile = «enabled → ключ включён», disabled → демон ничего не делает (ключ не
  трогает — прежние снимающие семантики D-8-4 сохраняются только в uninstall-revert).
- **D-8-5/REV:** аннулирована (regex-валидация и потолок 64 apps существовали ради списка).
- **D-8-6/REV:** набор магии не меняется (глобальный toolkit-accessibility + IsEnabled-пояс);
  default ON — осознанное решение владельца (не D-54-прецедент: a11y — не автокоррекция).
- WR-02 из 08-REVIEW.md закрывается этим решением (доки станут честными: матчинга нет —
  его и не будет). Todo a11y-apps-matching-semantics.md закрыт. Исполнение — дельта-план
  08-09 (spec-delta §4 до кода).
