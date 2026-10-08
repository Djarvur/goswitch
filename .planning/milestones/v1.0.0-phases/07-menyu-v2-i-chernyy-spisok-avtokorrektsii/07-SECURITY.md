---
phase: "07"
slug: "menyu-v2-i-chernyy-spisok-avtokorrektsii"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-10-05"
---

# Phase 07 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.
> Built from the `<threat_model>` blocks of all eight Phase 7 plans (07-01..07-08);
> verified at L1 grep/symbol depth per ASVS L1 with `go build ./...` + the full
> `-race` suite green at HEAD 4d29dd4 (2026-10-05; the single red,
> TestRun_OnConnHookCalledOnce, is the pre-existing timing flake of
> deferred-items.md — passed on the documented package re-run) and the code-review
> verdicts of 07-REVIEW.md folded in: CR-01 (b) + WR-01/WR-02/WR-03 fixed RED→GREEN
> (b708020/713d25a→80c99f2/a23a47c, d10671a→2504bf6, c70d620→d727ef4), the six Info
> findings documented, none a threat-register row.
> Privacy D-20/D-21 re-verified at HEAD by reading every emission site of the new
> phase-7 paths: menu/writer/sound/toggle/actor refusals carry slugs, counts,
> roles, paths, binary names and transport errors only — never the typed word
> (see the audit trail).

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| YAML-документ пользователя → конфиг/писатель | недоверенный ввод: битые/злонамеренные regex, гигантские списки, вырожденные пороги, старые ключи; тумблер пишет обратно тем же документом | config values, blocklist patterns |
| a11y-идентичность приложения → blocklist-гейт | идентичность может отсутствовать/ошибаться; blocklist запрещает только увиденное | app identity, AT-SPI role enum |
| Ролевый запрос (confirm) → конвейер замены | решение «стрелять» по живой роли; инверсия unknown⇒pass не должна ослабить fail-closed сегменты | role enum, armed payload |
| Session bus (клики /Menu) → демон | любой процесс сессии может слать Event — same-user доверенный контур, диспетчер обязан быть отказоустойчивым | menu clicks, item IDs |
| Клик «Настройки…»/sound-флип → exec | запуск внешних процессов из демона: argv-форма и дескрипторы пинены | config path, event name, wav path |
| ItemsPropertiesUpdated → gnome-shell | сигналы читает чужой JS-рендерер; метки несут конфиг-строки | property deltas, labels |
| Файловая система ($HOME) → демон | конфиг/кэш пользователя: гонки записей, права, symlink-мишени | config.yaml, tone wav |
| Секция sound конфига → argv субпроцесса | имя события из пользовательского YAML попадает в аргументы запускаемого бинарника | event string |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-07-01-01 | Repudiation | утрата аудит-трейла §11 при правке | medium | mitigate | старый вердикт сохранён дословно (SPEC.md:196 «отвергнута владельцем осознанно»), ревизия добавлена рядом (:217-221); append-only форма соблюдена | closed |
| T-07-01-02 | Tampering | расхождение формулировок SPEC §11 ↔ ADR-007 ↔ 07-CONTEXT | medium | mitigate | ADR-007 Amendment 2026-10-04 (:193-246) с перекрёстными ссылками; греп-гейты токенов — RELEASE-READINESS чек 2 PASS (ключи `apps_blocklist`×1, «2026-10-04»×3, слаг `app-blocked`×1) | closed |
| T-07-01-SC | Tampering (supply chain) | новые зависимости планом | low | accept | план документальный — go.mod не менялся (3 прямых + 1 indirect, tidy-diff гейт) | closed |
| T-07-02-01 | DoS | катастрофический бэктрекинг regex на живом boundary | low | mitigate | stdlib RE2, линеен — компиляция один раз на Load (config.go:367) + кэш актора, перекомпиляция только при смене списка (actor.go:2185-2212); regexp2/backtracking-движков в модуле нет | closed |
| T-07-02-02 | DoS | гигантский список паттернов | medium | mitigate | maxAutocorrectBlocklist=64 безусловен — проверка ДО early-return disabled (config.go:28,357-362); пин 65-паттернов в обеих состояниях (config_test.go:599-619) | closed |
| T-07-02-03 | Tampering | пустой паттерн «молча блокируй везде» | high | mitigate | сентинел пустого/whitespace-паттерна с индексом до компиляции (config.go:363-366, errAutocorrectBlocklistEmpty); CONFIG.md:60 документирует | closed |
| T-07-02-04 | Tampering | старый ключ тихо игнорируется после миграции | high | mitigate | strict decode KnownFields(true) — неизвестный ключ инвалидирует весь файл (load.go:28); пин TestLoad_StrictDecodeRejectsTypo (load_test.go:236); миграционная заметка README:201-205 + CONFIG.md Migration | closed |
| T-07-02-05 | Elevation | enabled:true с нулевыми порогами | medium | mitigate | ACTIVE-условие: пороги обязательны при enabled, отказ с именем поля (config.go:371-382); disabled-секция хранит zero-валидность | closed |
| T-07-02-06 | Information disclosure | док-примеры/логи валидации несут содержимое поля | low | accept | валидация оперирует паттернами (имена приложений, config.go:365), не словом; доки несут ключи/дефолты — AR-07-3 | closed |
| T-07-02-SC | Tampering (supply chain) | regexp — stdlib, новых пакетов нет | low | accept | go.mod без изменений (3 прямых deps) | closed |
| T-07-03-01 | Tampering | разорванная запись при падении mid-write | low | mitigate | temp в том же каталоге (dot-префикс) + os.Rename — POSIX-атомарность (writer.go:326-350); пин TestWriter_AtomicWrite — ни одного .tmp-артефакта (writer_test.go:181-204); rename не пишет сквозь symlink-мишень, temp — O_EXCL | closed |
| T-07-03-02 | Information disclosure | права на конфиг шире необходимых | medium | mitigate | 0600: CreateTemp + явный os.Chmod(tmp, 0o600) до rename (writer.go:44,342) — финальный файл всегда 0600, даже при замене более широкого существующего; пин mode==0600 (writer_test.go:203-204); каталог MkdirAll 0700 (main.go:421,438) | closed |
| T-07-03-03 | Elevation | битый дефолтный файл тихо заменяется дефолтами | high | mitigate | громкий отказ старта: только os.ErrNotExist обслуживает defaults (main.go:470-472), существующий-но-битый → load error → start refusal (:474-477); adoptLoad того же контракта для watcher (:405-417) | closed |
| T-07-03-04 | Tampering | гонка двух писателей (тумблер vs правка пользователя) | low | accept | last-writer-wins; rename не даёт битых файлов; окна миллисекундные — AR-07-4 | closed |
| T-07-03-05 | Denial of service | эхо-релоад зацикливает watcher | low | mitigate | запись только по клику; debounce 200 мс коалесцирует burst rename-эха, только события файла (watch.go:18,209-215); эхо — applySnapshot тех же значений (идемпотентный fold) | closed |
| T-07-03-06 | Information disclosure | журнальные записи писателя/adopt несут содержимое конфига | low | mitigate | writer.go — 0 вхождений slog; INFO несёт путь/исход (main.go:72-74); секционная семантика — только в сообщении применённого reload | closed |
| T-07-04-01 | Elevation (deception) | half-переворот: unknown тихо вернулся в запрет | high | mitigate | инверсия ровно одного сегмента: unknown identity PASSES оба гейта (actor.go:2133-2149 arm; :2412-2419 confirm — цикл только при known); пин TestAutoCorrect_IdentityUnknownNotProhibition (actor_test.go:5522); слаги/ветки равенства удалены — негативный греп no-apps/app-not-listed/app-unknown/app-changed = 0 по не-тестовому коду | closed |
| T-07-04-02 | Elevation | инверсия расползлась: роль/капсы/детектор стали fail-open | high | mitigate | роль-ошибка/таймаут → role-unknown/role-timeout отказ (actor.go:2326-2335); не-текстовая роль → role-forbidden (:2360-2361); no-caps → отказ (:2126-2131); unsure/короткий → отказ (:2159-2165); пин RoleDeadlineBounded (actor_test.go:5615) | closed |
| T-07-04-03 | Tampering | тихий остаток равенства (константа/ветка/тест/коммент) | high | mitigate | негативный греп по коду/константам/слагам = 0; TestAutoCorrect_ConfirmBlocklistRefuses/ConfirmUnknownPasses заменили старый WR-01 тест (actor_test.go:5795,5828); 06-регрессы CR-01/CR-02 выжили (review: «checked dimensions» п.1 SOUND) | closed |
| T-07-04-04 | Spoofing | fired «в чужое поле» при переносе фокуса без FocusOut | medium | accept | слоистая защита в коде: fired-порог переверифицирует буфер (payload-stale, actor.go:2393-2398; FocusOut hard-reset :743) + конвейер замены повторно проверяет поле перед Delete; риск мониторится матрицей 07-06 — AR-07-5 | closed |
| T-07-04-05 | Information disclosure | новые отказы/WARN несут набираемое слово | medium | mitigate | закрытый словарь слагов (actor.go:94-104); WARN — транспортный класс (warnACRole :2429-2439); ctlsvc рендерит ac_skip_<slug>=N (:167-169); non-leak пин в корпусе (actor_test.go:5889-5894) | closed |
| T-07-04-06 | Denial of service | blocklist-матчинг на каждом boundary | low | mitigate | компиляция один раз (кэш refreshACBlocklist), RE2 линеен, потолок 64 на Load — см. T-07-02-01/02 | closed |
| T-07-05-01 | Tampering | «Настройки…» как примитив исполнения команды | medium | mitigate | только xdg-open с ОДНИМ аргументом — закреплённый путь конфига, exec.Command без shell (main.go:289-291,315); $EDITOR не консультируется; fake-runner пинит argv | closed |
| T-07-05-02 | Denial of service | piped stdout/stderr лаунчера — дедлок демона | medium | mitigate | no-pipes форма: nil Stdout/Stderr, Start + go Wait (main.go:290,315-327); пин nil-pipes в корпусе | closed |
| T-07-05-03 | DoS (удалённый рендерер) | ItemsPropertiesUpdated с nil-removed паникует на живой шине | medium | mitigate | пустой срез-литерал []removedProps{} на КАЖДОМ эмите (menu.go:217,233,250,288,303; тип :95-103); recoverMenuCall сдерживает диспетчерскую сторону | closed |
| T-07-05-04 | Tampering | паника колбэка клика валит обработку меню | low | mitigate | recoverMenuCall оборачивает Event и Activate (menu.go:367; indicator.go:380,458-462) — распространяется на все новые пункты; nil-callbacks — silent no-op | closed |
| T-07-05-05 | Spoofing | рассинхрон снимка меню с фактическим состоянием | low | mitigate | пуш-точки: pushMenuSync на applySnapshot (actor.go:1271,1282) + SetMenuSync install-push (main.go:189); клик тумблера применяет синхронно (toggleReload+fold, main.go:160-166); e2e меню сверяет файл+статус+рестарт | closed |
| T-07-05-06 | Information disclosure | нотификация «О программе»/метки несут чувствительное | low | accept | notifyStatus — mode+version только (main.go:262-272); метки — имена клавиш конфига (не секреты) — AR-07-3 | closed |
| T-07-05-SC | Tampering (supply chain) | ноль новых зависимостей | low | accept | godbus/yaml/fsnotify в go.mod; xdg-open — системный | closed |
| T-07-06-01 | Information disclosure | текст слова в логах стенда/отчётах | medium | mitigate | witness-формы (длина/факт-совпадение: «witness TEXT:chars=0→7», RELEASE-READINESS:38) и счётчики-оракулы; дисциплина фазы 6 продолжена | closed |
| T-07-06-02 | Tampering | кейсы трогают реальный конфиг владельца | low | mitigate | все документы в s.tmpDir (case_menu.go:183), -config всегда явный (:198) | closed |
| T-07-06-03 | Denial of service | Event-флуд по /Menu вешает диспетчер | low | mitigate | единичные клики с ожиданием маркеров; recoverMenuCall на демоне; watchdog стенда | closed |
| T-07-06-04 | Repudiation | зелёный прогон не доказывает заявленное | medium | mitigate | fails_when называет наблюдаемый сигнал: fired=0, ровно одна ac_skip_app_blocked, witness до=после (RELEASE-READINESS:38 PASS 2026-10-05); кейс меню — файл+статус+рестарт | closed |
| T-07-06-SC | Tampering (supply chain) | стенд-зависимости | low | accept | новых пакетов нет; gdbus/mise — часть ОС/стенда | closed |
| T-07-07-01 | Repudiation | доки описывают семантику, которой нет в коде | high | mitigate | греп-гейты KEYS-MATCH/SOUND-DOC-KEYS/README-OK — RELEASE-READINESS чек 1 PASS (:105); CONFIG.md:60,65 зеркалят схему (потолок 64, дефолты) | closed |
| T-07-07-02 | Elevation (ложная готовность) | RELEASE-READY без реальных зелёных прогонов | high | mitigate | доказательства формой команда→дата→exit→HEAD→артефакт (RELEASE-READINESS:15-38; mise run ci exit 0 @a264a54; e2e-menu-v2 + blocklist-silent PASS); юнит-корпус переподтверждён -race на HEAD 4d29dd4 | closed |
| T-07-07-03 | Tampering | преждевременный тег/релиз | medium | mitigate | git tag --list = v1.0.0/v1.0.1 only (дофазовые); запрет в prohibits; пост-merge шаги описаны, не исполнены | closed |
| T-07-07-04 | Repudiation | тихий пропуск probe-строк в релизе | medium | mitigate | сверка 17 строк — RELEASE-READINESS чеки 3-5 PASS (:107-109: 15 пунктов меню, миграционная заметка, дефолты==validate()) | closed |
| T-07-07-SC | Tampering (supply chain) | релизный поезд не меняется | low | accept | goreleaser/mise пины дофазовые | closed |
| T-07-08-01 | Tampering | инъекция через имя события в argv | medium | mitigate | exec БЕЗ shell: событие — ОДИН argv-элемент [canberra-gtk-play -i <event>] (sound.go:172-185); paplay — только собственный кэш-путь демона; корпус пинит argv (TestPlayer_SetAutocorrectEvent, sound_test.go:139) | closed |
| T-07-08-02 | Denial of service | прокси-спавн на каждый флип | low | accept | флипы user-paced; потомки короткоживущие, reap go Wait; backend решается один раз и кэшируется — LookPath не на каждый вызов (sound.go:97-98,199-208) — AR-07-6 | closed |
| T-07-08-03 | Denial of service | дедлок на piped-дескрипторах потомка | medium | mitigate | no-pipes: Stdin/Stdout/Stderr nil (sound.go:92), Start + go Wait, никаких Run (:186-192) | closed |
| T-07-08-04 | Information disclosure | WARN/логи звука несут чувствительное | low | accept | WARN — reason (закрытые константы) + transport err (sound.go:340,344); слово в пакете отсутствует по построению — AR-07-3 | closed |
| T-07-08-05 | Tampering | подмена wav в кэше пользователя | low | accept | файл 0600 в каталоге 0700 того же пользователя (sound.go:44-45,246,251); содержимое — детерминированный синус; повреждённый файл не валиден paplay-критично — AR-07-6 | closed |
| T-07-08-SC | Tampering (supply chain) | ноль новых Go-зависимостей | low | accept | stdlib os/exec, encoding/binary, math; canberra/paplay — системные; go.mod без изменений | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on (high) count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-07-1 | CR-01 doc leg / role-61 residue (не строка реестра — ревью-находка) | KNOWN, non-blocklisted идентичность + роль 61 (GTK3-пароль в a11y-наблюдаемом приложении) по-прежнему стреляет — узкое владельцем санкционированное исключение. Код-митигация (b) закрыла UNKNOWN-ячейку fail-closed (role-ambiguous, actor.go:2405-2409, пин TestAutoConfirm_RoleTextAmbiguity); README:241-250 честно документирует ограничение и предписывает blocklist-паттерн. Датированная оговорка в SPEC §11 (:219) и ADR-007 (:224) — ОТКРЫТОЕ РЕШЕНИЕ ВЛАДЕЛЬЦА (UAT-07 тест 5, pending) | Owner (UAT-07 test 5 — pending verdict) | 2026-10-05 |
| AR-07-2 | white-list UX removal (trade-off ревизии D-53, 07-CONTEXT locked) | Удаление white-list сняло ногу безопасности «default пуст = выключено везде» (06 AR-3 компенсация) и сделало слой достижимым одним кликом при пустом blocklist. Взамен: opt-out не требует знания bridge-идентичности (закрыт UX-гэп 06 AR-4), роль+капсы+детектор остаются fail-closed, UNKNOWN+роль-61 отказывает (AR-07-1), README/CONFIG предписывают blocklist-паттерны для известных пароле-носителей. Осознанный владельцем вариант 1 | Owner (07-CONTEXT locked, ADR-007 Amendment Accepted) | 2026-10-04 |
| AR-07-3 | T-07-02-06, T-07-05-06, T-07-08-04 | Поверхности, не соприкасающиеся с текстом пользователя: валидация оперирует паттернами (имена приложений), нотификация — mode+version, звуковые WARN — имена бинаров/событий и transport-ошибки. D-20/D-21 тривиальны по построению пакета | Owner (plan dispositions) | 2026-10-05 |
| AR-07-4 | T-07-03-04 | Гонка двух писателей конфига: last-writer-wins, rename не даёт битых файлов, окна миллисекундные; merge-логика не строится | Owner (plan 07-03 disposition) | 2026-10-05 |
| AR-07-5 | T-07-04-04 | Перенос фокуса без FocusOut теоретически ведёт fired в чужое поле; слоистая защита в коде (payload-stale перевирификация буфера + повторная проверка поля конвейером замены перед Delete), риск мониторится матрицей 07-06 | Owner (plan 07-04 disposition) | 2026-10-05 |
| AR-07-6 | T-07-08-02, T-07-08-05 | Звуковой контур: спавн короткоживущего потомка на каждый флип (user-paced, reap go Wait, backend кэширован) и подменяемый пользователем же wav его собственного кэша (0600/0700, детерминированный синус без семантики) | Owner (plan 07-08 dispositions) | 2026-10-05 |
| AR-07-SC | T-07-01-SC, T-07-02-SC, T-07-05-SC, T-07-06-SC, T-07-07-SC, T-07-08-SC | Supply-chain: фаза не добавила ни одного Go-пакета — go.mod неизменен (godbus v5.2.2, yaml.v3 v3.0.1, fsnotify v1.10.1 + indirect x/sys); новые поверхности — stdlib os/exec/regexp и системные бинары (xdg-open, canberra-gtk-play, paplay) | Owner (plan dispositions; Package Legitimacy Audit 07-RESEARCH) | 2026-10-05 |

*Accepted risks do not resurface in future audit runs.*

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-05 | 45 | 45 (32 mitigated + 13 accepted) | 0 | gsd-security-auditor (L1 grep/symbol depth + build/-race at HEAD 4d29dd4 + 07-REVIEW verdicts folded in: CR-01(b)/WR-01/WR-02/WR-03 fixed RED→GREEN, 6 Info documented) |

L1 evidence spot-checks (grep/символьный уровень, все на HEAD):
- **Privacy D-20/D-21 (ядро фазы):** перечитан КАЖДЫЙ emission-сайт новых phase-7 путей — writer.go: 0 вхождений slog; sound.go: 2 WARN (закрытые reason-константы + transport err, sound.go:340,344); menu.go/indicator.go: emit/attach/panic-WARN — component+error только (menu.go:569; indicator.go:248,329,392,410,451,460); main.go новые записи — пути/причины/ошибки (:311,322,357,363,369,383,387); actor.go отказы — только слаги закрытого словаря (:94-104) через recordACAbstain, role-WARN — transport (:2429-2439); ctlsvc рендерит ac_skip_<slug>=N (:167-169); non-leak пин в живом корпусе (actor_test.go:5889-5894).
- **Критический контур CR-01 (b):** role-ambiguous fail-closed — actor.go:2405-2409 (роль 61 + UNKNOWN identity → отказ); KNOWN + 61 сохраняет blocklist-вердикт; 79/94 + unknown продолжают стрелять; пин-пара TestAutoConfirm_RoleTextAmbiguity (actor_test.go:5870-5915, включает non-leak assert); роль-гейт байт-не тронут (61/79/94 vs мир, actor.go:2337-2362).
- **Blocklist-гейт (07-04 инверсия):** unknown PASSES (actor.go:2133-2149 arm; :2412-2419 confirm — known-only цикл); compiled-cache только (refreshACBlocklist :2185-2212), RE2-подстрока; fail-closed сегменты нетронуты (роль-ошибка/таймаут :2326-2335, role-forbidden :2360-2361, no-caps :2126-2131, unsure :2159-2165); негативный греп остатков равенства = 0.
- **Конфиг (07-02):** потолок 64 безусловный ДО early-return (config.go:357-362, пин :599-619); пустой/whitespace-паттерн отвергается с индексом (:363-366); компиляция regex на Load (:367); strict decode (load.go:28); ACTIVE-пороги (:371-382).
- **Писатель (07-03):** same-dir dot-temp + os.Rename + 0600 (writer.go:326-350, пин TestWriter_AtomicWrite); writer молчит; громкий отказ битого adopt-файла (main.go:470-477); MkdirAll 0700 (:421,438); debounce-эхо идемпотентно (watch.go:18,209-215).
- **Меню/exec (07-05):** WR-01 atomic slot — Store после Attach, Load-guard в пушах (main.go:148,181,236,247; пин TestMenuTogglePushRacesAttachStore main_test.go:573); xdg-open — один аргумент, nil-pipes, Start+go Wait (main.go:289-291,315-327); []removedProps{} литерал на всех пяти эмитах (menu.go:217,233,250,288,303); recoverMenuCall на Event/Activate (menu.go:367; indicator.go:380); doubleUnderscores на всех конфиг-метках (menu.go:502,527,636-637); WR-03 empty-path guard (main.go:356-360, пин TestToggleEmptyPathRefused main_test.go:638).
- **Звук (07-08):** argv-пин [bin -i event] / [bin wav] (sound.go:172-185); no-pipes Start+reaped Wait (:92,186-192); backend кэширован (:199-208); кэш 0700/0600 (:44-45,246,251); WR-02 hot-событие — SetAutocorrectEvent под мьютексом (:152-157) + fold-пуш (TestNewActor_SoundEventHotReload main_test.go:881).
- **e2e/доки (07-06/07-07):** tmpDir+явный -config (case_menu.go:183,198); witness-формы и счётчики-оракулы (RELEASE-READINESS:38); KEYS-MATCH/SOUND-DOC-KEYS/README-OK PASS (:105); git tag --list — только дофазовые v1.0.x.
- **Гейты фазы:** `go build ./...` + `go test -race -count=1 ./...` на HEAD 4d29dd4 — все пакеты зелёные, кроме задокументированного дофазового флака TestRun_OnConnHookCalledOnce (deferred-items.md), прошедшего документированным package-рераном; go.mod — 3 прямых зависимости (tidy-diff).

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log (включая CR-01 role-61 residue — решение владельца открыто, UAT-07 тест 5)
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter
- [x] 07-REVIEW verdicts folded in (1C+3W fixed RED→GREEN: CR-01 role-ambiguous 80c99f2, WR-01 atomic slot a23a47c, WR-02 hot event 2504bf6, WR-03 empty-path guard d727ef4; 6 Info documented — IN-01 dead field, IN-02 early tone, IN-03 franken-node, IN-04 no-watcher status, IN-05 FocusOut pin, IN-06 fixture Wait; ни один не является строкой threat-реестра)

**Approval:** verified 2026-10-05
