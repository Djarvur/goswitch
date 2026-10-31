---
status: investigating
trigger: "в терминале в orca коррекция срабатывает только после второго double shift"
created: 2026-10-31T01:50:00+03:00
updated: 2026-10-31T13:58:00+03:00
symptoms_prefilled: true
goal: find_and_fix
---

## Round 2 — фикс 946b4cd недостаточен (2026-10-31, приёмка владельца: «работает со второго»)

Установка v1.2.0-alpha.2-12-g80f3955 (946b4cd в main через PR #24, fac7c89). Вердикт владельца 13:55: цикл «только со второго нажатия» СОХРАНИЛСЯ. Живой ряд (green106, goswitchd pid 787695):

```
13:55:36.868 action n=2 → verify_dead=true → skipped verify-timeout     (первое нажатие — как раньше)
13:55:40.516 action n=2 → burst 12/6 → commit-sent → done → flip-armed
13:55:40.946 flip-released(push) → mode to ru
13:55:40.948 engine created /goswitch/3 (ru)                            ← ре-минт
13:55:40.951 focus_out (en) / 40.952 focus_in (ru)                      ← пара ре-минта, 1 мс
13:55:40.954 switch_engine                                              ← немедленной стирки НЕТ (синтетика-карве-аут 946b4cd отработал)
13:55:41.847 autocorrect skipped terminal-veto                          ← нажатие владельца, класс ЕЩЁ ЖИВ
13:55:41.874 terminal class verify_dead=false                           ← СТИРКА через 923 мс после флипа
13:55:41.875 focus_out (ru) / 41.878 focus_in (ru)                      ← флап-пара ядра 3 мс, УЖЕ ПОСЛЕ стирки
```

**Гипотеза Round 2:** стирку в 41.874 делает ветка `LifecycleReset` HandleLifecycle (actor.go:1350-1352 — Reset стирает немедленно по контракту 10-07, лог-ряда у Reset нет). Через ~0.9 с после ре-минта КЛИЕНТ (Electron ibus-IM, медленный) асинхронно переподключает IM-контекст и доставляет Reset на новый engine-объект /goswitch/3 — фикс 946b4cd карве-аутит только synthetic FocusOut (ветка synthetic: actor.go:1301-1305 — кредит только для FocusOut, «A Reset kind never consumes a credit»), а клиентский Reset того же ре-минта стирает класс. Флап-пара 41.875/41.878 — следствие, не причина (идёт после стирки).

**Next:** (1) подтвердить — читать Reset-ветку: приходит ли LifecycleReset от клиента после ре-минта (DEBUG-ряд `lifecycle` kind=reset на -debug, либо лог-зонд); (2) RED-тест полного цикла «коррекция → флип → клиентский Reset → следующая коррекция»; (3) фикс — Reset, следующий за собственным флипом в пределах кредита/TTL (re-mint aftermath), не стирает класс, как synthetic FocusOut; настоящий Reset (реальная смена контекста) не трогать; (4) гейт, установка, пассивная верификация по журналу владельца (владелец освободил от участия: «проверяй сам в виртуалке, не дергай меня»).

Работа остановлена владельцем 2026-10-31 ~13:58 (перезагрузка машины). Сборка v1.2.0-alpha.2-12-g80f3955 установлена и останется активной; откат: `~/.local/share/goswitch/releases/v1.1.1-226-g8bd4edc/goswitchctl install`.

## Round 1 — resolution (архив)

root_cause: HandleLifecycle (internal/session/actor.go, ветка `synthetic` перехода FocusOut) стирал выученный поведенческий класс терминала (`setVerifyDead(false)`) на СОБСТВЕННОМ флипе демона: успешная терминальная коррекция ставит отложенный флип (deferCorrectionFlip terminal=true → flip-armed), следующий нажатый ключ его выпускает (releaseDeferredFlip("keystroke")), flipTo делает успешный SetGlobalEngine → flipCredit++ → ibus пере-минтует engine-объект → синтетическая пара focus_out/focus_in → кредитный FocusOut считался «смертью контекста» и класс умирал. Ошибка плана 10-07 — допущение «ре-минт = новое поколение контекста»: ре-минт обслуживает ТО ЖЕ САМОЕ поле ТОГО ЖЕ клиента (Orca IDE terminal, push-less) — отсутствие пушей SetSurroundingText свойство ПРИЛОЖЕНИЯ, а не engine-объекта. Цикл: нажатие 1 — контекст неклассифицирован → verify-раунд → Orca никогда не отвечает → VerifyExpiry учит verify_dead=true и пропускает (skipCorrection verify-timeout; флипа нет — settleCombo no-op, comboPending=false на пути double-shift) → нажатие 2 — едет выученным классом в терминальную ветку → доставка успешна → флип → класс стёрт → следующая коррекция снова платит verify-раунд. Стационарное «только со второго double-shift».
fix: ветка `synthetic` в HandleLifecycle больше не стирает класс — отменяет только отложенную очистку (cancelVerifyDeadWipe: снапшот старого engine-объекта после ре-минта никогда не завершится) и оставляет класс живым (карве-аут G-2-3 расширен с buffer hard-reset на behavioral class); Reset и НАСТОЯЩИЙ FocusOut без кредитов не изменены (отложенная очистка по-прежнему стирает класс при реальной смене контекста). Контрактные комментарии обновлены (HandleLifecycle doc, поле verifyDead, flap-окно, G-2-3 инлайн).
verification: TestTerminalClass_VerifyDeadSurvivesOwnFlip — полный цикл «learn → доставка → собственный флип → следующая коррекция» — RED до фикса (лог побайтно повторил журнал green106, включая forwards=12 keys=6), GREEN после; запиненный на старом поведении TestTerminalClass_VerifyDeadClearedBySyntheticFocusOut переписан в TestTerminalClass_VerifyDeadSurvivesSyntheticFocusOut; полный gate зелёный: go build ./..., go vet ./..., go test -race ./... (все пакеты), golangci-lint 0 issues, go mod tidy без дрейфа.
files_changed: [internal/session/actor.go, internal/session/actor_test.go]
prevention: why not caught — план 10-07 запинил мгновенную очистку юнит-тестом из проектного допущения («ре-минт = новое поколение»), а корпус не гонял ПОЛНЫЙ цикл «learn → успех → собственный флип → следующая коррекция» — двухнажатная петля была невидима для юнит-корпуса; e2e-кейсы term-word-* гоняют одиночный раунд коррекции на сессию. guard — TestTerminalClass_VerifyDeadSurvivesOwnFlip: полный цикл отныне запинен; маркер перехода `terminal class verify_dead=false` после flip-released в живом журнале = регрессия.

## Symptoms

- **Ожидаемое:** одно нажатие double-shift исправляет слово, набранное не в той раскладке.
- **Фактическое:** в интегрированном терминале Orca IDE первый double-shift не даёт видимого результата; коррекция выполняется только по ВТОРОМУ double-shift. Воспроизводится каждый раз — стационарный режим: каждая коррекция стоит двух нажатий.
- **Площадка:** хост green106, интегрированный терминал Orca IDE (Orca.AppImage 1.4.245, Electron 43.8.1, процесс orca-ide; терминал на xterm.js, AT-SPI-текста не отдаёт → работает терминальный путь доставки коррекции).
- **Демон:** goswitchd v1.1.1-226-g8bd4edc (сборка 8bd4edc), user-юнит (pid 2331418) живёт без рестартов с 2026-10-23 19:01:52 MSK.
- **Ошибок владелец не видит** — первый пропуск немой; единственный след — журнал демона.

## Evidence

- timestamp: 2026-10-31T01:02:39..01:34:34+03:00 — 10 повторов одного почерка за вечер. Каждый в точности одинаков: первый double-shift → `action n=2` (однажды n=3) → `terminal class verify_dead=true` → `correction skipped reason=verify-timeout` (~120 мс после action); второй double-shift через 1–2 с → `action` → `correction delivery` (burst-window open → chunks-sent → commit-sent → flip-armed/flip-released; латентность commit 274–765 мс) → `correction outcome=done changed=true`. Пример (01:02:39→01:02:41):
  ```
  action n=2
  terminal class verify_dead=true
  correction skipped reason=verify-timeout
  action n=2
  correction delivery stage=burst-window phase=open forwards=12 keys=6 emission-seq=27
  correction delivery stage=commit-sent latency_ms=397
  correction outcome=done changed=true
  mode to en / engine created /org/freedesktop/IBus/Engine/goswitch/94
  ```
- timestamp: 2026-10-31 — все 10 успешных коррекций вечера прошли терминальным путём доставки (10× `commit-sent`, emission-seq 27..36) — вся сессия была в терминале Orca; в обычных приложениях тем временем коррекции с одного нажатия идут штатно (01:17:17, 01:18:11, 01:25:07, 01:34:19, 01:35:07… без verify и delivery-стадий).
- timestamp: 2026-10-24..10-31 — частота `verify-timeout` в журнале: Oct 24 — 3, Oct 30 — 11, Oct 31 — 15 (Oct 25–29 — 0; терминальных коррекций в те дни не было). Первый пропуск: Oct 24 01:45:57, в первые часы жизни текущего демона (старт 10-23 19:01). Ранее механизма verify в проде не было (фаза 10, веха v1.2).
- timestamp: код — `internal/session/actor.go:1634-1654` (`VerifyExpiry`): бюджет верификационного раунда истёк без свежего ответа SetSurroundingText → `setVerifyDead(true)` → `skipCorrection("verify-timeout")`. Комментарий 1643–1648: класс терминала «живёт до конца жизни контекста, сбрасывается HandleLifecycle (FocusOut/Reset)». `actor.go:3494-3500` (`setVerifyDead`): INFO-маркер `terminal class` пишется только на ПЕРЕХОДЕ значения.
- timestamp: противоречие с тестами — `internal/session/actor_test.go:7294-7315` и `7378-7431` утверждают, что выученный класс переживает flap-пары и что «learned class routes the second round» без повторного verify-раунда. В живом журнале перед КАЖДЫМ пропуском стоит свежий переход `verify_dead=true` (10 переходов за 30 минут) — между коррекциями класс сбрасывается, обучение не приживается.
- timestamp: 2026-10-31 — код-подтверждение цепочки (чтение actor.go): classifyTerminal (3390-3410) отвечает «не терминал» пока verifyDead=false (identity Orca невидима, роль AT-SPI не терминальная) → verify-раунд (armVerifyDeadline 3985) → VerifyExpiry (1634-1654) учит класс и пропускает, ФЛИПА НЕТ (settleCombo 1911 no-op: comboPending=false — путь double-shift идёт через ExpiryAt→startCorrection, не через chord-ветку 3182) → следующее нажатие едет классом (3953-3963) → executeLevel2 откладывает флип (deferCorrectionFlip 4199, «flip-armed») → выпуск следующим нажатым ключом (feedKey 3176 «flip-released») → flipTo → успешный SetGlobalEngine → flipCredit++ (3001) → синтетический FocusOut → СТИРКА КЛАССА (бывшая строка 1340, ветка `synthetic`). Пин старого поведения: TestTerminalClass_VerifyDeadClearedBySyntheticFocusOut (примечание «the engine re-mint is a new context generation») — допущение, опровергнутое живым журналом.
- timestamp: 2026-10-31T02:13 — RED-воспроизведение TestTerminalClass_VerifyDeadSurvivesOwnFlip до фикса: лог теста побайтно повторил вечерний журнал green106 — action n=2 → verify_dead=true → skipped verify-timeout → action n=2 → burst-window open forwards=12 keys=6 → chunks-sent → commit-sent → done changed=true → flip-armed → flip-released reason=keystroke → mode to ru → `terminal class verify_dead=false` (стирка собственным флипом) → action n=2 → verify_dead=true → skipped — стационарная двухнажатная петля. После фикса тест GREEN (три жеста: skip, доставка, доставка без повторного verify).

## Current Focus

hypothesis: (рабочая гипотеза оркестратора, подтвердить/опровергнуть) каждый УСПЕШНЫЙ терминальный цикл сам стирает выученный класс: коррекция флипает режим → создаётся новый engine-инстанс (`/org/freedesktop/IBus/Engine/goswitch/111→112→113…`) с flap focus_out/focus_in → HandleLifecycle принимает это за смерть контекста (FocusOut/Reset) и сбрасывает verifyDead → следующая коррекция снова прогоняет verify-раунд, который терминал Orca (Electron/xterm.js, не отвечает SetSurroundingText — тот же класс терминалов, под который класс выучивался у wezterm) всегда проваливает → пропуск. Между двумя нажатиями подряд flap-а нет, поэтому ВТОРОЕ нажатие идёт выученным терминальным путём без verify и всегда успешно — отсюда устойчивое «только со второго раза».
test: журнал green106 (см. Evidence) + чтение HandleLifecycle: где именно verifyDead сбрасывается при собственном флипе/новом engine-инстансе; юнит-тест цикла «skip → успех+флип → следующая коррекция», показывающий повторный verify-раунд там, где тесты 7294-7431 ожидают выученный класс
expecting: HandleLifecycle сбрасывает verifyDead на фокус-флапе/новом контексте, порождённом собственным флипом демона; после фикса первая коррекция в терминале Orca срабатывает с одного double-shift, а класс терминала переживает собственный флип
next_action: resolved — корень найден и подтверждён RED→GREEN, фикс применён (см. Resolution)

## Notes

- Записи `autocorrect skipped` (role-unknown, dict-cur-hit, trigram-unsure, abstain-short) в том же журнале — отдельная подсистема автокоррекции на границах слов; к этому багу отношения не имеет, не смешивать.
- Сочетание «первое нажатие молчит, второе работает» стабильно для КАЖДОЙ коррекции (стационарный цикл), а не только для первой после фокуса — это отличает баг от простого «первый раз не выучили».
