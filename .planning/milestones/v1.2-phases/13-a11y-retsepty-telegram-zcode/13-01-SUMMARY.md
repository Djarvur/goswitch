---
phase: 13-a11y-retsepty-telegram-zcode
plan: 01
subsystem: a11y-probe
tags: [a11y, atspi, telegram, snap, flatpak, tarball, e2e-stand, probe-first, owner-verdict]

# Dependency graph
requires:
  - phase: 12-term-strova (запаркованный стенд goswitch-stand: юнит goswitch-rel, конфиг удалён, sources-takeover жив)
  - phase: 10-term-slovo (recipe подъёма 10-01/10-08: stand-env, rsync-зеркало, unit -debug, XAUTHORITY-находка; venue-протокол датированных рядов)
  - AppArmor-вердикт 2026-10-08 (snap 7.2.9 rev 7150) — предмет перепроверки D-01
provides:
  - scripts/a11y-selfcheck.py (D-06): внешний AT-SPI-зонд видимости текст-поля — роли по НОМЕРУ против канона internal/appid/appid.go:85-96 ({61,79,94} текст-класс, 40 = видимо-но-неисправимо), контракт 0/1/2, живость /proc без subprocess; инструмент пробы и фундамент self-check строки запуска (подключит 13-03)
  - Машинный фундамент A11Y-02: трёхслойная таблица трёх рунгов (snap 7.3.3 rev 7175 / flatpak 7.3.3 / официальный tarball) ×2 раунда каждая, ОДНОЙ Wayland-сессией — переключение ЗЕЛ везде, видимость/автокоррекция КРАСН везде
  - Вердикт владельца «Задокументировать» (развязка ветки (c)-2): A11Y-02 закрыт как недостижимый-by-evidence; D-04 (flatpak-кандидат) опровергнут; A5 (плашка диктора) подтверждена живьём — входы для рецептов 13-03
  - Стендовая механика: абсолютный канал клика QMP input-send-event (usb-tablet), относительность ydotool mousemove, GTK3-слепота событий сессии, сетевой стейджинг репо через хост
affects: [13-02 (стенд поднят и ждёт ZCode), 13-03 (рецепты пишут границу по рядам этого плана), phase-verification]

actuals:
  tokens: 9500         # chars/4 over the realized diff 8aa3882..3ec3e6b (270 insertions, 3 files)
  tasks: 3
  commits: 4           # MEASURED: git rev-list --count gsd-plan-head-before-13-01..HEAD (3 задачи + строка развязки; metadata-коммит далее)

tech-stack:
  added: []
  patterns:
    - "Трёхслойная проба сборки: key-trace флипа (журнал демона -debug) / номер роли сфокусированного поля от зонда (число, не имя) / дельты goswitchctl status на сентинел-корпусе — вердикт = пара согласных раундов"
    - "Роли по НОМЕРУ против канона appid.go — имена только для человекочитаемой печати (Pitfall 8)"
    - "Абсолютный драйв мыши ВМ: QMP input-send-event через usb-tablet (virsh qemu-monitor-command; press/release раздельными батчами) — ydotool 0.1.8 mousemove относительный"
    - "Host-стейджинг полезной нагрузки при ползущей сети ВМ: user-репо хоста → rsync зеркала → локальный file:// remote (refs→heads + build-update-repo)"

key-files:
  created:
    - "scripts/a11y-selfcheck.py (D-06, 224 строки: зонд 0/1/2, TEXT_ROLES {61,79,94}, walk() из focus_helper дословно)"
    - ".planning/phases/13-a11y-retsepty-telegram-zcode/13-01-SUMMARY.md"
  modified:
    - "docs/adr/adr-008-spike-log.md (+37 строк: секция «Фаза 13 A11Y-проба сборок (telegram snap/flatpak, zcode)» — 11 датированных рядов: подъём, калибровка, snap-механика, snap r1/r2, flatpak-механика, flatpak r1/r2, лестница tarball r1/r2, deb-рунг, итог, откаты, развязка)"
    - "test/e2e/spike_journal_test.go (константа secA11yProbe + запись в spikeSections() — пин заголовка секции в TestSpikeJournalSectionsExist)"

key-decisions:
  - "A11Y-02 закрыт вердиктом владельца «Задокументировать» (вариант 2, дословно): автокоррекция Telegram недостижима ни в одной форме — стаб идентичен даже unconfined-tarball (потолок Qt6-моста билда, не confinement); рецепты 13-03 пишут границу; граница фазы 2026-10-08 (демон не меняется) в силе"
  - "Слова владельца «переключенив телеграме работает» = слой 1 (IBus-хоткей), зелёный во всех рунгах ×2 — D-01/A6 решены машиной; AppArmor-вердикт 2026-10-08 по сути подтверждён на 7.3.3 rev 7175, диагноз расширен (стаб без confinement)"
  - "D-04 (flatpak как кандидат-рунг) опровергнут: a11y через xdg-dbus-прокси песочницы, дерево-стаб роль 75, ноль сигналов — прецедент A1 опровергнут на паре org.gnome.Platform 51 + Qt6-мост"
  - "Официального .deb Telegram не существует (HTML-страница/404) — deb-рунг лестницы закрыт за неимением легитимного источника (PPA вне Package Legitimacy Audit)"
  - "Стенд остаётся поднятым до конца фазы (нужен 13-02); откаты рунгов и стенда задокументированы дословно, host-стейджинг вычищен сразу"

patterns-established:
  - "Probe-first развилка с designed checkpoint: ВСЕ рунги красные → checkpoint:decision gate=blocking-human возвращён владельцу с машинной таблицей; вердикт владельца зафиксирован датированной строкой развязки"
  - "Позитив-контроль оракула в каждой пробе нового класса: zenity fired ×2 в той же сессии доказывает, что fail-closed молчание Telegram — сторона приложения, не сломанный оракул"

requirements-completed: [A11Y-02]  # владелец закрыл как недостижимый-by-evidence (вердикт 2026-11-03); рецепты-документация — 13-03/A11Y-03

coverage:
  - id: D1
    description: "scripts/a11y-selfcheck.py стоит с контрактом D-06: шебанг /usr/bin/python3, роли по НОМЕРУ {61,79,94} против appid.go:85-96, роль 40 отдельным вердиктом, walk() с поштучным try/except, живость /proc без subprocess (ASVS V5), коды 0/1/2; заголовок секции журнала запинен secA11yProbe"
    requirement: A11Y-02
    verification:
      - kind: unit
        ref: "verify задачи 1: test -x + шебанг + py_compile + greps TEXT_ROLES/Exit codes + git diff --exit-code -- internal/ cmd/ + mise run ci (build/vet/lint/test -race зелёные, TestSpikeJournalSectionsExist гоняет пин); негатив зонда на ВМ: nosuchapp → exit 2"
        status: pass
    human_judgment: false
  - id: D2
    description: "Стенд поднят (зеркало текущей ветки, unit -debug, autocorrect on, sources-takeover); калибровка ПЕРВЫМ рядом (gedit exit 0 роль 61; полный оракул-контроль на zenity: fired=+1); snap-рунг 7.3.3 rev 7175 измерен тремя слоями ×2"
    requirement: A11Y-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md: ряды «a11y-стенд подъём 15:27», «a11y-калибровка gtk 15:29-15:36», «a11y-проба telegram-snap r1 15:55 / r2 15:57-16:00»; verify задачи 2: systemctl is-active + selfcheck gedit exit 0 по ssh + grep рядов"
        status: pass
    human_judgment: false
  - id: D3
    description: "Flatpak-рунг ×2 + лестница D-03 (tarball ×2 unconfined, deb — нет легитимного источника) + сравнительная таблица + итог-ряд + откаты — всё датировано, ОДНОЙ сессией"
    requirement: A11Y-02
    verification:
      - kind: e2e
        ref: "docs/adr/adr-008-spike-log.md: ряды «a11y-механика flatpak install», «a11y-проба telegram-flatpak r1/r2», «a11y-проба лестница tarball r1/r2», «лестница deb», «a11y-проба итог», «Откаты»; verify задачи 3: flatpak list показывает telegram + grep рядов"
        status: pass
    human_judgment: false
  - id: D4
    description: "Вердикт-развилка A11Y-02 исполнена дословно: ВСЕ рунги красные по слоям 2-3 → checkpoint:decision gate=blocking-human возвращён (ветка правки демона среди вариантов отсутствует); владелец разрешил вариантом 2 «Задокументировать» — A11Y-02 закрыт как недостижимый-by-evidence"
    requirement: A11Y-02
    verification:
      - kind: other
        ref: "docs/adr/adr-008-spike-log.md: ряд «a11y-проба развязка 2026-11-03 (вердикт владельца)» — вердикт дословно, следствия по решению"
        status: pass
    human_judgment: true
    rationale: "Развилка A11Y-02 по требованию REQUIREMENTS/плана — решение владельца на машинных рядах; formальная классификация края (unclassified/unresolved SPEC без ## Edge Coverage) остаётся за фазовой верификацией"

duration: 138min
completed: 2026-11-03
status: complete
---

# Phase 13 Plan 01: Живая ВМ-проба Telegram (D-01/D-02/D-03) Summary

**Трёхслойная проба ТРЁХ сборок Telegram (snap 7.3.3 rev 7175 — побайтовая реплика хоста владельца, flatpak 7.3.3, официальный tarball unconfined) одной Wayland-сессией дала машинную истину: переключение раскладки (IBus-хоткей) работает во всех рунгах ×2, автокоррекция недостижима ни в одном ×2 — дерево-стаб (1 нода, роль 75) идентичен даже без confinement, потолок — экспозиция Qt6-моста самого билда; владелец вердиктом «Задокументировать» закрыл A11Y-02 как недостижимый-by-evidence, D-04 (flatpak-кандидат) опровергнут, плашка диктора подтверждена живьём — входы рецептов 13-03.**

## Performance

- **Duration:** 138 мин (15:18-17:36 UTC, включая ожидание вердикта владельца по чекпойнту)
- **Tasks:** 3/3 (задача 1 tracer — зонд+секция+пин; задача 2 — стенд+калибровка+snap ×2; задача 3 — flatpak ×2+лестница+итог+откаты+развязка)
- **Files:** 3 репозиторных (scripts/a11y-selfcheck.py новый, adr-008-spike-log.md +37 строк, spike_journal_test.go +9)

## Accomplishments

- **Задача 1 (tracer):** scripts/a11y-selfcheck.py — внешний AT-SPI-зонд D-06: роли ТОЛЬКО по номеру int(node.get_role()) против канона internal/appid/appid.go:85-96 (текст-класс {61,79,94}; 40 = «видимо, но неисправимо»; прочие без предположения), приложения()/walk() из focus_helper дословно (поштучный try/except), живость процесса чтением /proc (ASVS V5: имя приложения — данные, никогда команда), контракт 0/1/2 в докстринге, /usr/bin/python3; секция журнала Фазы 13 открыта, secA11yProbe запинен в spikeSections() — mise run ci зелёный; tracer-гейт: verify перезапущен end-to-end зелёным.
- **Задача 2:** стенд поднят (rsync-зеркало ветки 449057a, gap-сборка, unit dist+drop-in -debug вместо stale юнита фазы 12, конфиг autocorrect on, sources-takeover жив); калибровка ПЕРВЫМ рядом: gedit exit 0 роль 61 — но GTK3-слепота событий (0 из 386 сигналов) нашла резерв: zenity — ПОЛНАЯ цепочка оракула зелёная (прямое bridge-событие + GetRole + fired=+1); snap-рунг: L1 ЗЕЛ ×2, L2 КРАСН ×2 (стаб, регистрация :1.29 без контента, 0 сигналов), L3 КРАСН ×2 (abstain role-unknown, fail-closed).
- **Задача 3:** flatpak-рунг установлен host-стейджингом (сеть ВМ ползла — Pitfall 5 живьём): L1 ЗЕЛ ×2, L2 КРАСН ×2 (стаб через xdg-dbus-прокси, ПЛАШКА ДИКТОРА живьём), L3 КРАСН ×2; лестница D-03: tarball unconfined ×2 — идентичный стаб (ключевой факт: потолок не confinement), deb-рунг закрыт (официального .deb нет); итог-ряд с таблицей, откаты дословно, откат стенда отложен (нужен 13-02).
- **Развязка:** checkpoint:decision gate=blocking-human (ветка (c) — все рунги красные) возвращён владельцу; вердикт «Задокументировать» (вариант 2) зафиксирован датированной строкой развязки.

## Task Commits

1. **Задача 1 (tracer): зонд + секция + пин** — `449057a` (feat)
2. **Задача 2: стенд + калибровка + snap ×2** — `83bbe86` (docs)
3. **Задача 3: flatpak ×2 + лестница + итог + откаты** — `07e93ae` (docs)
4. **Развязка: вердикт владельца** — `3ec3e6b` (docs)

**Plan metadata:** финальный docs-коммит (SUMMARY+STATE+ROADMAP+REQUIREMENTS) — далее.

## Files Created/Modified

- `scripts/a11y-selfcheck.py` — новый (D-06): зонд видимости текст-поля, exit 0/1/2, роли-по-номерам
- `docs/adr/adr-008-spike-log.md` — секция «Фаза 13 A11Y-проба сборок (telegram snap/flatpak, zcode)»: 11 датированных рядов (подъём, калибровка, snap-механика, snap r1/r2, flatpak-механика, flatpak r1/r2, лестница tarball r1/r2, deb, итог, откаты, развязка)
- `test/e2e/spike_journal_test.go` — константа secA11yProbe + spikeSections()
- VM-local (не репо, с documented rollback): snap telegram-desktop 7.3.3, flatpak + org.telegram.desktop 7.3.3 + flathub/flathub-mirror remotes, ~/tg-tarball-unpack, зеркало /tmp/flathub-mirror, конфиг стенда, stand-env.sh

## Decisions Made

См. key-decisions. Ядро: A11Y-02 закрыт вердиктом владельца как недостижимый-by-evidence (демон не тронут — граница фазы в силе); D-04 опровергнут; рецепты 13-03 пишут границу; стенд ждёт 13-02.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] gedit отсутствовал на ВМ (плановое калибровочное поле)**
- **Found during:** Задача 2
- **Issue:** план предполагал gedit предустановленным; резерв zenity --entry оставался
- **Fix:** sudo apt-get install -y gedit 46.2 из архива Ubuntu (прецедент wl-clipboard 10-01); наблюдённая роль поля — 61 (план ожидал 94; номер решает проба — Pitfall 8)
- **Commit:** 83bbe86 (строка подъёма)

**2. [Rule 3 - Blocking] flatpak install с ВМ ползёт и умирает (Pitfall 5 живьём)**
- **Found during:** Задача 3
- **Issue:** dl.flathub.org доступен, но ~150 КБ/с + мёртвые CLOSE-WAIT соединения (88 МБ за ~12 мин, затем 0 КБ/с) — прямая установка не завершается
- **Fix:** host-стейджинг (прецедент D-03 tarball применён к flatpak-рунгу): хост качает 5.5 МБ/с → flatpak install --user (репо 1.9 ГБ) → rsync зеркала на ВМ → локальный file:// remote (refs/remotes→refs/heads + flatpak build-update-repo для summary) → полностью офлайн-установка; create-usb невозможен (нет collection-ID)
- **Commit:** 07e93ae (строка механики)

**3. [Rule 3 - Blocking] flatpak remote-add требует sudo**
- **Found during:** Задача 3
- **Issue:** плановая команда remote-add в user-контексте отвергнута («ConfigureRemote not allowed for user»)
- **Fix:** sudo для системных операций remote-add/install (соответствует sudo-форме apt install плана)
- **Commit:** 07e93ae

**4. [Rule 1 - Bug] клики ydotool не попадали в цель (дрейф указателя)**
- **Found during:** Задача 2 (snap-рунг)
- **Issue:** ydotool 0.1.8 mousemove — ОТНОСИТЕЛЬНЫЙ (флага -a в сборке нет) — накопленный дрейф отправлял клики мимо окон (открывали overview)
- **Fix:** открыт абсолютный канал: QMP input-send-event через usb-tablet ВМ (virsh qemu-monitor-command; координаты px*32767/1279 / py*32767/799; press и release РАЗДЕЛЬНЫМИ батчами — Qt игнорирует пару одним батчем); калиброван кликом по часам top-bar
- **Commit:** 83bbe86 (механика-ряд snap)

**5. [Rule 3 - Honest-redo] испорченные раунды перезапущены чисто (не ретрай-до-зелёного)**
- **Found during:** Задача 2
- **Issue:** (а) первый r2-заезд snap: zenity-диалог перекрыл окно — корпус ушёл в zenity (fired=+1 = лишний позитив оракула, НЕ вердикт telegram); (б) промежуточный корпус в ru-раскладке дал dict-cur-hit abstain
- **Fix:** оба заезда задокументированы честными строками и перезапущены: (а) zenity закрыт, поле телефона сфокусировано заново; (б) режим проверен статусом до корпуса, корпус в en
- **Commit:** 83bbe86 (ряды r2)

**Total deviations:** 5 автофиксов. **Impact:** ни один не ослабил контракт задач или честность вердиктов; находки 2/4 закреплены как стендовая механика в журнале.

## Issues Encountered

- GTK3-приложения сессии не эмитят StateChanged (gedit: дерево живо, события ноль) — калибровка честно перешла на zenity (GTK4 эмитит прямые bridge-события); на рунги не влияет (Qt/Electron — свой мост), но механика задокументирована.
- DC Telegram с ВМ недоступны (149.154.167.51:443/80 FAIL) — QR не рендерится; форма телефона рендерится офлайн; слой 3 измерен на поле авторизации (QR-честность — в итог-ряду и чекпойнте).
- Строковые артефакты конвейера (tail-маскировка exit-кода зонда в одном выводе) — истинный код перепроверен отдельным прогоном (exit 1), задокументирован в ряду.

## User Setup Required

None — вердикт владельца получен и зафиксирован; стенд поднят и ждёт 13-02 (ZCode), откаты — в конце фазы.

## Next Phase Readiness

- 13-02: стенд готов (демон active -debug на зеркале ветки, autocorrect on, каналы драйва закреплены — send-key + QMP-абсолют); ZCode прибудет rsync /opt/ZCode с хоста.
- 13-03: recipes пишут границу по рядам этого плана — «переключение работает / автокоррекция недостижима ни в одной форме», плашка диктора (A5) — ожидаемый побочный эффект, residual-risk PRIMARY — в README; scripts/a11y-selfcheck.py остаётся инструментом — 13-03 подключает строкой запуска.
- Для верификатора: формальная классификация края A11Y-02 (unclassified/unresolved — SPEC без ## Edge Coverage) остаётся за фазовой верификацией (D4 rationale); flagged assumptions A1 (опровергнут), A3 (использован экран авторизации), A5 (подтверждена flatpak/tarball, не snap), A6 (подтверждён) закрыты живыми рядами.

---
*Phase: 13-a11y-retsepty-telegram-zcode*
*Completed: 2026-11-03*

## Self-Check: PASSED

SUMMARY существует; все 4 коммита от базы плана 8aa3882 верифицированы в git log (449057a / 83bbe86 / 07e93ae / 3ec3e6b; ledger gsd-plan-head-before-13-01 на диске, rev-list count = 4); файлы созданы (a11y-selfcheck.py, SUMMARY) и модифицированы (adr-008-spike-log.md, spike_journal_test.go) — все FOUND; пин secA11yProbe в spikeSections() зелёный в mise run ci (задача 1, tracer-гейт перезапущен end-to-end); демон нетронут на каждом коммите (git diff --exit-code -- internal/ cmd/); счётчик рядов a11y-пробы в журнале = 8 (подъём/калибровка/механика ×2/рунги ×6/итог); stub-скан чист (TODO/FIXME/placeholder — ноль); host-стейджинг вычищен; стенд нетронут (ждёт 13-02).
