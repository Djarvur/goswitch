---
phase: 10
plan: 18
subsystem: host-reclosure-round4
tags: [gap-closure-round4, host-install, g-10-1, replay-before-stand, drain-row, delivery-marks, owner-reacceptance-pending, read-only-host, human-check-pending]
requires:
  - "Ряды 19:21:26/19:36:00 (DEBUG round-2 блок + дополнения) и бисекция 2026-10-16 — против чего round-4"
  - "10-17 (машина round-4: drain-settle, флип вне окна, delivery-метки, кейс term-word-drain) и 10-16 (DST-гард) — устанавливаемый контент"
  - "Процесс установки 10-06/10-09/10-12 (build → release-dir → goswitchctl install) и точка отката v1.1.1"
provides:
  - "Ряд repin-round4 (21:32–21:38 UTC): реплей-гейты 10-17 + mise run ci + go.mod заморожены — ЗЕЛЁНЫЕ на HEAD bffefc2 ДО всякого стенда (урок 2026-10-12/09:39 исполнен)"
  - "Стенд re-pin round-4 (21:39–21:48 UTC): term-word-vte / term-word-vte-flip / term-word-wezterm / term-negative-witnesses байт-точные зелёные + живой drain-ряд term-word-drain (pty-порядок при каждой комбинации матрицы default/tight/loose; delivery-метки упорядочены; cause-honest=true; CLIPBOARD нетронут); ноль DST-эмиссий, ноль крахов shell"
  - "Round-4 сборка v1.1.1-180-gefac750 установлена и активна на живом столе владельца (НОВЫЙ release-dir; v1.1.1, v1.1.1-124, v1.1.1-142 нетронуты); фикс. путь ctl; откат назван ДО приёмки; машинные предпроверки read-only зелёные (счётчики на нуле, corr_dst_guarded=0)"
  - "Round-4 блок Closure validation в DEBUG-файле + UAT Current Test №5: критерии журнальных рядов (а)-(е), гейт «слово в wezterm ОСТАЁТСЯ», суженная семантика флипа и честный стек доказательств предъявлены явно; вердикт владельца — честное ОЖИДАНИЕ (не синтезируется)"
affects:
  - "G-10-1 closure: закрывающий оракул — вердикт владельца по round-4 (approved → ряд закрытия; rejected → честные ряды, СТОП без правок кода вне нового цикла)"
  - "Фазовая верификация 10 (TERM-02, UAT end-of-phase) и harvest вехи v1.2"
tech-stack:
  added: []
  patterns:
    - "Round-N ритуал дословно: реплей-ряд → стенд-ряды → установка в новый release-dir → откат-до-приёмки → машинные предпроверки → вердикт владельца не синтезируется"
    - "Живой drain-ряд как proof-форма: pty-порядок (0x7f прежде слова) при явной матрице pacing + порядок delivery-меток + честная причина flip-релиза"
key-files:
  created:
    - ".planning/phases/10-term-slovo/10-18-SUMMARY.md"
  modified:
    - "docs/adr/adr-008-spike-log.md (ряд repin-round4 + секция «Фаза 10 re-pin round-4 G-10-1» + ряд host-install-gap-round4)"
    - ".planning/debug/DEBUG-term-host-live-failures.md (round-4 блок Closure validation; frontmatter updated)"
    - ".planning/phases/10-term-slovo/10-UAT.md (Current Test №5 + тест 5 + Summary counts)"
    - "test/e2e/case_term_word_vte_flip.go (оракул: round-4 release-cause keystroke)"
key-decisions:
  - "Оракул term-word-vte-flip переведён на round-4 release-семантику (Rule 3): терминальный слот флипа (10-17) не имеет grace/verify-релизов — verify-timeout это заключение, не причина; хвостовой newline кейса = честная keystroke-причина, ждётся flip-released(reason=keystroke) между settlement и mode; порядок усилен done<settlement<keystroke-release<mode — с.e3-инвариант сохранён и усилен"
  - "Порядок раунда исполнен дословно: реплей-ряд 21:32 → стенд 21:39 → установка 00:51; ни одного заезда/установки без зелёного ряда (урок 2026-10-12/09:39, T-10-18-03)"
  - "Установка — санкционированным процессом в НОВЫЙ release-dir v1.1.1-180-gefac750 (git describe); каталоги отката не перезаписаны (test -x v1.1.1 в verify); после установки хост строго read-only (T-10-18-01/04)"
  - "Вердикт владельца round-4 НЕ синтезируется: честное ОЖИДАНИЕ с датированной доставкой инструкции 00:54 MSK; суженная семантика флипа (флип на следующей клавише; правило 2026-09-28 отложено-не-отменено) и честный стек доказательств (машинного чтения поля хоста нет: стенд pty-drain + журнальные delivery-метки + жест владельца) предъявлены в инструкции ЯВНО (T-10-18-05)"
patterns-established:
  - "Дрейф оракула замороженного кейса за машиной: при смене release-семантики флипа стенд-кейс, пиннивший старые релизы, обязан получить release-cause-шаг нового раунда (аборт заезда-1 — честный маркер дрейфа)"
requirements-completed:
  - "requirements: [TERM-02] (см. assumptions плана: край TERM-02 принят в формулировке живых проверок UAT, формальная классификация — за владельцем, UAT Test 2 pending; закрывающий оракул G-10-1 — вердикт владельца round-4, pending)"
coverage:
  human_judgment: true
  rationale: "Машино-верифицируемая часть ВСЯ зелёная и задокументирована рядами: реплей-гейт на HEAD ДО стенда (repin-round4, -race ok), mise run ci 21 пакет REAL_EXIT=0 (×2 на финальном HEAD), go.mod/go.sum заморожены, четыре стенд-кейса байт-точные + drain-ряд вся матрица (ряды вербатим с зеркала), ноль DST-эмиссий и ноль крахов shell, установка с предпроверками (unit/exe/journal/status/счётчики), откат-каталог нетронут (test -x). НЕ закрывает план сам по себе: вердикт владельца round-4 (единственный закрывающий оракул G-10-1) — human_judgment, честно ОЖИДАЕТ; маршрутизация — UAT end-of-phase."
actuals:
  tokens: 14101    # chars/4 over the realized diff (56405 chars, 4 файла, +277/−24)
  tasks: 3
  commits: 4       # MEASURED: git rev-list --count gsd-plan-head-before-10-18..HEAD (fb8e403, bf70ff8, efac750, c416504)
  plan_head_before: bffefc2dda09e88bbc4b4185912c75869ff1577e
duration: 25min
completed: 2026-10-16
status: complete
---

# Phase 10 Plan 18: Round-4 закрытие G-10-1 — реплей-гейт, стенд re-pin + drain-ряд, установка, приёмка владельца Summary

**Round-4 разрыва G-10-1 доставлен по всей цепочке с дословным порядком ритуала: реплей-гейты 10-17 зелёные на HEAD ДО всякого стенда (ряд repin-round4), четыре стенд-кейса байт-точные + живой drain-ряд term-word-drain при всей матрице pacing (pty-порядок 0x7f-прежде-слова, delivery-метки упорядочены, причина флипа честна, CLIPBOARD нетронут), сборка v1.1.1-180-gefac750 установлена на живой стол владельца санкционированным процессом (новый release-dir, откат v1.1.1 назван до приёмки, предпроверки read-only зелёные) — вердикт владельца по гейту «слово в wezterm ОСТАЁТСЯ» честно ОЖИДАЕТ (human_judgment, UAT end-of-phase): закрывающий оракул G-10-1 — только его ответ.**

## Performance

| Метрика | План | Факт |
|---|---|---|
| Токены (chars/4 по диффу 56405 байт) | 45000 | ~14101 |
| Задачи | 3 | 3 |
| Коммиты | — | 4 (измерено: rev-list bffefc2..HEAD) |
| Длительность | — | 25 мин |

## Accomplishments

- **Реплей-гейт round-4 ДО стенда (Task 1, fb8e403):** ряд repin-round4 2026-10-16 21:32–21:38 UTC — `ReplayWezterm1921|ReplayHost1936|ReplayHost0939|TwoPhaseLevel1|FlapWindowContextIdentity|Level1ConfirmFlush|TestDstGuard` зелёные на HEAD bffefc2 (-race, 3.767s), `mise run ci` зелёный (21 пакет, REAL_EXIT=0; одиночный флак прогона 1 задокументирован — deferred-класс v1.0.0, два повторных прогона зелёные), go.mod/go.sum заморожены. Контроль сборки: HEAD = финальный HEAD 10-17, пост-плановых правок нет.
- **Стенд re-pin round-4 + живой drain-ряд (Task 2, efac750):** зеркало re-sync рецептом 10-08, unit -debug на зеркале HEAD (engines:2, goswitch-en), сессионный env по находке round-3 (XAUTHORITY в /run/user/1000). Пять заездов: term-word-vte PASS ×1 (bash/zsh/mid-line/clip-frozen, pty_hex=d0bfd180d0b8d0b2d0b5d1820a); term-word-vte-flip — честный abort заезда-1 (21:40:26) → фикс оракула (bf70ff8) → PASS заезд-2 (flip-word-bash delivery_order=done<settlement<mode + flip-clip-frozen pty_hex=67686264746e0a); term-word-wezterm — честный FAIL заезда-1 (префиксный мусор 0x78, класс repin-mechanics 10-08) → PASS заезд-2 (caps 0x9, word-a+word-b ×2, mid-line, clip-frozen, leg-probe red-замороженный); term-negative-witnesses PASS (vim/htop молчат fired=0, veto Δ=1, positive-control зелёный; sub-форма — документированный blocked-артефакт); **term-word-drain PASS со всей матрицей** (default/tight 5 мс/loose 50 мс+settle 400: каждый 0x7f строго прежде байтов слова, marks chunks→commit→flip-armed→flip-released упорядочены, cause-honest=true, CLIPBOARD-сентинел байт-нетронут). Честные ряды заездов-1 сохранены дословно (восстановлены из сохранённой копии — rsync --delete промежуточной синхронизации затёр зеркальные ряды, прецедент 10-06 учтён). Ноль DST-эмиссий (probe-вербы не использовались), ноль крахов shell, ядро дефолтное.
- **Установка round-4 + инструкция приёмки (Task 3, c416504):** гейты перед установкой повторены на финальном HEAD (реплей -race ok 3.827s, ci 21 пакет, go.mod чист); build из рабочего дерева → НОВЫЙ release-dir v1.1.1-180-gefac750 → goswitchctl install (инсталлятор сам daemon-reload + enable + restart; sources wrapped; engine activated). Предпроверки read-only: unit active, старт 00:51:27 MSK, /proc/2670465/exe = новый release-dir, journal поднятия чист (component registered engines:2), status: config_valid=true, счётчики на нуле (corrections_done=0, autocorrect_fired=0, ac_role_errors=0, corr_dst_guarded=0). Откат назван ДО приёмки: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install` (test -x зелёный, mtime Oct 8 нетронут). Round-4 блок Closure validation в DEBUG-файле: критерии (а)-(е), инструкция (гейт ОДИН + опциональные пункты), суженная семантика флипа и честный стек доказательств предъявлены ЯВНО, доставка инструкции 00:54 MSK, вердикт — честное ОЖИДАНИЕ. UAT Current Test №5 создан.

## Task Commits

| Task | Name | Commit |
| ---- | ---- | ------ |
| 1 | Реплей-гейт round-4 на HEAD — ряд repin-round4 ДО стенда | fb8e403 |
| 2a | fix: term-word-vte-flip оракул — round-4 release-cause (keystroke) | bf70ff8 |
| 2b | Стенд re-pin round-4 — четыре кейса + живой drain-ряд | efac750 |
| 3 | Установка round-4 на хост + инструкция re-acceptance | c416504 |

## Files

- `docs/adr/adr-008-spike-log.md` — ряд repin-round4 (секция round-4 re-pin, 21:32 UTC) + секция «Фаза 10 re-pin round-4 G-10-1» (сводный ряд repin-round4-stand + ряды пяти заездов вербатим с зеркала) + ряд host-install-gap-round4 (00:51 MSK)
- `.planning/debug/DEBUG-term-host-live-failures.md` — секция «Closure validation round-4 (G-10-1)»: сборка/фикс. путь/откат, машинные предпроверки 00:51–00:53 MSK, критерии рядов (а)-(е), инструкция с гейтом «слово ОСТАЁТСЯ», предъявленные семантика флипа и стек доказательств, доставка 00:54 MSK, вердикт ОЖИДАНИЕ
- `.planning/phases/10-term-slovo/10-UAT.md` — Current Test №5 (round-4 форма), тест 5, Summary total:5/pending:2
- `test/e2e/case_term_word_vte_flip.go` — оракул release-cause keystroke (Rule 3, см. Deviations)

## Decisions Made

См. key-decisions. Ядро: порядок раунда реплей→стенд→установка исполнен дословно; установка в новый release-dir с откатом-до-приёмки; хост после установки строго read-only; вердикт владельца не синтезируется, инструкция предъявляет суженную семантику флипа и честный стек доказательств открыто.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Оракул term-word-vte-flip переведён на round-4 release-семантику**
- **Found during:** Task 2, заезд-1 term-word-vte-flip (честный abort 21:40:26: «mode record after the settlement: timeout — 0 occurrences»)
- **Issue:** замороженный оракул кейса 10-08 ждал mode-запись сразу после settlement (verify-таймаут как релиз — семантика round-2); машина round-4 (10-17, юнит-пины FlipImmuneToVerify*) держит терминальный флип армед ПОСЛЕ verify-заключения по построению — релиз только по следующей клавише/подтверждающему пушу. Кейс не мог пройти зелёным на правильной машине без правки оракула.
- **Fix:** хвостовой newline кейса (pty-флаш) подан как честная keystroke-причина ДО ожидания mode; кейс ждёт flip-released(reason=keystroke) между settlement и mode/switch; порядок оракула усилен done<settlement<keystroke-release<mode; pty-оракул и CLIPBOARD-сентинел не тронуты. go build/vet/lint (0 issues) + go test ./test/e2e зелёные; заезд-2 PASS.
- **Files modified:** test/e2e/case_term_word_vte_flip.go
- **Commit:** bf70ff8

**2. [Rule 3 - Blocking] Восстановление затёртых rsync-ом зеркальных рядов**
- **Found during:** Task 2 (промежуточная синхронизация зеркала после фикса кейса затёрла 4 ряда VTE-заезда-1 и abort-ряд flip — прецедент 10-06 «журнал хоста перезаписал зеркальный»)
- **Fix:** ряды восстановлены ДОСЛОВНО из сохранённой до re-sync копии (/tmp) в свои секции; дальнейший порядок «сначала сохранить ряды, потом rsync» соблюдён
- **Files modified:** docs/adr/adr-008-spike-log.md (в составе efac750)
- **Commit:** efac750

**Всего отклонений: 2 авто-фикса (Rule 3, оба — форма исполнения, не семантика плана). Порядок ритуала, гейты и критерии приёмки исполнены дословно.**

## Issues Encountered

- **Task 3 human-check — вердикт ОЖИДАНИЕ (не отклонение):** план и human_check_mode запрещают синтез вердикта владельца. Grep-гейты Task 3 технически матчатся с ТЕКСТОМ ИНСТРУКЦИИ и ОЖИДАНИЯ (`approved|rejected`, `number: 5`, `round-4`) — записанного вердикта НЕТ; честный статус: unrun до ответа владельца (прецеденты 10-09 D4 / 10-12 13:00; оговорка записана в round-4 блоке DEBUG-файла). Routing: UAT end-of-phase.
- Одиночный флак первого прогона `mise run ci` (пакет не идентифицирован — окно вывода; два повторных прогона зелёные, REAL_EXIT=0) — до-существующий deferred-класс «ctlsvc/config/appid флаки под -race нагрузкой» (v1.0.0), вне скоупа плана, код между прогонами не менялся.
- Заезд-1 term-word-wezterm: префиксный инъекционный мусор 0x78 (класс repin-mechanics 10-08) — повторный заезд зелёный; маршрутизация/доставка отказом не затронуты (ни одного verify-timeout/verify-mismatch во всех заездах).
- Трасса round-4 на стенде честно показала себя: drain-ряд нёс латентность ~136 мс chunks→commit (окно 150 мс), delivery-метки упорядочены при всех комбинациях — машина работает как запинено реплеем.

## User Setup Required

**Owner re-acceptance round-4 (human-check закрывающего оракула G-10-1) — на живом столе, сборка v1.1.1-180-gefac750 активна. ГЕЙТ ОДИН:**
1. **wezterm:** набрать «ghbdtn», двойной правый Shift → «привет» — и слово **ОСТАЁТСЯ** на экране (не «появляется и удаляется»).

Опционально (всё записывается): 2. GNOME Terminal — то же; 3. флап-бурст + СРАЗУ жест; 4. середина фразы («ghbdtn » + хвостовое слово, курсор в конце); 5. «ytn?» → «нет,»; 6. автокоррекция молчала (машинно: `/home/nil/.local/share/goswitch/releases/v1.1.1-180-gefac750/goswitchctl status` — autocorrect_fired не вырос).

**Явно предъявлено:** (А) суженная семантика флипа — режим переключается на СЛЕДУЮЩЕМ нажатии клавиши (первая клавиша печатает уже в новом режиме), не мгновенно; правило 2026-09-28 отложено, не отменено — вердикт владельца его валидирует; (Б) честный стек доказательств — машинного чтения поля хоста нет, персистентность доказывается стопкой (стенд pty-drain + журнальные delivery-метки + этот жест).

Ответ «approved» или описание отклонений. **Откат по запросу — немедленно: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install`.**

## Next Phase Readiness

- **G-10-1:** при «approved» — ряд закрытия (gap closed by owner re-acceptance round-4, дата) дописывается в round-4 блок DEBUG, spike-log и UAT; при «rejected» — честные журнальные ряды, СТОП без правок кода вне нового цикла планирования, следующий цикл обязан реплеем НОВЫХ рядов доказать фикс до owner-запуска; откат по запросу немедленно.
- **Фазовая верификация 10 / UAT end-of-phase:** Current Test №5 ждёт вердикт; флагнутое допущение (Test 2, формальная классификация края TERM-02) остаётся за владельцем — никогда не снимается автоматически.
- Верификатору: сверка окна приёмки — `journalctl --user -u goswitchd` против критериев (а)-(е) round-4 блока; счётчики чистые с 00:51 MSK; grep-оговорка о вердикте записана в блоке.
- Блокеров нет; TERM-02 остаётся Complete (shared-ID gate, открыт 10-06); хост несёт round-4 сборку (машинное поведение корректно; видимый дефект проверяет владелец).

## Self-Check: PASSED

SUMMARY существует; все 4 задачных коммита (fb8e403, bf70ff8, efac750, c416504) найдены в git log; commits=4 измерено `git rev-list --count bffefc2..HEAD`; файлы существуют (spike-log с repin-round4/host-install-gap-round4, DEBUG с round-4 блоком, UAT с number: 5, case_term_word_vte_flip.go); release-dir v1.1.1-180-gefac750 с обоими бинарниками на месте, откат-каталог v1.1.1 нетронут (test -x); unit active на новом бинарнике (/proc exe сверен); пост-коммитная проверка удалений — пусто (diff без deletions файлов); stub-скан новых файлов — ноль; чужие untracked-файлы не тронуты (.gsd/, .zcode/, draft.txt, test/e2e/e2e, .planning/milestone.lock, .planning/debug/2026-10-16-*, .planning/phases/09-*/evidence-*.log, 10-PATTERNS.md — вне коммитов).
