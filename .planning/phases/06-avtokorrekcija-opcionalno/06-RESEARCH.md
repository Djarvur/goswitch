# Phase 6: Автокоррекция опционально - Research

(ренумерация 2026-09-28: фаза вставилась как 6-я после новой Фазы 5 «Интеграция с
GNOME» по директиве владельца; артефакты перенесены 05→06, ADR автокоррекции
ренумерован 006→007 — ADR-006 занят ревизией ADR-001 двухисточниковой схемы)

**Researched:** 2026-09-27
**Domain:** wrong-layout detection (dictionary + trigram scoring), AT-SPI role-based safety policy (app × role), config/status surface, e2e negative cases (password/terminal), licensing of donor algorithms and dictionary data
**Confidence:** HIGH (детектор-алгоритмы и лицензии проверены по первоисточникам; AT-SPI-политика проверена живьём на целевой машине; размеры бинарника измерены сборкой)

## Summary

Phase 5 возвращает осознанно отвергнутую в v1 автокоррекцию в форме, предписанной самим SPEC §11: opt-in, default off, white-list. Всё исследование опиралось на живую целевую машину (этот стол = GNOME 46 / IBus 1.5.29, goswitchd v1.0.0 запущен) и на первоисточники доноров.

**Детектор.** xneur (GPL-2+, прочитан вживую по репо `AndrewCrewKuznetsov/xneur-devel`) показывает, что его ядро — НЕ словарь, а дешёвые эвристики: собственные словари xneur — 8–9 regex'ов на язык, реальное покрытие делегировано опциональному enchant/aspell, а основная работа — proto-списки «невозможных» 2-буквенных сочетаний (ru: 203 пары) + правило «первый словарь, содержащий слово». Easy Switcher (GPL-2.0, Pascal — прочитан вживую) словарного слоя НЕ имеет вовсе: это ручной корректор по клавише (буфер сканкодов + ReplaceWord) — премисса D-52 о нём как о «словарном референсе» фактически не подтверждается. Оба донора читались read-only, код не копируется (GPL).

Отсюда форма нашего детектора (гибрид D-52): (а) словарный путь на настоящих словарях hunspell — уже установленных на этой машине (ru_RU: 146 269 слов, лицензия custom-BSD Lebedev; en_US: 79 013 слов, SCOWL-permissive — обе совместимы с MIT-репо), (б) триграммный fallback **собственной мини-модели**, обученной на тех же словарях, а не whatlanggo: whatlanggo (MIT, v1.0.1 2019, zero-dep) спроектирован под предложения из 84 языков и на слове 2–10 символов деградирует, стоит +750 КБ бинарника (измерено) против ~50–80 LOC и десятков КБ данных своей модели. Размер вшитых словарей измерен сборкой: полные ru+en как sorted slice = **+7.9 МБ** бинарника и **ноль** рантайм-кучи (rodata), против +10–12 МБ RSS у map при том же размере бинарника — при бюджете памяти SPEC §5 < 50 МБ sorted slice + binary search — однозначный выбор.

**Политика безопасности (D-53).** Живая проверка AT-SPI на этой машине дала критический факт: **роль парольного поля не различается между тулкитами**. GTK4 `GtkPasswordEntry` честно отдаёт `password-text` (40), но GTK3-пароль (zenity --password, GtkEntry visibility=off) отдаёт роль 61 «text box» — неотличимую от обычного поля. Следствие: ролевый фильтр necessary-but-not-sufficient; безопасность фичи держится на конъюнкции D-53 в целом — white-list приложений (default пустой = выключено везде) + роль + surrounding-text + уверенность детектора. wezterm вообще отсутствует на a11y-шине (в перечне 25 приложений его нет) → «нет AT-SPI-дерева → молчание» работает для него автоматически. Живой вызов GetRole по пути из фокус-события верифицирован: `org.a11y.atspi.Accessible.GetRole` на `/org/gnome/Zenity/a11y/<uuid>` из другого соединения отвечает мгновенно — расширение `internal/appid` (хранить sender+path последнего фокус-события) и шов `RoleSource` — прямая дорога.

**Primary recommendation:** spec-delta + ADR-007 (ADR-006 занят Фазой 5) до кода (D-55); затем генератор `layouts/dictgen` (hunspell → golden sorted slices, полный ru 146k + полный en 79k), pure-пакет `internal/detect` (словарный путь + своя триграммная модель + golden-корпус), расширение `internal/appid` живым GetRole, конфиг-секция `autocorrect` (enabled:false, apps:[]) + ctl-счётчики, хук в `feedKey` на границе слова с деградацией в молчание, e2e-кейсы с GTK4-фикстурой парольного поля (gir1.2-gtk-4.0 установлен; фиксatura ~30 строк python3 проверена живьём в этом исследовании) и терминальным негативом через счётчики ctl.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions (D-51..D-55, дословно по `.planning/phases/06-avtokorrekcija-opcionalno/06-CONTEXT.md (ранее 05-CONTEXT.md)`)

- **D-51:** Автокоррекция открывает milestone v1.1.0 — не фаза внутри v1.0.0 (спека v1 автокоррекцию исключает). Формальный new-milestone-переключатель (PROJECT.md/STATE/архивация фаз) выполняется при complete-milestone v1.0.0; до тех пор артефакты v1.1 ведутся аддитивно (ROADMAP-секция, папка 05).
- **D-52:** Гибридный детектор, решение владельца 2026-09-27: (а) словарный путь — на границе слова «набранное слово НЕ в словаре текущей раскладки И его ремап по таблице ЙЦУКЕН↔QWERTY ЕСТЬ в словаре другой» → высокая уверенность ошибки; (б) триграммный скоринг как fallback для OOV-слов (имена, опечатки) — сравнение правдоподобия слова в обеих раскладках. Доноры: xneur (GPL — ТОЛЬКО референс алгоритма, goibus-прецедент, код не копируется), Easy Switcher (GPL-2.0-only — тоже только референс), whatlanggo (MIT — допустим как зависимость; финальное решение «своя мини-модель vs зависимость» — за research по размеру/аудиту). Словарные данные (hunspell ru/en, частотные списки) запекаются go:generate-генератором в golden-паттерне таблиц Фазы 1; лицензии конкретных источников проверяет research ДО выбора.
- **D-53:** Политика срабатывания — конъюнкция ВСЕХ условий: (1) приложение в white-list политики (по bridge-namespace из internal/appid), (2) роль сфокусированного AT-SPI-объекта — текстовый ввод (entry/document text; НЕ password text, НЕ terminal, НЕ canvas), (3) приложение отдаёт surrounding text (verify-rung применим), (4) детектор дал уверенный вердикт, (5) слово ≥ минимальной длины. ЛЮБОЕ «неизвестно» (нет AT-SPI-дерева, нет фокус-события, роль неопределима) → молчание. Роль запрашивается живым GetRole по объекту из фокус-события В МОМЕНТ решения — кэш appid (последнее известное приложение) advisory, не основание для срабатывания.
- **D-54:** Default off везде: конфиг-секция `autocorrect` в YAML-схеме (strict decode D-33, hot reload D-32 распространяются автоматически), вшитые дефолты — enabled: false + пустой white-list. goswitchctl status показывает состояние автокоррекции (enabled/политика/последние счётчики срабатываний — аналогично MACRStats). Публичный конфиг-контракт docs/CONFIG.md дополняется.
- **D-55:** SPEC правится spec-delta внутри фазы ДО кода детектора (ADR нумерация обновлена: 006→007): §2 (класс 3 больше не «вне объёма»), §10 (снять пункт), §11 (записать решение возврата: opt-in, default off, white/black list — как предписано; новая дата решения владельца). ADR-007 (ренумерован с 006) — механика: AT-SPI роль источником решения, отказ от кэш-семантики, лицензионная позиция по донорам.

### Claude's Discretion (дословно)
- Имена конфиг-ключей секции autocorrect, форма white/black list (flat list vs map app→roles) — в рамках D-31..D-33/D-54
- Размер и состав запекаемых словарей (top-N частотных, порог размера бинарника) — research предлагает, планировщик пинирует
- Триграммная модель: whatlanggo как зависимость vs собственная мини-модель на тех же частотных данных — research сравнивает, решение по критериям аудита/размера
- Порог уверенности детектора, минимальная длина слова, дебаунс срабатываний — методикой при реализации, золотой корпус юнит-тестов обязателен
- Ролевый словарь AT-SPI (какие роли считать «текстовым вводом») — research верифицирует живьём на GNOME 46 (GTK4, Chromium/Electron, VTE-терминал, парольные поля)
- Разбиение на планы/волны, порядок spec-delta vs detector-package

### Deferred Ideas (OUT OF SCOPE)
- Терминальный класс (clipboard-уровень для терминалов) и OSD — остаются во v2 (D-51-контекст фазы; автокоррекция в терминалах не делается — только «НЕ трогаем»)
</user_constraints>

<phase_requirements>
## Phase Requirements

Финальные REQ-ID фиксируются в REQUIREMENTS.md при new-milestone (ROADMAP §Phase 5); ниже провизорные AC из ROADMAP.

| ID | Description | Research Support |
|----|-------------|------------------|
| AC-01 | Гибридный детектор (словарь + ремап + триграммный fallback) | Алгоритм xneur прочитан (detection.c); xneur-словари — 8–9 regex'ов, реальное покрытие — spellchecker: наш словарный путь строится на hunspell (146k/79k, лицензии проверены); триграммная мини-модель дешевле и точнее по размеру, чем whatlanggo (+750 КБ измерено); размеры binary измерены |
| AC-02 | Политика app×role на AT-SPI-данных, unknown → молчание | Живые роли GNOME 46: GTK4 password-text=40 vs GTK3 password → 61 «text box» (не различимо!); wezterm отсутствует на a11y-шине; GetRole по пути фокус-события работает из любого соединения (busctl live); шов AppidSource (actor.go:301-303) расширяется RoleSource |
| AC-03 | Конфиг `autocorrect` (strict decode, hot reload) + ctl-статус с счётчиками | Образцы: MACR-секция config.go:83-88 + Defaults:118-123; applySnapshot/SetOptions (actor.go:277,659); MACRCounters (actor.go:348) как шаблон; docs/CONFIG.md:60-84 — форма секции |
| AC-04 | e2e-доказательства безопасности (пароль/терминал НЕ трогаются) | Реестр кейсов caseSpec (main.go:66-79); GTK4-фикстура GtkPasswordEntry проверена живьём (~30 строк python3, role password-text); zenity --password для ролевого гейта непригоден (роль 61); терминальный негатив — через счётчики ctl + -debug-лог |
| AC-05 | spec-delta §2/§10/§11 + ADR-007 | Текст §10/§11 прочитан (SPEC.md:177-188); лицензионная позиция по донорам собрана (таблица ниже) |
</phase_requirements>

## Project Constraints (from AGENTS.md / CONVENTIONS.md — no CLAUDE.md exists)

1. **Strict TDD** (D-07) — каждая задача поведения red → green → refactor; золотой корпус детектора обязателен (Discretion).
2. **Green iteration** (D-08) — `go build ./...`, `go test -race ./...`, `golangci-lint run` чистые на каждой итерации.
3. **mise, not make** (D-09) — задачи dictgen/e2e — в `mise.toml`.
4. **golangci-lint maximal-strict** с Фазы 1 (`.golangci.yml`, v2, `default: all`).
5. **Runtime-зависимости**: stdlib + godbus + fsnotify + yaml.v3 только; whatlanggo если бы его выбрали — единственный кандидат на исключение (см. Standard Stack; рекомендация — НЕ добавлять).
6. **Приватность логов** (D-20/D-21): детектор НЕ логирует набранное слово — только счётчики и причины.
7. **Golden-паттерн генераторов** (layouts/generator/main.go:420 `//go:generate go run ./generator`; generated file committed; CI гоняет только golden-тесты) — словарный генератор обязан следовать ему.
8. **Git workflow** (user AGENTS.md): фаза уже на ветке `gsd/phase-04-postavka-i-priemka`; новых веток research не создавал.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Wrong-layout детекция (словарь + триграммы) | Daemon pure-пакет (`internal/detect`) | `layouts/dictgen` (генератор данных) | Чистая функция над рунами + вшитые данные; TDD без D-Bus; данные — golden через go:generate |
| Политика app×role (живой GetRole) | Daemon (`internal/appid` + `internal/session`) | — | appid уже владеет a11y-шиной и фокус-событиями; расширяется хранением (sender,path) и Role-запросом |
| Решение о срабатывании (конъюнкция D-53) | Daemon (`internal/session/actor.go` — feedKey) | — | Граница слова живёт в feedKey; актор держит buf/mode/caps/appid; correction pipeline переиспользуется |
| Замена слова | Daemon (`internal/correct` plan/verify/ladder) | — | Единственный путь замены (D-52 reuse); никакого второго механизма |
| Конфиг + ctl-поверхность | Daemon (`internal/config`, `internal/ctlsvc`) + ctl | — | Секция autocorrect по образцу macr; счётчики по образцу MACRStats |
| Негативные e2e-кейсы | e2e stand (`test/e2e/case_autocorrect.go` + GTK4-фикстура) | goswitchctl status (счётчики-оракул) | Стенд уже умеет spawn/focus/witness; новых механизмов нет |
| Spec-delta + ADR-007 | docs (`docs/SPEC.md`, `docs/adr/ADR-007-*.md`) | — | D-55: до кода детектора |

## Standard Stack

### Core

| Library/Tool | Version | Purpose | Why Standard |
|--------------|---------|---------|--------------|
| hunspell-ru (данные, не зависимость) | 1:24.2.1-1 (Ubuntu 24.04, установлен) `[VERIFIED: dpkg -l]` | Словарь ru для детекции: `ru_RU.dic` = 146 269 лемм (`head -1` = «146269»), 3.4 МБ | Уже на целевой машине и в CI-образе не нужен вовсе (golden-данные коммитятся); лицензия custom-BSD (см. Licensing) — MIT-совместима |
| hunspell-en-us (данные) | 1:2020.12.07-2 (установлен) `[VERIFIED: dpkg -l]` | Словарь en: `en_US.dic` = 79 013 лемм, 844 КБ | SCOWL-лицензия permissive («use, copy, modify, distribute and sell … without fee») — MIT-совместима с атрибуцией |
| Собственная триграммная мини-модель (in-repo) | n/a (~50–80 LOC + сгенерированные таблицы) | Fallback-скоринг OOV-слов (D-52б) | Решение research (Discretion): чтоlanggo не подходит по размеру/профилю задачи — см. ниже |
| `internal/appid` + godbus GetRole | godbus v5.2.2 (уже в go.mod) `[VERIFIED: STACK.md]` | Живой запрос роли по (sender,path) фокус-события (D-53) | Форма вызова верифицирована живьём busctl: `org.a11y.atspi.Accessible.GetRole` → `(u)`; новых зависимостей нет |

### Supporting

| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| `python3` + gir1.2-gtk-4.0 | 3.12.3 / 4.14.5 `[VERIFIED: dpkg -l]` | GTK4-фикстура парольного поля для e2e | Живая проверка роли password-text; фиксatura коммитится в test/e2e/fixtures/ |
| `focus_helper.py` witness | in-repo | Роль+chars фокус-инпута (уже различает PASSWORD_TEXT) | Оракул негативных кейсов (`focused-input-pid`, surface.go) |
| `busctl --user --address=<a11y>` / `gdbus` | systemd 255 | Отладка AT-SPI-ролей при разработке | Формула адреса: `org.a11y.Bus.GetAddress` → `unix:path=/run/user/1001/at-spi/bus` (live) |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Своя триграммная модель | `github.com/abadojack/whatlanggo` v1.0.1 (MIT, zero-dep) `[VERIFIED: proxy.golang.org @latest + GitHub API license]` | +750 КБ бинарника (измерено: 1 970 336 vs 1 220 768 байт, -s -w), 1 918 LOC аудита, профили под 84 языка и предложения — на слове 2–10 символов (0–8 триграмм) уверенность деградирует; заморожен с 2019. Своя модель: обучается на тех же словарях, ~десятки КБ, аудируется за час, пороги под наш корпус. Рекомендация: своя модель |
| Полные hunspell-словари (+7.9 МБ) | top-N частотные подмножества | Требуют частотного источника: OpenCorpora недоступен в этой сессии (HTTP 521), лицензия не верифицирована → `[ASSUMED]` CC-BY-SA; подмножество режет покрытие (каждый словарный hit — подавление коррекции, т.е. полнота = безопасность). Полные словари рекомендованы для v1.1 |
| Runtime map | sorted []string + `sort.SearchStrings` | Бинарник одинаков (литерал доминирует), но map = +10–12 МБ RSS (измерено 15.9 МБ VmRSS процесса против 14.4 МБ у слайса, а вшитый литерал вообще в rodata ≈ 0 кучи) — SPEC §5 < 50 МБ |
| AT-SPI роль по имени (GetRoleName) | AT-SPI роль по enum (GetRole → u) | Имена локализованы/различаются («text box» vs gi-nick «text»); enum стабилен: 40/60/61/79/94 — верифицированы |

**Installation:**
```bash
# НОВЫХ Go-зависимостей НЕТ (рекомендация). Словарные данные ставятся один раз на dev-машине:
sudo apt install hunspell-ru hunspell-en-us   # уже установлены на целевой машине (VERIFIED dpkg)
# dev-пакеты для GTK4-фикстуры e2e (уже установлены):
# gir1.2-gtk-4.0 (4.14.5), python3-gi (3.48.2)
```

## Package Legitimacy Audit

> Seam `package-legitimacy check` поддерживает только npm/pypi/crates (как и в Фазе 4) — Go-верификация по proxy.golang.org + GitHub API.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| github.com/abadojack/whatlanggo | proxy.golang.org + GitHub API | v1.0.1 — 2019-03-06; последний push 2023-03-28 | 696 stars (Go-статистика загрузок N/A) | github.com/abadojack/whatlanggo (license: MIT по GitHub API) | OK (двухисточниковая верификация) | НЕ одобрен к использованию — рекомендация research: своя мини-модель (см. Alternatives); если планировщик/владелец выберет его — добавить checkpoint:human-verify |
| hunspell-ru / hunspell-en-us | Ubuntu 24.04 archive (данные, не Go-зависимость) | 24.2.1 / 2020.12.07 | distro | libreoffice-dictionaries + SCOWL | OK | Approved — данные запекаются генератором, лицензии верифицированы по Debian copyright |
| github.com/Djarvur/goswitch (self) | — | — | — | this repo | OK | Approved |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Research Findings

### Q1. xneur: алгоритм детекции (GPL-2+ — read-only референс, код НЕ копируется)

Источник: репо `AndrewCrewKuznetsov/xneur-devel` (master, push 2024-10-18; tarballs 0.19.0/0.20.0 в `dists/`; license GPL-2+ в шапках файлов) `[VERIFIED: gh api + raw.githubusercontent.com, файлы прочитаны этой сессией]`. Ключевой файл `xneur/lib/ai/detection.c` (444 строки, прочитан целиком).

Алгоритм `check_lang` (detection.c:365-443):

1. **Буфер во всех языках одновременно.** Буфер набранного хранится реконвертированным в КАЖДЫЙ настроенный язык (`p->i18n_content[lang].content` — remap по keymap-таблицам). Последнее слово добывается в каждом языке (`get_last_word`), с ведущими пунктуациями/цифрами срезанными (offset до первой не-punct/не-digit) и хвостовыми цифрами отрезанными (`del_final_numeric_char`).
2. **Словарный проход — «первый по порядку» без порогов** (`get_dictionary_lang`, detection.c:46-63): для каждого языка `dictionary->exist(word, BY_REGEXP)`; первый язык, чей словарь содержит слово, выигрывает. Сами словари xneur — крошечные regex-списки: ru — 9 строк вида `(?i)^а$`, `(?i)\.рф`; en — 8 строк `[VERIFIED: xneur/share/languages/{ru,en}/dictionary fetched]`. Это НЕ полноценные словари — только однозначные короткие слова.
3. **Spellchecker-проход (опционально, WITH_ENCHANT/WITH_ASPELL)**: сначала текущий язык, потом остальные; гейт «слово длиннее одного символа» (`word_len / sym_len[lang][0] > 1`); опять первый hit.
4. **«Misprint»-проход (опционально)** (`get_similar_words`, detection.c:245-363): spellchecker-подсказки + минимум Левенштейна, порог `LEVENSHTEIN_LEN = 3`; язык ближайшей подсказки.
5. **Proto-проход — ядро дешёвой эвристики** (`get_proto_lang`, detection.c:193-243): на каждый язык список «прото» — 2-символьных последовательностей (`PROTO_LEN = 2`) и 3-символьных (`BIG_PROTO_LEN = 3`; в ru big_proto ПУСТ — 0 строк, en тоже). Правило: если слово СОДЕРЖИТ прото-последовательность текущего языка → подозрение; потом первый ДРУГОЙ язык, чьё прото НЕ содержит ни одной последовательности слова → вердикт. Данные: ru proto = 203 пары («аъ», «аы», «аь», «бй», «вй», «гй»…), en proto = 69 `[VERIFIED: fetched]`.
6. **Границы слова**: `word_len > 250 || word_len < 2 → skip` (detection.c:269).

**Уроки для нашего детектора** (алгоритмические, не код):
- «Первый словарь, содержащий слово» — нуль порогов и нуль тюнинга; у нас словарей ровно два, и правило D-52а («не в текущем И в ремапе-другом») сильнее — оно требует ОБЕ стороны, а не одну.
- Дешёвый n-gram-фильтр (proto) — хорош как быстрый отсекатель, но сам по себе даёт ложные срабатывания на именах/аббревиатурах; у нас его роль играет словарь+триграммы, а proto-подход можно не переносить.
- Xneur верит spellchecker'у больше, чем своим спискам: наш аналог веры — полные hunspell-словари, запечённые заранее.
- Сравнение через remap-буферы обоих языков — тот же примитив, что наш `correct.Convert` + `layouts.ENToRU/RUToEN` — переносимо один-в-один.

### Q1b. Easy Switcher: словарного слоя НЕТ (коррекция премиссы D-52)

Репо `freemind001/easy-switcher` (Pascal, GPL-2.0, 121 звезда, push 2025-02-03; единственный исходник `easy-switcher.lpr`, 1139 строк, прочитан) `[VERIFIED: gh api + raw fetch]`. Это **ручной** корректор по горячей клавише (Pause/Break): буфер сканкодов evdev, `GetBufferAction` → KeepBuffer при <2 клавишах, ReplaceWord/ReplaceAll по последовательностям нажатия replace-key и Shift. Греп по `dict|dictionary|wordlist` — ноль вхождений. Архитектура uinput (собственный known-bug: «Doesn't work correctly together with key remappers such as keyd»). **Для детектора автокоррекции донорских уроков не даёт** — кроме подтверждения нижней границы «слово < 2 символов не трогаем». Отметить в ADR-007 честно: референс просмотрен, словарного алгоритма не содержит.

### Q2. Триграммный fallback: whatlanggo vs своя мини-модель

whatlanggo `[VERIFIED: proxy.golang.org/github.com/abadojack/whatlanggo/@latest → v1.0.1, 2019-03-06; @v/v1.0.1.mod — без require-строк (zero-dep); GitHub API license=MIT; пакет скачан и распакован этой сессией]`:

- 1 918 LOC Go суммарно; профили языков — Go-константы (`lang.go` 179 КБ исходника), 84 языка.
- **Стоимость бинарника измерена**: пустой main = 1 220 768 байт; main + `whatlanggo.Detect("привет")` = 1 970 336 байт (go1.27.1, `-ldflags="-s -w"`, linux/amd64) → **+747 КБ**.
- Пригодность к одному короткому слову: `Detect` создан для предложений; на слове 2–10 символов у нас всего 0–8 триграмм — базовая слабость триграммных моделей на коротком входе; уверенность (`Confidence`) на коротких строках нестабильна — это известное свойство класса моделей, whatlanggo не исключение `[ASSUMED — классическое свойство триграммного скоринга; численный порог пинируется golden-корпусом при реализации]`.
- Определение скрипта (Cyrillic/Latin) нам не нужно — режим и так известен (a.mode), кандидатов ровно два.

Своя мини-модель (рекомендация): trigram-счётчики, обученные генератором на тех же hunspell-словах; таблица «top-K триграмм на язык» (~2–4 К записей на язык ≈ десятки КБ бинарника); скоринг = сумма log-вероятностей триграмм слова по обеим таблицам, нормировка длиной; вердикт — порог разницы + порог абсолютной уверенности. ~50–80 LOC + golden-тесты. **Ни одной новой зависимости, полный аудит, пороги под наш корпус.** Решение за планировщиком/владельцем (Discretion), рекомендация: своя модель.

### Q3. Словарные данные: источники, лицензии, размеры

**Что уже установлено на этой машине** `[VERIFIED: dpkg -l + ls, эта сессия]`:
- `hunspell-ru 1:24.2.1-1` → `/usr/share/hunspell/ru_RU.dic` (146 269 лемм с affix-флагами, 3.4 МБ; после снятия флагов и фильтра len≥2 — 146 261 уникальных слов — измерено)
- `hunspell-en-us 1:2020.12.07-2` → `/usr/share/hunspell/en_US.dic` (79 013 записей, 844 КБ; после обработки — 78 951)
- `wamerican/wbritish 2020.12.07-2` → `/usr/share/dict/american-english` (104 334 слова) — запасной en-источник
- aspell-en, libhunspell-1.7 — сами спеллчекеры НЕ нужны (данные запекаются, рантайм-зависимости нет)

**Лицензии** `[VERIFIED: /usr/share/doc/hunspell-ru/copyright и /usr/share/doc/hunspell-en-us/copyright прочитаны этой сессией]`:

| Источник | Лицензия | Правообладатель | MIT-репо |
|----------|----------|-----------------|----------|
| ru_RU.aff/dic | `custom-bsd-4-clauses` (пункт 4 снят правообладателем ispell-линии; ствол «Redistribution and use … with or without modification, are permitted provided …») — дословно в copyright: «Files: dictionaries/ru_RU/* / Copyright: 1997-2008 Alexander I. Lebedev / License: custom-bsd-4-clauses» | Alexander I. Lebedev | Совместимо при сохранении копирайт-нотиса в файле данных/доках |
| en_US.aff/dic | SCOWL permissive: «Permission to use, copy, modify, distribute and sell these word lists … for any purpose is hereby granted without fee, provided that the above copyright notice appears …» + компоненты Moby — public domain | Kevin Atkinson (SCOWL), Alan Beale и др. (нотисы в copyright) | Совместимо при сохранении нотисов |
| OpenCorpora частотники | НЕ ВЕРИФИЦИРОВАНО: opencorpora.org недоступен (HTTP 521, эта сессия); общепринятое мнение CC-BY-SA — `[ASSUMED]` | — | Избежать в v1.1; если частотные подмножества понадобятся — верифицировать лицензию отдельной сессией |

**Размеры бинарника — измерены сборкой** (go1.27.1, `-ldflags="-s -w"`, linux/amd64; база — «пустой main» 1 220 768 байт) `[VERIFIED: измерено этой сессией, файлы-генераторы в /tmp/dictsize]`:

| Вшитые данные | Представление | Бинарник | Дельта |
|---------------|---------------|----------|--------|
| ru 146 261 слово (полный) | []string литерал + runtime map build | 6 881 440 B | **+5.66 МБ** |
| ru 60 000 | map | 3 715 232 B | +2.5 МБ |
| ru 30 000 | map | 2 617 504 B | +1.4 МБ |
| ru 10 000 | map | 1 884 320 B | +0.66 МБ |
| en 78 951 (полный) | map | 3 424 416 B | **+2.2 МБ** |
| whatlanggo | import | 1 970 336 B | +747 КБ |

Рантайм-память: процесс с runtime-map из 146k слов = VmRSS 15 932 КБ; со слайсом, собранным в рантайме, = 14 360 КБ; **вшитый []string литерал живёт в rodata — прирост кучи ≈ 0** (lookup — `sort.SearchStrings`, ~17 сравнений на 146k). Вывод: **sorted slice + binary search**; map в рантайме не строить.

**Рекомендация**: запечь ПОЛНЫЕ ru_RU + en_US как отсортированные срезы (≈ +7.9 МБ суммарно) — каждый словарный hit ПОДАВЛЯЕТ коррекцию, т.е. полнота словаря = безопасность фичи; частотные подмножества режут покрытие ради размера, требуя при этом неверифицированный источник. Ё-вопрос: ru_RU — «е»-словарь; слова с «ё» (`ёлка` vs `елка`) могут давать ложный miss → генератору добавить нормализацию ё→е при построении таблицы ИЛИ продублировать ё-формы — пинируется golden-корпусом (риск R4).

### Q4. AT-SPI роль: живая верификация (эта машина, эта сессия)

**Шина и события** `[VERIFIED: живые пробы busctl/gdbus, эта сессия]`:
- Адрес: `gdbus call --session --dest org.a11y.Bus --object-path /org/a11y/bus --method org.a11y.Bus.GetAddress` → `('unix:path=/run/user/1001/at-spi/bus,guid=…',)` — тот же двухшаговый паттерн, что appid.Start (appid.go:118-158).
- Фокус-события — broadcast-сигналы `org.a11y.atspi.Event.Object / StateChanged` на пути бридж-объекта `/org/gnome/Zenity/a11y/<uuid>`; тело `siiva{sv}`: `STRING "focused"; INT32 1` — байт-в-байт совпадает с wire-facts appid.go:9-17.
- Match-rule `interface=org.a11y.atspi.Event.Object` получает события любого приложения (проверено busctl monitor: Zenity, gnome-shell, at-spi registry).

**GetRole по пути из фокус-события** `[VERIFIED: живой вызов]`:
```text
D_DATA7qK2mV_START (живой вывод busctl этой сессии)
[entry] GetRole on event path: u 61
[entry] GetRoleName:            s "text box"
[entry] GetAttributes:          a{ss} 1 "toolkit" "GTK"
D_DATA7qK2mV_END
```
Вызов с ЛЮБОГО a11y-соединения (busctl = отдельный процесс) по пути, полученному из события, отвечает мгновенно — объект жив и отвечает. Именно это делает D-53 («живой GetRole в момент решения») реализуемым без нового механизма.

**Ролевый enum** `[VERIFIED: /usr/bin/python3 + gi Atspi 2.52.0, эта сессия; значения и имена — дословный вывод]`:
```text
D_DATA9wR4tX_START
PASSWORD_TEXT 40 password text
ENTRY 79 entry
DOCUMENT_TEXT 94 document text
TERMINAL 60 terminal
TEXT 61 text
CANVAS 6 canvas
PARAGRAPH 73 paragraph
APPLICATION 75 application
FRAME 23 frame
D_DATA9wR4tX_END
```
Решение принимать по enum-значению GetRole (u), не по имени-строке: GetRoleName(61) = "text box", а gi-nick того же значения = "text" — строки расходятся, числа стабильны.

**Реальные роли по тулкитам** (живая сессия):
| Поверхность | Роль | Вердикт |
|---|---|---|
| GTK4 `GtkEntry` / `GtkTextView` (питон-фикстура с gir1.2-gtk-4.0 4.14.5) | `'text'` (61), focused, editable | «текстовый ввод» — разрешено `[VERIFIED: живой обход дерева]` |
| GTK4 `GtkPasswordEntry` | `'password-text'` (40) | Пароль — запрещено, различимо `[VERIFIED: живой обход]` |
| GTK3 zenity `--entry` | u 61 «text box» | Разрешено `[VERIFIED: busctl GetRole]` |
| **GTK3 zenity `--password` (GtkEntry visibility=off)** | **u 61 «text box» — НЕ различимо по роли** | **Критично: ролевый гейт не ловит GTK3-пароли; безопасность — на white-list + конъюнкции D-53** `[VERIFIED: живой обход полного дерева + GetRole]` |
| GNOME Shell свои entry (overview/lockscreen) | PASSWORD_TEXT (прецедент фазы 01-03) `[VERIFIED: test/e2e/case_m1.go:27-33]` | Запрещено, различимо |
| gnome-text-editor | app «gnome-text-editor»; фокус-инпут = ':TEXT:' (ожидание стенда) `[VERIFIED: test/e2e/case_word.go:348, surface.go:52]` | Разрешено |
| Electron (code, zcode), Chrome | application+frame есть; контент-дерево лениво (в момент обхода пусто) | Роль пинировать на живом окне при реализации; чтение page-input Chromium стендом доказано в Фазах 2–4 `[VERIFIED: обход 25 приложений + прецедент стенда]` |
| Qt (telegram-desktop) | только application-узел, дерева нет | Роль неизвестна → молчание (D-53) `[VERIFIED: живой обход]` |
| **wezterm (терминал)** | **ОТСУТСТВУЕТ на a11y-шине полностью** (0 вхождений в перечне 25 приложений) | Нет фокус-событий и роли → молчание автоматически; негативный кейс — по счётчикам ctl `[VERIFIED: живой перечень]` |
| gnome-terminal/VTE | не запущен; enum TERMINAL=60 верифицирован, роль VTE — документирована | Пинировать при реализации `[ASSUMED → MEDIUM]` |

**Q4(d): объект из фокус-события отвечал GetRole стабильно во всех пробах** (вызов шёл через секунды после события; путь бриджа стабилен, пока виджет жив) — HIGH с оговоркой, что между событием и вызовом в проде пройдут миллисекунды, а не секунды.

### Q5. Точка интеграции в акторе

**Граница слова** — `feedKey` (actor.go:1225-1284, прочитан):
- CORR-09-сбросы: Enter/KP-Enter/Tab/Escape (isResetKeyval actor.go:1597-1604) → `a.buf.HardReset()` (actor.go:1248-1255) — перед сбросом последнее завершённое слово доступно из буфера.
- Разделитель (пробел, запятая…): печатаемый не-token символ уходит в `a.buf.Push(r)` — именно Push завершает токен (buffer.go:38-49: «a separator finishes the current token», координаты в lastStart/lastLen). Актор сейчас НЕ получает уведомления о факте завершения токена — хуку нужен сигнал «токен только что завершился»: либо сравнение `Token()` до/после Push, либо минимальное расширение Buffer (событие/флаг). Backspace (buffer.go:55-61) и HardReset честно переустанавливают координаты — состояние детектора следует за буфером.
- FocusOut/Reset (HandleLifecycle actor.go:481-488) — HardReset; границей слова не считаем (слово не «завершено» нажатием).

**Состояние, нужное детектору** (всё уже в акторе, caller holds the mutex): токен+хвост (`buf.Token()/Tail()`), режим (`a.mode`), caps surrounding-text (`a.caps`, verify-rung применимость — D-53 условие 3), движок (`a.eng`), appid (`a.appid.FocusedApp()`), конфиг-снапшот (`a.opts`). Новое: **роль** — `AppidSource` (actor.go:301-303) сейчас отдаёт только имя:
```go
D_DATA2nF8hQ_START (actor.go:301-303, дословно)
type AppidSource interface {
	FocusedApp() (string, error)
}
D_DATA2nF8hQ_END
```
Observer выбрасывает сигнал после извлечения имени (appid.go:192-197) — для D-53 нужно хранить (sender, path) последнего focus-gain и отдавать Role-запрос; шов по образцу AppidSource: интерфейс `RoleSource`/расширение, определённый в точке использования, тест-даблы вместо живой шины.

**Исполнение коррекции** — только существующий конвейер: план→verify→лестница (`startRangeCorrection`/`executeLevel2`, D-27 кап; buffer.ReplaceToken/ReplacePhrase держат буфер зеркальным). Никакого второго пути замены (D-52 reuse).

**Горячий путь**: решение происходит только на границе слова (не на каждой клавише). Стоимость: binary-search по 146k ≈ микросекунды; GetRole = один D-Bus round-trip по unix-сокету ≈ доли мс (живые вызовы мгновенны); триграммы — только при словарном miss. Граничная клавиша транзитит как раньше (семантика CORR-09: «the reset is engine state, never consumption», actor.go:1249-1252). Бюджет SPEC §5 < 50 мс не под угрозой.

**Деградация — направление инвертировано относительно MACR**: macrTargetActive при недоступности appid ДЕГРАДИРУЕТ К ГЛОБАЛЬНОМУ правилу (permissive, actor.go:1045-1066) с WARN-once «app identity unavailable» (actor.go:1071-1082). Автокоррекция при ЛЮБОМ неизвестном обязана молчать (D-53) — переносится паттерн warn-once (по эпизоду, не на клавише), НЕ переносится ранг деградации. Направление безопаснее: fail-closed вместо fail-open; в ADR-007 зафиксировать отличия.

### Q6. e2e-поверхность негативных кейсов

**Реестр стенда** `[VERIFIED: test/e2e/main.go:66-101]` — `caseSpec{fn, standalone, watchdog}`, диспатч по `-case`, справочник имён в caseListUsage; новый файл `case_autocorrect.go` по образцу case_macr.go (спайк-кейс «passes on the fact of execution», негатив — валидный исход). Конфиг демона под кейс: записать YAML с секцией autocorrect + `restartDaemonWithArgs -config …` (main.go:409-415; образец — MACR-кейсы case_macr.go:477,718 с inline-конфигом `verify_wait_ms: 100`).

**Парольный негатив (ролевой гейт работает)** — GTK4-фикстура: gir1.2-gtk-4.0 4.14.5 установлен `[VERIFIED: dpkg]`; питон-фикстура ~30 строк (GtkWindow + GtkEntry + GtkPasswordEntry) **проверена живьём в этом исследовании**: окно мапится и берёт фокус как свежепорождённое (тот же механизм, что startZenity surface.go), AT-SPI-дерево отдаёт `'text'` и `'password-text'`, фокус-событие приходит. Оракул «ничего не изменилось»: witness `focused-input-pid` (role:chars до/после) + фикстура печатает содержимое поля на выходе (stdout-оракул в духе zenity). Коммитится как `test/e2e/fixtures/password_entry.py`.

**GTK3-пароль (zenity --password)** — роль 61 «text box»: для ролевого гейта НЕПРИГОДЕН как негатив (роль indistinguishable); использовать только как документирующий кейс white-list-гейта (по умолчанию zenity не в white-list → нет коррекции) — по решению планировщика.

**Терминальный негатив** — wezterm на a11y-шине отсутствует → AT-SPI-readback невозможен. Оракул: счётчики автокоррекции в ctl status (D-54) = 0 + отсутствие коррекционных записей в -debug-логе + инжекция слова с границей в сфокусированный wezterm. gnome-terminal/VTE на машине не запущен — если появится в матрице, роль TERMINAL=60 пинировать живьём.

**Chromium-пароль** — `test/e2e/fixtures/input.html` добавить `<input type="password">`; роль Chromium для него живьём не снята `[ASSUMED — static knowledge: Chromium мапит password-инпут в password-роль; пинировать при реализации по дисциплине «pin the actual»]`.

## Detector Algorithm Spec Sketch (рекомендация для планировщика)

**Входы** (на границе слова): токен `tok []rune` (buf.Token), режим `mode` (EN/RU), конфиг (пороги), dictionaries (запечённые ru/en sorted slices), trigram-таблицы, (вне детектора — политика: app white-list, роль, caps).

**Процедура** (чистая функция, TDD-корпус):
1. Нормализация: len(tok) < `min_word_len` (пинировать, кандидат 3–4; xneur использует 2 — наш порог выше из-за цены ошибки) → `abstain`. Токен без букв целевой стороны → `abstain`.
2. Словарный путь (D-52а): `cur_in = dict[mode](tok)`; `conv, ok = correct.Convert(tok, dir)`; `other_in = ok && dict[other](conv)`.
   - `!cur_in && other_in` → вердикт «wrong-layout», high confidence.
   - `cur_in` (слово валидно в текущей раскладке) → `no` — НИКОГДА не корректировать (важно: «ghbdtn»-случаи почти всегда miss в обоих, см. 3).
   - `cur_in && other_in` (случайные анаграммы) → `no`.
3. Триграммный fallback (только `!cur_in && !other_in` или `cur_in && other_in` с противоречием): скоринг `score(lang) = Σ log P(trigram | lang)` по таблицам, нормировка длиной; вердикт «wrong-layout, low confidence» если `score(other) − score(cur) > θ_diff` и `score(other) > θ_abs` (пороги пинируются golden-корпусом — Discretion).
4. Дебаунс: не повторять коррекцию одного и того же слова/позиции (ReplaceToken уже делает повтор безопасным — toggle-инвариант buffer.go:97-111; но детектор не должен пере-срабатывать на скорректированном слове: после коррекции слово валидно → ветка 2 → `no`).
5. Выход: `{verdict, confidence, direction}` — вердикты логируются счётчиком+причиной, БЕЗ слова (D-20/D-21).

**Пороги для пинирования при реализации** (Discretion): min_word_len; θ_diff; θ_abs; топ-K триграмм на язык; ё-нормализация (да/нет); поведение при mixed-токенах (reuse correct.Dir-отказа).

## Licensing Verdict Table (все кандидаты данных/кода)

| Кандидат | Тип | Лицензия (верификация) | Вердикт для MIT-репо |
|----------|-----|------------------------|----------------------|
| hunspell ru_RU (.dic/.aff) | данные | custom-BSD-4(-4 снят), © 1997-2008 Alexander I. Lebedev `[VERIFIED: /usr/share/doc/hunspell-ru/copyright]` | Разрешено: запекать производный словарь с нотисом (файл LICENSE-data/docs) |
| hunspell en_US (SCOWL) | данные | permissive «…distribute and sell … without fee…» © Kevin Atkinson + PD Moby `[VERIFIED: /usr/share/doc/hunspell-en-us/copyright]` | Разрешено с сохранением нотисов |
| /usr/share/dict/american-english | данные | SCOWL-семейство (не читал copyright этого пакета — MEDIUM) `[ASSUMED: та же семья SCOWL]` | Запасной вариант; предпочесть en_US |
| OpenCorpora freq lists | данные | НЕ верифицировано (сайт 521) `[ASSUMED: CC-BY-SA]` | Не использовать в v1.1 без отдельной верификации |
| xneur (detection.c и пр.) | код-референс | GPL-2+ `[VERIFIED: шапки файлов]` | Только чтение алгоритма; НИ СТРОКИ кода (goibus-прецедент) |
| Easy Switcher (easy-switcher.lpr) | код-референс | GPL-2.0-only `[VERIFIED: GitHub API license]` | Только чтение; словарного алгоритма не содержит (см. Q1b) |
| whatlanggo | Go-зависимость | MIT `[VERIFIED: GitHub API + proxy]` | Юридически допустим (D-52); технически НЕ рекомендован (размер/профиль задачи) |

## Architecture Patterns

### System Architecture Diagram

```text
TYPING (mode EN или RU, автокоррекция enabled и слово в white-list-приложении)
  keypress → ibus-daemon → goswitchd ProcessKeyEvent → actor.HandleKey → feedKey
    ├─ separator/Enter/Tab/Escape → ГРАНИЦА СЛОВА
    │     ├─ (1) app white-list?  ← appid.FocusedApp() (bridge-namespace)
    │     ├─ (2) роль текстового ввода? ← НОВОЕ: живой GetRole по (sender,path)
    │     │        последнего фокус-события (appid расширяется хранением пути)
    │     ├─ (3) caps surrounding-text? (verify-rung применим)
    │     ├─ (4) детектор: dict[mode](tok)? → dict[other](Convert(tok))? → trigram?
    │     └─ (5) len(tok) ≥ min?
    │   ВСЕ «да» → существующий конвейер: plan → verify → ladder → Delete/Commit
    │   ЛЮБОЕ «нет/неизвестно» → молчание (ключ транзитит, ничего не происходит)
    └─ обычная клавиша → buf.Push / transit (как в Фазах 2–4, без изменений)

OFF-THE-HOT-PATH
  layouts/dictgen (dev, go:generate) ← /usr/share/hunspell/{ru_RU,en_US}.dic
    → committed golden: sorted []string + trigram-таблицы (tables-стиль Фазы 1)

E2E SAFETY PROOFS
  GTK4-фикстура (GtkPasswordEntry, role=password-text) → инжекция «ghbdtn» + пробел
    → witness chars до=после + stdout фиксатуры → кейс зелёный = пароль не тронут
  wezterm → инжекция → ctl-счётчики autocorrect=0, -debug без коррекций
```

### Recommended Project Structure (additions only)

```
layouts/dictgen/          # генератор: hunspell → golden sorted slices + trigram-таблицы
layouts/                  # + сгенерированные dict_ru.go / dict_en.go / trigrams.go (committed)
internal/detect/          # pure: словарный путь + скоринг + пороги + Verdict/Confidence
internal/appid/           # + хранение (sender,path) последнего фокуса, RoleSource-запрос
internal/session/         # + хук границы слова в feedKey, конъюнкция D-53, счётчики
internal/config/          # + секция autocorrect (enabled/apps/пороги, валидация, дефолты off)
internal/ctlsvc/ + cmd/goswitchctl/  # + autocorrect-статус (образец MACRStats)
docs/CONFIG.md            # + секция autocorrect (D-54)
docs/adr/ADR-007-*.md     # механика (D-55): роль как источник решения, fail-closed, лицензии
docs/SPEC.md              # spec-delta §2/§10/§11 (D-55) — ДО кода детектора
test/e2e/case_autocorrect.go        # кейсы: fires / password-silent / terminal-silent
test/e2e/fixtures/password_entry.py # GTK4-фикстура (проверена живьём, см. Q6)
test/e2e/fixtures/input.html        # + password input для chromium-поверхности
```

### Anti-Patterns to Avoid

- **Роль по строковому имени** — GetRoleName(61)="text box", gi-nick="text"; только enum (u).
- **Ролевый гейт как единственная защита** — GTK3-пароли проходят его (живой факт); только конъюнкция D-53.
- **Кэш роли** — D-53 прямо запрещает: живой GetRole в момент решения; кэш appid — advisory.
- **Runtime map словарей** — +10–12 МБ RSS; только вшитые sorted slices.
- **Логирование слова** — только счётчики/причины (D-20/D-21).
- **Второй путь замены текста** — только plan/verify/ladder internal/correct.
- **Срабатывание при «probably»** — любое «неизвестно» = молчание (fail-closed); MACR-ранг «деградация к глобальному правилу» сюда НЕ переносится.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Словарные данные | ручные списки слов в исходниках | hunspell → go:generate golden | Полнота=безопасность, лицензии чистые, golden-тесты бесплатны |
| Проверка «слово существует» | префиксные деревья/блум-фильтры | `sort.SearchStrings` по срезу | 146k ≈ 17 сравнений — микросекунды; ноль кода, ноль кучи |
| Ремап токена | своя таблица | `correct.Convert` + layouts-таблицы | Уже есть, регистр-сохраняющий, протестирован |
| Замена в поле | второй механизм | план/верифай/лестница internal/correct | Единственный путь (D-52); caps/кап/хвост уже решены |
| Наблюдение фокуса | новый подписчик | internal/appid Observer | Работает с Фазы 3; расширяется, не дублируется |
| Роль поля | свой обход дерева | GetRole по (sender,path) из фокус-события | Живой вызов дешевле и точнее обхода; D-53 его и требует |

**Key insight:** весь «новый» engineering фазы — чистый детектор и его данные; всё остальное (шина, фокус, замена, конфиг, статус) уже построено предыдущими фазами и расширяется точечно.

## Runtime State Inventory

> Не rename/refactor-фаза. Словарные данные запекаются в бинарник (новое «runtime state» = сам бинарник), инсталлированное состояние не меняется (install/uninstall не трогаются). Проверено:

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | Нет нового пользовательского состояния; словари — в бинарнике | none |
| Live service config | goswitchd v1.0.0 запущен на столе; e2e-кейсы пишут свой -config (прецедент MACR) | e2e управляет своим экземпляром демона |
| OS-registered state | none — верифицировано: фаза не добавляет юнитов/XML | none |
| Secrets/env vars | none — новых нет | none |
| Build artifacts | Новый генератор dictgen → committed golden; CI гоняет только golden-тесты (паттерн layouts) | goldens коммитить; генератор — dev-only |

## Common Pitfalls

### Pitfall 1: GTK3-парольное поле проходит ролевый гейт
**What goes wrong:** авто-коррекция срабатывает в GTK3-пароле white-list-приложения.
**Why:** GTK3 GtkEntry visibility=off отдаёт роль 61 «text box» (живой факт Q4).
**How to avoid:** рассматривать роль как necessary-not-sufficient; white-list держать узким; документировать ограничение в ADR-007 и CONFIG.md; e2e-фикстура — GTK4 (там роль честная).
**Warning signs:** желание «упростить» политику до одной роли.

### Pitfall 2: Срабатывание на корректных словах с OOV-именами
**What goes wrong:** триграммный fallback «уверен» в имени собственном («Москва» в EN-раскладке Vfrcbv → …).
**Why:** короткие слова слабо различимы триграммами; смешанные тексты (код, транслит) шумят.
**How to avoid:** словарный путь приоритетен; fallback только при miss в ОБОИХ словарях; пороги — по golden-корпусу с обязательными OOV-кейсами; минимальная длина слова выше xneur-овских 2.
**Warning signs:** ложные срабатывания в корпусе → поднять пороги/длину, не «тюнить на живую».

### Pitfall 3: ё/е в ru-словаре
**What goes wrong:** слово с «ё» не находится в е-словаре → ложный miss → неверное срабатывание словарного пути.
**Why:** ru_RU hunspell — «е»-написание (Ё опционально; xneur-репо даже несёт автоматизацию hunspell-ru-ie-yo).
**How to avoid:** генератор нормализует ё→е (или дублирует формы) — решается золотым корпусом до кода детектора.
**Warning signs:** корpus-тест с «ёлка/ещё» красный.

### Pitfall 4: Стейл appid-кэш при фокусе на приложении без AT-SPI
**What goes wrong:** детектор верит последнему известному приложению (кэш) и срабатывает там, где дерева нет.
**Why:** Observer хранит последнее имя (appId Watch Out; wezterm вообще не на шине).
**How to avoid:** D-53 дословно: живой GetRole по (sender,path); роль недоступна → молчание; хранить (sender,path) вместе с именем и сбрасывать при loss-событиях.
**Warning signs:** тест «фокус на wezterm после zenity» срабатывает.

### Pitfall 5: Латентность решения на границе слова
**What goes wrong:** страх «D-Bus в горячем пути».
**Why:** GetRole — межпроцессный вызов.
**How to avoid:** измерить: локальный unix-сокет round-trip — доли мс (живые вызовы мгновенны); решение только на границе; клавиша транзитит до/независимо от решения; общий бюджет SPEC §5 50 мс с запасом.
**Warning signs:** perf-прогон Фазы 4 после фазы — сравнить p95.

### Pitfall 6: Приватность слова в логах/статусе
**What goes wrong:** слово пользователя утекает в INFO-лог или ctl-статус.
**How to avoid:** счётчики и причины (dict-hit/other-hit/trigram-low/…), само слово — никогда (D-20/D-21); corpus-тесты фиксируют это свойство детектора.

### Pitfall 7: Дублирование конвейера замены
**What goes wrong:** детекторная ветка делает «свой» delete+commit мимо verify/капа.
**How to avoid:** единственный путь internal/correct; хук актора строит range и зовёт существующие старты (startRangeCorrection-семейство).

## Code Examples

### Живой GetRole по пути фокус-события (форма вызова верифицирована busctl)

```go
// Source: живая проба этой сессии: busctl --user --address=$A11Y call :1.110 \
//   /org/gnome/Zenity/a11y/<uuid> org.a11y.atspi.Accessible GetRole  → "u 61"
// appid.Start уже держит соединение (appid.go:118-158); сигнал несёт sig.Sender+sig.Path.
const accessibleIface = "org.a11y.atspi.Accessible"

// Role запрашивает enum роли у объекта последнего фокус-события (D-53: живой запрос
// в момент решения). Роли: 40=password text, 60=terminal, 61=text, 79=entry,
// 94=document text (VERIFIED: gi Atspi 2.52.0, live).
func (o *Observer) Role(ctx context.Context) (uint32, error) {
	o.mu.Lock()
	sender, path := o.focusSender, o.focusPath // хранить с последнего focus-gain
	o.mu.Unlock()
	if path == "" {
		return 0, ErrBusClosed // «неизвестно» → политика молчит
	}
	var role uint32
	err := o.conn.Object(sender, path).CallWithContext(
		ctx, accessibleIface+".GetRole", 0).Store(&role)
	return role, err
}
```

### Детектор: словарный путь (форма, не реализация)

```go
// Source: D-52а; примитивы — correct.Convert (convert.go, прочитан) и golden-срезы dictgen.
package detect

// Verdict — исход детекции; Reasons для счётчиков без слова (D-20/D-21).
type Verdict struct {
	WrongLayout bool
	Confident   bool // словарный путь = true; триграммный = по порогам
	Dir         correct.Dir
	Reason      string // "dict-cur-miss-other-hit" | "trigram" | "abstain-len" | ...
}

// Check реализует D-52: (а) не в словаре текущей И в словаре другой после ремапа;
// (б) иначе триграммное сравнение правдоподобия. tok — буферный токен; mode — текущий
// режим актора; dicts/trigrams — запечённые данные dictgen.
func Check(tok []rune, mode string, d Data, t Trigrams, p Params) Verdict { /* … */ }
```

### GTK4-фикстура парольного поля (проверена живьём, ~30 строк)

```python
# Source: живая проба этой сессии (gir1.2-gtk-4.0 4.14.5); роль дерева — 'password-text'.
# test/e2e/fixtures/password_entry.py: печатает FINAL:<entry text> на выходе —
# stdout-оракул «ничего не изменилось» в духе zenity-поверхности стенда.
import gi
gi.require_version('Gtk', '4.0')
from gi.repository import Gtk, GLib

win = Gtk.Window(title="gs-e2e-password")
box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=6)
plain = Gtk.Entry()
secret = Gtk.PasswordEntry()          # ← роль AT-SPI: password-text (40)
box.append(plain); box.append(secret)
win.set_child(box)
win.present()
loop = GLib.MainLoop()
GLib.timeout_add_seconds(int(__import__('sys').argv[1]), loop.quit)
loop.run()
print(f"FINAL:{secret.get_text()}", flush=True)
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| xneur: proto-списки + опциональный spellchecker | Полноценные словари, запечённые в бинарник + триграммы | этот этап | Ноль рантайм-зависимостей (enchant/aspell не нужны), детерминизм для тестов |
| Ручные корректоры (Easy Switcher, буфер сканкодов + клавиша) | Автокоррекция за конъюнкцией политик | этот этап | Клавиатурный буфер не нужен — границу слова видит сам IME |
| Кэш приложения как основа решения | Живой AT-SPI GetRole в момент решения | D-53 | Отказ от кэш-семантики фиксируется в ADR-007 |
| whatlanggo как «дефолт» триграмм в Go | Своя 2-язычная мини-модель под короткие слова | это исследование | −750 КБ бинарника, −1900 LOC аудита, пороги под корпус |

**Deprecated/outdated:**
- Лор «zenity --password даст password-роль» — опровергнут живьём (роль 61); фаза 01-03-факт (shell-entries = PASSWORD_TEXT) остаётся верным и не обобщается на GTK3-приложения.
- «Xneur использует настоящие словари» — его словари 8–9 regex'ов; покрытие даёт spellchecker.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | OpenCorpora freq lists ≈ CC-BY-SA (сайт недоступен, HTTP 521, лицензию не читал) | Q3 / Licensing | Если понадобятся частотные подмножества — лицензионный риск; обход: полные hunspell (рекомендовано) |
| A2 | Chromium `<input type=password>` отдаёт password-роль в AT-SPI | Q6 | Негативный chromium-кейс не сработает как ожидалось; пинировать живьём при реализации |
| A3 | VTE (gnome-terminal) отдаёт роль TERMINAL=60 | Q4 | Терминальный кейс через роль не построить; оракул всё равно счётчики; пинировать при необходимости |
| A4 | Триграммная уверенность на 2–10-символьных словах слабая (свойство класса моделей) | Q2 | Пороги могут оказаться нечувствительными → golden-корпус скорректирует или fallback отключается порогом |
| A5 | /usr/share/dict/american-english — SCOWL-семейство permissive | Q3 | Запасной en-источник; основной en_US уже верифицирован — риск не затрагивает рекомендацию |
| A6 | GetRole на пути из фокус-события отвечает и через миллисекунды после события (проба отвечала через секунды) | Q4(d) | Если объект умрёт мгновенно — вызов вернёт ошибку → политика молчит (безопасно по построению) |
| A7 | Бинарный вес своей триграммной модели ≈ десятки КБ (не измерялось — таблицы ещё не генерировались) | Q2 | Влезет в любой бюджет; при генерации проверить golden-размер |

## Open Questions

1. **Бюджет размера бинарника (+7.9 МБ за полные словари)** — Discretion явно отдаёт предложение research, пинирует планировщик/владелец. Что известно: числа измерены; полнота словаря = безопасность. Рекомендация: принять полные словари; checkpoint:human-verify на выбор владельца, если важнее компактность.
2. **Оракул терминального негатива** — счётчики ctl vs AT-SPI (wezterm нечитаем). Рекомендация: счётчики + -debug; планировщик фиксирует в кейсе.
3. **Ё-нормализация** — генератором (ё→е) или дубль-формами; решит golden-корпус (Pitfall 3).
4. **Порядок spec-delta vs detector-package** — Discretion планировщика; D-55 требует «ДО кода детектора» только для spec-правок, ADR-007 может идти параллельно с dictgen.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| hunspell-ru + hunspell-en-us | dictgen (dev, разово) | ✓ установлены `[VERIFIED: dpkg]` | 24.2.1 / 2020.12.07 | скачать от upstream (лицензии те же) |
| gir1.2-gtk-4.0 + python3-gi | e2e password-фикстура | ✓ `[VERIFIED: dpkg]` | 4.14.5 / 3.48.2 | нет нужды — есть |
| a11y-шина + GetRole | политика роли | ✓ живая `[VERIFIED: пробы]` | at-spi2 2.52.0 | молчание (fail-closed by design) |
| wezterm | терминальный негатив | ✓ запущен (вне a11y) `[VERIFIED]` | — | счётчики-оракул |
| proxy.golang.org | верификация whatlanggo | ✓ `[VERIFIED: эта сессия]` | — | — |
| opencorpora.org | частотные списки | ✗ HTTP 521 `[VERIFIED: попытка]` | — | полные hunspell (рекомендовано) |
| gnome-terminal/VTE | опциональный кейс | ✗ не запущен | — | wezterm-кейс; VTE пинировать позже |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** opencorpora (обход — полные словари), VTE-кейс (обход — wezterm).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + golden files; strict TDD (`tdd_mode: true`) |
| Config file | `mise.toml` + `.golangci.yml` |
| Quick run command | `mise run ci` (build+vet+lint+test -race) |
| Full suite command | `mise run ci` + live e2e-кейсы на столе владельца |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| AC-01 | dictgen: hunspell → golden срезы/триграммы; ё-нормализация | unit+golden | `go test ./layouts/ -run TestDictGolden` | ❌ Wave 0 |
| AC-01 | детектор: словарный путь, fallback, пороги; приватность логов | unit (golden-корпус слов — обязателен) | `go test -race ./internal/detect/` | ❌ Wave 0 |
| AC-02 | appid хранит (sender,path); Role(); отказ → молчание | unit (синтетические сигналы — шов feed) | `go test -race ./internal/appid/` | ❌ Wave 0 |
| AC-02/05 | конъюнкция D-53 в акторе; unknown → молчание; счётчики | unit (фейки RoleSource/AppidSource) | `go test -race ./internal/session/ -run AutoCorrect` | ❌ Wave 0 |
| AC-03 | схема autocorrect: strict decode, дефолты off, hot reload | unit | `go test -race ./internal/config/ -run AutoCorrect` | ❌ Wave 0 |
| AC-04 | e2e: fires (GTE) / password-silent (GTK4-фикстура) / terminal-silent | e2e (live) | `go run ./test/e2e -case autocorrect-fires` (+ -password-silent, -terminal-silent) | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `mise run ci`
- **Per wave merge:** `mise run ci` + соответствующие live-кейсы
- **Phase gate:** полный зелёный набор + perf-сравнение p95 (Pitfall 5) перед verify-work

### Wave 0 Gaps
- [ ] `layouts/dictgen/` + golden-данные + `layouts`-тесты
- [ ] `internal/detect/` пакет + golden-корпус (включая ё-слова, OOV-имена, mixed)
- [ ] `internal/appid` расширение + тесты фида
- [ ] `internal/config` autocorrect-секция + валидация
- [ ] `test/e2e/fixtures/password_entry.py` (прототип проверен живьём)
- [ ] `test/e2e/case_autocorrect.go` + регистрация имён кейсов

## Security Domain

`security_enforcement: true`, ASVS L1 (`.planning/config.json`).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | нет auth-поверхностей |
| V3 Session Management | no | — |
| V4 Access Control | partial | fail-closed политика D-53: любое «неизвестно» → молчание; white-list default пуст; роль — живой запрос, не кэш |
| V5 Input Validation | yes | конфиг-секция autocorrect под strict decode D-33 с потолками (образец: `maxMACRApps = 64` — config.go:25); пороги детектора в диапазонах; словарные данные валидируются генератором (golden) |
| V6 Cryptography | no | — |
| V14 Config | yes | дефолты enabled:false + apps:[] (D-54); hot reload публикует last-good снапшот (D-32) — битый autocorrect-блок инвалидирует весь файл (D-33) |

### Known Threat Patterns for автокоррекция IME

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Пароль, набранный в wrong layout, «исправлен» и закоммичен | Tampering/Information disclosure | конъюнкция D-53 + GTK4-роль; GTK3-ограничение задокументировано; white-list default off |
| Утечка набранного слова через логи/статус | Information disclosure | только счётчики/причины (D-20/D-21); корпус-тест на приватность |
| DoS через конфиг (гигантские списки/пороги) | DoS | потолки списка/диапазоны порогов в Validate (прецедент MACR) |
| Испорченный golden-словарь (supply chain данных) | Tampering | данные проходят генератор + golden-тесты в CI; лицензионные нотисы в репо |

## Implications for Planning

**Рекомендуемая нарезка** (порядок: spec-delta первым — D-55):
1. **План A (docs):** spec-delta §2/§10/§11 + ADR-007 (роль как источник решения; отказ от кэша; fail-closed; лицензионная позиция: hunspell BSD/SCOWL, доноры GPL read-only, whatlanggo отклонён технически) + docs/CONFIG.md-секция.
2. **План B (данные):** `layouts/dictgen` — генератор (hunspell-парсер: снять флаги, len≥2, сортировка, ё-нормализация) → golden `dict_ru.go`/`dict_en.go`/триграммные таблицы + golden-тесты; mise-задача regen.
3. **План C (детектор):** `internal/detect` — чистый Verdict/Check + golden-корпус (корпус обязателен по Discretion; включает OOV/имена/ё/mixed).
4. **План D (роль):** appid хранит (sender,path)+Role-запрос (шов `UseAppidStarter`-стиль); тесты на синтетическом фиде.
5. **План E (актор+конфиг+ctl):** autocorrect-секция конфига, SetOptions-расширение, хук границы слова (сигнал завершения токена из Buffer), конъюнкция D-53, счётчики + ctl/status.
6. **План F (e2e):** фиксатура + case_autocorrect.go (fires / password-silent / terminal-silent) + спайк-пин chromium-пароля.

**Риски:** R1 GTK3-пароли вне ролевого гейта (мифигируется white-list+дока); R2 триграммные пороги потребуют тюнинга корпуса (Discretion это и предполагает); R3 ё-нормализация (Pitfall 3); R4 wezterm-оракул только по счётчикам; R5 +7.9 МБ бинарник — подтвердить владельцем (checkpoint при желании).

**Прецеденты для задач:** спайк-кейс «passes on fact of execution» (case_macr.go), warn-once деградация (actor.go:1071), strict-decode потолки (config.go), golden-генератор (layouts/generator).

## Sources

### Primary (HIGH confidence — живые пробы/чтение этой сессии)
- AT-SPI live: GetAddress (gdbus), StateChanged-монитор (busctl), GetRole/GetRoleName/GetAttributes по путям событий, обход 25 приложений (Electron/Qt/GTK3/GTK4/shell), GTK4-фикстура с GtkPasswordEntry, отсутствие wezterm — все пробы воспроизведены на целевой машине
- Ролевый enum: `/usr/bin/python3` + gi Atspi 2.52.0 (значения процитированы дословно в Q4)
- Лицензии: `/usr/share/doc/hunspell-ru/copyright`, `/usr/share/doc/hunspell-en-us/copyright` (прочитаны; ключевые строки процитированы)
- xneur: `xneur/lib/ai/detection.c` (444 строки, прочитан), proto/dictionary-файлы fetched; шапки GPL-2+
- Easy Switcher: `easy-switcher.lpr` (прочитан), README; GitHub API license GPL-2.0
- Измерения: proxy-скачивание whatlanggo v1.0.1; сборки-замеры бинарника/RSS (go1.27.1) — команды и числа в тексте
- In-repo (прочитаны этой сессией): internal/session/actor.go (feedKey/CORR-09/AppidSource/MACR-деградация/Options), internal/correct/{buffer,convert,plan}.go, internal/appid/appid.go, internal/config/config.go, layouts/generator/main.go, test/e2e/{main,matrix,case_macr,case_m1,case_word,surface}.go, focus_helper.py, docs/SPEC.md §2/§10/§11, docs/CONFIG.md, docs/adr/ADR-005, .planning/ROADMAP.md §Phase 5, phase CONTEXT

### Secondary (MEDIUM confidence)
- [AndrewCrewKuznetsov/xneur-devel](https://github.com/AndrewCrewKuznetsov/xneur-devel) — upstream-репо (launchpad-тарболы в dists/)
- [freemind001/easy-switcher](https://github.com/freemind001/easy-switcher) — актуальный Easy Switcher (admig/easy-switcher = 404)
- [abadojack/whatlanggo](https://github.com/abadojack/whatlanggo) + [proxy.golang.org](https://proxy.golang.org/github.com/abadojack/whatlanggo/@latest) — версия/лицензия/пакет
- [SCOWL/wordlist](http://wordlist.sourceforge.net/) — семейство en-списков (пакетный copyright первичен)

### Tertiary (LOW confidence)
- OpenCorpora лицензия (сайт недоступен) — A1
- Chromium/VTE парольные/терминальные роли — A2/A3 (пинировать живьём)

## Metadata

**Confidence breakdown:**
- Детектор-домен (xneur/доноры/данные/размеры): HIGH — первоисточники прочитаны, числа измерены
- AT-SPI политика: HIGH на GTK4/GTK3/shell/wezterm (живьём), MEDIUM на Chromium-пароль/VTE (A2/A3)
- Интеграция в актора: HIGH — код прочитан, швы названы
- Триграммные пороги: по определению MEDIUM — пинируются golden-корпусом при реализации (Discretion)

**Research date:** 2026-09-27
**Valid until:** 2026-10-27 (стабильный домен; версии словарных пакетов — по needed)
