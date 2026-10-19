---
status: diagnosed
trigger: "Source-only investigation (no stand runs): explain from exact sources why the host defect (6x ForwardKeyEvent(BackSpace) + CommitText('привет') → text appears, then ~1 s later gets deleted AGAIN) double-applies the six backspaces in wezterm-nightly 20261010-083911-e2a98e1b on GNOME 46, while the stock-kernel stand never reproduces. Sources pinned: wezterm e2a98e1b8af156988ad0ac778658ac27b60a55fb, mutter tag 46.2 (020504148), gnome-shell tag 46.0 (046351145)."
created: 2026-10-19T00:00:00Z
updated: 2026-10-19T00:00:00Z
symptoms_prefilled: true
goal: find_root_cause_only
---

# DEBUG: wezterm «повторное применение» BackSpace — разбор по исходникам (sourcedig)

- **Date:** 2026-10-19
- **Method:** read-only source dig; clones at exact commits: wezterm `e2a98e1b8af156988ad0ac778658ac27b60a55fb` (= nightly `20261010-083911-e2a98e1b`), mutter `46.2`, gnome-shell `46.0` (Ubuntu 24.04 base of the owner host; host mutter build `46.2-1ubuntu0.24.04.18` — same upstream 46.2 sources, Ubuntu patch drift not inspected, flagged below). No builds, no runs, host untouched.
- **Inputs:** owner video (twice: «ghbdtn» deleted, «привет» appears, ~1 s later deleted too), journal emission-seq proof (exactly ONE burst from goswitch), 2026-10-16 bisection (mutter crash pinned at `meta-wayland-text-input.c:254`), 2026-10-19 round-6 mandate.

## 1. Verdict in one paragraph

**Двойной доставки не существует.** Вторая волна из шести удалений — это **клиентский автоповтор wezterm**, взведённый нашими OWN форвардами: `replayBurst` шлёт `ForwardKeyEvent(BackSpace, 14, state=0)` — **только press, без release** (`internal/session/actor.go:4252`), а wezterm на каждый press повторяемой клавиши ставит `KeyRepeatState` и планировщик повтора, который снимается ТОЛЬКО release'ом той же клавиши, press'ом другой клавиши либо Enter/Leave клавиатурного фокуса. Release никогда не приходит → через `delay` (GNOME default 500 мс, доставляется нам через `wl_keyboard.repeat_info`) планировщик начинает синтезировать `WindowEvent::KeyEvent(BackSpace)` с частотой ~30 Гц — это и съедает «привет» после commit'а. На стенде не воспроизводится, потому что драйвер e2e вмешивается (клавиша/фокус) внутри 500-мс окна взвода; на хосте владелец смотрит на результат ≥500 мс — повтор стреляет. H1 и H3 опровергнуты структурой кода (никаких очередей/replay в mutter/gnome-shell для key-событий), H2 в исходной формулировке тоже, но её носителем оказался настоящий механизм — см. §4.

## 2. Delivery chain: goswitch → wezterm (каждый прыжок, file:line)

Путь **ForwardKeyEvent** (ключевой канал):

1. **goswitch engine** — `internal/session/actor.go:4252` (`replayBurst`): `eng.ForwardKeyEvent(KeyBackSpace=0xff08, 14, 0)` — сигнал `org.freedesktop.IBus.Engine.ForwardKeyEvent` на IBus-сокете; `state=0` ⇒ только press (`MaskRelease = 1<<30`, `internal/engine/keys.go:11`, никогда не ставится в этом вызове). Пары press+release в кодовой базе ЕСТЬ и правильные: `internal/session/probe.go:179-183` (`ProbeForward`), `actor.go:2669-2672` (`forwardCtrlV`), `actor.go:2862-2865` — терминальная ветка единственная, кто нарушает дисциплину пар.
2. **ibus-daemon 1.5.29** — релей сигнала в сфокусированный input context (контекст `gnome-shell`, создан в gnome-shell, см. п.3). Источник ibus-daemon не читался (за рамками бюджета; имя сигнала подтверждается `connect('forward-key-event', ...)` на стороне приёмника). Drift: помечен явно.
3. **gnome-shell 46.0** — `js/misc/inputMethod.js:87` (`connect('forward-key-event')`) → `_onForwardKeyEvent` (строки 172-184): `this.forward_key(keyval, keycode + 8, state & MODIFIER_MASK, time, press)` — один вызов на один сигнал, в `Clutter.InputMethod.forward_key`.
4. **mutter 46.2** — `clutter/clutter/clutter-input-method.c:467-502` (`clutter_input_method_forward_key`): синтезирует `ClutterKeyEvent` c флагом `CLUTTER_EVENT_FLAG_INPUT_METHOD` и **ровно один** `clutter_event_put()` (строка 500). Важно: `if (!priv->focus) return;` (483-484) — форвард в окно focus-out **молча теряется** (это может ронять release'ы, но НЕ является механизмом дубля).
5. Событие уходит в очередь clutter → глобальный фильтр MetaDisplay (`src/core/events.c:477-489` → `meta_display_handle_event`): key-события с флагом INPUT_METHOD **не захватываются** IM-фильтром (`clutter-input-method.c:459-460` — защита от петли) и проваливаются в доставку (`events.c:454-465`) → `meta-wayland.c:538` → `meta-wayland-seat.c:663` → `meta-wayland-input.c:376` → `meta-wayland-seat.c:496-500` → **`meta-wayland-keyboard.c:673-711`** (`meta_wayland_keyboard_handle_event` → `notify_key`:312 → `broadcast_key`:268 → `wl_keyboard_send_key`:304) — **один send на один put**, в `focus_resource_list`.
6. **wezterm e2a98e1b** — `window/src/os/wayland/keyboard.rs:14-115` (Dispatch<WlKeyboard>) → `window/src/os/wayland/window.rs:1188-1213` (`keyboard_event`, Key) → `window/src/os/x11/keyboard.rs:183-198` (`process_wayland_key`) → `:272-457` (`process_key_event_impl`). На press с `want_repeat` (xkb `key_repeats(BackSpace)`=true; `keyboard.rs:710-714`): dispatch события в GUI + `return Some(event)` (`keyboard.rs:450-453`) → **`window.rs:1194-1206`: установка `KeyRepeatState` + `KeyRepeatState::schedule`**.

Путь **CommitText** (текстовый канал, параллельный, не пересекается с ключевым):

1. `actor.go:4167` `eng.CommitText(...)` → ibus-daemon → gnome-shell `inputMethod.js:82,120-122` (`_onCommitText` → `this.commit(text)`) → `clutter-input-method.c:266-273` (`clutter_input_method_commit` → `CLUTTER_IM_COMMIT` event_put) → `meta-wayland-seat.c:509-513` → `meta-wayland-text-input.c:892-937` (`meta_wayland_text_input_handle_event`) → `clutter-input-focus.c:206-225` → `meta-wayland-text-input.c:277-292` (`focus_commit_text`) → `zwp_text_input_v3_send_commit_string` (строка 288) — один send на resource; затем отложенный `.done` (`defer_done`:196-217).
2. **wezterm**: `inputhandler.rs:174-179` (CommitString → pending) → `:180-198` (Done → dispatch `WindowEvent::KeyEvent(KeyCode::composed(text))`) → `termwindow/keyevent.rs:599-746` (`Key::Composed` → `pane.send_composed_text`, строка 738). К key_repeat/повторам этот путь **никогда не прикасается**.

Путь **focus_out/focus_in пар** (джиттер хоста, для полноты): смена клавиатурного фокуса окна → `meta-wayland-keyboard.c:776-838` (wl_keyboard leave/enter) и `meta-wayland-text-input.c:382-443` (zwp leave/enter + `clutter_input_method_focus_out/in`) → gnome-shell `inputMethod.js:186-218` (`vfunc_focus_in/out` → `context.focus_in()/focus_out()`) → ibus-daemon → FocusOut/FocusIn движка (то, что видит наш журнал). Клиентское зеркало: wezterm `keyboard.rs:25-52` — на Enter делает `input.enable()+commit()`, на Leave `input.disable()+commit()` — каждая пара и включается, и выключается.

## 3. Механизм двойного применения (полный причинный ряд)

1. `replayBurst` (`actor.go:4240-4267`) шлёт N=6 press-only форвардов с интервалом `level2ForwardGap` (default 20 мс, `actor.go:109`; на хосте видео-пейсинг ~120 мс — debug-override `GOSWITCH_DEBUG_LEVEL2_FORWARD_GAP_MS`, `cmd/goswitchd/main.go:122`); затем `drainSettleWindow` (default 150 мс, `actor.go:126`) и `CommitText` (`actor.go:4167`).
2. Каждый press BackSpace доходит до wezterm как `wl_keyboard.key(14, pressed)` и **заводит/заменяет** `key_repeat` (`window.rs:1198` — `.replace()`), т.е. после 6-го press'а живёт ровно один повторный планировщик — для BackSpace.
3. Release нет. Единственные снятия взвода: release той же клавиши (`window.rs:1207-1212`), press другой клавиши (там же, через replace), **Enter/Leave клавиатурного фокуса** (`window.rs:1158` — `emit_focus` делает `self.key_repeat.take()`), уничтожение окна. Commit-канал взвод не трогает (§2).
4. `KeyRepeatState::schedule` (`window.rs:96-177`): ждёт `delay` (из `wl_keyboard.repeat_info`; mutter читает gsettings `org.gnome.desktop.peripherals.keyboard`: `repeat`, `delay` (default 500 мс), `repeat-interval` (default 33 мс → rate 30) — `meta-wayland-keyboard.c:485-516, 553`), затем в цикле каждые ~33 мс re-dispatch'ит сохранённый `WindowEvent::KeyEvent(BackSpace)` (строки 156-167).
5. Итог по таймингу: последний press ≈ N·gap (при 120 мс ≈ 720 мс), commit ≈ +150 мс settle; повтор стартует в last_press+500 мс ≈ commit+~350 мс и за ~200 мс (6 символов × 33 мс) съедает «привет»; дальнейшие повторы упираются в пустую строку промпта и **невидимы** — наблюдатель видит «ровно те же 6 BackSpace, применённые второй раз». "~1 s" владельца = старт + прогресс удаления + восприятие; при большем gsettings `delay` совпадение точнее.
6. Осталось бы стрелять вечно — останавливается первым же: нажатием любой клавиши, следующим Enter/Leave (фокусная пара — та самая корреляция журнала: **одна и та же пара Leave/Leave→Enter одновременно = FocusOut/FocusIn в журнале движка И `key_repeat.take()` в wezterm**), либо закрытием окна.

**Почему стенд не воспроизводит (инференс, помечен честно):** взвод детерминирован на ЛЮБОМ стеке — нужен лишь ≥`delay` (500 мс) покой после последнего press'а без нажатий и без смен фокуса. На стенде драйвер сразу после коррекции предпринимает следующее действие (чтение/следующий кейс/selftype-набор/закрытие) — нажатие или фокусовый переход снимают взвод внутри окна; byte-exact оракул `0x7f`-серии потому и сходится. На хосте владелец ~секунду просто смотрит — повтор успевает отработать. Код e2e-драйвера в этом дайджесте не читался — это единственная некодовая экстраполяция.

## 4. Арбитраж гипотез

- **H1 (mutter/gnome-shell очередь ForwardKeyEvent-событий + flush на focus_in) — ОПРОВЕРГНУТ структурой.** Во всей цепочке нет ни одной очереди с отложенным повторным сбросом для key-событий: `clutter_input_method_forward_key` — один `event_put`, без очереди (`clutter-input-method.c:467-502`); доставка — один `wl_keyboard_send_key` на событие (`meta-wayland-keyboard.c:268-309`); на рефокус `wl_keyboard.enter` несёт только массив `pressed_keys` (state, не события — `keyboard.c:765-773`), и wezterm их игнорирует кроме trace-лога (`window.rs:1177-1184`). Единственный «повтор» в mutter — `clutter_input_method_notify_key_event` (`clutter-input-method.c:321-352`) — re-put ФИЗИЧЕСКОЙ клавиши, не обработанной IM (асинхронный ответ ibus), по одному разу на событие; флаг INPUT_METHOD исключает повторную фильтрацию (459-460). Фокусный джиттер в этой архитектуре может только ПОТЕРЯТЬ форвард (окно `!priv->focus`, 483-484), не удвоить.
- **H2 (wezterm дублирует в pty-write при redraw/resize-гонке) — ОПРОВЕРГНУТ в исходной форме, но носитель найден рядом.** Путь записи дублирования не содержит: `WindowEventSender::dispatch` — прямой колбэк без очереди (`window/src/lib.rs:213-235`); `key_event_impl` → один `pane.key_down`/`send_composed_text` (`keyevent.rs:699, 738`). Настоящий механизм — не гонка, а **детерминированный взвод клиентского автоповтора** press-only форвардами (§3): `window.rs:90-177, 1194-1206` + `keyboard.rs:450-453`.
- **H3 (два канала доставки одновременно) — ОПРОВЕРГНУТ.** Канал ровно один для клавиш (wl_keyboard через virtual re-inject в clutter-очередь) и один для текста (text-input-v3 через mutter-мост); wezterm-wayland не содержит ни строчки IBus-кода (grep по `window/`, `wezterm-gui/` — только X11-комментарий), IM-петля разомкнута флагом INPUT_METHOD (`clutter-input-method.c:459`). Асинхронный replay физических клавиш и форварды движка — один и тот же единственный канал доставки, а не два параллельных.

## 5. Round-6 вопрос: DeleteSurroundingText + CommitText вне replay-пути?

**Да — и ещё дальше: вне пути ПРИМЕНЕНИЯ вообще.** wezterm e2a98e1b в обработчике `ZwpTextInputV3` матчит только `PreeditString`, `CommitString`, `Done`; `DeleteSurroundingText` попадает в `_ => {}` — **игнорируется полностью** (`window/src/os/wayland/inputhandler.rs:166-200`). Конструкция round-6 в этом клиенте удалит 0 символов и просто вставит «привет» позади «ghbdtn» (`ghbdtnпривет`). Ключевой канал (форварды, взвод повтора) text-input-сообщения не задевают ни в одной точке — структурно вне replay-пути — но это палка о двух концах: удалять они в wezterm тоже не умеют. Вдобавок остаётся kills-фактор 2026-10-16: DST в virgin-окне роняет mutter 46.2 (`meta-wayland-text-input.c:254`, `strlen(NULL+мусорный курсор)` — см. note 2026-10-16). Round-6-матрица на стене это и должна показать: для wezterm применим только CommitText; DST мёртв дважды (no-op клиенту, SIGSEGV композитору в virgin-окне).

## 6. Проверяемые предсказания (для следующего прогона на стенде/хосте)

1. Вторая волна стартует в `last_backspace_press + gsettings delay` (default 500 мс), каденция = `repeat-interval` (default 33 мс) — видна в RAW-pty захвате как продолжение `0x7f` после паузы.
2. `gsettings set org.gnome.desktop.peripherals.keyboard repeat false` (rate=0 → mutter шлёт `repeat_info rate=0` → wezterm-планировщик выходит сразу, `window.rs:109-111`) — дефект исчезает без единой правки goswitch.
3. Любая клавиша (или фокусовый переход) в течение `delay` после последнего форварда — вторая волна не наступает; это объясняет зелёность стендовых орокулов.
4. Правка в goswitch: пресс+релиз-пара в `replayBurst` (по образцу `probe.go:179-183`, `actor.go:2669-2672`) — release гасит взвод (`window.rs:1207-1212`) и дефект исчезает при любом gsettings.
5. На хосте повтор должен наблюдаться и БЕЗ commit'а: чистый `probe-delete`-стайл burst press-only форвардов + 600 мс покоя → 0x7f-хвост в pty (если получится безопасно измерить).

## 7. Ограничения и drift

- Исходники ibus-daemon 1.5.29 не вскрывались (релей сигнала принят по приёмнику `connect('forward-key-event')`); journal уже доказал single-emission со стороны goswitch, дублей релея код не показывает.
- Host mutter = Ubuntu-патченный `46.2-1ubuntu0.24.04.18`; читался апстрим 46.2. Crash-фрейм 2026-10-16 совпадает со строкой 254 апстрима — патч-дрейф по этим файлам ничтожен, но не проверен построчно.
- «Stand не воспроизводит, потому что драйвер вмешивается < 500 мс» — инференс (код драйвера не читался), остальные звенья — код.
- Оценка "~1 с" из видео не разложена покадрово; формула предсказания 1 точнее видео.

## Resolution

root_cause: терминальная коррекция шлёт press-only ForwardKeyEvent(BackSpace) (actor.go:4252, state=0, без MaskRelease); wezterm 20261010-083911-e2a98e1b на каждый press повторяемой клавиши взводит клиентский автоповтор (window.rs:1194-1206 + keyboard.rs:450-453), снимаемый только release/другой клавишей/сменой клавиатурного фокуса; release никогда не приходит → через repeat-delay (GNOME default 500 мс) wezterm сам синтезирует вторую волну BackSpace с частотой ~30 Гц. Никакой повторной доставки форвардов нет; mutter/gnome-shell структурно не содержат replay-очередей для key-событий (H1/H3 опровергнуты), pty-write wezterm не дублирует (H2 в исходной форме опровергнута; настоящий механизм — детерминированный взвод повтора).
fix: (диагноз-only по мандату) кандидат — пары press+release в replayBurst по образцу ProbeForward (probe.go:179-183); предсказания §6 для верификации.
verification: §6 (predictions 1-5), параллельный DST-probe не тронут.
files_changed:
  - .planning/debug/2026-10-19-wezterm-repeat-sourcedig.md (этот файл)
