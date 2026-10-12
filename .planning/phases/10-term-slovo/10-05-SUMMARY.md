---
phase: 10-term-slovo
plan: 05
subsystem: live-e2e-vte
tags: [vte-live-rows, pty-oracle, terminal-veto, a2-key-loss, pacing, a11y-registry, clipboard-sentinel, vm-stand]

requires:
  - phase: 10-term-slovo plan 01
    provides: вердикт площадки vm-primary, стенд goswitch-stand (goswitchd -debug, зеркало, /tmp/stand-env.sh), стенд-механика (арминг по key-trace, Enter через type)
  - phase: 10-term-slovo plan 04
    provides: терминальная ветка записи (чанкированный level-2, verify-after noRung), вердикты OQ 1 unified / OQ 2 contract, журнальные оракулы done+level-2
provides:
  - Кейс term-word-vte: живые ряды слова в GNOME Terminal по pty-оракулу — bash и zsh пан-программы, мид-лайн контракт-ряд (OQ 2), замороженный CLIPBOARD-сентинел; done+level-2 грейпы пиннят терминальную ветку
  - Кейс term-negative-witnesses: замороженные негативные ряды vim-insert/htop ВНУТРИ gnome-terminal (fired=0, terminal-veto воздержания растут, буфер vim дословен) при зелёном positive-control в том же классе
  - ОТВЕТ A2 (вопрос 10-04): тесный бурст ForwardKeyEvent+CommitText переставляется в VTE (клиент разбирает форварды лениво, коммит синхронен) — executeLevel2 теперь пакирует провод (sleep 20ms перед каждым форвардом и перед коммитом); после фикса ряды байт-точные
  - Наблюдатель демона регистрирует собственный интерес в at-spi реестре (RegisterEvent) — мосты эмитят фокус-события только пока кто-то держит регистрацию; на голом столе роль-источник больше не слепнет
  - mise-задачи e2e-term-word-vte / e2e-term-negative-witnesses (NOT in ci); датированные ряды в adr-008-spike-log.md с маркером venue: vm
affects: [10-06 (амендмент ADR-008: закрепить A2-ответ и pacing в D3/D4; полный wezterm re-pin), phase-verification (критерии 2/3/5 ROADMAP закрыты живыми рядами)]

# Actuals (#2632) — pairs with the plan's `estimate` (55000) to calibrate future estimates.
actuals:
  tokens: 21101        # chars/4 over the realized diff 1b7771b..HEAD (84406 chars, 18 files)
  tasks: 3
  commits: 11          # MEASURED: git rev-list --count gsd-plan-head-before-10-05..HEAD

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "pty-оракул с канонической строкой: коммитted-текст сидит в line buffer до type $'\\n' — capture содержит ФИНАЛЬНУЮ строку (байт-точная земля без a11y)"
    - "арминг рядов по key-trace с самоуничтожающимся пробником (x + BackSpace): доставка доказана, буфер/строка не загрязнены; named-key Enter форм на ydotool 0.1.8 стенда не существуют"
    - "at-spi регистрация интереса: raw bus match слеп на голом столе — мосты эмитят только при живой регистрации в реестре (GetRegisteredEvents-гейт)"
    - "счётчики-оракул вето: ac_skip_terminal_veto парсится из goswitchctl status, дельта до/после набора отделяет вето конкретного свидетеля от накопленных воздержаний"

key-files:
  created:
    - "test/e2e/case_term_word_vte.go (кейс term-word-vte: VTE-фикстура, pty-оракул, CLIPBOARD-сентинел, venue-маркер, 4 ряда)"
    - "test/e2e/case_term_negative.go (кейс term-negative-witnesses: vim-insert/htop свидетели, counters-оракул, positive-control, re-pin ряд)"
    - "internal/appid/appid_registry_test.go (корпус регистрации интереса — RED/GREEN фикса a11y-слепоты)"
  modified:
    - "internal/session/actor.go (executeLevel2 пакирует провод: level2ForwardGap 20ms перед каждым форвардом и перед коммитом; seam UseLevel2Sleeper)"
    - "internal/session/actor_test.go (TestTerminalBranch_Level2PacesWire + fakeSink.commitHook)"
    - "internal/appid/appid.go (Start регистрирует Object:state-changed в at-spi реестре)"
    - "test/e2e/main.go (регистрация кейсов, caseRegistry, caseSpec.noSurfacePreflight, startSetupDaemon unit-snapshot до reap, restoreUnit с awaitNameFree)"
    - "test/e2e/preflight.go (skipSurfaces — chromium/gte/gedit проверки только для кейсов, которые их используют)"
    - "test/e2e/case_autocorrect.go (spawn-лестница активирует движок с ретраями)"
    - "test/e2e/case_switch.go (restoreUnit ждёт освобождения имени компонента)"
    - "test/e2e/spike_journal_test.go (секция живых рядов в корпусе секций)"
    - "mise.toml (e2e-term-word-vte, e2e-term-negative-witnesses — NOT in ci)"
    - "docs/adr/adr-008-spike-log.md (секция «Фаза 10 живые ряды VTE»: 9 датированных рядов)"

key-decisions:
  - "A2 ОТВЕЧЕН живым рядом (вопрос 10-04 «потеря клавиш в очереди IME решает живой ряд»): тесный бурст уровня wire переставляется в VTE — форварды стоят в GDK-очереди клиента, коммит исполняется синхронно; живой hex «ghbdtn»→«ghприв» (порядок e1..e4, COMMIT, e5..e6). Фикс — pacing 20ms перед каждой эмиссией executeLevel2; wire-порядок и геометрия чанков не тронуты (op-log пины зелёные), после фикса ряды байт-точные («привет\n» ровно)"
  - "Наблюдатель a11y обязана регистрировать интерес в реестре (RegisterEvent Object:state-changed): мосты приложений спрашивают GetRegisteredEvents и эмитят ТОЛЬКО при живой регистрации (живая проба: 0 StateChanged до, 153 после) — raw bus match слеп на голом столе; отказ регистрации валит Start (честная деградация WARN вместо тихой слепоты)"
  - "Арминг живых рядов — ПО KEY-TRACE с самоуничтожающимся пробником (x + BackSpace): план предписывал key-trace-гейт, pty-пробная форма armGate оказалась непереносимой (named-key Enter не резолвится ydotool 0.1.8 стенда и печатает 'k'); доставка остаётся за venue-probe r1/r2 и pty-оракулом самих рядов"
  - "Мид-лайн ряд по вердикту OQ 2 «contract»: префикс+слово набираются одним флоу (курсор не двигается), оракул — байт-смежные «echo привет\\n» в capture; контрактная граница D7 не нарушается"
  - "Re-pin autocorrect-terminal-silent — строка журнала, никогда вердикт кейса (план: «стендовый wezterm не нужен для VTE-части… полный wezterm re-pin — план 10-06»); живой прогон честно записан blocked (zenity-свидетель фокуса прибит к gnome-shell — известная недостоверность AT-SPI)"

patterns-established:
  - "VTE-фикстура с уникальным tmp-маркером в cmdline pane-программы: teardown pgrep-ом по маркеру + reap клиента (EOF-чорда на стенде не трейсится)"
  - "venue-маркер ряда: systemd-detect-virt внутри ВМ (stand-env живёт на хосте, не в ВМ) — каждая строка несёт venue: vm (kvm)"
  - "noSurfacePreflight: поверхностные драйвер-проверки preflight выполняют только кейсы, которые используют эти поверхности (философия matrixPreflight)"

requirements-completed: [TERM-02]  # declared by ALL six 10-xx plans — shared-ID gate keeps it open until 10-06 (ready-ids: 0/1)

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "Кейс term-word-vte: VTE-фикстура + pty-оракул + сентинел CLIPBOARD + ряды bash/zsh/мид-лайн, регистрации в main.go (реестр+usage+error), mise-задача NOT in ci"
    requirement: TERM-02
    verification:
      - kind: other
        ref: "mise run ci зелёный на HEAD (build+vet+golangci-lint v2 strict 0 issues+test -race); grep term-word-vte в main.go, e2e-term-word-vte в mise.toml"
        status: pass
      - kind: unit
        ref: "test/e2e/spike_journal_test.go — секция живых рядов в корпусе (заголовок — контракт)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Критерий 2 ROADMAP: слово исправляется в GNOME Terminal по pty-оракулу в bash и zsh на живой сессии ВМ (вердикт 10-01 vm-primary)"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#word-bash-191739 — pty_hex=d0bfd180d0b8d0b2d0b5d1820a («привет\\n» ровно); exit кейса 0; done+level-2 грейпы в журнале демона"
        status: pass
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#word-zsh-191741 — pty_hex=d0bfd180d0b8d0b2d0b5d1820a (тот же протокол, pane zsh -f)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Критерий 3 ROADMAP: замороженный ряд CLIPBOARD-сентинел — сентинел GS-SPIKE-класса байт-нетронут терминальной коррекцией (затирание пред-стейта задокументировано в ряду)"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#clip-frozen-191746 — clip_before=clip_after=GS-SPIKE-SENTINEL, pty_hex содержит полный «привет»"
        status: pass
    human_judgment: false
  - id: D4
    description: "Мид-лайн ряд по семантике вердикта OQ 2 «contract»: набор сразу в позицию курсора через разделитель, префикс не тронут, байт-смежная земля"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#mid-line-191743 — pty_hex=6563686f20d0bf…d1820a («echo привет\\n» ровно)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Критерий 5 ROADMAP: замороженные негативные ряды — автокоррекция молчит в vim-insert и htop ВНУТРИ gnome-terminal (fired=0, terminal-veto воздержания Δ=1 в каждом, буфер vim дословен) при зелёной ручной коррекции в том же классе (positive-control байт-точный)"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#witness-vim-192928 — buffer_hex=67686264746e200a («ghbdtn␠» дословно), fired=0, veto Δ=1"
        status: pass
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#witness-htop-192931 — fired=0, veto Δ=1, htop жив; positive-control 19:29:34 — pty_hex=полный «привет»"
        status: pass
    human_judgment: false
  - id: D6
    description: "A2-ответ и фикс: тесный бурст write-ветки переставляется в VTE (A2 живым рядом), executeLevel2 пакирован (20ms перед каждой эмиссией), юнит пиннит paced-таймлайн sleep→forward→…→sleep→commit; после фикса живые ряды байт-точные"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/session/actor_test.go — TestTerminalBranch_Level2PacesWire (RED_EVIDENCE_OK → GREEN); op-log пины 10-04 (burstOps) зелёные без правок"
        status: pass
      - kind: e2e
        ref: "ряды word-bash/word-zsh/mid-line после фикса — байт-точные hex (D2/D4 выше); до фикса честный FAIL «ghприв» задокументирован в term-word-vte-mechanics"
        status: pass
    human_judgment: false
  - id: D7
    description: "Наблюдатель a11y самодостаточна на голом столе: Start регистрирует Object:state-changed в at-spi реестре (мосты эмитят только при живой регистрации — проба 0→153 StateChanged); отказ регистрации валит Start (WARN-деградация вызывающей стороны)"
    requirement: TERM-02
    verification:
      - kind: unit
        ref: "internal/appid — TestRegisterEventInterest_PinsRegistryCall (RED_EVIDENCE_OK → GREEN): ровно один RegisterEvent с именем события, отказ — обёрнутая именованная ошибка"
        status: pass
      - kind: e2e
        ref: "роль 60 видима на ВМ без сторонних a11y-потребителей: вето/маршрутизация рядов D2/D5 сработали через классификационный шов (term-word-vte-mechanics, находка 3)"
        status: pass
    human_judgment: false
  - id: D8
    description: "Re-pin ряд autocorrect-terminal-silent на площадке — план предписывал прогон с оговоркой «полный wezterm re-pin — план 10-06»; живой прогон честно записан blocked (zenity-свидетель фокуса прибит к gnome-shell — известная недостоверность AT-SPI, журнал спайка 2026-10-11)"
    requirement: TERM-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md#autocorrect-terminal-silent-repin-195937 — Verdict: blocked, ошибка дословно в ряду"
        status: unknown
    human_judgment: true
    rationale: "Blocked-ряд — запроектированный исход (план отнёс полный wezterm re-pin к 10-06); вердикт о достаточности VTE-пиннинга негатива — за фазовой верификацией/владельцем"
  - id: D9
    description: "Flagged assumption плана: формальная классификация края TERM-02 остаётся unresolved (SPEC без ## Edge Coverage) — живые ряды покрывают край (bash/zsh/мид-лайн/клипборд/негатив), но вердикт классификации — за фазовой верификацией"
    requirement: TERM-02
    verification:
      - kind: other
        ref: ".planning/phases/10-term-slovo/10-05-PLAN.md — Flagged Assumptions: {\"requirement_id\":\"TERM-02\",\"category\":\"unclassified\",\"status\":\"unresolved\"}"
        status: unknown
    human_judgment: true
    rationale: "Flagged assumption плана никогда не авто-разрешается: verifier обязан классифицировать край явно, не тихим pass"

# Metrics
duration: 102min
completed: 2026-10-12
status: complete
---

# Phase 10 Plan 05: живые ряды VTE — слово и молчание на площадке vm-primary Summary

**Критерии 2/3/5 ROADMAP закрыты живыми рядами на ВМ: слово исправляется в GNOME Terminal байт-точно по pty-оракулу (bash, zsh, мид-лайн «echo привет\n»), CLIPBOARD байт-нетронут (сентинел GS-SPIKE), автокоррекция молчит в vim-insert/htop при зелёной ручной в том же классе; попутно A2 отвечен живым рядом — бурст write-ветки переставлялся в VTE («ghbdtn»→«ghприв»), теперь executeLevel2 пакирован и ряды байт-точны, а наблюдатель a11y сама держит регистрацию интереса в реестре (голый стол больше не слепит роль-источник).**

## Performance

- **Duration:** 102 min
- **Started:** 2026-10-12T18:25:33Z
- **Completed:** 2026-10-12T20:07:40Z
- **Tasks:** 3/3 (Task 1 tracer — код кейса, гейт пройден расширенным verify; Tasks 2-3 — живые прогоны на ВМ)
- **Files modified:** 12 репозиторных (2 новых кейса, 1 новый тест-файл, actor/appid фиксы, реестр/префлайт/mise/журнал) + 1 evidence-JSON в .planning

## Accomplishments

- **Критерий 2 (слово в терминале):** word-bash и word-zsh — pty_hex=d0bfd180d0b8d0b2d0b5d1820a («привет\n» РОВНО, байт-точно) на живой сессии goswitch-stand; грейпы done+level-2 пиннят именно терминальную ветку записи 10-04; exit кейса 0.
- **Критерий 3 (CLIPBOARD):** clip-frozen — сентинел GS-SPIKE-SENTINEL до и после коррекции байт-равен (wl-paste --no-newline), слово исправлено; затирание пред-стейта задокументировано в ряду (D5).
- **Критерий 5 (молчание):** witness-vim (buffer_hex=67686264746e200a — «ghbdtn␠» дословно) и witness-htop (процесс жив) — fired=0, terminal-veto воздержания Δ=1 в каждом; positive-control в том же классе — «привет\n» байт-точный: молчание есть вето 10-03, не глухота канала.
- **A2 отвечен:** живой ряд поймал перестановку бурста в VTE («ghbdtn»→«ghприв», порядок e1..e4/COMMIT/e5..e6); фикс — pacing executeLevel2 (sleep 20ms перед каждым форвардом и перед коммитом), юнит пиннит paced-таймлайн, op-log пины 10-04 зелёные без правок, после фикса все ряды байт-точные.
- **a11y-наблюдатель самодостаточна:** Start регистрирует Object:state-changed в at-spi реестре (мосты эмитят только при живой регистрации — проба 0→153); на голом столе ВМ роль 60 стала видима без сторонних потребителей, классификационный шов (вето + ветка записи) работает из коробки.
- **Стенд-механика закреплена датированным рядом:** re-sync рецепт, DISPLAY/XAUTHORITY для ssh-активации, GTK_MODULES=gail:atk-bridge для терминал-сервера ВМ, KEY_ENTER-ловушка ydotool 0.1.8 (арминг по key-trace с пробником x+BackSpace).
- Строгий TDD на обоих кодовых фикса (RED_EVIDENCE_OK для каждого), mise run ci зелёный на каждом коммите, ноль новых модулей (go.mod/go.sum нетронуты).

## Task Commits

1. **Task 1 (tracer): кейс term-word-vte + регистрации + mise + секция журнала** - `b4f7f35` (feat); tracer-гейт: verify перезапущен end-to-end зелёным (auto_advance) — расширение авторизовано
2. **Task 2: живой прогон term-word-vte на площадке** - `e5357b3` (fix: три блокера харнесса), `ce15b57` (fix: арминг по key-trace), `667b0a4`+`5f7837d` (test+feat: регистрация интереса a11y), `00e2799`+`710b6d4` (test+feat: pacing write-ветки), `7c58f2c` (docs: 4 зелёных ряда + механика)
3. **Task 3: замороженные негативные ряды** - `1fbc4f9` (feat: кейс term-negative-witnesses), `4b39678` (fix: ретрай активации в spawn-лестнице), `e8ad4db` (docs: 4 ряда — свидетели/контроль/re-pin)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP)

_Note: два кодовых фикса несли строгий TDD (RED_EVIDENCE_OK → GREEN); живые ряды — verify-свидетельство плана, не юнит-тесты_

## Files Created/Modified

- `test/e2e/case_term_word_vte.go` — кейс term-word-vte: 4 ряда (bash/zsh/мид-лайн/клип), termArmGate по key-trace, pty-оракул, venue-маркер, honest-abort строки
- `test/e2e/case_term_negative.go` — кейс term-negative-witnesses: vim/htop свидетели, counters-оракул (ac_skip_terminal_veto Δ), positive-control, re-pin строка
- `internal/session/actor.go` — level2ForwardGap=20ms + paced executeLevel2 + seam UseLevel2Sleeper (UseRole-форма)
- `internal/appid/appid.go` — Start регистрирует интерес в at-spi реестре (RegisterEvent Object:state-changed), отказ валит Start
- `test/e2e/main.go` — реестр/usage/error-список (2 кейса), caseRegistry (funlen), caseSpec.noSurfacePreflight, startSetupDaemon: unit-snapshot ДО stale-reap, restoreUnit: awaitNameFree перед стартом
- `test/e2e/preflight.go` — skipSurfaces: chromium/gte/gedit только для кейсов, которые их используют
- `test/e2e/case_autocorrect.go` — spawn-лестница: activateGoswitchRetry вместо одиночного SetGlobalEngine
- `test/e2e/case_switch.go` — restoreUnit ждёт освобождения имени компонента
- `test/e2e/spike_journal_test.go`, `mise.toml`, `docs/adr/adr-008-spike-log.md` — корпус секций, 2 mise-задачи, 9 датированных рядов
- `.planning/phases/10-term-slovo/evidence-20261012-10-05-red-{appid,task2}.json` — TDD RED-свидетельства

## Decisions Made

См. key-decisions. Ядро: A2-ответ (перестановка в VTE) зафиксирован живым hex и починен пакированием провод-эмиссий без изменения wire-порядка и геометрии чанков; a11y-наблюдатель получает собственную регистрацию интереса (правильная интеграция любого a11y-потребителя); арминг живых рядов — key-trace с пробником x+BackSpace (named-Enter формы на стенде несуществующи); мид-лайн — контракт-семантика OQ 2 байт-смежным оракулом; re-pin — строка журнала, не вердикт кейса.

## Deviations from Plan

**[Rule 3 - Blocking] Харнесс e2e не был готов к стенду ВМ (три блокера первого прогона)**
- **Found during:** Task 2, прогон 1
- **Issue:** (а) unit-демон ВМ несёт -debug (10-01 оракул) → pgrep stale-обхода убивал его ДО снапшота → teardown молча пропускал restore (стенд остался без демона); (б) restoreUnit стартовал юнит против ещё не освобождённого имени (находка 2026-10-13) — демон-рестарт; (в) preflight требовал chromium/gte/gedit — на ВМ их нет, а честного архивного пути для chromium не существует (snap)
- **Fix:** снапшот unitState ПЕРЕД stale-reap (на хостах байт-эквивалентно); awaitNameFree перед стартом юнита; caseSpec.noSurfacePreflight (философия matrixPreflight — поверхностные проверки несут кейсы, которые их используют)
- **Files modified:** test/e2e/main.go, test/e2e/preflight.go, test/e2e/case_switch.go
- **Verification:** последующие прогоны: preflight зелёный, юнит восстановлен после каждого teardown
- **Commit:** e5357b3

**[Rule 3 - Blocking] Арминг-форма armGate непереносима на стенд: named-key Enter не резолвится**
- **Found during:** Task 2, прогон 3
- **Issue:** ydotool 0.1.8 стенда не знает KEY_ENTER — печатает первую букву ('k' в key-trace), каноническая строка не смывается, pty-пробник читался как ложный stale-context
- **Fix:** termArmGate по KEY-TRACE (буквально по плану): пробник x + BackSpace (venue-mechanics: BackSpace доходит, 0xff08), буфер чист; доставка остаётся за venue-probe r1/r2 и pty-оракулом самих рядов
- **Files modified:** test/e2e/case_term_word_vte.go
- **Commit:** ce15b57

**[Rule 2 - Missing critical] Наблюдатель a11y слепа на голом столе (роль-источник фазы зависел от сторонних a11y-потребителей)**
- **Found during:** Task 2, прогон 4 (role-unknown на жесте → verify-timeout вместо терминальной ветки)
- **Issue:** мосты GTK-приложений спрашивают GetRegisteredEvents и эмитят StateChanged ТОЛЬКО при живой регистрации в реестре; raw bus match демона регистрации не делает — на хосте работало по случайности (сторöne потребители держат регистрацию)
- **Fix:** Start регистрирует Object:state-changed (RegisterEvent); отказ валит Start → честная WARN-деградация вызывающей стороны; TDD RED (667b0a4, RED_EVIDENCE_OK) → GREEN (5f7837d); живая проба 0→153 StateChanged
- **Files modified:** internal/appid/appid.go, internal/appid/appid_registry_test.go, internal/appid/export_test.go
- **Commit:** 667b0a4, 5f7837d

**[Rule 1 - Bug] Тесный бурст write-ветки переставляется в VTE (A2: ответ «да, ключи теряются»)**
- **Found during:** Task 2, прогон 5 (живой hex «ghbdtn»→«ghприв»; порядок e1..e4, COMMIT, e5..e6 — клиент разбирает форварды лениво, коммит синхронен)
- **Issue:** executeLevel2 эмитировал 6 форвардов + коммит подряд — перестановка портила исправленную строку (критерий 2 не выполнялся бы никогда)
- **Fix:** pacing — sleep level2ForwardGap (20ms) перед КАЖДЫМ форвардом и перед коммитом; wire-порядок и чанк-геометрия не тронуты; TDD RED (00e2799, RED_EVIDENCE_OK) → GREEN (710b6d4); после фикса ряды байт-точные
- **Files modified:** internal/session/actor.go, internal/session/actor_test.go
- **Commit:** 00e2799, 710b6d4

**[Rule 3 - Blocking] Одиночный SetGlobalEngine в spawn-лестнице конфиг-демона гонял свежий движок**
- **Found during:** Task 3, прогон 1 («Operation was cancelled» сразу после readiness-марок)
- **Fix:** activateGoswitchRetry (3 попытки, форма case_resilience) вместо одиночного вызова; первый успех байт-как-сегодня
- **Files modified:** test/e2e/case_autocorrect.go
- **Commit:** 4b39678

**Расширение files_modified Task 1 (журнал + corpus секций):** секция «Фаза 10 живые ряды VTE» создана пустой в Task 1 вместе с константой и записью в spikeSections — TestSpikeJournalSectionsExist требует существующий заголовок (контракт секции); предписано контрактом пин-теста, не планом. Commit b4f7f35.

**Стендовые мутации вне репо (не отклонения плана — доводка внутренностей,venue-mechanics-прецедент):** zsh поставлен из архива (ряд word-zsh); gnome-terminal-server ВМ перезапущен с GTK_MODULES=gail:atk-bridge (a11y-мост); DISPLAY/XAUTHORITY в ssh-сессиях активации; все задокументированы рядом term-word-vte-mechanics.

**Total deviations:** 6 auto-fixed (1 Rule 1 — A2-перестановка с TDD; 1 Rule 2 — a11y-регистрация с TDD; 4 Rule 3 — харнесс/арминг/ретрай/секция-журнала). **Impact:** нулевое на контракт задач — все три задачи исполнены, критерии 2/3/5 закрыты живыми рядами; два фикса демона (pacing, a11y-регистрация) усиливают фазу и вынесены в ADR-008-амендмент 10-06.

## Issues Encountered

- Шесть заходов живого прогона term-word-vte до зелёного (18:40 env, 18:47 preflight, 19:06 арминг, 19:25 активация, 19:47 — honest-FAIL ghприв; 20:06 — зелёный): каждый FAIL — честная строка/девиация с первопричиной, ни одного ретрай-цикла одной и той же формы; брутфорс живой сессии отсутствует (T-10-05-04).
- Re-pin autocorrect-terminal-silent на площадке записан blocked: zenity-свидетель фокуса waitZenityEntry не увидел фокуса диалога (witness прибит к gnome-shell — известная глобальная недостоверность AT-SPI-репортинга); полный wezterm re-pin — план 10-06 (плановая развилка, ряд в журнале).
- Прогон term-negative-witnesses ожидаемо потратил 30-минутный бюджет waitZenityEntry на blocked re-pin (19:29→19:59) — бюджетAutorизации кейса (autocorrectWatchdog) выдержал.

## User Setup Required

None - no external service configuration required. (Площадка vm-primary: владелец не нужен; все мутации стенда задокументированы в журнале, откаты — стандартные: sources [("xkb","us")], GTK_MODULES без override при пересоздании сессии ВМ.)

## Next Phase Readiness

- План 10-06 получает: закрытые критерии 2/3/5 (живые ряды в журнале с venue: vm), A2-ответ для амендмента ADR-008 (D3/D4 дополняются pacing-фактом), готовый стенд (a11y-мост, zsh, htop, vim, observer-регистрация) и терминированную форму Enter (type $'\n').
- Полный wezterm re-pin 10-06: wezterm-nightly на борту ВМ; zenity-свидетель на ВМ недостоверен — фокус-гейты мызтера должны идти по daemon focus_in (waitFocusInFrom), не по AT-SPI-свидетелю (запрокомментировано ошибкой re-pin).
- Блокеров нет; TERM-02 остаётся открытой до 10-06 (shared-ID, шесть планов; ready-ids: 0/1); flagged assumption плана (D9) — за фазовой верификацией.

---
*Phase: 10-term-slovo*
*Completed: 2026-10-12*

## TDD Gate Compliance

- RED: `test(10-05)` commits 667b0a4 (TestRegisterEventInterest_PinsRegistryCall), 00e2799 (TestTerminalBranch_Level2PacesWire) — каждый с проверенным свидетельством RED_EVIDENCE_OK (gsd check tdd-red-evidence, TAP 1:1)
- GREEN: `feat(10-05)` commits 5f7837d, 710b6d4 — целевые тесты зелёные, пакеты зелёные под -race, mise run ci зелёный
- REFACTOR: отдельных refactor-коммитов нет (REFACTOR опционален; чистка уезжала в GREEN)
- e2e-кейсы (Task 1/3) — не behavior-задачи плана (tdd-атрибутов нет): их live-верктики — verify-свидетельство плана

## Self-Check: PASSED

SUMMARY exists; 11 коммитов от 1b7771b (gsd-plan-head-before-10-05) verified in git log; deletion-scan чист; untracked-скан: новых артефактов не оставлено (pty-артефакты tmp-only в ВМ, журнал — hex/счётчики); stub-скан новых файлов чист (TODO/FIXME/placeholder — ноль); mise run ci зелёный на HEAD; go.mod/go.sum нетронуты.
