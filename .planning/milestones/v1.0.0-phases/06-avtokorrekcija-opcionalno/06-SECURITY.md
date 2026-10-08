---
phase: "06"
slug: "avtokorrekcija-opcionalno"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-10-04"
---

# Phase 06 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.
> Built from the `<threat_model>` blocks of all ten Phase 6 plans (06-01..06-10);
> verified at L1 grep depth per ASVS L1 with the full `-race` suite green at HEAD
> (go build + go test -race -count=1 ./... — 15/15 packages ok, 2026-10-04) and the
> code-review verdicts of 06-REVIEW.md / 06-REVIEW-FIX.md folded in (2 Critical +
> 1 Warning fixed with RED→GREEN evidence; 4 Info documented, none a threat-model row).
> Privacy D-20/D-21 re-verified at HEAD by reading every emission site of the new
> phase-6 paths (see the audit trail).

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| a11y-шина → Observer/Role (appid) | чужие приложения отвечают на GetRole; ответ — недоверенный uint32 enum, гейтится политикой потребителя | role enum, focus (sender, path) |
| буфер пользователя → detect.Check | токен — недоверенный ввод любой природы (пароль/код/имя); ошибка детектора = порча текста | typed word (никогда не покидает процесс) |
| YAML-документ → конфиг (autocorrect) | недоверенный ввод: гигантские списки, вырожденные пороги, опечатки | config values |
| клавиатура → feedKey | горячий путь ввода всего стола; async confirm добавляет in-flight окно | keystrokes, armed payloads |
| системные hunspell-файлы → dictgen (dev-side) | данные стороннего происхождения запекаются в поставку — supply chain словарей | dictionary data |
| armed payload → confirm-горутина | решение уходит из-под мьютекса и возвращается ≤25 мс спустя — стейл-стейт класс | correction range, app identity |
| фикстура пароля → stdout-оракул (e2e) | FINAL-печать — канал наблюдения, не утечки | fixture field content |
| фазовые доказательства → Accepted-вердикт (ADR-007) | слабые/подменённые доказательства легализуют небезопасную фичу | matrix runs, perf records |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-06-01-01 | Tampering | docs/SPEC.md §11 — потеря старого вердикта при дельте | medium | mitigate | старый вердикт сохранён дословно («отвергнута владельцем осознанно», SPEC.md:196), решение добавлено рядом (SPEC.md:201-205); spec-delta 0d3ed3d | closed |
| T-06-01-02 | Tampering (legal) | GPL-код из xneur/Easy Switcher в MIT-репо | high | mitigate | ADR-007:48-49,112 — «НИ СТРОКИ кода, только чтение алгоритма» (goibus-прецедент); кодовые планы не имеют GPL-источников в read_first; dictgen без донорских ссылок (grep=0) | closed |
| T-06-01-03 | Repudiation | коллизия ADR-нумерации (006 занят) | low | mitigate | ADR-007:27-28 документирует ренумерацию 006→007; ADR-006 остаётся файлом Фазы 5 | closed |
| T-06-02-01 | Tampering (supply chain) | испорченный/подменённый golden-словарь | high | mitigate | данные только через dictgen (hard errors errUnsortedWords/errUnknownName «has no license notice», dictgen/main.go:53,278,349); golden-инварианты layouts/dict_test.go (sorted/disjoint/no-yo/envelope/non-positive); нотисы в docs/LICENSE-data.md и заголовках dict_ru.go | closed |
| T-06-02-02 | DoS | +7.9 МБ бинарника / рантайм-map (бюджет <50 МБ) | medium | mitigate | отсортированные []string, lookup только sort.SearchStrings (detect.go:256-260, «zero heap» dict_ru.go); потолок-тест maxTrigramsPerLang=4096 (dict_test.go:15,86-90); RSS в perf-таблице ADR-007 (VmHWM) | closed |
| T-06-02-03 | Tampering (legal) | GPL-код доноров в генераторе | high | mitigate | генератор с нуля по пересказу 06-RESEARCH; ADR-007 позиция; read_first планов без GPL-исходников | closed |
| T-06-02-04 | DoS | тихий пропуск аномальных строк .dic → дырявый словарь | medium | mitigate | hard errors: errBadCountLine/errCountMismatch/errEmptyDict (dictgen/main.go:50-52,119,145); пин TestParseDIC (dictgen/main_test.go:49) | closed |
| T-06-03-01 | Spoofing | стейл-пара (объект умер/путь переиспользован) → Role не за то поле | medium | mitigate | Role — живой GetRole В МОМЕНТ решения по свежему gain (appid.go:249-259; actor autoConfirm:2032); ошибка/несоответствие → молчание (fail-closed); живой e2e-пин witness-гейта (case_autocorrect.go:226) | closed |
| T-06-03-02 | DoS | GetRole-затык в горячем пути | medium | mitigate | Role не держит блокировок во время вызова — снапшот пары, затем call (appid.go:250-252); жёсткий дедлайн задаёт потребитель autoRoleTimeout=25ms (actor.go:74,2030) | closed |
| T-06-03-03 | Information Disclosure | логирование сендера/пути/роли с контентом | low | mitigate | Role never logs (в appid.go 0 вхождений slog); потребитель — warn-once без контента (actor.go:2109-2123) | closed |
| T-06-03-04 | Tampering | строковое сравнение ролей (локаль) обходит гейт пароля | high | mitigate | ТОЛЬКО enum uint32: getRoleMethod — полный wire literal (appid.go:40), uint32 Store (appid.go:220-222); греп-гейт GetRoleName == 0 по всему репо; константы-пин в тестах (actor_test.go:4937-4938) | closed |
| T-06-04-01 | DoS | гигантский apps-список / вырожденные пороги | medium | mitigate | maxAutocorrectApps=64 безусловно (config.go:27,292-294); диапазоны [2,16], floor>0, margin≥floor (config.go:298-306); каждая ошибка именует поле (D-33) | closed |
| T-06-04-02 | Tampering | опечатка тихо меняет семантику | high | mitigate | strict decode KnownFields(true) — неизвестный ключ инвалидирует весь файл (load.go:28); TestLoad_StrictDecodeRejectsTypo (load_test.go:234), TestLoad_UnknownKeyRejected | closed |
| T-06-04-03 | Elevation | enabled:true без apps «включает везде» | high | mitigate | конъюнкция: пустой white-list = нигде — потребитель отказывает no-apps ДО детектора (actor.go:1887-1891); пин валидности TestValidate_AutocorrectDormantShapesValid (config_test.go:495) + default off TestDefaults_AutocorrectOff | closed |
| T-06-04-04 | Repudiation | расхождение docs/CONFIG.md и схемы | medium | mitigate | CONFIG.md:42-46 зеркалит имена/дефолты/диапазоны схемы (min_word_len 4/[2,16], margin 2.0, floor 1.0); таблица полноты секции (CONFIG.md:22) | closed |
| T-06-05-01 | Tampering | порча легитимного слова ложным wrong-layout | high | mitigate | вето dict-cur-hit и both-hit ДО вердикта (detect.go:150-153); mixed — abstain (161-162); минималка 4; бюджет ложных = 0 — TestCorpus_LegitWordsZeroFalsePositives (detect_test.go:653) | closed |
| T-06-05-02 | Information Disclosure | утекание набранного слова через Verdict/логи | high | mitigate | Verdict структурно без токен-полей (detect.go:66-71); slog в internal/detect = 0; замкнутый набор причин (detect.go:33-58, TestCorpus_ReasonVocabularyClosed); журнальные пины потребителя (actor_test.go:5067-5068,5327-5328) | closed |
| T-06-05-03 | DoS | вырожденный вход (мегабайтный «токен») в скоринге | low | accept | длина ограничена буфером исправления (CapSurroundingText/фразовые рамки); скоринг O(длина) — см. Accepted Risks AR-06-1 | closed |
| T-06-05-04 | Tampering (data) | дырявый/подменённый словарь → систематические ложные срабатывания | medium | mitigate | данные пинены golden-инвариантами 06-02; корпус на реальных данных (TestCorpus_*) ловит деградацию | closed |
| T-06-06-01 | Tampering | автокоррекция «исправляет» пароль (role 40) и коммитит | critical | mitigate | role 40 запрещён: разрешены ТОЛЬКО 61/79/94 (actor.go:81-85, switch default → role-forbidden :2067-2068); white-list default пуст (config Defaults); fail-closed на любой ошибке; silence-matrix ячейки role 40/60 (actor_test.go:5096,5100); живой e2e password-silent с тремя оракулами (case_autocorrect.go) — UAT-4 зелёный 2026-10-04 | closed |
| T-06-06-02 | Tampering | GTK3-пароль (роль 61) проходит ролевой гейт | high | mitigate/accept | вторая ступень: узкий white-list точного равенства, без префиксов (actor.go:1905-1911), default пуст; ограничение задокументировано (README.md:190-196, CONFIG.md, ADR-007); остаточный риск принят владельцем (AR-06-3, ADR-007 Accepted 2026-10-04) | closed |
| T-06-06-03 | Information Disclosure | слово пользователя в логах/status | high | mitigate | все новые emission-сайты несут slug/count/error без контента (actor.go:1973-1977,1990,2065,2119-2123; ctlsvc.go:160-167 ac-токены — счётчики); Reasons — замкнутые слаги; журнальный корпус-пин (TestAutoCorrect_CountersAndReasons + non-leak pins) | closed |
| T-06-06-04 | DoS | GetRole-затык замораживает ProcessKeyEvent | high | mitigate | решение вне a.mu (armed-payload: armAutoCorrect уходит горутиной :1869), дедлайн autoRoleTimeout=25ms (actor.go:74,2030-2032), вызов только на границах; пин TestAutoCorrect_RoleDeadlineBounded (actor_test.go:5220); perf-гейт p95 дельта −3.3 мс ≤ +5 мс (ADR-007:174) | closed |
| T-06-06-05 | DoS | шторм границ плодит горутины confirm | medium | mitigate | генерационный счётчик acGeneration (actor.go:1867-1868): новая граница отменяет устаревший confirm (:2036-2037); роль — одна за раз | closed |
| T-06-06-06 | Tampering | стейл-payload корректирует чужое поле | medium | mitigate | re-check enabled под мьютексом на re-entry (actor.go:2039-2043, TestAutoCorrect_ReEntryRechecks:5253); диапазон снимается в момент границы; CR-01-расширение: content-revalidation expectToken/expectTail (acConfirmRefusals actor.go:2087-2093, TestAutoConfirm_RevalidatesArmedPayload:5293) — живая UAT-3 проба 18/18 чисто | closed |
| T-06-07-01 | Tampering | password-silent зелёный без witness-подтверждения | high | mitigate | три независимых оракула (witness role+chars, FINAL-stdout, fired=0); отсутствие фокуса = FAIL, не skip (case_autocorrect.go:218-226 waitFixtureWitness) | closed |
| T-06-07-02 | Repudiation | v3-строки «перенесены» с изменениями | medium | mitigate | v3 встроена байт-дословно (программная сверка: 31/32 имён в v4; CASES-FROZEN git diff-гейты 06-08-SUMMARY:112,148); единственное имя вне v4 — super-space-alive, САНКЦИОНИРОВАННЫЙ вывод владельцем 2026-10-04 (UAT-1, заголовок v4, commit 1eda53d) — поствердиктная правка владельца, не исполнительская | closed |
| T-06-07-03 | Information Disclosure | фикстура печатает введённый пароль | low | mitigate | фикстура никогда не вводит настоящий секрет (password_entry.py:28); FINAL печатает содержимое только для проверки неизменности (ожидается неизменённое) | closed |
| T-06-07-04 | DoS | ночной dispatch дважды красный из-за нового default v4 | medium | mitigate | шум материализовался и был честно триангулирован: drift воспроизведён БАЙТ-ИДЕНТИЧНО на до-фазовом дереве d96dcb2 (06-07-SUMMARY:77), корень найден (G-6-1 самоблокировка фабрики) и ИСПРАВЛЕН gap-планами 06-09/06-10; после фикса 32/35×2 идентично, все 9 drift-строк зелёные; WINDOWS #13 закрыт вердиктом владельца (UAT-1 pass); v1/v2/v3 регресс-пути сохранены (e2e-matrix.yml:45-46) | closed |
| T-06-08-01 | Repudiation | Accepted без двойного регресса | high | mitigate | гейт удержан: ADR-007 честно ОСТАВАЛСЯ Proposed при красном двойном прогоне (06-08-SUMMARY:52 — «исполнитель НЕ решает вопрос гейта владельца»); перевод в Accepted — только вердиктом владельца 2026-10-04 на verify-work после закрытия строки 5 фиксом G-6-1 (32/35×2; ADR-007:5, UAT-2 pass) | closed |
| T-06-08-02 | Spoofing | README умалчивает GTK3-ограничение | high | mitigate | честная оговорка в README.md:190-196 («Known limitation: GTK3 password fields» + «never add password prompts or terminals»); геп-гейт FINALIZE-GREPS-OK (06-08-SUMMARY:87) | closed |
| T-06-08-03 | DoS | латентность GetRole смещает p95 за бюджет | medium | mitigate | perf-гейт записан в ADR-007:146-157,174: baseline p95 180.2 мс → ON p95 176.9 мс, дельта −3.3 мс ≤ +5 мс — ПРОЙДЕН; абсолютный бюджет-фейл окна честно зафиксирован (санкционированный исход 04-07) | closed |
| T-06-08-04 | Tampering | преждевременный милстоун-свитч | medium | mitigate | MILESTONE-FROZEN: git diff пуст на PROJECT/STATE/REQUIREMENTS/release.yml/.goreleaser.yaml (06-08-SUMMARY:109-112,148); REQUIREMENTS.md не тронут ни одним коммитом фазы (git log e13af69..HEAD — только SPEC.md дельта 0d3ed3d, санкционированная T-06-01-01); REQ-ID фиксация отложена на new-milestone v1.1.0 (D-51) | closed |
| T-06-09-01 | Tampering | гонка параллельного AttachEngine/чтения эмиттера | high | mitigate | atomic.Pointer[emitterSlot] (actor.go:161) — атомарный store/load, без мьютекса актора; корпус под -race зелёный (в т.ч. TestActor_AttachEngineWhileFlipInFlight); полный -race прогон 15/15 пакетов на HEAD | closed |
| T-06-09-02 | Denial of Service | шторм минта движков без сериализации | low | accept | AttachEngine — O(1) store указателя; профиль памяти не меняется, churn существовал до фикса — см. Accepted Risks AR-06-2 | closed |
| T-06-09-03 | Repudiation | правка противоречит пину 05-03 молча | medium | mitigate | датированная Amendment 2026-10-02 в ADR-006 (ADR-006:14-16,227-231): «„deadline-under-mutex“ СТОИТ и подтверждается, не заменяется» | closed |
| T-06-09-SC | Tampering (supply chain) | npm/pip/cargo-установки | high | mitigate | новых зависимостей нет: только stdlib sync/atomic; go.mod не меняется (3 прямых deps + 1 indirect, tidy-diff гейт mise.toml:31-33) | closed |
| T-06-10-01 | Repudiation | «зелёный» кейс с ослабленными ожиданиями | high | mitigate | инжекция строго БЕЗ паузы перед буквой (суть кейса; «any pause BEFORE the letter is forbidden» — 06-10-SUMMARY); ожидания из пробы (в/d); двойная зелень D-48-формой 24/24+12/12 | closed |
| T-06-10-02 | Denial of Service | экологический WARN шины трактован неверно | medium | mitigate | гейт — только раунды кейса: round-scoped WARN baseline (switchEngineWarnCount с базы раунда; 06-10-SUMMARY patterns); внераундовые WARN'ы фиксируются честно | closed |
| T-06-10-SC | Tampering (supply chain) | npm/pip/cargo-установки | high | mitigate | новых зависимостей нет; go.mod не меняется (tidy-diff в mise run ci) | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on (high) count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-06-1 | T-06-05-03 | Вырожденно длинный «токен» в скоринге: длина ограничена буфером исправления (CapSurroundingText, фразовые рамки Фаз 2-3 — вход Check никогда не превышает токен поля); скоринг O(длина), детектор чист и синхронен | Owner (plan 06-05 disposition) | 2026-10-04 |
| AR-06-2 | T-06-09-02 | Шторм минта движков (churn контекстов) без сериализации мьютексом: AttachEngine — store указателя O(1), движки минтит ibus per-контекст как раньше; churn существовал до фикса G-6-1 | Owner (plan 06-09 disposition) | 2026-10-04 |
| AR-06-3 | T-06-06-02 (residual) | GTK3-парольные поля сообщают роль 61 «text» — неотличимы от обычного текста; ролевой гейт их не видит. Остаточный риск ограничен: слой off по умолчанию, white-list пуст по умолчанию, точное равенство имён; пользователю предписано не вносить недоверенные приложения (README/CONFIG.md/ADR-007) | Owner (ADR-007 Accept, verify-work) | 2026-10-04 |
| AR-06-4 | white-list UX gap (06-UAT Deferred Follow-Up test 3, не строка реестра) | Заполнение white-list сегодня требует знания bridge-идентичности приложения (доступной через кейс-харнесс); неверная догадка отказывает fail-closed (ac_skip_app_not_listed) — security-экспозиции нет, риск чисто UX. Backlog v1.1.x: токен последней наблюдаемой идентичности в goswitchctl status | Owner (UAT verdict 2026-10-04) | 2026-10-04 |

*Accepted risks do not resurface in future audit runs.*

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-04 | 40 | 40 (38 mitigated + 2 accepted) | 0 | gsd-security-auditor (L1 grep depth + full -race suite green at HEAD + 06-REVIEW/06-REVIEW-FIX verdicts folded in) |

L1 evidence spot-checks (grep/символьный уровень, все на HEAD):
- **Privacy D-20/D-21 (ядро фазы):** перечитан КАЖДЫЙ emission-сайт новых phase-6 путей — internal/detect: 0 вхождений slog, Verdict без токен-полей (detect.go:66-71); actor.go новые записи — slug/error-only (:1973-1977, :1990, :2065, :2119-2123); ctlsvc ac-токены — только счётчики (:160-167); журнальные non-leak пины (actor_test.go:5067-5068, :5327-5328). Единственный контент-несущий канал — предфазовые DEBUG-записи (surrounding push :1112, logCorrectionDone :2445), исключительно за `-debug` (logging.Setup; флаг сам предупреждает «passwords become visible», cmd/goswitchd/main.go:31) — санкционированный carve-out.
- **Критический контур D-53:** role 40 в default-ветке отказа (actor.go:2055-2068); разрешены только 61/79/94; GetRoleName = 0 по репо; white-list exact-match (:1907); fail-closed инверсия MACR-деградации (TestAutoCorrect_FailClosedNoFailOpen); 25ms дедлайн (:74) с пином RoleDeadlineBounded; async-окно закрыто CR-01/CR-02/WR-01 фиксами (acConfirmRefusals :2087-2107; acBoundary always-verify :2339) с RED→GREEN доказательствами 06-REVIEW-FIX и живой UAT-3 пробой 18/18.
- **Конфиг:** KnownFields(true) (load.go:28); потолок 64 безусловный; default off (TestDefaults_AutocorrectOff); CONFIG.md зеркален.
- **Данные:** LICENSE-data.md + заголовки; golden-инварианты dict_test.go (включая потолок 4096); TestParseDIC.
- **Гейты фазы:** полный `go build ./...` + `go test -race -count=1 ./...` зелёный на HEAD (15/15 пакетов); go.mod — 3 прямых зависимости (tidy-diff).

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter
- [x] 06-REVIEW verdicts folded in (2C+1W fixed RED→GREEN; 4 Info documented — IN-01 counter semantics, IN-02 INFO noise, IN-03 godbus pending-call accumulation, IN-04 fixture Wait race; ни один не является строкой threat-реестра)

**Approval:** verified 2026-10-04
