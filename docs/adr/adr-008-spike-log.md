# ADR-008: журнал спайка — терминальный транспорт (wezterm + GNOME Terminal/VTE)

**Вопрос спайка (фаза 09):** какой канал доставляет демонскую коррекцию и флип в терминалы, и что именно ломает переключение раскладки в wezterm (дефект владельца, 2026-10-08 — ежедневный терминал)?

**Критерий доказательства — только НАБЛЮДАЕМОЕ (прецедент d01):** каждая строка ниже — датированный вердикт живой пробы, которую исполнил кейс (`mise run e2e-spike-smoke`, `mise run e2e-spike-wezterm`, `mise run e2e-spike-caps`, `mise run e2e-wezterm-flip`). Веб-источники строк не пишут (Evidence-правило ADR-008, T-09-02-05): апстрим-ловушки упоминаются hazard-строками без ссылок, выбор ветки фикса — не здесь, а у владельца (чекпойнт 09-03).

**Дисциплина корпуса:** только синтетический текст (`ghbdtn`, SENTINEL-токены); pty-захваты живут во временных файлах — в журнал и коммиты идут hex и счётчики, никогда содержимое (Pitfall 6, -debug «пароли становятся видимыми»); `wl-paste` всегда с `--no-newline`; пустое чтение PRIMARY — определённое состояние, не ошибка (Pitfall 5); каждая select-chord проба идёт с сентинелом PRIMARY (известный пред-стейт → утверждение смены на ожидаемый префикс).

**Формат:** таблицы создаются пустыми, строки дописывают кейсы; колонка «Наблюдаемое» содержит наблюдения, «Вердикт» — вывод только из них.

## Средовой блокиратор активации — живой бисект (2026-10-09)

Сессия утрачивает выдачу фокуса новым окнам на мапе; инъекция в несфокусированную поверхность запрещена, поэтому все пробы фазы идут под **ручной активацией** (owner-authorized): окно открывается — владелец кликает в него (продуктивный жест), инъекция ждёт маркер фокуса (AT-SPI witness ИЛИ свежая focus_in-пара демона), каждая строка журнала несёт `activation: manual-click`.

| Факт | Наблюдаемое | Статус |
|---|---|---|
| Свежесессионное средство | relogin 10:13:59–10:14:12 не вылечил: пробы 10:39 (возраст сессии 25 мин — внутри D-48 гейта ≤30 мин), 10:48, 11:01 (простаивающий стол) и 13:21 — 4/4 с одинаковой сигнатурой (witness gnome-shell:WINDOW:chars=-1, диалог глух к Enter, ноль focus_in от окна пробы) | подтверждено |
| Bitwarden (главный подозреваемый) | автостартовал headless `--ozone-platform=x11` каждую сессию (почему relogin не лечил); закрыт владельцем (7 процессов → 0) с явного одобрения — проба 13:21 всё ещё падает; микродельта: пары focus_in появились на мапе окна (в пробах 10:48/11:01 их не было) — динамика изменилась, фокус не удерживается | исключён как триггер, микродельта записана |
| Демон | goswitchd остановлен + `ibus engine xkb:us::eng` → проба всё ещё падает | исключён |
| Расширения GNOME | все 6 отключены → проба всё ещё падает | исключены |
| Клиентский путь | GDK_BACKEND=x11 (XWayland) zenity — идентичный отказ нативно-Wayland форме | исключён |
| Чтение /dev/input | per-device phantom capture заблокирован uaccess для non-root shell — ограничение инструмента, не свидетельство | не-свидетельство |
| Корреляция загрузки | все 12 отказов фокуса — на текущей загрузке (boot-0); ядро 7.2.9-2-t2 (обновлено 2026-10-08) — оставшаяся дельта; reboot-проба отклонена владельцем | открытая гипотеза (для t2-linux-отчёта) |
| Живучесть машины фокуса | ввод владельца течёт через демон (10:14:20, 10:33:02 — автокоррекционные записи юнит-журнала; 21:22:39 — живая коррекция «ytn ybrfrjuj jryf цуяеукь» → «нет никакого окна wezterm» в собственном терминале владельца, уровень 1) | подтверждено |
| A3 (форма спавна) | `wezterm start --config-file …` отвергается CLI целиком («unexpected argument», окно не рождается вовсе) — aborted-строки 21:17/21:22 это форма спавна, а не фокус; верная форма: `wezterm --config-file … start --always-new-process -- PROG` (21:25: pane жив, readiness-маркер создан) | закреплено |
| Клик-активация (чистый прогон 21:37–21:41, оркестратор hands-off; владелец подтвердил: «вижу все окна, кликнул в каждое, включая wezterm-фикстуры») | окно смаплено, pane-программа работает (readiness-маркер), НО за все 3 минуты клик-ожидания — НОЛЬ новых IM-контекстов любого класса (0 caps-записей; witness gnome-shell неизменен): контекст wezterm не минтится даже по прямому клику владельца | дефект глубже активации |
| Загрязнение 21:26–21:35 | параллельные пробы оркестратора (zenity ENTER-TEST + Enter-бёрсты) — маркеры этого окна (4 чужих caps-записи, engine/4 goswitch-ru в 21:34:04) не сигнал фикстуры | помечено загрязнением |
| Пути активации (пробы 2026-10-09/10) | atspi-grabFocus ✗ (30× поков, фокус не выдаётся) / map-auto-grant флакующий (13:21 — пары на мапе; чистый прогон 21:37 — ноль) / EWMH xdotool windowactivate на XWayland ✗ (Enter не доставляется) → клик владельца — единственный работающий грант активации | закреплено (ADR-008 Evidence + t2-linux-отчёт) |
| Лаунч-атрибуция (2026-10-11, живая проба владельца) | владелец кликнул в ожидающее окно фикстуры и НАБРАЛ текст — ничего не появилось; суммарно: auto-grant ✗/флак, grab ✗, EWMH ✗, клик-фокус ✗ (демон не видит focus-событий по клику, только map-пары). ВЫВОД: окна, запущенные из ФОНОВЫХ процессов (агентских), никогда не получают клавиатурной интерактивности на ядре 7.2.9-2 — ни авто, ни по клику; окна, запущенные самим владельцем, работают (их собственный терминал жив). Ретроспективно объясняет WINDOWS #7 (фоновая клик-активация): старое ядро терпело фоновые окна, новое запрещает им клавиатуру. | центральное свидетельство ADR-008 D2 + ядро t2-linux-отчёта; валидационная проба owner-launched фикстуры — следующая строка |

Вывод для ADR-008: выдача фокуса срывается на уровне композитора; userspace исключён живым бисектом. Терминальные вердикты фазы от этого не зависят содержательно — транспорт и посадка измеряются на surfaces с ручной активацией; map-time auto-grant каждого терминала фиксируется отдельной строкой (без инъекции).

## Дым пробы

| Проба | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| spike-smoke 23:32:45 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 23:33:26 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 23:38:52 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 23:42:00 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 23:44:55 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 23:47:50 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 23:50:45 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 23:53:36 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 09:22:53 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 09:24:01 | `gather smoke evidence` | case aborted before the verdict | Verdict: aborted |
| spike-smoke 19:56:00 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 21:18:12 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 21:22:46 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 21:34:09 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 21:39:18 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 18:10:39 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 18:26:47 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 18:40:26 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 18:59:42 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 19:04:46 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 19:46:36 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 20:06:26 | `gather smoke evidence` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-smoke 21:22:55 | `armGate probe (ydotool type 1234 + Enter), 1 попытки` | activation: manual-click; engine_traced=true (probe-key records delta 2); pty неизменна — движок доставил клавиши устаревшему контексту (pty_path=/tmp/gs-owner-spike-smoke-in) | Verdict: reproduced: engine→stale-context desync (XIM) |
| spike-smoke 22:26:02 | `armGate probe (ydotool type 1234 + Enter), 1 попытки` | activation: focus-detected; engine_traced=true (probe-key records delta 2); pty неизменна — движок доставил клавиши устаревшему контексту (pty_path=/tmp/goswitch-e2e-3656689991/spike-smoke-in) | Verdict: reproduced: engine→stale-context desync (XIM) |

























## HALT-запись батча 2 (2026-10-11 вечер, решение владельца)

Батч 2 (attach-режим owner-launch) прерван владельцем после R1 FAIL («останови писк, тесты не работают»); раунды R2–R6 не исполнялись. Все записанные строки остаются свидетельствами. R1-наблюдение дословно (журнал батча /tmp/gs-batch3.log):

- 19:02:17 — кейс spike-smoke (attach) ожидает owner-launch;
- ~19:04 — «КЛИК 1 из 6 — spike-smoke engaged (daemon focus_in above case-start baseline) — owner click registered»: клик подтверждён парой focus_in демона (после патча гейта: AT-SPI-свидетель исключён из подтверждения клика — он прибит к gnome-shell для всего);
- затем FAIL: `arming gate: no key-trace record after 5 modifier taps` — при подтверждённом клике и ВИДИМОМ В ОКНЕ инжектированном вводе (владелец видел ввод в фикстуре) журнал демона -debug НЕ содержит key-записей.

Слои дефекта на текущей сессии (кумулятивно): (1) фоновый спавн → клавиатура мертва (лаунч-атрибуция); (2) клик-фокус подтверждается только focus_in демона; (3) инъекция доходит до ОКНА, но не до движка (key-trace пуст при видимом вводе) — разрыв между клиентом и движком/armGate-каналом; это подрывает и armGate-доказательство, и все инъекционные оракулы. Формальный инструмент isolации слоя (3) — физический ввод владельца против ydotool-инъекции на одной поверхности (не исполнено, сессия остановлена владельцем).

Возобновление: R2–R6 по owner-runbook (команды в журнале батча; owner launch + click + видимый ввод уже доказаны), блокирует только чин слоя (3) в гейтинге кейсов. Процессы остановлены, рабочий стол тих.

## Лестница первопричины P0–P5 (wezterm, базовый транспорт)

| Проба | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| P0 18:10:57 | `/proc/<wezterm-gui>/environ + xlsclients` | activation: manual-click; pids=[935482]; WAYLAND_DISPLAY=wayland-0; DISPLAY=:0; xlsclients="green106  gsd-xsettings\ngreen106  ibus-x11\ngreen106  mutter-x11-frames" | Verdict: native-wayland (WAYLAND_DISPLAY есть — XIM-гипотеза опровергнута, лестница идёт под h1/h2/h3) |
| P0 21:56:06 | `/proc/<wezterm-gui>/environ + xlsclients` | activation: manual-click; pids=[1104052]; WAYLAND_DISPLAY=wayland-0; DISPLAY=:0; xlsclients="green106  gsd-xsettings\ngreen106  mutter-x11-frames" | Verdict: native-wayland (WAYLAND_DISPLAY есть — XIM-гипотеза опровергнута, лестница идёт под h1/h2/h3) |
| P0 22:26:19 | `/proc/<wezterm-gui>/environ + xlsclients` | activation: focus-detected; pids=[1131666]; WAYLAND_DISPLAY=wayland-0; DISPLAY=:0; xlsclients="green106  gsd-xsettings\ngreen106  mutter-x11-frames" | Verdict: native-wayland (WAYLAND_DISPLAY есть — XIM-гипотеза опровергнута, лестница идёт под h1/h2/h3) |




## Проба нативно-Wayland рецепта

| Проба | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| P0' 18:14:10 | `/proc/<wezterm-gui>/environ под wayland-фикстурой` | activation: manual-click; pids=[936856]; WAYLAND_DISPLAY=wayland-0; DISPLAY=:0; xlsclients="green106  gsd-xsettings\ngreen106  ibus-x11\ngreen106  mutter-x11-frames" | Verdict: транспорт сменился (WAYLAND_DISPLAY есть — рецепт применился) |
| spike-wezterm 18:17:16 | `рецепт-проба` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-wezterm 19:10:40 | `рецепт-проба` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| spike-wezterm 22:29:14 | `рецепт-проба` | activation: focus-detected; case aborted before the verdict | Verdict: aborted |
| spike-wezterm 22:29:33 | `рецепт-проба` | activation: focus-detected; case aborted before the verdict | Verdict: aborted |
| spike-wezterm 22:55:27 | `рецепт-проба` | activation: focus-detected; case aborted before the verdict | Verdict: aborted |







## Вердикт-матрица первопричины

| P0 | P1 | P2 | P3 | P4 | Вердикт |
|---|---|---|---|---|---|

## Capability-таблица терминалов

| Строка | Терминал | Проба | Наблюдаемое | Вердикт |
|---|---|---|---|---|
| VTE a11y 18:19:12 | `focus_helper.py witness при сфокусированном VTE` | activation: manual-click; witness=gnome-shell:WINDOW:chars=-1 | Verdict: форма идентичности записана (наблюдаемое выше) |
| VTE доставка 18:19:15 | `ydotool type ghbdtn, счётчик key-записей` | activation: manual-click; key_records=0 | Verdict: клавиши НЕ доставлены (движок глух — диагностика армирования) |
| VTE select-chord (a) 18:19:27 | `ydotool shift+home после набора корпуса` | activation: manual-click; sentinel=GS-SPIKE-SENTINEL; primary_after=GS-SPIKE-SENTINEL | Verdict: выделения нет (sentinel не сменился) — kill-критерий Фазы 12: строка сужается до wezterm, вердикт о сужении — владелец (чекпойнт 09-03) |
| VTE посадка+шорткат 18:19:29 | `флип → RU → набор ghbdtn → pty hex` | activation: manual-click; flip_pair=false; pty_hex= | Verdict: шорткат НЕ работает (pty != привет — словесная коррекция VTE требует полного канала) |
| VTE caps 18:19:29 | `SetCapabilities в журнале демона для контекста VTE` | activation: manual-click; caps=0x29 (0x29-класс объявляет surrounding — недоверие, Pitfall 3) | Verdict: значение записано, решение на caps не ключуется |
| VTE итог 18:19:29 | `VTE раунд` | activation: manual-click; все пробы раунда исполнены, строки выше | Verdict: executed |
| spike-caps 22:30:05 | `armGate probe (ydotool type 1234 + Enter), 2 попытки` | activation: focus-detected; engine_traced=true (probe-key records delta 2); pty неизменна — движок доставил клавиши устаревшему контексту (pty_path=/tmp/goswitch-e2e-3388380942/spike-caps-in) | Verdict: reproduced: engine→stale-context desync (XIM) |
| VTE итог 22:30:05 | `VTE раунд (ветка арминга)` | activation: focus-detected; инъекционные раунды не исполнялись — арминг не доказал доставку в окно | Verdict: arming branch (раунды пропущены, наблюдение записано) |









## TERM-04 негатив (attach-семейство, rev D-55 §7.2)

| Проба | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| term04-negative 23:43:46 | `негатив` | activation: focus-detected; case aborted before the verdict | Verdict: aborted |
| term04-negative 23:45:17 | `негатив` | activation: focus-detected; case aborted before the verdict | Verdict: aborted |
| term04-negative 01:46:07 | `негатив` | activation: focus-detected; case aborted before the verdict | Verdict: aborted |
| term04-negative 01:56:44 | `негатив` | activation: focus-detected; case aborted before the verdict | Verdict: aborted |





## Оракул wezterm-flip

| Прогон | Раунд | pty hex | Вердикт |
|---|---|---|---|
| wezterm-flip 18:14:07 | `flip rounds` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| wezterm-flip 19:08:01 | `flip rounds` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| wezterm-flip раунд 1 21:34:26 | `armGate probe (ydotool type 1234 + Enter), 1 попытки` | activation: manual-click; engine_traced=true (probe-key records delta 2); pty неизменна — движок доставил клавиши устаревшему контексту (pty_path=/tmp/gs-owner-wezterm-flip-r1-in) | Verdict: reproduced: engine→stale-context desync (XIM) |
| wezterm-flip 21:34:29 | `flip rounds` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| wezterm-flip раунд 1 21:51:11 | `armGate probe (ydotool type 1234 + Enter), 1 попытки` | activation: manual-click; engine_traced=true (probe-key records delta 2); pty неизменна — движок доставил клавиши устаревшему контексту (pty_path=/tmp/gs-owner-wezterm-flip-r1-in) | Verdict: reproduced: engine→stale-context desync (XIM) |
| wezterm-flip 21:55:44 | `flip rounds` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| wezterm-flip 21:58:33 | `flip rounds` | activation: manual-click; case aborted before the verdict | Verdict: aborted |
| wezterm-flip раунд 1 22:28:44 | `armGate probe (ydotool type 1234 + Enter), 1 попытки` | activation: focus-detected; engine_traced=true (probe-key records delta 2); pty неизменна — движок доставил клавиши устаревшему контексту (pty_path=/tmp/goswitch-e2e-3380261729/wezterm-flip-r1-in) | Verdict: reproduced: engine→stale-context desync (XIM) |
| wezterm-flip раунд 2 22:28:54 | `armGate probe (ydotool type 1234 + Enter), 1 попытки` | activation: focus-detected; engine_traced=true (probe-key records delta 2); pty неизменна — движок доставил клавиши устаревшему контексту (pty_path=/tmp/goswitch-e2e-3380261729/wezterm-flip-r2-in) | Verdict: reproduced: engine→stale-context desync (XIM) |
| wezterm-flip итог 22:28:54 | `два раунда` | activation: focus-detected; ни одного pty_hex==привет (латиница/пусто — дефект воспроизведён, честный RED фазы) | verdict: defect-reproduced |











## A3 и hazard-строки

| Метка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| map-auto-grant wezterm 21:18:00 | `wezterm start --always-new-process (xim fixture), observation only` | activation: manual-click; auto-grant marker=focus_in within 12s | Verdict: map-time auto-grant present (focus_in) |
| A3 indirect 21:18:00 | `wezterm start --config-file gs-e2e-wezterm-xim.lua` | activation: manual-click; spawn up and owner-clicked; the option did not refuse the path (direct WAYLAND_DISPLAY re-probe belongs to the wayland fixture run) | Verdict: config-file accepted (indirect) |
| map-auto-grant wezterm 21:22:46 | `wezterm start --always-new-process (xim fixture), observation only` | activation: manual-click; no auto-grant within 12s (no witness name, no focus_in above base) | Verdict: map-time auto-grant absent (manual-click protocol required) |
| A3 indirect 21:22:46 | `wezterm start --config-file gs-e2e-wezterm-xim.lua` | activation: manual-click; spawn up and owner-clicked; the option did not refuse the path (direct WAYLAND_DISPLAY re-probe belongs to the wayland fixture run) | Verdict: config-file accepted (indirect) |
| map-auto-grant wezterm 21:31:02 | `wezterm --config-file ... start --always-new-process (xim fixture), observation only` | activation: manual-click; auto-grant marker=caps-0x9 within 12s | Verdict: map-time auto-grant present (caps-0x9) |
| map-auto-grant wezterm 21:36:12 | `wezterm --config-file ... start --always-new-process (xim fixture), observation only` | activation: manual-click; auto-grant marker=caps-0x9 within 12s | Verdict: map-time auto-grant present (caps-0x9) |
| map-auto-grant wezterm 18:07:30 | `wezterm --config-file ... start --always-new-process (xim fixture), observation only` | activation: manual-click; auto-grant marker=caps-0x9 within 12s | Verdict: map-time auto-grant present (caps-0x9) |
| map-auto-grant wezterm 18:10:57 | `wezterm --config-file ... start --always-new-process (xim fixture), observation only` | activation: manual-click; no auto-grant within 12s (no caps-0x9 fingerprint above base) | Verdict: map-time auto-grant absent (manual-click protocol required) |
| map-auto-grant wezterm 18:33:21 | `wezterm --config-file ... start --always-new-process (xim fixture), observation only` | activation: manual-click; no auto-grant within 12s (no caps-0x9 fingerprint above base) | Verdict: map-time auto-grant absent (manual-click protocol required) |
| A3 indirect 19:04:33 | `wezterm start --config-file gs-e2e-wezterm-xim.lua` | activation: manual-click; spawn up and owner-clicked; the option did not refuse the path (direct WAYLAND_DISPLAY re-probe belongs to the wayland fixture run) | Verdict: config-file accepted (indirect) |
| A3 indirect 20:06:12 | `wezterm start --config-file gs-e2e-wezterm-xim.lua` | activation: manual-click; spawn up and owner-clicked; the option did not refuse the path (direct WAYLAND_DISPLAY re-probe belongs to the wayland fixture run) | Verdict: config-file accepted (indirect) |
| A3 indirect 21:22:52 | `wezterm start --config-file gs-e2e-wezterm-xim.lua` | activation: manual-click; spawn up and owner-clicked; the option did not refuse the path (direct WAYLAND_DISPLAY re-probe belongs to the wayland fixture run) | Verdict: config-file accepted (indirect) |
| map-auto-grant wezterm 21:56:06 | `wezterm --config-file ... start --always-new-process (xim fixture), observation only` | activation: manual-click; no auto-grant within 12s (no caps-0x9 fingerprint above base) | Verdict: map-time auto-grant absent (manual-click protocol required) |
| A3 indirect 22:26:00 | `wezterm start --config-file gs-e2e-wezterm-xim.lua` | activation: focus-detected; spawn up, pane ready; the option did not refuse the path (direct WAYLAND_DISPLAY re-probe belongs to the wayland fixture run) | Verdict: config-file accepted (indirect) |















## 2026-10-11 ~19:15 — Остановка пакета batch2 (owner-launch attach mode) владельцем

- **19:04:46 R1 spike-smoke: FAIL** — «arming gate: no key-trace record after 5 modifier taps». Парадокс зафиксирован дословно: клик владельца подтверждён (журнал демона: пары focus 18:57:50), инъекция видна в окне («там появился ввод» — владелец), но key-trace-записи в демоне нет. Это — предмет разбора следующего захода.
- **Раунды R2–R6 не исполнены.** Владелец остановил пакет из-за звуковых циклов и незакрывающихся окон; все процессы пакета убиты оркестратором ~19:10, таблица ниже не дополнялась.
- **Центральное свидетельство среды (питает ADR-008 D2 и отчёт о регрессии ядра):** окна, запущенные из фоновых процессов, не получают клавиатуру никаким способом — map-time auto-grant нестабилен (21:18 есть / 21:22 нет), AT-SPI grabFocus ✗, EWMH windowactivate ✗, клик владельца ✗ (фокус-пары демона при клике возникают, ввод в окно не попадает — «поле не печатает»). Окна, запущенные самим владельцем, клавиатуру принимают (валидация 18:50: буквы появляются; заголовочное окно attach-режима). Вывод: инъекция возможна только в «пользовательские» окна; запуск фикстур должен принадлежать владельцу, а стенд — подключаться к ним, а не порождать.
- **Runbook подтверждён работоспособным** на шаге доставки: владелец вставляет команду → окно владельца → клик → инъекция видна. Не решённым остаётся только армирование/трейс на стороне кейса.
- AT-SPI-репортинг фокуса на этой машине недостоверен глобально (свидетель всегда «gnome-shell:WINDOW» даже для окон с работающим вводом владельца) — свидетеля убрать из всех маркеров стенда.

## 2026-10-12 ~02:00 — Вердикт: нативный Wayland-рецепт мёртв на этой сборке (дифференциальный тест)

- `wezterm --config enable_wayland=true` умирает при старте: «Error while flushing display: Broken pipe» — у владельца (01:56:50) и у оркестратора (01:55:59, одинаковая сигнатура). mutter журналирует «WL: error in client communication» ровно на PID обоих wezterm'ов.
- Дифференциальный контроль: другой Wayland-клиент (GDK_BACKEND=wayland zenity) живёт минуты — композитор здоров, принимает и держит клиентов.
- Вердикт: wezterm данной сборки (апстрим-тег 20240203) несовместим с Wayland этого GNOME — фикс флипа через enable_wayland=true невозможен без обновления wezterm. Остаётся ветка фикса в движке goswitchd (контекст-трекинг XIM-клиентов — см. вердикты desync выше), опирающаяся на reproduced-строки этого журнала, либо обновление wezterm.

## 2026-10-12 ~02:20 — Лестница «обновление wezterm»: ночная сборка 20261010 живёт на Wayland этого GNOME; установлена, рецепт применён

Датированные пробы оркестратора (поручение владельца «починить работу в wezterm, а если невозможно — подобрать замену»; деталь — .planning/debug/2026-10-12-wezterm-nightly-probe.md, незакоммичено).

| Проба | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| P1 02:06 | `wezterm-nightly_20261010 .deb → dpkg-deb -x /tmp → wezterm-gui --config 'enable_wayland=true'` (бок-о-бок, установленный не тронут) | жив 10 с / 75 с / ~3 мин; stderr чист (краш-сигнатуры нет); сокет-дескриптор = `wayland-0`; окно+панель живы (`wezterm cli list`: 80x24, шелл); ежедневный ~/.wezterm.lua владельца загружен без ошибок; goswitchd записал focus-пары на спавн (02:06:24) | Verdict: ночная сборка — нативный Wayland-клиент, стабилен |
| P2 02:07 | `wezterm-gui --config 'enable_wayland=false'` (ночник, негатив-контроль) | жив на 30 с | Verdict: X11-режим ночника не деградировал |
| P3 02:15 | `sudo apt-get install wezterm-nightly` (замена стабильной: Conflicts) + `apt-mark hold`; ~/.wezterm.lua:78 → `enable_wayland = true`; контрольная проба установленного `/usr/bin/wezterm-gui` без переопределений | жив 8 с / 63 с; панель жива; GUI-процесс держит `wayland-0`; симлинк фронта `wayland-wayland-0-org.wezfurlong.wezterm` создан | Verdict: установленная ночная сборка работает в ежедневных условиях |

- Дифференциал к вердикту 02:00: та же машина, тот же GNOME/mutter, тот же конфиг-путь — стабильная 20240203 умирает при старте, ночная 20261010 живёт минутами. Несовместимость была свойством сборки, не среды.
- XIM-путь больше не задействуется: нативный Wayland-клиент идёт text-input → mutter → IBus — канал, на котором продукт работает сегодня (GTK-приложения).
- НЕ доказано здесь (недоказуемо без владельца при клавиатурно-мёртвых фоновых окнах ядра 7.2.9-2-t2): флип печатью в реальном окне. Доказательство — обычная работа владельца в свежезапущенном wezterm (не тест-протокол).
- Откат задокументирован: строка 78 → false, либо `apt-mark unhold wezterm-nightly && apt-get install wezterm=20240203-110809-5046fc22`.
- Версия закреплена (`apt-mark hold`) — плавающие ежедневные ночники не меняют бинарник владельца без явного шага.

## 2026-10-12 ~13:42-14:20 UTC — Wave-0 площадка Фазы 10 (план 10-01): готовность стенда goswitch-stand, клавиатурная проба, вердикт площадки

Площадка живых рядов 10-05/10-06. Проба оркестратора ~15:30 (дневник debug-файла) уже дала KEYBOARD_ALIVE=true; этот блок воспроизводит пробу собственными датированными строками и ставит вердикт площадки. Дисциплина: маркер армирования — ТОЛЬКО key-trace демона -debug (никогда AT-SPI-свидетель, Pitfall 7); корпус только синтетический (ghbdtn-класс), pty-артефакты tmp-only, в журнал — hex и счётчики.

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| venue-ready 13:49 | `stand-env (/tmp/stand-env.sh: VM_ADDR=192.168.122.150, VM_USER=stand, VM_REPO=/home/stand/goswitch) + rsync + unit -debug` | goswitch-stand (стоковое ядро 6.8.0-142-generic): goswitchd — сборка рабочего дерева хоста 87908b0 — `systemctl --user is-active` = active, unit dist/systemd/user/goswitchd.service + drop-in stand-debug.conf (ExecStart=%h/go/bin/goswitchd -debug); компонент зарегистрирован (engines:2), глобальный движок переключён `ibus engine goswitch-en` (journal демона: engine created → focus_in → enable → caps 0x29); ydotool 0.1.8 + input-группа + /dev/uinput 0660 root:input (udev-правило 60-goswitch-stand-uinput.rules, запись открытием подтверждена); gnome-terminal (VTE) на месте; wl-clipboard 2.2.1-1build1 доставлен из архива Ubuntu 24.04; зеркало ~/goswitch = рабочее дерево хоста 87908b0 (rsync -a, исключения .git/.zcode/node_modules; stale .git удалён; MIRROR_BUILDS go1.27.1); спавн окон из ssh-сессии доходит до GNOME-сессии (WAYLAND_DISPLAY=wayland-0 + XDG_RUNTIME_DIR + DBUS bus: gnome-terminal pane-программа исполнилась, /tmp/spawn-check=ok); хостовый goswitchd владельца не тронут (0 рестартов юнита за окно задачи) | Verdict: стенд готов — внутренности доведены, зеркало актуально, оракул -debug активен |
| venue-mechanics 13:49-14:10 | `доводка стендовой механики перед пробой (4 находки)` | (1) сессия ВМ САЛЯБИЛАСЬ: autologin-сессия простояла 55 мин → idle-delay 300s + lock-enabled=true → LockedHint=yes, ScreenSaver active — первые попытки пробы шли ПОД ЭКРАНОМ БЛОКИРОВКИ (окна мапятся, key-trace и pty мертвы — та же сигнатура «инъекция видима/трейс пуст», но причина — lock, не демоны); разблокировано `loginctl unlock-session 1`, на стенде idle-lock выключен (lock-enabled=false, idle-delay=0) — иначе живые ряды 10-05/10-06 упрутся в то же; (2) sources-takeover (эквивалент install-шага ADR-001/D-40): `gsettings set org.gnome.desktop.input-sources sources "[('ibus','goswitch-en')]"` (прежнее значение [('xkb','us')] задокументировано для отката) — БЕЗ него gsd re-apply на churn uinput-устройства ydotool сбрасывает глобальный движок в foreign xkb:us::eng (WARN «foreign engine»), с ним движок переживает churn; (3) ydotool 0.1.8 named-key таблица: модификаторы НЕ доходят до ProcessKeyEvent (5 tap-ов shift = 0 записей), «BackSpace» — доходит (0xff08); `key Return` НЕ резолвится и падает в посимвольный набор букв r-e-t-u-r-n (keyval 0x72 в трейсе) — Enter доставляется `type $'…\n'` (traced 0xff0d); (4) каждое ydotool-вызывается = create/destroy uinput-устройства = bounce IM-фокуса (пары focus_out/focus_in) — доставка ключей при этом не страдает | Verdict: механика стенда закреплена для живых рядов 10-05/10-06 (unlock-gate, sources-takeover, arming-ключ BackSpace, corpus одной type-командой с \\n) |
| venue-probe r1 2026-10-12 14:12 | `armed: BackSpace x3 (key-trace delta 6, 0xff08 x3 пары) → ydotool type --key-delay 40 $'ghbdtn\n' (одна команда, 7 клавиш)` | venue: vm; activation: stand-spawn (окно заспавнено из ssh-сессии, фокус получен при мапе — стоковое ядро); критерий (а) key-trace демона: delta 14 записей «msg":"key» (6 букв + Return, press+release) — ПРОЙДЕН; критерий (б) pty-байты: hex /tmp/probe-capture-r1.bin = 67686264746e0a (ровно «ghbdtn»+LF, синтетический корпус) — ПРОЙДЕН; контроль флипа: `key shift_r` (traced) → журнал демона: action n=1 → mode to ru → switch_engine goswitch-ru → повтор набора теми же клавишами: hex += d0bfd180d0b8d0b2d0b5d1820a («привет»+LF, кириллица в pty — полный канал ProcessKeyEvent→commit) — ПРОЙДЕН | Verdict: keyboard-live (round 1 из 2 подряд) |
| venue-probe r2 2026-10-12 14:16 | `тот же протокол на свежем окне (reap r1 по pid), armed: BackSpace x3 (delta 6)` | venue: vm; activation: stand-spawn; критерий (а) key-trace: delta 14 — ПРОЙДЕН; критерий (б) pty hex = 67686264746e0a — ПРОЙДЕН; контроль флипа: action n=1 → mode to ru → switch_engine goswitch-ru; hex += d0bfd180d0b8d0b2d0b5d1820a — ПРОЙДЕН; два подряд зелёных раунда (дисциплина SPEC §7.2 rev D-55, применённая к самой пробе); воспроизводит пробу оркестратора ~15:30 (debug-дневник: KEYBOARD_ALIVE=true на том же стенде); teardown: окна закрыты, pane-процессы reap по pid, pty-артефакты удалены из /tmp ВМ (hex и счётчики — единственный след); EOF-чорда ctrl+d на стенде не трейсится (доставка чорды не доказана — teardown fallback: reap по pid; учтено для кейсов 10-05/10-06) | Verdict: keyboard-live (round 2 из 2 подряд) |
| venue-verdict 2026-10-12 14:20 | `вердикт площадки по venue-probe r1 2026-10-12 14:12 + venue-probe r2 2026-10-12 14:16 (два подряд зелёных раунда — дисциплина двух подряд зелёных SPEC §7.2 rev D-55 применена к самой пробе)` | клавиатурность стенд-spawn фоновых окон ВМ goswitch-stand (стоковое ядро 6.8.0-142-generic) доказана двухступенчатой пробой дважды подряд: key-trace демона -debug + байты корпуса в pty, обе ступени зелёные в обоих раундах; закрывает Wave-0 требование VALIDATION Фазы 10 и строку «to be proven» Environment Availability RESEARCH. Операционное следствие для планов 10-05/10-06: живые ряды VTE и wezterm исполняются ВНУТРИ ВМ через /tmp/stand-env.sh (ssh stand@192.168.122.150, зеркало ~/goswitch, goswitchd -debug как оракул, арминг по key-trace) — мир пассивен для владельца; хост остаётся вне живых рядов (фоновые окна ядра 7.2.9-2-t2 клавиатурно мертвы — блокиратор 2026-10-11 выше), owner-UAT на реальном столе — финальный зачёт вне этих планов. Откаты стенда задокументированы: sources [("xkb","us")], lock-enabled=true/idle-delay=300 при необходимости | Verdict: vm-primary |

## Фаза 10 живые ряды VTE (term-word-vte / term-negative-witnesses)

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| word-bash 19:17:39 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| word-zsh 19:17:41 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| mid-line 19:17:43 | `armGate + type echo ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| clip-frozen 19:17:46 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено) |
| term-word-vte-mechanics 19:05-19:18 | `re-sync зеркала рецептом 10-01 Task 1 → go build ./... → unit-демон пересобран из кода фазы и перезапущен; четыре захода кейса до зелёного (18:40/18:47/19:06 — честные FAIL-строки заезда, 19:17 — зелёный)` | механика живых рядов VTE закреплена четырьмя находками: (1) ssh-сессии для `ibus engine` нужны DISPLAY=:0 + XAUTHORITY (mutter-Xwaylandauth.*), иначе setxkbmap-refusal; (2) a11y-шина ВМ была СЛЕПАЯ: GTK-приложения без atk-bridge (GTK_MODULES пуст) не эмитят StateChanged — gnome-terminal-server перезапущен с GTK_MODULES=gail:atk-bridge через activation-environment, после этого дерево живо (states: terminal frame active, terminal focused); (3) наблюдатель демона получила собственную регистрацию интереса в at-spi-реестре (RegisterEvent — мосты эмитят сигналы ТОЛЬКО пока кто-то держит регистрацию; живая проба: 0 StateChanged до регистрации, 153 после) — internal/appid 5f7837d; (4) A2 ОТВЕЧЕН (вопрос 10-04): тесный бурст ForwardKeyEvent+CommitText ПЕРЕСТАВЛЯЕТСЯ в VTE — клиент разбирает форварды лениво (GDK-очередь), коммит садится синхронно; живой hex «ghbdtn»→«ghприв» (порядок e1..e4, COMMIT, e5..e6); фикс — pacing executeLevel2 (sleep 20ms перед каждым форвардом и перед коммитом), после него ряды байт-точные: «привет\n» ровно (710b6d4); armGate KEY_ENTER на ydotool стенда не резолвится (печатает 'k') — арминг рядов по key-trace (x+BackSpace), форма Enter только type $'\n' | Verdict: механика закреплена; критерии 2-3 (слово в bash+zsh по pty-оракулу, CLIPBOARD байт-нетронут) — ЗЕЛЁНЫЕ на живой сессии |
| witness-vim 19:29:28 | `vim insert-mode + type ghbdtn␠ + esc + :wq (counters + buffer-file oracle)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; buffer_hex=67686264746e200a | Verdict: pass (fired=0, terminal-veto воздержания >0, буфер не изменён) |
| witness-htop 19:29:31 | `htop + type ghbdtn␠ (counters oracle + process alive)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; htop жив (pid 21884); контент-свидетельство несёт witness-vim — у TUI нет текстового буфера | Verdict: pass (fired=0, terminal-veto воздержания >0, процесс жив) |
| positive-control 19:29:34 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| autocorrect-terminal-silent-repin 19:59:37 | `runAutocorrectTerminalSilent in-process (existing case)` | activation: focus-detected; case error: zenity entry did not take focus (witness "gnome-shell:WINDOW:chars=-1") | Verdict: blocked (полный wezterm re-pin остаётся за планом 10-06) |
| word-bash 21:30:49 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| word-zsh 21:30:51 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| mid-line 21:30:53 | `armGate + type echo ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| clip-frozen 21:30:56 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено) |
| witness-vim 21:32:13 | `vim insert-mode + type ghbdtn␠ + esc + :wq (counters + buffer-file oracle)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; buffer_hex=67686264746e200a | Verdict: pass (fired=0, terminal-veto воздержания >0, буфер не изменён) |
| witness-htop 21:32:27 | `htop + type ghbdtn␠ (counters oracle + process alive)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; htop жив (pid 5989); контент-свидетельство несёт witness-vim — у TUI нет текстового буфера | Verdict: pass (fired=0, terminal-veto воздержания >0, процесс жив) |
| positive-control 21:32:29 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| autocorrect-terminal-silent-repin 21:32:33 | `runAutocorrectTerminalSilent in-process (existing case)` | activation: focus-detected; case error: autocorrect-terminal-silent: 1 fired records — the field must never fire | Verdict: blocked (полный wezterm re-pin остаётся за планом 10-06) |


## Фаза 10 re-pin после классификационных правок 10-07 (G-10-1 причина A)

Ряды re-pin замороженных стенд-рядов после правок Tasks 1-3 плана 10-07 (флап-стойкое
стирание verifyDead, дефолтный identity-якорь correction.terminals, громкая/липкая
role-диагностика). Зеркало ресинчено рецептом 10-01 Task 1, демон пересобран и
перезапущен (unit -debug, is-active = active, engine reactivated goswitch-en,
18:57:44Z).

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| term-word-vte 18:58:20 | `VTE word round` | activation: focus-detected; case aborted: activate goswitch-en via SetGlobalEngine: ibus engine goswitch-en: exit status 1: (process:56236): IBUS-WARNING **: 18:58:20.683: main.vala:277: Execute setxkbmap failed: Authorization required, but n | Verdict: aborted |
| word-bash 18:58:48 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| word-zsh 18:58:50 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| mid-line 18:58:52 | `armGate + type echo ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| clip-frozen 18:58:55 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено) |
| witness-vim 18:59:10 | `vim insert-mode + type ghbdtn␠ + esc + :wq (counters + buffer-file oracle)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; buffer_hex=67686264746e200a | Verdict: pass (fired=0, terminal-veto воздержания >0, буфер не изменён) |
| witness-htop 18:59:14 | `htop + type ghbdtn␠ (counters oracle + process alive)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; htop жив (pid 57606); контент-свидетельство несёт witness-vim — у TUI нет текстового буфера | Verdict: pass (fired=0, terminal-veto воздержания >0, процесс жив) |
| positive-control 18:59:16 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| autocorrect-terminal-silent-repin 18:59:21 | `runAutocorrectTerminalSilent in-process (existing case)` | activation: focus-detected; case error: autocorrect-terminal-silent: 2 correction records — nothing may enter the ladder | Verdict: blocked (полный wezterm re-pin остаётся за планом 10-06) |
| term-negative-witnesses 18:59:21 | `go run ./test/e2e -case term-negative-witnesses на площадке (ssh-env DISPLAY=:0 + XAUTHORITY)` | activation: focus-detected; exit кейса 0 — witness-vim/htop молчали под вето (fired=0, veto Δ=1 у каждого), позитивный контроль зелёный байт-точно; ряды выше (18:59:10-18:59:16) — его | Verdict: pass — замороженные негативные свидетели байт-точные зелёные после правок 10-07 |
| autocorrect-terminal-silent-repin 19:07:31 | `go run ./test/e2e -case autocorrect-terminal-silent на площадке — STANDALONE форма re-pin 10-06 (свежий журнал кейса)` | activation: focus-detected; venue: vm (kvm); zenity control engaged (daemon focus_in); wezterm negative surface engaged (daemon focus_in); terminal silent — fired=0, zero correction records under both IME delivery modes; exit кейса 0 | Verdict: pass — канонический counters-оракул негатива чист при коде 10-07 (якорь + флап-окно + role-счётчик); встроенная форма (ряд 18:59:21, blocked «2 correction records») — до-существующий артефакт общего журнала кейса: sub-run считает correction-записи ПОСЛЕДНЕГО прогона целиком, и в них попадают done+verify-after самого позитивного контроля свидетелей (10-05-форма блокировалась раньше — на zenity-гейте, до этой проверки); вердикт негатива несёт STANDALONE форма |
| term-word-vte-repin-mechanics 18:57-19:08 | `re-sync зеркала рецептом 10-01 Task 1 (план 10-07: код Tasks 1-3) → go build → unit -debug перезапущен → is-active; term-word-vte, term-negative-witnesses, autocorrect-terminal-silent` | три захода прогонов: заход 1 (18:58:20) — setup-отказ до строк: ssh-прогон без XAUTHORITY падает в setxkbmap-refusal (повтор терминологии term-word-vte-mechanics — env найден, форма прогона дополнена, ряд не страдал); заходы 2-3 — term-word-vte и term-negative-witnesses зелёные С ПЕРВОЙ попытки, ряды байт-точные (pty_hex и counters дословно как 19:17/19:29 10-05); все correction-вердикты гонялись кодом 10-07: verifyDead переживает флап-пары (юнит-корпус), identity-якорь [gnome-terminal, wezterm] в силе (конфиги кейсов без ключа terminals получили якорь — классификация VTE-контуров стала identity-истинной, маршрутизация терминальной ветки не изменилась), role-счётчик в status | Verdict: жёсткое ограничение разрыва удержано — замороженные ряды term-word-vte и term-negative-witnesses байт-точные зелёные после классификационных правок причины A (план 10-07) |




## Фаза 10 живые ряды wezterm (term-word-wezterm)

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| caps re-pin 20:34:27 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=0x9 (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |
| word-a 20:34:29 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| word-b 20:34:31 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| mid-line 20:34:32 | `armGate + type echo ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| clip-frozen 20:34:45 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); learning=false; clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено) |
| leg-probe 20:34:47 | `sentinel → goswitchctl probe-forward Home, shift+End → wl-paste --primary` | activation: focus-detected; venue: vm (kvm); sentinel=GS-SPIKE-SENTINEL; primary_after=GS-SPIKE-SENTINEL; token_hex=67686264746e | Verdict: leg-probe red (лег не подключён — буфер-доверие остаётся единственным механизмом по вердикту OQ 2 «contract»; граница D7 задокументирована) |
| term-word-wezterm-mechanics 20:30-20:35 | `re-sync зеркала рецептом 10-01 Task 1 → go build ./... → unit -debug перезапущен; четыре захода кейса до зелёного (20:30:26 ssh-env, 20:30:49 armGate-floor, 20:33:52 oracle-форма — честные FAIL-заезды, 20:34:29 — зелёный)` | механика живых рядов wezterm закреплена четырьмя находками: (1) ssh-прогонам кейса нужны DISPLAY=:0 + XAUTHORITY (mutter-Xwaylandauth.*) — без них активация `ibus engine` падает в setxkbmap-refusal «Cannot open display» (заход 20:30:26, setup-отказ до всяких строк; повтор терминологии term-word-vte-mechanics); (2) uinput-churn первых ydotool-вызовов реминтит движок — BackSpace press заезда 20:30:49 сел в disable-окно старого мина, трейснулся только release → арминг-гейт wezterm-ряда несёт ОГРАНИЧЕННУЮ ретрай-лестницу (3 попытки, план §7.2 «ограниченные ретраи»; попытка 2 зелёная); (3) оракул дельты — hex-ТЕКСТ, завершение строки едет литерой «0a», не байтом \n (заход 20:33:52: pty_delta_hex был байт-точен, want был ошибочен — честный FAIL формы оракула, не канала); строки заездов 20:30:49/20:33:52 затёрты повторным rsync зеркала (журнал хоста перезаписал зеркальный) — причины восстановлены дословно из сессионного вывода, ни один вердикт ими не опирается; (4) ночник на стенде минтит контекст с caps=0x9 (ряд caps re-pin) — окружающий-бит НЕ анонсируется, коррекция едет без-капс веткой level-2 ПЕРВЫМ же жестом (learning=false во всех рядах), рукопожатие обучения остаётся терпящей формой; автокоррекция на границе слова воздержалась no-caps («autocorrect skipped reason no-caps» на Enter-границе — 06-07-негатив жив в том же прогоне) | Verdict: механика закреплена; критерий 1 (слово в wezterm по pty-оракулу, два подряд зелёных раунда) — ЗЕЛЁНЫЙ на живой сессии; leg-probe red — граница D7 без механизма подтверждена пробой |
| caps re-pin 21:17:14 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=0x9 (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |
| word-a 21:17:16 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| word-b 21:17:18 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| mid-line 21:17:20 | `armGate + type echo ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| clip-frozen 21:17:22 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); learning=false; clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено) |
| leg-probe 21:17:23 | `sentinel → goswitchctl probe-forward Home, shift+End → wl-paste --primary` | activation: focus-detected; venue: vm (kvm); sentinel=GS-SPIKE-SENTINEL; primary_after=GS-SPIKE-SENTINEL; token_hex=67686264746e | Verdict: leg-probe red (лег не подключён — буфер-доверие остаётся единственным механизмом по вердикту OQ 2 «contract»; граница D7 задокументирована) |
| autocorrect-terminal-silent-repin 21:13:40 | `go run ./test/e2e -case autocorrect-terminal-silent на площадке (существующий кейс; гейты управления переведены на daemon focus_in — waitFocusInFrom, коммит fix(10-06))` | activation: focus-detected; venue: vm (kvm); zenity control engaged (daemon focus_in); wezterm negative surface engaged (daemon focus_in); terminal silent — fired=0, zero correction records under BOTH IME delivery modes; exit кейса 0; закрывает blocked-ряд 19:59:37 (VTE-секция): zenity-свидетель AT-SPI прибит к gnome-shell, демон focus_in гейты предписаны передачей 10-05 | Verdict: pass (полный counters-оракул негатива воспроизводим машинно на площадке; закрывает re-pin, оставленный плану 10-06) |
| autocorrect-terminal-silent-repin 19:07:31 | `go run ./test/e2e -case autocorrect-terminal-silent на площадке (STANDALONE форма re-pin; код 10-07 в силе)` | activation: focus-detected; venue: vm (kvm); zenity control engaged (daemon focus_in); wezterm negative surface engaged (daemon focus_in); terminal silent — fired=0, zero correction records under both IME delivery modes; exit кейса 0; см. секцию «Фаза 10 re-pin … 10-07» ниже | Verdict: pass (канонический counters-оракул негатива чист при коде 10-07) |
| host-install 00:19 | `goswitchctl install из рабочего дерева (release-dir v1.1.1-94-ge62a5e4: go build goswitchd+goswitchctl, unit re-point, ibus restart, sources wrapped (goswitch-en, goswitch-ru), engine activated goswitch-en)` | хост green106: unit active, ExecStart=$HOME/.local/share/goswitch/releases/v1.1.1-94-ge62a5e4/goswitchd, journal 00:19:41 «component registered» (engines:2), status config_valid=true; сборка фазы 10 (терминальная классификация D-55-вето, чанкированный level-2, конвертируемый пробег) вернулась в ежедневную работу владельца поверх v1.1.1-9-gb1dcb05; откат A7: `~/.local/share/goswitch/releases/v1.1.1/goswitchctl install` (релизный каталог v1.1.1 на месте); owner-зачёт — human-check 5 пунктов VALIDATION (ghbdtn→привет в wezterm и GNOME Terminal, середина строки, «ytn?»→«нет,», молчание автокоррекции), harvest конца фазы | Verdict: установлено (автоматизированная часть зелёная; ручная приёмка критерия 1 — за владельцем) |
| caps re-pin 21:31:29 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=(none) (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |
| word-a 21:31:32 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=true; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| word-b 21:31:33 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| mid-line 21:31:35 | `armGate + type echo ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| clip-frozen 21:31:38 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); learning=true; clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_delta_hex=67686264746e6e6e6e6e6e6e6e6e6e6e6e6e6ed0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: FAIL (клипборд изменён или слово не исправлено) |
| leg-probe 21:31:40 | `sentinel → goswitchctl probe-forward Home, shift+End → wl-paste --primary` | activation: focus-detected; venue: vm (kvm); sentinel=GS-SPIKE-SENTINEL; primary_after=GS-SPIKE-SENTINEL; token_hex=67686264746e | Verdict: leg-probe red (лег не подключён — буфер-доверие остаётся единственным механизмом по вердикту OQ 2 «contract»; граница D7 задокументирована) |


## Фаза 10 живой ряд VTE flip (term-word-vte-flip)

План 10-08: механизм ряда отложенного флипа (G-10-1 причина B, вердикт E3).
Секция создана планом до прогона (контракт секций журнала); строки пишет
кейс term-word-vte-flip (mise e2e-term-word-vte-flip, NOT in ci).

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| flip-word-bash 21:32:54 | `armGate + type ghbdtn + double Shift_R + journal order oracle (done<settle<mode) + pty-hex` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a; delivery_order=done<settlement<mode | Verdict: pass (pty байт-точен; флип освободился подтверждением доставки — E3 порядок держится) |
| flip-clip-frozen 21:32:57 | `printf %s GS-SPIKE-SENTINEL | wl-copy → flip word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=67686264746e0a (набор шёл в перевернутом режиме — коррекция вернула латиницу; CLIPBOARD перезаписан сентинелом — синтетика) | Verdict: pass (CLIPBOARD байт-нетронут и при активной машине флипа; слово исправлено) |

## Фаза 10 дрейф клиента wezterm и амендмент семантики (план 10-08, G-10-1 причина B, 2026-10-13)

Датированные свидетельства дрейфа стенда и вердикт владельца по контрольной точке 10-08.

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| wezterm-abort 19:36:06 и 19:37:11 | `go run ./test/e2e -case term-word-wezterm (код амендмента ещё не написан; два заезда)` | оба: FAIL «neither done+level-2 nor the verify-dead learning within 5s»; журналы демона: action n:2 → RequireSurroundingText → surrounding push text="x" cursor=1 (ПРОТУХШИЙ кэш арминг-пробы — набранное слово клиентом не пушится) → correction skipped verify-mismatch за 2 мс → рукопожатие обучения (verify-timeout → verifyDead) не срабатывает — раунд закрыт раньше бюджета | Verdict: подпись дрейфа зафиксирована — nightly минтит caps 0x29 (окружающий-бит АНОНСИРУЕТСЯ, против запиненных 10-06 0x9) и отвечает на Require протухшим кэшем; честные aborted-ряды заездов |
| drift-probe 19:43-19:50 | `ручной спавн wezterm-nightly + чтение mints демона; ibus restart; virsh reboot ВМ` | mints engine/2 (21 мс после спавна) и engine/3: caps=0x29 КАЖДЫЙ раз; ibus restart и полный reboot ВМ подпись НЕ меняют; dpkg: ibus 1.5.29-2 / mutter 46.2 / gnome-shell 46.0-15 установлены 2026-10-12 12:29 — ДО зелёных рядов 10-06 (20:34/21:17); wezterm-nightly held, dpkg -V чист; sources-takeover пережив reboot ([('ibus','goswitch-en')]) | Verdict: дрейф 0x9→0x29 — окруженческий факт, корень из стенда не идентифицируется; гипотезы пакетов/сессии/sources сняты |
| checkpoint 20:0x | `контрольная точка Task 2 владельцу` | варианты: (1) семантический амендмент, (2) ручная активация, (3) искать корень, (4) закрыть с хвостом | Verdict: владелец одобрил вариант 1; 2/3/4 отвергнуты дословно («никакой ручной активации не будет», «искать источник — бесполезно», «закрыть с хвостом — нет, мне нужен рабочий инструмент»); критерий приёмки — рабочий инструмент |
| amendment 21:0x | `TDD RED b53b44e → GREEN 442f69d (pendingStaleCache: nil-cache word-раунд + ответ без набранного пробега = класс учится, раунд едет в терминальную ветку); fail-closed для кэш-свидетельств сохранён (юзнит-пин)` | юнит-корпус: 3 целевых RED-теста + защитный пин честного GTK mismatch; мигрированные пины VerifyPaths/PhraseVerifyMismatch/RefusalsNeverFlip переведены на honest-push мир; RED_EVIDENCE_OK | Verdict: амендмент в коде, владелец одобрил (Rule 4, план-дельта в 10-08-PLAN.md, амендмент в ADR-008 секции 2026-10-13) |
| learning=true word-a 21:31:32 | `term-word-wezterm, ряд word-a` | learning=true — ПЕРВЫЙ жест ночника прошёл stale-cache маршрут амендмента: класс выучен, коррекция исполнена В ТОМ ЖЕ раунде, pty_delta_hex байт-точен; word-b learning=false (класс уже жив) | Verdict: амендмент жив на стенде — ночной 0x29-контекст корректируется первым же жестом |

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| repin-mechanics 21:08-21:36 | `re-sync зеркала рецептом 10-01 Task 1 (код амендмента + кейс flip) → go build → unit -debug → is-active; term-word-vte, term-word-wezterm, term-negative-witnesses, term-word-vte-flip` | хронология вечера: batch-1 (21:08-21:15) — все три замороженных кейса зелёные с первого захода, но ряды затёрты моим промежуточным rsync (прецедент терминологии 10-06 «журнал хоста перезаписал зеркальный»; verb(reason) восстановлены дословно из сессионного вывода); batch-2 (21:14-21:16) — мой поллинг читал сталые exit-файлы и запустил три кейса ПАРАЛЛЕЛЬНО — гонка за единственное имя демона («Set global engine failed»), честные aborted-ряды гонки в журнале зеркала затёрты следующим rsync; runs 2-6 term-word-vte (21:18-21:26) — пять честных FAIL-заездов деградировавшей uinput-доставки (жест 0xffe2 не трейсится / type недодаёт байты; класс 10-06 mechanics «uinput-churn»); reboot ВМ (boot 21:27:35) → term-word-vte 21:30 PASS с первой попытки → wezterm 21:31 PASS → witnesses 21:32 PASS → flip 21:32:57 PASS — ВСЕ ЧЕТЫРЕ на одной свежей сессии; ряды выше (21:30-21:32) — их; гейт фокуса несёт машинный nudge (ydotool-клик, сигнал продолжения — focus_in демона; tonight 0 срабатываний — гейты прошли сами); арминг-гейт VTE несёт ограниченную ретрай-лестницу wezterm-формы (SPEC §7.2) | Verdict: жёсткое ограничение разрыва удержано — term-word-vte, term-word-wezterm, term-negative-witnesses байт-точные зелёные при коде амендмента и машины отложенного флипа; flip-кейс зелёный с журнальным порядком доставки; наблюдение: clip-frozen wezterm-ряда нёс префиксный мусор в pty-потоке (16 байт 0x6e перед «привет\n») при пройденном contains-оракуле — задокументировано, вердикт ряда не опирается на чистоту префикса |

## Фаза 10 round-3 caps-проба доставки (план 10-13): P1/P2/P3 — DST honored, receipt, drain-proof

Секция создана планом до прогона (контракт секций журнала; строки пишет кейс
term-cap-delivery, mise e2e-term-cap-delivery, NOT in ci). Машина маршрутизации
10-14 читает ТОЛЬКО этот блок: три пробы × два клиента (wezterm-nightly,
gnome-terminal/VTE). Пробы: P1 — клиент применяет DeleteSurroundingText (поле
сокращается ровно на nchars); P2 — ответ RequireSurroundingText отражает только
что закоммиченный текст; P3 — ответ Require отражает применённые
ForwardKeyEvent(BackSpace). Опрос — верб probe-delete-surrounding 0 0 (DST(0,0)
нет-op + Require), решение исполнителя по read_first плана 10-13; бюджет опроса
— внутренняя константа кейса (15 опросов × 20 мс), исчерпание = честный
never-ряд, не зависание. Критерий «отражает» — курсорная арифметика против
базового опроса (курсор сдвинулся ровно на n действия) плюс отсутствие/наличие
токена в ответе; ряды несут исходы/счётчики/латентности, никогда содержимое
(D-20/D-21 — содержимое живёт только в DEBUG-журнале демона стенда).

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| term-cap-delivery 17:01:32 | `term-cap-delivery probe round` | activation: focus-detected; case aborted: activate goswitch-en via SetGlobalEngine: ibus engine goswitch-en: exit status 1: (process:2522): IBUS-WARNING **: 17:01:32.860: main.vala:277: Execute setxkbmap failed: Authorization required, but no | Verdict: aborted |
| term-cap-delivery 17:03:27 | `term-cap-delivery probe round` | activation: focus-detected; case aborted: term-cap-delivery: P1 action: exit status 1 | Verdict: aborted |
| term-cap-delivery 17:09:43 | `term-cap-delivery probe round` | activation: focus-detected; case aborted: term-cap-delivery: P1 action: exit status 1 | Verdict: aborted |
| term-cap-delivery 17:21:04 | `term-cap-delivery probe round` | activation: focus-detected; case aborted: timeout after 10s: daemon log has 0 occurrences of "\"msg\":\"key\"", want >= 12 | Verdict: aborted |
| term-cap-delivery 17:24:04 | `term-cap-delivery probe round` | activation: focus-detected; case aborted: term-cap-delivery: P1 action: exit status 1 | Verdict: aborted |
| term-cap-delivery 17:32:35 | `term-cap-delivery probe round` | activation: focus-detected; case aborted: timeout after 10s: daemon log has 0 occurrences of "\"msg\":\"key\"", want >= 12 | Verdict: aborted |
| term-cap-delivery 17:35:46 | `term-cap-delivery probe round` | activation: focus-detected; case aborted: term-cap-delivery: P1 action: exit status 1 | Verdict: aborted |
| session-regression 16:58 | `diag: найденное состояние стенда перед прогоном (ssh stand@192.168.122.150, ~/goswitch зеркало, rsync рецептом 10-08)` | сессия пользователя stand МЕРТВА с 2026-10-14 18:44: gnome-shell нет, владелец seat0 — gdm (его ibus), IBus-адрес пользователя указывал на мёртвый pid 1417; юнит goswitchd active; recipe: virsh reboot goswitch-stand (прецедент 10-08/10-12) → свежая Wayland-сессия; сессионный env прогона: DISPLAY=:0 + XDG_RUNTIME_DIR + DBUS_SESSION_BUS_ADDRESS + XAUTHORITY=/run/user/1000/.mutter-Xwaylandauth.* (НЕ в $HOME — находка этой сессии, дополнение к механике 10-06) | Verdict: стенд поднят (лучше найденного состояния — живая сессия вместо мёртвой) |
| shell-crash-correlation 17:03-17:36 | `diag по журналам four run: журнал демона (метки probe delete-surrounding / key) + journalctl --user (строки gnome-shell crashed with signal 11) — четыре прогона, ВСЕ четыре краха в ОДНОЙ точке` | 4/4 краха gnome-shell signal 11 СЕКУНДОЙ ПОЗЖЕ эмиссии пробы P1 (базовый опрос probe-delete-surrounding 0 0 = DST(0,0)+Require): заезд 17:03:22 (ask 17:03:22.380 → crashed 17:03:22), заезд 17:09:38 (ask 17:09:38.807 → crashed 17:09:38), заезд 17:23:59 (12 key-записей набора, crashed 17:23:59 — непосредственно после «Could not release device /dev/input/event4»), заезд 17:35:45 (crashed 17:35:45.96); КОНТРОЛЬ: заезды 17:21:04 и 17:32:35, где набор НЕ дошёл до демона (0 key-записей — эмиссии пробы НЕ БЫЛО) — shell ПЕРЕЖИЛ прогон; тяжёлая нагрузка того же класса (клики/набор/probe-forward) на этой же ВМ 2026-10-14 09:44-09:52 — 8 минут зелёных без краха; stack trace краха вырожден (только init.js), faulting-компонент журналами не называется | Verdict: крах воспроизводимо коррелирует с эмиссией DST+Require из движка демона — новое поведение, зелёными сессиями никогда не exercised (они несли только probe-forward); триггерный компонент (ibus-daemon→shell-панель vs mutter IM vs клиент) журналами не различим |
| kernel-experiment 17:29-17:35 | `sudo grub-reboot «Advanced options… 6.8.0-146-generic» + virsh reboot (одноразовая загрузка предыдущего ядра)` | заезд на 6.8.0-146: 17:32:35 набор не дошёл (miss), заезд 17:35:46 — набор дошёл, крах ТОЙ ЖЕ сигнатуры в точке P1 → гипотеза «регресс ядра 142→146/147 умеренна» СНЯТА: крах не ядерный | Verdict: ядро не триггер; одноразовая grub-установка откатилась сама (stand возвращён на дефолт -147 reboot'ом 17:40) |
| typed-miss-class 17:21/17:32 | `diag: два заезда с 0 key-записей при пройденном фокус-гейте` | свежая сессия: окно фикстуры открывается не в хардкод-фолбэке клика (факт харнесса WR-04 10-REVIEW: клик всегда едет по хардкод-фолбэку) → клик мимо терминала, набор уходит в пустоту; фокус-гейт проходит на focus_in самого открытия окна; повторный заезд лечит (placement-гонка свежей сессии, прецедент WR-04-замечания 10-12) | Verdict: вторичная климатическая особенность свежих сессий, не блокер сама по себе |
| environment-window 17:15 | `diag: /var/log/apt/history.log (unattended-upgrades 2026-10-15 06:53) + dpkg -l + journalctl всех загрузок` | Oct 15 06:53 unattended-upgrades: ядро 6.8.0-142→147 (142 УДАЛЁН), systemd-userland (libsystemd0/libudev1/libpam/systemd-shared), bluez, curl — ГРАФИЧЕСКИЙ стек нетронут (gnome-shell 46.0-15/mutter 46.2-18/wezterm-nightly 20261010 как 10-12); ИСТОРИЯ КРАХОВ: gnome-shell signal 6 — 2026-10-14 18:44 на СТАРОМ ядре БЕЗ всякого кейс-загруза (этим крахом сессия и умерла до сегодняшнего захода); итого: shell на этой ВМ падал и до новых пакетов, но КОРРЕЛЯЦИЯ с эмиссией пробы — новая, сегодняшняя, 4/4 | Verdict: окно деградации стенда — [2026-10-14 18:44, сейчас]; проектируемая причина краха — эмиссия, не пакеты |
| BLOCKED-VERDICT 17:45 | `итог плана 10-13 Task 3: mise run e2e-term-cap-delivery × 6 заездов (2 miss-набора, 4 краха shell) + 2 reboot + grub-эксперимент` | ШЕСТЬ поведенческих рядов P1/P2/P3 × 2 клиента НЕ СУЩЕСТВУЮТ: ни один прогон не дошёл до первого отражающего ответа — 4 краха compositor в точке первой пробы wezterm, 2 miss-набора; вердикт-таблица маршрутизации, drain/receipt-бюджеты и строка об ограничении переноса НЕ ВЫВЕДЕНЫ — выводить их не из чего; гейт секции (≥6 vm-primary P[123]-рядов) НЕ насыщен; 10-14 ЛИШЁН журнального входа маршрутизации; ЭСКАЛАЦИЯ ВЛАДЕЛЬЦУ (Rule 4): измерительный инструмент в спроектированной форме (эмиссия DST+Require движком) валит gnome-shell стенда — Require-доказуемая доставка (drain-proof P3, receipt-окно P2) НЕПРОВЕРЯЕМА как минимум на стенде, кандидат-маршруты: (а) Require-свободные пробы с pty-оракулом для P1/P3 — но вопрос P2 (receipt) без Require неформулируем, (б) композиция пробы (раздельная эмиссия DST и Require для локализации триггера), (в) разбор краха на стенде вне кейса; VTE-first-порядок не пробовался (крах в wezterm-точке до VTE; эмиссия клиент-агностична) | Verdict: BLOCKED — честный отказ плана 10-13 Task 3; никакого синтеза рядов/таблицы; работы Tasks 1-2 (вербы + кейс) в коде и зелёные |
| term-cap-delivery 18:09:29 | `term-cap-delivery probe round (-log /tmp → ~/bisect/R2-daemon.log), уже под systemd-coredump` | activation: focus-detected; case aborted: term-cap-delivery: P1 action: exit status 1; daemon-журнал: baseline-опрос DST(0,0)+Require 18:09:24.562 (0.48 с после focus_in фикстуры), НОЛЬ surrounding-push ответов, focus_out/focus_in 26.824 — shell умер в «девственном окне» ~2.3 с после эмиссии; ядро снято (coredumpctl, 39 МБ) — КРАХ 5/5 | Verdict: aborted — воспроизведение под coredump (ряд с ВМ-зеркала перенесён дословно) |
| core-stack 18:09-18:51 | `diag: coredumpctl info × 3 ядра (18:09:26 pid1292 пара, 18:45:12 pid1285 пара, 18:47:39 pid1308 DST-only) + gdb с libmutter-14-0-dbgsym 46.2-1ubuntu0.24.04.18 (ddebs, build-id 359f1d98…)` | ТРИ независимых ядра — ОДИН кадр: #4 __strlen_avx2 ← #5 meta_wayland_text_input_focus_delete_surrounding (libmutter-14.so.0+0x15a605, src/wayland/meta-wayland-text-input.c:254; локали start=0x0, focus offset=0 len=0) ← #6 clutter_input_focus_delete_surrounding ← #7 meta_wayland_text_input_handle_event ← … _clutter_event_process_filters; strlen-аргумент rdi=0x50c (дикий указатель = мусорный cursor поверх NULL-текста surrounding-состояния); GNOME Shell crashed with signal 11, сессия целиком вниз, degenerate GJS-трейс ui/init.js:21 | Verdict: краш-кадр назван — mutter 46.2 delete_surrounding на неинициализированном surrounding-состоянии (мы, goswitch, только триггер; дефект в mutter) |
| manual-runs 18:01-18:06 | `diag: два ручных заезда probe-delete-surrounding 0 0 по ssh (wezterm, юнит-демон БЕЗ рестарта)` | ОБА гейта фокуса/минта не сошлись (нового минта нет: минт принадлежит мостовому контексту gnome-shell и случается на рестарте демона/ibus, не на спавне окна); эмиссии уехали в СТАРЫЙ отвечающий контекст (push-ответы текут) — shell ПЕРЕЖИЛ оба | Verdict: ряды НЕВАЛИДНЫ как fresh-тесты (честно: попытки мимо окна триггера); вынесли факт — на тёплом контексте пара DST+Require безвредна |
| DST-only slow 18:36:56 | `probe-delete 0 6 (новый сплит-верб, DST без Require) на контексте +1.5 с после focus_in, wezterm` | rc=0 emitted; keys_delta=12; push-ов до эмиссии ноль; shell жив 10 с наблюдения | Verdict: survive — DST alone на НЕсовсем свежем контексте не валит |
| pair slow 18:40:01 | `probe-delete-surrounding 0 0 (пара) на контексте +23 с после реминта (рестарт демона), wezterm` | rc=0 emitted; keys_delta=12; shell жив 10 с | Verdict: survive — пара на тёплом состоянии безвредна |
| gate-design 18:42:37 | `run3 заезд: гейт на НОВЫЙ минт после спавна окна — abort БЕЗ эмиссии` | минт не случился (минт мостового контекста уже в baseline — случился на рестарте демона); гейт переведён на focus_in + машинный нудж (ydotool центр-клик, рецепт 10-08) | Verdict: aborted без эмиссии (дизайн-ошибка гейта, честный abort) |
| Require-only fast 18:44:33 | `probe-require (Require без DST) на 0.44 с после focus_in (нудж-гейт прошёл), wezterm, keys_delta=12` | rc=0 emitted; shell жив 10 с | Verdict: survive — Require ALONE в крах-форме БЕЗОПАСЕН (согласуется с прод-путём GTK-verify) |
| pair fast 18:45:10 | `probe-delete-surrounding 0 0 на 0.46 с после focus_in, wezterm, keys в окне` | rc=0 emitted → shell умер ~2.5 с спустя (18:45:12, ядро pid1285 — тот же кадр); БЕЗ e2e-харнесса, чисто ручная форма | Verdict: CRASH 6/6 семейства — пара в быстрой форме валит воспроизводимо |
| DST-only fast 18:47:39 | `probe-delete 0 0 (no-op DST БЕЗ Require) на 0.49 с после focus_in, wezterm` | rc=0 emitted → shell умер ~2.5 с спустя (18:47:39, ядро pid1308 — тот же кадр); параметры не важны (0,0 ≡ 0,6) | Verdict: CRASH — ДОСТАТОЧЕН DST alone в девственном окне; Require НЕ нужен |
| BISECT-VERDICT 18:51 | `итог бисекции: 6 валидных заездов матрицы (пара/DST/Require × fast/slow, wezterm; 3 reboot) + 2 невалидных ручных + воспроизведение e2e` | ТРИГГЕР: DeleteSurroundingText (любые параметры, включая no-op 0,0) на мостовом контексте text-input-v3 в «девственном окне» ≲1.5 с после focus_in свежей генерации (рестарт демона/ibus) — strlen на неинициализированном surrounding-состоянии mutter 46.2 → SIGSEGV gnome-shell, сессия вниз; Require сам по себе безопасен (fast-заезд пережит); на тёплом состоянии DST безвреден (slow-заезды пережиты); VTE-плечо НЕ понесло затрат (крах вне VTE-точки не требовал проверки — клиент-агностичность эмиссии осталась из 17:0x рядов); верб-инструментарий сплита (probe-delete/probe-require) в коде, зелёный (build+test -race+lint) | Verdict: DST — compositor-killer на этом стеке в девственном окне; round-4 маршрутизация — в .planning/debug/2026-10-16-stand-gnome-shell-dst-crash.md (verdict a/b/c) |

## Фаза 10 host-ревалидация G-10-1 (план 10-09): установка gap-сборки, 2026-10-14

Установка сборки закрытия разрыва на живой стол владельца (процесс 10-06 Task 3).
После установки хост строго read-only: инъекций нет, окон не создаётся, чужие сервисы
не перезапускаются — жесты приёмки исполняет владелец (секция Closure validation в
.planning/debug/DEBUG-term-host-live-failures.md).

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| host-install-gap 2026-10-14 00:46 | `go build goswitchd+goswitchctl из рабочего дерева (HEAD 7ace3fd = v1.1.1-124-g7ace3fd, mise run ci зелёный; оба корневых фикса G-10-1: 10-07 флап-стойкая классификация + identity-якорь [gnome-terminal, wezterm] + ac_role_errors, 10-08 defer-until-confirmed доставка + stale-cache амендмент) → release-dir v1.1.1-124-g7ace3fd → goswitchctl install (unit re-point, ibus restart, sources wrapped (goswitch-en, goswitch-ru), engine activated goswitch-en)` — **gap G-10-1 closure build** | unit active, ExecStart=$HOME/.local/share/goswitch/releases/v1.1.1-124-g7ace3fd/goswitchd (старт 00:46:57 MSK), journal: config loaded → connected → engine created goswitch-en → component registered (engines:2) → ctl service listening → engine reactivated — ошибок регистрации нет; status (фикс. путь) config_valid=true config_error= (пуст), счётчик role-ошибок 10-07 виден: ac_role_errors=0, autocorrect_fired=0; фикс. путь ctl для ВСЕХ последующих машинных проверок (не на PATH, новейший releases по mtime): /home/nil/.local/share/goswitch/releases/v1.1.1-124-g7ace3fd/goswitchctl; откат по запросу владельца (назван ДО приёмки): ~/.local/share/goswitch/releases/v1.1.1/goswitchctl install (релизный каталог v1.1.1 нетронут) | Verdict: установлено — машинная часть зелёная; re-acceptance критерия 1 (5 исходных пунктов + флап-бурст + молчание автокоррекции) — за владельцем, human-check: секция Closure validation (G-10-1) в DEBUG-term-host-live-failures.md |

## Фаза 10 re-pin round-2 G-10-1 (план 10-12): четыре терминальных кейса + реплей-гейт на коде 10-10/10-11, 2026-10-14

Re-pin замороженных стенд-рядов на коде round-2 (answer-only маршрут 10-10 + двухфазный уровень-1 10-11). Порядок исполнения — урок 2026-10-12: РЕПЛЕЙ-ГЕЙТ первым — юнит-корпус реальных хост-рядов 09:39 (ReplayHost0939*, TwoPhaseLevel1*) зелёный на HEAD 2c1ae04 до всякого стенд/хост-запуска (`go test ./internal/session/ -race -run 'ReplayHost0939|TwoPhaseLevel1'` → ok). Стенд re-sync рецептом 10-08 Task 2 (rsync -a --delete, исключения .git/.zcode/node_modules → go build goswitchd+goswitchctl → рестарт unit -debug → is-active; journal: component registered engines:2, ошибок нет). Терминальные ряды идут уровнем-2 (двухфазность уровня-1 и answer-only-маршрут их пинов не касаются: их конфиги армятся с классом/сценарием, где кэш-семантика прежняя); flip-ряд несёт delivery_order=done<settlement<mode. Механика сессии: ssh-прогонам кейса нужен сессионный env (DISPLAY=:0 + XAUTHORITY=.mutter-Xwaylandauth.*, XDG_RUNTIME_DIR, DBUS bus — повтор находки 1 term-word-wezterm-mechanics 10-06); WR-04 (10-REVIEW): xrandr-проба нуджа мертва — клик всегда едет по хардкод-фолбэку, факт харнесса; сегодня гейты фокуса прошли сами (0 ручных нуджей), отказ заезда 1 wezterm был НЕ на шаге нуджа.

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| term-word-vte 09:44:32 | `go run ./test/e2e -case term-word-vte (первый заезд БЕЗ сессионного env — честный setup-отказ)` | activation: focus-detected; case aborted: activate goswitch-en via SetGlobalEngine: setxkbmap failed: Cannot open display "default display" (повтор находки 1 механики 10-06) | Verdict: aborted (окружение ssh-сессии, не кейс; повторный заезд с env — зелёный) |
| word-bash 09:45:44 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) — **round-2 G-10-1 re-pin** |
| word-zsh 09:45:47 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass — **round-2 G-10-1 re-pin** |
| mid-line 09:45:49 | `armGate + type echo ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass — **round-2 G-10-1 re-pin** |
| clip-frozen 09:45:52 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией) — **round-2 G-10-1 re-pin** |
| flip-word-bash 09:46:09 | `armGate + type ghbdtn + double Shift_R + journal order oracle (done<settle<mode) + pty-hex` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a; delivery_order=done<settlement<mode | Verdict: pass (pty байт-точен; флип освободился подтверждением доставки — E3 порядок держится на коде 10-11) — **round-2 G-10-1 re-pin** |
| flip-clip-frozen 09:46:12 | `printf %s GS-SPIKE-SENTINEL | wl-copy → flip word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=67686264746e0a (набор шёл в перевернутом режиме — коррекция вернула латиницу; CLIPBOARD перезаписан сентинелом — синтетика) | Verdict: pass — **round-2 G-10-1 re-pin** |
| caps re-pin 09:47:10 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=0x9 (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая); дрейф 0x29 вечера 10-08 на свежей загрузке НЕ воспроизвёлся — подпись 10-06 |
| word-a 09:47:22 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle (первый заезд wezterm)` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d187d0bfd180d0b8d0b2d0b5d1820a — в поле xghbdtn: ЛИШНИЙ байт 0x78 («x») перед словом; журнал демона: correction done changed=true level=2 runes=7 source=xghbdtn result=чпривет latency_ms=174 — коррекция исполнена ТЕРМИНАЛЬНОЙ веткой на фактическом содержимом поля | Verdict: FAIL (pty-дельта не байт-точна) — инъекционный мусор (префиксный байт до набора; класс наблюдения repin-mechanics 10-08 «префиксный мусор в pty-потоке»), НЕ маршрутизация и НЕ доставка round-2: ни одного verify-timeout/verify-mismatch, done+level-2 честный; заезд 2 зелёный |
| term-word-wezterm 09:47:23 | `wezterm word round` | activation: focus-detected; case aborted: capture delta d187d0bfd180d0b8d0b2d0b5d1820a, want exactly d0bfd180d0b8d0b2d0b5d1820a | Verdict: aborted (следствие ряда word-a выше; WR-04-замечание: отказ НЕ на шаге нуджа — гейты фокуса прошли сами) |
| caps re-pin 09:48:36 | `SetCapabilities в журнале демона для контекста wezterm-nightly (заезд 2)` | activation: focus-detected; venue: vm (kvm); caps=0x9 | Verdict: подпись стабильна между заездами |
| word-a 09:48:38 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) — **round-2 G-10-1 re-pin** |
| word-b 09:48:40 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (второй зелёный подряд — дисциплина ×2 Pitfall 1) — **round-2 G-10-1 re-pin** |
| mid-line 09:48:42 | `armGate + type echo ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass — **round-2 G-10-1 re-pin** |
| clip-frozen 09:48:45 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); learning=false; clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика) | Verdict: pass — **round-2 G-10-1 re-pin** |
| leg-probe 09:48:46 | `sentinel → goswitchctl probe-forward Home, shift+End → wl-paste --primary` | activation: focus-detected; venue: vm (kvm); sentinel=GS-SPIKE-SENTINEL; primary_after=GS-SPIKE-SENTINEL; token_hex=67686264746e | Verdict: leg-probe red (замороженное состояние — лег не подключён; граница D7 задокументирована) |
| witness-vim 09:49:13 | `vim insert-mode + type ghbdtn␠ + esc + :wq (counters + buffer-file oracle)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; buffer_hex=67686264746e200a | Verdict: pass (fired=0, terminal-veto воздержания >0, буфер не изменён) — **round-2 G-10-1 re-pin** |
| witness-htop 09:49:16 | `htop + type ghbdtn␠ (counters oracle + process alive)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; htop жив; контент-свидетельство несёт witness-vim | Verdict: pass — **round-2 G-10-1 re-pin** |
| positive-control 09:49:19 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (позитивный контроль в том же классе) — **round-2 G-10-1 re-pin** |
| autocorrect-terminal-silent-repin 09:49:23 | `runAutocorrectTerminalSilent in-process (existing case, sub-run witnesses)` | activation: focus-detected; case error: autocorrect-terminal-silent: 2 correction records — nothing may enter the ladder | Verdict: blocked — до-существующий артефакт общего журнала кейса (прецеденты 18:59:21/19:59:37/21:32:33: sub-run считает correction-записи ПОСЛЕДНЕГО прогона целиком, в них попадают done+verify-after позитивного контроля); вердикт негатива несёт STANDALONE форма (ряд ниже) |
| autocorrect-terminal-silent-repin ~09:51 | `go run ./test/e2e -case autocorrect-terminal-silent на площадке — STANDALONE форма re-pin (свежий журнал кейса)` | activation: focus-detected; venue: vm (kvm); zenity control engaged (daemon focus_in); wezterm negative surface engaged (daemon focus_in); terminal silent — fired=0, zero correction records under both IME delivery modes; exit кейса 0 | Verdict: pass (канонический counters-оракул негатива чист на коде round-2 — прецедент 19:07:31) — **round-2 G-10-1 re-pin** |
| repin-round2 09:44-09:52 | `реплей-гейт ReplayHost0939*/TwoPhaseLevel1* на HEAD (ok, -race) → re-sync зеркала (rsync --delete → go build → unit -debug → is-active) → term-word-vte, term-word-vte-flip, term-word-wezterm (2 заезда), term-negative-witnesses, autocorrect-terminal-silent STANDALONE` | все пять прогонов зелёные; honest-ряды отказов заездов-1 (setxkbmap-env, stray-x) записаны выше дословно; mise run ci на HEAD — зелёный; go.mod/go.sum нетронуты | Verdict: жёсткое ограничение разрыва удержано на коде round-2 — term-word-vte, term-word-vte-flip, term-word-wezterm, term-negative-witnesses байт-точные зелёные; терминальная ветка и delivery-порядок round-2 не тронули замороженные пины; реплей-гейт исполнен ДО стенда и установки (урок 2026-10-12) |
| host-install-gap-round2 2026-10-14 12:55 MSK | `go build goswitchd+goswitchctl из рабочего дерева (HEAD 2fbfee1 = v1.1.1-142-g2fbfee1, mise run ci зелёный; round-2 фиксы G-10-1: 10-10 answer-only маршрут на обеих ветках вердикта + WR-01/WR-02, 10-11 двухфазный уровень-1 (delete → подтверждение → commit) + отложенный флип уровня-1 + WR-03 same-context) → НОВЫЙ release-dir v1.1.1-142-g2fbfee1 → goswitchctl install (unit re-point, ibus restart, sources wrapped (goswitch-en, goswitch-ru), engine activated goswitch-en)` — **round-2 G-10-1 closure build** | unit active, ExecStart=$HOME/.local/share/goswitch/releases/v1.1.1-142-g2fbfee1/goswitchd (старт 12:55:53 MSK), journal: connected → ctl service listening → component registered (engines:2) → engine created — ошибок регистрации нет; status (фикс. путь) парсится: config_valid=true config_error= (пуст), счётчики на нуле (свежий демон — чистая база приёмки): autocorrect_fired=0, ac_role_errors=0; фикс. путь ctl для ВСЕХ машинных проверок round-2 (не на PATH, новейший releases по mtime): /home/nil/.local/share/goswitch/releases/v1.1.1-142-g2fbfee1/goswitchctl; откат по запросу владельца (назван ДО приёмки): ~/.local/share/goswitch/releases/v1.1.1/goswitchctl install (релизный каталог v1.1.1 нетронут; прежний release-dir v1.1.1-124-g7ace3fd не перезаписывался) | Verdict: установлено — машинная часть зелёная (реплей-гейт и стенд-repin round-2 выше); re-acceptance критерия 1 (шесть пунктов round-2) — за владельцем, human-check: round-2 блок Closure validation в DEBUG-term-host-live-failures.md |
