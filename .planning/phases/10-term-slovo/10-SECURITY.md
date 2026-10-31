---
phase: "10"
slug: "term-slovo"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-10-31"
---

# Phase 10 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.
> Register authored at plan time (16 планов, блоки `<threat_model>`); верификация закрытий — gsd-security-auditor, L1/L2 grep-depth, ASVS 1, block_on: high. Планы 10-14/10-15 retired (status: superseded, #2349) — из активного реестра исключены.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| демон ↔ терминальные клиенты (VTE/wezterm) | собственный канал IBus (ForwardKeyEvent+commit_text), без внешних инжекторов | синтетические клавиши коррекции; CLIPBOARD никогда не пишется |
| -debug трейс / pty-захваты → журналы | трассировка клавиш экспонирует ввод | INFO-ряды без содержимого (текст только на DEBUG, пин тестом); pty — tmp-only, hex |
| пробы/вербы → рабочий стол | debug-пробы выдают синтетические клавиши | только под -debug, fail-closed nil-dep; инъекция только в поверхности стенда (ВМ) |
| демон → gsettings/GNOME | переключение раскладки, чтение sources | индексы раскладок, не секреты |
| установка release → хост владельца | новый release-dir, старые пути отката нетронуты | бинарник + unit; откат назван до приёмки |

---

## Threat Register

| Threat ID | Severity | Disposition | Status | Evidence |
|-----------|----------|-------------|--------|----------|
| T-10-01-01 | medium | mitigate | closed | 10-01-SUMMARY.md:43 (hex/счётчики в журнал, pty tmp-only); test/e2e/case_spike_caps.go:94 |
| T-10-01-02 | high | mitigate | closed | 10-01-SUMMARY.md:37 (venue vm-primary; host only owner-UAT; stand daemon inside VM) |
| T-10-01-03 | medium | mitigate | closed | 10-01-SUMMARY.md:135 (apt — только архив Ubuntu 24.04) |
| T-10-SC (планы 01-05, 07-12) | low | accept | accepted | npm/pip/cargo инсталляции отсутствуют — Go stdlib, ноль новых модулей (go.mod проверен замороженным); rationale в threat_model планов |
| T-10-02-01 | low | mitigate | closed | internal/config/config.go:35,416-427 (потолок 64, regex+empty validation, sentinel errors) |
| T-10-02-02 | medium | mitigate | closed | internal/config/load.go:28 (KnownFields strict), :45-46 (nil→defaults); docs/CONFIG.md:84 |
| T-10-02-03 | low | mitigate | closed | синтетический корпус ghbdtn-класс (internal/correct/runs_test.go:29-38); D-20/D-21 нетронуты |
| T-10-02-04 | medium | mitigate | closed | internal/clipboard/clipboard.go:42,178-186 (литеральные флаги; контент через pipe, не argv) |
| T-10-03-01 | high | mitigate | closed | internal/session/actor.go:3387-3422 classifyTerminal — verifyDead-first fail-closed без identity; вето на арме :3358; TestAutoCorrect_TerminalVetoBeforeDetector (actor_test.go:7774) |
| T-10-03-02 | medium | mitigate | closed | actor.go:3511-3517 (INFO-маркер смены класса); стирание только HandleLifecycle :1351-1359 |
| T-10-03-03 | medium | mitigate | closed | actor.go:3410-3417 (GetRole под autoRoleTimeout-дедлайном; err→warn, класс не стирается) |
| T-10-03-04 | low | mitigate | closed | actor.go:3567-3575 recordACAbstain (закрытые слаги + INFO без содержимого); terminal-veto :3359 |
| T-10-04-01 | high | mitigate | closed | actor.go:4246-4254,4293-4296 (free-abort до первого форварда); chunks ≤8 :63,4290 + перечек chunkBoundaryMoved :4255,4310; verify-after пины actor_test.go:8259,8330 |
| T-10-04-02 | high | mitigate | closed | actor.go:793-794 (paste запрещён by construction), :2503-2509 (ранг недостижим); TestTerminalBranch_VerifyAfterNoRungOnMismatch :8259 |
| T-10-04-03 | medium | mitigate | closed | replayBurst форма пинов (actor.go:4266-4300); живые pty-порядок ряды case_term_word_drain.go:210-214,384 |
| T-10-04-04 | low | mitigate | closed | actor.go:4521-4528 (changed=true) vs :3861,:3951 (changed=false) — D-24-дискриминатор |
| T-10-04-05 | medium | mitigate | closed | logCorrectionDone: INFO без содержимого, текст только DEBUG (actor.go:4521-4528) |
| T-10-05-01 | medium | mitigate | closed | 10-01-SUMMARY.md:43 + 10-05-SUMMARY.md:61 (hex-журнал, tmp-only) |
| T-10-05-02 | high | mitigate | closed | venue-вердикт vm-primary (10-01-SUMMARY.md:37); venue-маркер в каждой строке (10-05-SUMMARY.md:61) |
| T-10-05-03 | medium | mitigate | closed | арминг-гейт key-trace (10-05-SUMMARY.md:31,55; case_term_word_vte.go:205-209) |
| T-10-05-04 | medium | mitigate | closed | honest-abort ряды (10-12-SUMMARY.md:31; 10-18-SUMMARY.md:69) |
| T-10-06-01 | high | mitigate | closed | машина 10-04 в коде (см. T-10-04-01) + pane-pty дельта-оракул ×2 (10-06-SUMMARY.md:17) |
| T-10-06-02 | medium | mitigate | closed | сентинел clip-frozen (case_term_word_wezterm.go:343-355; 10-06-SUMMARY.md:17); leg-probe red — лег не подключён (:19,55) |
| T-10-06-03 | medium | mitigate | closed | 10-06-SUMMARY.md:270 (pty tmp-only в ВМ, журнал hex/счётчики) |
| T-10-06-04 | medium | mitigate | closed | откат A7 назван (10-06-SUMMARY.md:21,56) |
| T-10-06-05 | high | mitigate | closed | ×2-двухподряд-дисциплина (10-06-SUMMARY.md:5,17) |
| T-10-SC (10-06, wezterm .deb) | medium | mitigate | closed | 10-06-SUMMARY.md:15 (wezterm-nightly в ВМ, hold; сетевых apt нет) |
| T-10-07-01 | medium | mitigate | closed | focusFlapWindow 10 мс (actor.go:73); отложенное стирание/переарм :1351-1359,:1743-1759; TestFlapWindow* в repin-корпусе |
| T-10-07-02 | high | mitigate | closed | DefaultTerminals gnome-terminal/wezterm (config.go:97-105); substring RE2 (actor.go:3432-3441); `[]` opt-out (docs/CONFIG.md:46,84); verify-after второй пояс |
| T-10-07-03 | medium | mitigate | closed | acRoleErrors счётчик (actor.go:520-529,:3742-3756, status :1198); переарм-WARN |
| T-10-07-04 | low | mitigate | closed | одно переармиваемое окно (actor.go:1743-1759); гард устаревшего колбэка |
| T-10-08-01 | high | mitigate | closed | deferredFlip defer-until-confirmed (actor.go:799-814,:417-428); release-сайты вне FocusOut — TestTerminalBranch_FlipDeferredUntilConfirmed :8512, FlipSurvivesFocusOutGraceReleases :8774; живой ряд term-word-vte-flip |
| T-10-08-02 | medium | mitigate | closed | flipConfirmGrace 300 мс (actor.go:106); ровно один флип, слот очищается (flush nil-check :1846-1851) |
| T-10-08-03 | medium | mitigate | closed | switchTimeout 150 мс wedge-гард (actor.go:149-161) |
| T-10-08-04 | medium | mitigate | closed | журнальный гейт порядка done<settlement<mode (case_term_word_vte_flip.go:10-31; 10-18-SUMMARY.md:32) |
| T-10-08-05 | high | mitigate | closed | три замороженных ряда re-pin байт-точные (10-08-SUMMARY.md:15,88); flip-кейс — отдельный конфиг-документ |
| T-10-09-01 | high | mitigate | closed | откат назван ДО human-check (10-09-SUMMARY.md:42) |
| T-10-09-02 | high | mitigate | closed | declared human-check, вердикт не синтезируется (10-09-SUMMARY.md:20); rejected→STOP исполнен (10-UAT.md:32,36) |
| T-10-09-03 | medium | mitigate | closed | INFO-ряды D-20/D-21 (критерии Closure validation, 10-09-SUMMARY.md:20) |
| T-10-09-04 | low | accept | accepted | rationale в плане (остаточный средовой риск флап-пункта) |
| T-10-10-01 | medium | mitigate | closed | pendingStaleCache требует cacheNil при арме; selection/acBoundary исключены (actor.go:2425-2453); амендмент ADR-008 |
| T-10-10-02 | high | mitigate | closed | VerifyExpiry epoch no-op (actor.go:1653-1656); TestVerifyExpiryStaleRound_DoesNotEatNextRound :9241; TestTerminalClass_StaleVerifyExpiryLearnsNothing :7685 |
| T-10-10-03 | high | mitigate | closed | level2Busy check-and-set :4348, очистка на всех выходах :4353,:480; busy-skip :2996; пин второго входа в реплей-корпусе |
| T-10-10-04 | high | mitigate | closed | TestReplayHost0939_MatchAnswerRoutesTerminalBranch :9033; RejectionSignaturesAbsent :9133; TestReplayWezterm1921_* :7592 |
| T-10-11-01 | high | mitigate | closed | двухфазный уровень-1: armLevel1Confirm :1801, фолбэк-бюджет воспроизводит пару с громкой меткой :1855-1871; реплей-гейт ReplayHost0939/TwoPhaseLevel1 обязателен (repin-ряды) |
| T-10-11-02 | critical | mitigate | closed | flushLevel1Confirm коммитит через CAPTURED снапшот w.eng, не emitter() (actor.go:1846-1880, комментарий цитирует T-10-11-02); пины подмены эмиттера :8057-8097, :9731-9755 |
| T-10-11-03 | medium | mitigate | closed | FocusOut/Reset флашат немедленно по снапшоту (actor.go:1313); бюджет 60 мс (:97) |
| T-10-11-04 | medium | mitigate | closed | громкий INFO «correction confirm outcome=timeout» (actor.go:1861) |
| T-10-11-05 | low | mitigate | closed | flipCredit synthetic-ветка не тронута (actor.go:1296-1304); TestTerminalClass_* зелёные в repin |
| T-10-12-01 | high | mitigate | closed | реплей-гейт 09:39 ДО стенда/установки (10-12-SUMMARY.md:15,29) |
| T-10-12-02 | high | mitigate | closed | новый release-dir, откат назван, host read-only (10-12-SUMMARY.md:17,30) |
| T-10-12-03 | high | mitigate | closed | вердикт «честное ОЖИДАНИЕ (не синтезируется)» (10-12-SUMMARY.md:18); исторические тексты в 10-UAT дословно |
| T-10-12-04 | medium | mitigate | closed | D-20/D-21 (критерии Closure validation) |
| T-10-12-05 | low | accept | accepted | rationale в плане (средовой остаточный риск) |
| T-10-13-01 | high | mitigate | closed | пробные вербы ТОЛЬКО под -debug (cmd/goswitchd/main.go:312-341); nil-dep fail-closed errProbeDisabled (internal/ctlsvc/ctlsvc.go:47-50) |
| T-10-13-02 | medium | mitigate | closed | probe DEBUG-ряды несут runes, не текст (internal/session/probe.go:214-228); содержимое — только pty-оракул стенда |
| T-10-13-03 | medium | mitigate | closed | capsPollAsks — внутренний бюджет опроса (test/e2e/case_term_cap_delivery.go:58,348) |
| T-10-13-SC | medium | mitigate | closed | go.mod/go.sum заморожены (содержимое = заявленный стек; git diff --exit-code гейт в планах 13/16/17/18) |
| T-10-16-01 | critical | mitigate | closed | DST-гард executeCorrection: эмиссия только при surrSeen && surrEng==eng, иначе форварды (actor.go:4037-4062); TestDstGuard_NeverMaterializedContextGetsForwards :9789 |
| T-10-16-02 | high | mitigate | closed | свидетельство умирает в FocusOut/Reset-ветке (actor.go:1322-1328: surrSeen=false, surrEng=nil, ctxEpoch++); TestDstGuard_EvidenceDiesWithGeneration :9891 |
| T-10-16-03 | medium | mitigate | closed | TestDstGuard_MaterializedContextByteUnchanged :9846; запрет в must_haves.prohibitions плана |
| T-10-16-04 | low | mitigate | closed | corrDstGuarded счётчик (actor.go:352-358, status :1149-1199) + INFO «dst-guarded» :4057 |
| T-10-16-05 | low | mitigate | closed | docs/adr/adr-008-spike-log.md — датированные ряды, только версии/хеши/источники |
| T-10-16-SC | medium | mitigate | closed | go.mod заморожен, состав соответствует заявленному в плане |
| T-10-17-01 | high | mitigate | closed | level2Busy накрывает всё окно вкл. release-ожидание (actor.go:4321,:4427,:4461); ctxEpoch-перечек chunkBoundaryMoved :4238-4255,:4310; пин в реплей-корпусе |
| T-10-17-02 | high | mitigate | closed | delivery-метки: emission-seq/t_rel_ms/брекетинг/монотонность (deliverymarks_test.go); TestDeliveryBurst_RowsAreContentFree :262; honest-uncertainty (cause-honest=true, drain-ряд) |
| T-10-17-03 | critical | mitigate | closed | терминальная ветка структурно без DST: все non-test call-site — guarded level-1 (actor.go:4064) и debug-пробы (probe.go:209,226); press+release пары round-6 (actor.go:4266-4300); TestDstGuard в repin-рампе |
| T-10-17-04 | medium | mitigate | closed | debugPacing гейтится флагом -debug, env игнорируется (main.go:72-75); TestDebugPacing_FailClosedWithoutDebug (main_test.go:1027); громкий отказ на мусор |
| T-10-17-05 | medium | mitigate | closed | defaultTerminalDrainSettle константа (actor.go:136); flip_pending epoch-ноу-оп (deferredFlip :799-805) + lifecycle-флаш |
| T-10-17-06 | medium | mitigate | closed | метки — исходы/латентности (TestDeliveryBurst_RowsAreContentFree :262); чтений поля на хосте нет |
| T-10-17-SC | medium | mitigate | closed | go.mod/go.sum заморожены |
| T-10-18-01 | critical | mitigate | closed | приборные ряды — только штатные INFO production-демона; предпроверки read-only (10-18-SUMMARY.md:13,70); -debug unit только на зеркале стенда (:69) |
| T-10-18-02 | high | mitigate | closed | гвард 10-16 в установленной сборке; ноль крахов shell, corr_dst_guarded=0 (10-18-SUMMARY.md:12-13); крах=STOP в критериях (а)-(е) |
| T-10-18-03 | high | mitigate | closed | порядок исполнен дословно: реплей 21:32 → стенд 21:39 → установка 00:51 (10-18-SUMMARY.md:33) |
| T-10-18-04 | high | mitigate | closed | НОВЫЙ release-dir v1.1.1-180; v1.1.1/-124/-142 нетронуты; test -x отката (10-18-SUMMARY.md:13,70) |
| T-10-18-05 | high | mitigate | closed | вердикт не синтезируется — честное ОЖИДАНИЕ (10-18-SUMMARY.md:14,35); закрывающий вердикт владельца — UAT Тест 6 approved 2026-10-20 (10-UAT.md:76) |
| T-10-18-06 | high | mitigate | closed | тишина=провал, критерии (а)-(е) (10-18-SUMMARY.md:14); honest FAIL/abort ряды сохранены (:69) |
| T-10-18-07 | medium | mitigate | closed | синтетический корпус only; ряды без содержимого (D-20/D-21) |
| T-10-18-SC | medium | mitigate | closed | go.mod заморожен; mise ci ×2 REAL_EXIT=0 (10-18-SUMMARY.md:68) |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-10-01 | T-10-SC (планы 01-05, 07-12) | npm/pip/cargo инсталляции отсутствуют: Go stdlib + заявленный стек, ноль новых модулей (go.mod заморожен, гейт git diff --exit-code в планах) | владелец (threat_model планов) | 2026-10-08..19 |
| AR-10-02 | T-10-09-04 | остаточный средовой риск: флап-пункт приёмки не полностью воспроизводит окно kernel-флапа 16:20:05 | владелец (план 10-09) | 2026-10-14 |
| AR-10-03 | T-10-12-05 | остаточный средовой риск того же класса (closure validation) | владелец (план 10-12) | 2026-10-14 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-31 | 79 | 76 (+3 accepted) | 0 | оркестратор /gsd-manager через gsd-security-auditor (register authored at plan time; ASVS 1, block_on: high) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-10-31 (L1/L2, ASVS 1 — threats_open: 0, block_on: high; SECURED-вердикт аудитора)
