---
status: investigating
trigger: "в терминале в orca коррекция срабатывает только после второго double shift"
created: 2026-10-31T01:50:00+03:00
updated: 2026-10-31T23:30:00+03:00
symptoms_prefilled: true
goal: find_and_fix
---

## Round 2 — закрытие (2026-10-31, гипотеза клиентского Reset подтверждена и починена)

### Подтверждение гипотезы (2026-10-31 23:20, машинная элиминация по журналу дня)

Гипотеза Round 2 (клиентский `LifecycleReset` после ре-минта стирает класс) ПОДТВЕРЖДЕНА
машинной элиминацией по полному журналу дня green106 (обе жизни демона: до- и
после-перезагрузочной). Механика элиминации: `setVerifyDead(false)` в коде ровно два
сайте — ветка `LifecycleReset` HandleLifecycle (actor.go:1350-1352 — немедленная стирка)
и `verifyDeadWipeExpiry` (отложенная стирка через focusFlapWindow=10 мс после РЕАЛЬНОГО
FocusOut, который всегда пишет INFO-ряд `focus_out`). Для каждого из 22 рядов
`terminal class verify_dead=false` за день проверено: НИ перед одним нет ряда `focus_out`
в пределах 15 мс до стирки → отложенная ветка исключена → все 22 стирки исполнила ветка
Reset, т.е. `LifecycleReset` ДОСТАВЛЯЛСЯ в каждом случае:

```
13:55:41.874  RESET-BRANCH (приёмка владельца, +1.6ms после стирки — флап-пара)
14:22:52.104  RESET-BRANCH    14:36:37.383  RESET-BRANCH    14:55:36.132  RESET-BRANCH
15:49:10.228  RESET-BRANCH    16:05:37.917  RESET-BRANCH    16:14:07.484  RESET-BRANCH
16:26:07.172  RESET-BRANCH    17:36:37.977  RESET-BRANCH    17:39:29.432  RESET-BRANCH
17:58:16.792  RESET-BRANCH    18:03:53.645  RESET-BRANCH    18:13:26.998  RESET-BRANCH
18:22:01.415  RESET-BRANCH    18:31:36.080  RESET-BRANCH    18:43:54.364  RESET-BRANCH
18:45:48.016  RESET-BRANCH    19:03:01.820  RESET-BRANCH    19:06:21.553  RESET-BRANCH
19:23:34.079  RESET-BRANCH    20:30:33.709  RESET-BRANCH    20:37:52.048  RESET-BRANCH
= элиминация держится 22/22
```

Подпись ряда стирки стабильна весь день: стирка → через 0.5–3 мс (иногда десятки мс)
флап-пара focus_out/focus_in ТОГО ЖЕ engine — стирка всегда ПЕРВАЯ, пара — следствие.
Атрибуция «клиент (Electron ibus-IM Orca, медленный) переподключает IM-контекст после
ре-минта»: в приёмочном ряду 13:55 владелец печатал в ТОМ ЖЕ терминале за 27 мс до
стирки (41.847 terminal-veto — контекст жив, смены окна нет), пара после стирки —
same-engine 3 мс. Прямой DEBUG-ряд `lifecycle kind=reset` решено не добывать отдельным
probe-инсталлом: инструмент встроен в сам фикс (INFO-ряд `lifecycle kind=reset
aftermath=true` на каждый съеденный эхо-Reset) — живые ряды после установки несут
прямой маркер (см. install/journal-verify ниже).

### Round 2 install (2026-10-31 23:11–23:12 MSK)

- Фикс: fix/orca-terminal-client-reset-aftermath — RED f1f0019 (полный цикл
  TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip + fail-closed пин
  ClearedByResetOutsideAftermath) → GREEN 2c00c10 (карве-аут aftermath: Reset в окне
  flipAftermathWindow=2 с после собственного флипа не стирает класс; вне окна —
  байт-прежняя немедленная стирка; комментарий контракта обновлён). CI PR зелёный
  (build/vet/lint/test -race/tidy/govulncheck), слит в защищённый main через PR #25
  (merge 58b07cc; PR #24 не переоткрывался).
- Установка: `~/.local/share/goswitch/releases/v1.2.0-alpha.2-15-g2c00c10/`
  (goswitchd + goswitchctl, build из дерева fix-ветки HEAD 2c00c10, GOFLAGS=-trimpath),
  `goswitchctl install` + рестарт unit'а 23:12:00 MSK, `is-active` = active
  (pid 581041, engine goswitch-en активирован). Все прежние release-каталоги нетронуты.
- Путь отката (назван ДО пассивной верификации):
  `~/.local/share/goswitch/releases/v1.2.0-alpha.2-12-g80f3955/goswitchctl install`.

### Round 2 journal verify (2026-10-31) — ОЖИДАЕТ живого цикла (честный ряд)

Состояние на 2026-10-31 23:45 MSK: сборка Round 2 установлена и активна
(23:12:00, pid 581041 → после logout'а владельца перезапущена 23:28:06, pid 592485),
unit active, демон в reconnect-лупе (Discover() пере-читает адрес при КАЖДОЙ попытке —
подключится сам, когда владелец заново войдёт в сессию и новый ibus напишет свежий
адрес-файл; ручного вмешательства не нужно).

Ход машинной стороны 23:12–23:45: владелец был активен до 23:11:19; в 23:21:15 сессия
GNOME ЗАВЕРШЕНА (systemd: Stopping gnome-session-manager@ubuntu / IBus.session.GNOME /
dbus — честный logout; к 23:40 в loginctl только gdm-гricular экран). Драйв цикла
исполнителем невозможен и не предпринимался: окно Orca — Wayland-нативное (X11-тулинг
его не видит, второй инстанс с --user-data-dir отказал на Wayland-connect; AT-SPI из
сторонних клиентов недостоверен — фаза 9; инъекция в живую сессию владельца —
T-09-02-03). К 23:40 в loginctl только gdm-экран входа. Пассивная верификация
естественно продолжится на СЛЕДУЮЩЕЙ сессии владельца: каждое использование терминала
Orca производит полный цикл (вечер дня дал 16 циклов learn/wipe).

Оракул ряда (без изменений): следующая коррекция с ОДНОГО нажатия — нет повторного
verify-раунда; маркер регрессии — строка `terminal class verify_dead=false` между
flip-released и следующей коррекцией; прямой положительный маркер карве-аута — INFO-ряд
`lifecycle kind=reset aftermath=true`. Владелец от Round 2 освобождён («проверяй сам в
виртуалке, не дергай меня»); by-ear зачёт — UAT end-of-phase (прецедент G-10-1),
вердикт не синтезируется. Откат (если потребуется):
`~/.local/share/goswitch/releases/v1.2.0-alpha.2-12-g80f3955/goswitchctl install`.

### Round 2 VM-попытка (2026-11-01 20:05 UTC) — четвёртый канал (virsh send-key) сработал, цикл закрыт живьём

Стенд goswitch-stand подготовлен по плану: релиз v1.2.0-alpha.2-15-g2c00c10 скопирован и
установлен (`goswitchctl install` после возврата xkb us/ru; вычищен мусорный drop-in
`goswitchd.service.d/stand-debug.conf`, тянувший ExecStart на бинарь ~/go/bin от 19 окт —
без этого юнит поднимал ДО-фиксную сборку), unit active на релизном бине, движок
goswitch-en активирован, Orca поднят из squashfs-root (--ozone-platform=wayland, CDP
9222), worktree round2, терминал создан и сфокусирован (CDP .focus() на
.xterm-helper-textarea + активация окна).

Три канала НЕ доносят клавиши до IBus-пути (отказ специфичен для нативно-Wayland
Electron-окна; клики мышью до стола доходят, клавиши — нет):

- ydotool/ydotoold (удерживаемое устройство): type/key в окно Orca не доходят.
- Свежее uinput-устройство internal/uinject (одноразовый пробник, тот же API, что
  e2e-самодрайв): Type("ghbdtn") в сфокусированный xterm — терминал чист, журнал пуст.
- CDP Input.dispatchKeyEvent (double right-Shift, корректные коды): заходит в рендерер
  в обход браузерного IME-моста, IBus её не видит (by design). Orca CLI terminal send
  пишет в PTY под GUI — клавиатурным путём не является вовсе.

Рабочий канал — **virsh send-key** (клавиатура виртуальной машины как настоящее
устройство: композитор доверяет ей и даёт раскладку, uinput-классу — нет). Доставка
доказана побочным эффектом (Super открыл Обзор — диф скриншотов ~50% пикселей), после
чего фокус окна восстанавливался Esc и контролировался через document.hasFocus (CDP) и
AT-SPI-список окон (на ВМ разрешено). Уроки для будущих прогонов стенда:
(1) gnome-shell держит собственную сфокусированную поверхность — перед печатью
проверять активность окна, Esc закрывает перекрытие; (2) committed-текст в
встроенный терминал Orca при этом драйве не отрисовывался (DOM-фокус xterm не
держится под чарном активации) — журнальный оракул от этого не зависит.

Живой ряд на сборке фикса (daemon pid 3025, g2c00c10, engine goswitch-en→ru→en):

```
20:26:24  action n=2 → correction skipped empty-buffer     (проба без слова — штатно)
20:27:19  action n=2 → terminal class verify_dead=true →
          correction skipped verify-timeout                (ЕДИНСТВЕННЫЙ раунд обучения
                                                           нового контекста — по дизайну)
20:27:45  action n=2 → burst forwards=12 keys=6 → chunks → commit →
          done changed=true → flip-armed → flip-released(push) → mode to ru →
          engine re-mint /3 → flap pair → switch_engine   (коррекция №2: со 2-го нажатия
                                                           — воспроизведение исходного бага)
20:28:05  action n=2 → burst forwards=24 keys=12 → … → done changed=true →
          mode to en → engine /4                          (коррекция №3: С ОДНОГО нажатия)
20:28:40  … done changed=true                             (№4: с одного)
20:28:45  … done changed=true                             (№5: с одного)
маркеры по всей сессии: verify_dead=false — 0 строк (регрессии НЕТ);
lifecycle kind=reset aftermath=true — не наблюдался: эхо-Reset от ВМ-клиента не пришёл
(нечего глотать — класс выжил и без карве-аута; само проглатывание pinned юнит-парой
SurvivesClientResetAfterOwnFlip / ClearedByResetOutsideAftermath, GREEN на HEAD).
```

Вердикт: стационарная двухнажатная петля на сборке фикса ОТСУТСТВУЕТ — после одного
штатного раунда обучения каждая следующая коррекция (включая следующие за
собственным флипом демона и ре-минтом движка) срабатывает с одного double-shift.
Хостовой пассивный ряд продолжает накапливаться естественно (каждая сессия владельца).

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
