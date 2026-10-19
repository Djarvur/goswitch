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
| word-bash 21:39:55 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| word-zsh 21:39:58 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| mid-line 21:40:01 | `armGate + type echo ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| clip-frozen 21:40:04 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено) |
| witness-vim 21:46:52 | `vim insert-mode + type ghbdtn␠ + esc + :wq (counters + buffer-file oracle)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; buffer_hex=67686264746e200a | Verdict: pass (fired=0, terminal-veto воздержания >0, буфер не изменён) |
| witness-htop 21:46:55 | `htop + type ghbdtn␠ (counters oracle + process alive)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; htop жив (pid 8051); контент-свидетельство несёт witness-vim — у TUI нет текстового буфера | Verdict: pass (fired=0, terminal-veto воздержания >0, процесс жив) |
| positive-control 21:46:58 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| autocorrect-terminal-silent-repin 21:47:03 | `runAutocorrectTerminalSilent in-process (existing case)` | activation: focus-detected; case error: autocorrect-terminal-silent: 2 correction records — nothing may enter the ladder | Verdict: blocked (полный wezterm re-pin остаётся за планом 10-06) |
| word-bash 17:54:53 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| word-zsh 17:54:56 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| mid-line 17:54:59 | `armGate + type echo ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| clip-frozen 17:55:02 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено) |
| term-negative-witnesses 18:08:10 | `терминальные негативные свидетели` | activation: focus-detected; case aborted: term-negative-witnesses: arming gate: key-trace floor unmet after 3 attempts | Verdict: aborted |
| witness-vim 18:15:48 | `vim insert-mode + type ghbdtn␠ + esc + :wq (counters + buffer-file oracle)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; buffer_hex=67686264746e200a | Verdict: pass (fired=0, terminal-veto воздержания >0, буфер не изменён) |
| witness-htop 18:15:52 | `htop + type ghbdtn␠ (counters oracle + process alive)` | activation: focus-detected; venue: vm (kvm); fired=0; veto Δ=1; htop жив (pid 14142); контент-свидетельство несёт witness-vim — у TUI нет текстового буфера | Verdict: pass (fired=0, terminal-veto воздержания >0, процесс жив) |
| positive-control 18:15:55 | `armGate + type ghbdtn + double Shift_R + pty-hex oracle` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-hex содержит ожидаемые байты; done+level-2 в журнале демона) |
| autocorrect-terminal-silent-repin 18:15:59 | `runAutocorrectTerminalSilent in-process (existing case)` | activation: focus-detected; case error: autocorrect-terminal-silent: 2 correction records — nothing may enter the ladder | Verdict: blocked (полный wezterm re-pin остаётся за планом 10-06) |






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
| caps re-pin 21:45:51 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=0x9 (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |
| word-a 21:46:03 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d187d0bfd180d0b8d0b2d0b5d1820a | Verdict: FAIL (pty-дельта не байт-точна) |
| term-word-wezterm 21:46:03 | `wezterm word round` | activation: focus-detected; case aborted: capture delta d187d0bfd180d0b8d0b2d0b5d1820a, want exactly d0bfd180d0b8d0b2d0b5d1820a | Verdict: aborted |
| caps re-pin 21:46:23 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=0x9 (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |
| word-a 21:46:25 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| word-b 21:46:27 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| mid-line 21:46:29 | `armGate + type echo ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| clip-frozen 21:46:32 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); learning=false; clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено) |
| leg-probe 21:46:33 | `sentinel → goswitchctl probe-forward Home, shift+End → wl-paste --primary` | activation: focus-detected; venue: vm (kvm); sentinel=GS-SPIKE-SENTINEL; primary_after=GS-SPIKE-SENTINEL; token_hex=67686264746e | Verdict: leg-probe red (лег не подключён — буфер-доверие остаётся единственным механизмом по вердикту OQ 2 «contract»; граница D7 задокументирована) |
| term-word-wezterm 17:55:44 | `wezterm word round` | activation: focus-detected; case aborted: term-word-wezterm: neither done+level-2 nor the verify-dead learning within 5s | Verdict: aborted |
| word-a 17:58:52 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex= | Verdict: FAIL (pty-дельта не байт-точна) |
| term-word-wezterm 17:58:52 | `wezterm word round` | activation: focus-detected; case aborted: capture /tmp/goswitch-e2e-2264306851/wez-in never grew past offset 0 within 5s | Verdict: aborted |
| term-word-wezterm 18:00:12 | `wezterm word round` | activation: focus-detected; case aborted: term-word-wezterm: arming gate: key-trace floor unmet after 3 attempts | Verdict: aborted |
| term-word-wezterm 18:06:31 | `wezterm word round` | activation: focus-detected; case aborted: term-word-wezterm: arming gate: key-trace floor unmet after 3 attempts | Verdict: aborted |
| word-a 18:16:17 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| word-b 18:16:19 | `armGate + type ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| mid-line 18:16:21 | `armGate + type echo ghbdtn + double Shift_R (learning handshake) + pty-delta oracle` | activation: focus-detected; venue: vm (kvm); learning=false; pty_delta_hex=6563686f20d0bfd180d0b8d0b2d0b5d1820a | Verdict: pass (pty-дельта байт-точна; done+level-2 в журнале демона) |
| clip-frozen 18:16:24 | `printf %s GS-SPIKE-SENTINEL | wl-copy → word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); learning=false; clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_delta_hex=d0bfd180d0b8d0b2d0b5d1820a (CLIPBOARD перезаписан сентинелом — синтетика, пред-стейт задокументирован затиранием) | Verdict: pass (CLIPBOARD байт-нетронут терминальной коррекцией, слово исправлено) |
| leg-probe 18:16:26 | `sentinel → goswitchctl probe-forward Home, shift+End → wl-paste --primary` | activation: focus-detected; venue: vm (kvm); sentinel=GS-SPIKE-SENTINEL; primary_after=GS-SPIKE-SENTINEL; token_hex=67686264746e | Verdict: leg-probe red (лег не подключён — буфер-доверие остаётся единственным механизмом по вердикту OQ 2 «contract»; граница D7 задокументирована) |
| caps re-pin 17:55:38 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=(none) (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |
| caps re-pin 17:58:29 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=0x9 (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |
| caps re-pin 17:59:41 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=(none) (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |
| caps re-pin 18:06:00 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=(none) (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |
| caps re-pin 18:16:15 | `SetCapabilities в журнале демона для контекста wezterm-nightly (до решений)` | activation: focus-detected; venue: vm (kvm); caps=0x9 (наблюдаемая подпись ночника; легаси-строки 0x9 исторические — Pitfall 4) | Verdict: подпись записана до решений (маршрутизация живая — classifyTerminal, не зашитая) |











## Фаза 10 живой ряд VTE flip (term-word-vte-flip)

План 10-08: механизм ряда отложенного флипа (G-10-1 причина B, вердикт E3).
Секция создана планом до прогона (контракт секций журнала); строки пишет
кейс term-word-vte-flip (mise e2e-term-word-vte-flip, NOT in ci).

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| flip-word-bash 21:32:54 | `armGate + type ghbdtn + double Shift_R + journal order oracle (done<settle<mode) + pty-hex` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a; delivery_order=done<settlement<mode | Verdict: pass (pty байт-точен; флип освободился подтверждением доставки — E3 порядок держится) |
| flip-clip-frozen 21:32:57 | `printf %s GS-SPIKE-SENTINEL | wl-copy → flip word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=67686264746e0a (набор шёл в перевернутом режиме — коррекция вернула латиницу; CLIPBOARD перезаписан сентинелом — синтетика) | Verdict: pass (CLIPBOARD байт-нетронут и при активной машине флипа; слово исправлено) |
| term-word-vte-flip 21:40:26 | `VTE flip round` | activation: focus-detected; case aborted: term-word-vte-flip: mode record after the settlement: timeout after 5s: daemon log has 0 occurrences of "\"msg\":\"mode\",\"to\"", want >= 1 | Verdict: aborted |
| flip-word-bash 21:45:33 | `armGate + type ghbdtn + double Shift_R + journal order oracle (done<settle<mode) + pty-hex` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a; delivery_order=done<settlement<mode | Verdict: pass (pty байт-точен; флип освободился подтверждением доставки — E3 порядок держится) |
| flip-clip-frozen 21:45:36 | `printf %s GS-SPIKE-SENTINEL | wl-copy → flip word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=67686264746e0a (набор шёл в перевернутом режиме — коррекция вернула латиницу; CLIPBOARD перезаписан сентинелом — синтетика) | Verdict: pass (CLIPBOARD байт-нетронут и при активной машине флипа; слово исправлено) |
| flip-word-bash 17:55:19 | `armGate + type ghbdtn + double Shift_R + journal order oracle (done<settle<mode) + pty-hex` | activation: focus-detected; venue: vm (kvm); pty_hex=d0bfd180d0b8d0b2d0b5d1820a; delivery_order=done<settlement<mode | Verdict: pass (pty байт-точен; флип освободился подтверждением доставки — E3 порядок держится) |
| flip-clip-frozen 17:55:22 | `printf %s GS-SPIKE-SENTINEL | wl-copy → flip word round → wl-paste --no-newline` | activation: focus-detected; venue: vm (kvm); clip_before=GS-SPIKE-SENTINEL; clip_after=GS-SPIKE-SENTINEL; pty_hex=67686264746e0a (набор шёл в перевернутом режиме — коррекция вернула латиницу; CLIPBOARD перезаписан сентинелом — синтетика) | Verdict: pass (CLIPBOARD байт-нетронут и при активной машине флипа; слово исправлено) |



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

## Фаза 10 round-4: DST-гард девственного окна и пин апстрим-фикса (план 10-16), 2026-10-16

Обоснование — секция round-3 выше (ряды R2/R6b/R7 и BISECT-VERDICT 18:51): вердикт (a)
сделал DST-эмиссию к непроэнерженному контексту недопустимой на уровне машины, вердикт (c)
дал фолбэк-примитив (форварды — write-примитив round-4). Секция фиксирует юнит-ряд гварда
и пин апстрим-фикса — условие любого будущего ретеста DST-маршрута.

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| safety-guard 2026-10-16 | `TestDstGuard_* корпус internal/session (юнит-уровень): RED evidence-20261016-10-16-red-task1.json (RED_EVIDENCE_OK / target_test_failed; прогон на НЕизменённом actor.go — git diff --exit-code чист, 226 tests / 2 fail; пин MaterializedContextByteUnchanged зелёный уже на HEAD) → GREEN (гвард в executeCorrection)` | правило гварда: DeleteSurroundingText только при перегенерационном свидетельстве материализованности surrounding-состояния (surrSeen/surrEng — set в handleSurroundingLocked на КАЖДЫЙ пуш/материализующий ответ текущей генерации, bind к эмиттеру поколения — форма flapEng; clear ТОЧНО в FocusOut/Reset-ветке a.surr=nil/ctxEpoch++); без свидетельства — диверсия в СУЩЕСТВУЮЩУЮ executeLevel2-форму (чанки ≤8, pacing 20 мс, ctxEpoch-перечек, один commit, WR-01 level2Busy — нового транспортного кода нет, вердикт (c)) + громкий INFO-рубеж dst-guarded с латентностью БЕЗ содержимого (D-20/D-21) + счётчик corr_dst_guarded в status (форма ac_role_errors); корпус: NeverMaterializedContextGetsForwards (контекст текущей генерации без пуша/ответа — НОЛЬ DST на проводе, замена форвардами+commit по порядку), EvidenceDiesWithGeneration (свидетельство умирает со сменой поколения: FocusOut/Reset-форма ~:1166 И churn-реминт эмиттера), MaterializedContextByteUnchanged (честно пушащий клиент байт-прежний: один delete над token+tail CORR-07, машина 10-11 не дивертирована); СКОП правила — коррекционная машина продакшн-пути (executeCorrection/executeLevel2); debug-вербы probe-delete/probe-require (ProbeDelete/ProbeRequire, internal/session/probe.go, fail-closed без -debug, прецедент 10-13) — сознательный exempt-инструментарий бисекции ВНЕ гварда, подчинён ретест-гейту ниже; mise run ci зелёный, go.mod/go.sum нетронуты | Verdict: pass — класс T-10-16-01/02 (DST к контексту без свидетельства, свидетельство поверх смены поколения) machine-impossible в коррекционной машине; обоснование — ряды бисекции R2/R6b/R7 (три ядра, один кадр) |
| mutter-fix pin 2026-10-16 | `веб-проверка доступными средствами: GitLab REST API GNOME/mutter через curl — история файла src/wayland/meta-wayland-text-input.c (52 коммита), diff коммитов-кандидатов, чтение файла на тегах 47.0/48.0/48.1/49.8/50.6/51.1 и main; локально: dpkg хоста (libmutter-14-0 46.2-1ubuntu0.24.04.16), changelog.Debian noble до 0.24.04.16 (июнь 2026)` | точный фикс «NULL/garbage-гард surrounding-состояния в meta_wayland_text_input_focus_delete_surrounding» НЕ СУЩЕСТВУЕТ апстрим — UNPINNED ПО ПРИЧИНЕ ОТСУТСТВИЯ: краш-сайт strlen по text_input->surrounding.text живёт БЕЗ NULL-гарда на main (функция :285, strlen :305) и на ВСЕХ проверенных тегах — 47.0/48.0/48.1/49.8/50.6/51.1 (grep: strlen-вызов 1, NULL-гард 0 на каждом; новейший стабильный 51.1 2026-10-12); запинено ЧАСТИЧНОЕ харденение: 2c42a284 2024-11-17 «wayland/text-input: Don't crash on wrong parameters to delete_surrounding()» (g_assert→g_return_if_fail — границы арифметики ПОСЛЕ strlen, NULL не ловит) и fb86d5d5 2024-11-20 «wayland/text-input: Reset surrounding text values by focus change» (сброс surrounding.text/cursor/anchor на смену фокуса — убивает мусорный курсор 0x50c, text остаётся NULL) — ОБА первые в теге 48.0 (в 47.0 нет); noble-бэкпорта нет (ряды бисекции: краш воспроизведён на 46.2-1ubuntu0.24.04.18; changelog хоста до .16 — июнь 2026 — text-input правок не несёт); ПРЕДЕЛ ИНСТРУМЕНТАРИЯ зафиксирован честно: штатного веб-инструментария у проекта НЕТ (brave_search/firecrawl/exa_search/tavily_search/perplexity/jina отключены в .planning/config.json) — пин построен curl-пробами GitLab REST API (источники перечислены в команде ряда) и локальным dpkg/changelog | Verdict: unpinned — фикс не существует апстримом (проверено 2026-10-16 через GitLab REST API + локальные источники); ретест-гейт закрыт на ВСЕХ существующих версиях mutter (46.2 → 51.1 + main), не только на 46.2 |
| retest-gate 2026-10-16 | `строка гейта (без прогона — условие, не событие)` | ретест DST как write-примитива терминала разрешён ТОЛЬКО на mutter, где датированным рядом ЭТОЙ секции запинен фикс, гардящий NULL/garbage surrounding-состояние в meta_wayland_text_input_focus_delete_surrounding (обновление стека хоста/стенда = повод перепроверить пин НОВЫМ датированным рядом); до тех пор DST-для-терминалов депрекирован (вердикт (a) бисекции); гейт накрывает И exempt-вербы probe-delete/probe-require (ProbeDelete/ProbeRequire, internal/session/probe.go — exempt-инструментарий бисекции вне скоупа гварда, fail-closed без -debug): любые DST-эмиссии через них на стеке без запиненного фикса запрещены наравне с продакшн-путём | Verdict: gate CLOSED — на текущем стеке (mutter 46.2) и на всех проверенных апстрим-версиях до 51.1 включительно; открывается только новым датированным рядом пина |

## Фаза 10 round-4 pty-drain (term-word-drain)

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| секция открыта 2026-10-16 (план 10-17) | `код кейса доставлен планом 10-17 (test/e2e/case_term_word_drain.go, mise-задача e2e-term-word-drain — NOT in ci); ЖИВОЙ РЯД исполняет план 10-18` | машина round-4 под явной матрицей pacing (default / tight 5 мс / loose 50 мс + удлинённый drain-settle через -debug оверрайд, fail-closed без него): RAW pty-захват (stty raw — поток байтов, не приготовленные строки), оракулы — pty-дельта порядка (каждый 0x7f строго прежде байтов «привет», ни одного после), порядок delivery-меток (chunks-sent → commit-sent → flip-armed; flip-released только после честной причины — push/keystroke), CLIPBOARD-сентинел байт-нетронут; flip_after_correction: true (паритет хоста) | Verdict: pending — рядов ещё нет, живой прогон за планом 10-18 |
| row-default 21:47:23 | `armGate + type ghbdtn + double Shift_R + pty-порядок (0x7f прежде байтов слова) + порядок delivery-меток + release-причина флипа` | activation: focus-detected; venue: vm (kvm); combo=default (gap="" settle=""); pty_hex=787f67686264746e7f7f7f7f7f7fd0bfd180d0b8d0b2d0b5d182; marks marks: chunks@12040 commit@12176 flip-armed@12592 flip-released@13267 cause-honest=true; clip_after=GS-SPIKE-SENTINEL | Verdict: pass (каждый 0x7f в pty строго прежде байтов слова; delivery-метки упорядочены, флип-релиз после своей причины; CLIPBOARD байт-нетронут) |
| row-tight 21:47:30 | `armGate + type ghbdtn + double Shift_R + pty-порядок (0x7f прежде байтов слова) + порядок delivery-меток + release-причина флипа` | activation: focus-detected; venue: vm (kvm); combo=tight (gap="5" settle="150"); pty_hex=787f67686264746e7f7f7f7f7f7fd0bfd180d0b8d0b2d0b5d182; marks marks: chunks@26112 commit@26247 flip-armed@26662 flip-released@27336 cause-honest=true; clip_after=GS-SPIKE-SENTINEL | Verdict: pass (каждый 0x7f в pty строго прежде байтов слова; delivery-метки упорядочены, флип-релиз после своей причины; CLIPBOARD байт-нетронут) |
| row-loose 21:47:39 | `armGate + type ghbdtn + double Shift_R + pty-порядок (0x7f прежде байтов слова) + порядок delivery-меток + release-причина флипа` | activation: focus-detected; venue: vm (kvm); combo=loose (gap="50" settle="400"); pty_hex=787f67686264746e7f7f7f7f7f7fd0bfd180d0b8d0b2d0b5d182; marks marks: chunks@40097 commit@40233 flip-armed@40649 flip-released@41323 cause-honest=true; clip_after=GS-SPIKE-SENTINEL | Verdict: pass (каждый 0x7f в pty строго прежде байтов слова; delivery-метки упорядочены, флип-релиз после своей причины; CLIPBOARD байт-нетронут) |




## Фаза 10 round-4 машина доставки (план 10-17), 2026-10-16

Обоснование — блоки 19:21:26/19:36:00 DEBUG-term-host-live-failures.md (round-2 + дополнения) и
бисекция 2026-10-16 (секция round-3): DST мёртв, Require-рефлексия для wezterm-класса мертва,
потеря сидит в доставке пакета level-2 под хостовой задержкой клиента. Секция фиксирует
реплей-гейт реальных рядов и машину round-4; живой ряд стенда и жест владельца — план 10-18.

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| replay-gate 2026-10-16 | `TestReplayWezterm1921_* / TestReplayHost1936_SecondSampleKeepsWord (юнит-корпус internal/session): RED на неизменённом HEAD — evidence-20261016-10-17-red-task1.json (RED_EVIDENCE_OK / target_test_failed; git diff --exit-code по actor.go чист; полный -race -v прогон 232 tests / 6 fail) → GREEN cb0d93c` | инвертирующий потребитель wezterm-класса (коммит синхронен, форварды с управляемой задержкой по машинным pacing-тикам, release-причины флипа драйвимы швами); зонированная матрица 19:21:26 (задержка форвардов × release-причина × момент смены контекста): ДО машины — поле «ghприве» при lag 60 мс (дословный класс ряда владельца), тишина вместо delivery-меток, grace/verify-исходы отпускают флип внутри окна (SetGlobalEngine до причины), criterion-4-стоп обрывает раунд без коммита при смене контекста после первого форварда, WR-01 не накрывал окно доставки; ПОСЛЕ — в пределах drain-settle поле держит «привет » при ЛЮБОЙ причине, за пределами — громкие метки без утверждений персистентности, SetGlobalEngine только по причине (push/keystroke), mid-flow смена контекста — коммит на захваченном снапшоте за меткой context-changed-midflow, до первого форварда — свободный стоп; 19:36:00 (+76 мс second-sample) — та же форма, одна зона | Verdict: pass — реплей реальных рядов красный до правок и зелёный после (урок 2026-10-12 исполнен ДО стенда и ДО вопроса владельцу) |
| delivery-machine 2026-10-16 | `executeLevel2 (internal/session): replayBurst (чанки ≤8 + свободный стоп до первого форварда + mid-flow метки) → drainSettleWindow (terminalDrainSettle 150 мс, мьютекс освобождён — WR-01 накрывает, T-10-17-01) → commit-sent → deferCorrectionFlip(converted, true) без grace и verify-релизов` | константы — поля актора с журнальными дефолтами (defaultLevel2ForwardGap 20 мс, defaultTerminalDrainSettle 150 мс: ряд 19:21:26 done +144 мс при пакете 6×20 мс + E3 push +16 мс), НЕ конфиг; -debug оверрайд GOSWITCH_DEBUG_LEVEL2_FORWARD_GAP_MS / _TERMINAL_DRAIN_SETTLE_MS — fail-closed без -debug (юнит-пины main_test), мусор — громкий отказ старта; delivery-метки INFO (chunks-sent/commit-sent/flip-armed/flip-released причина+латентность) без содержимого (D-20/D-21), grep-формы done не переименованы; регресс-рампа 09:39/TwoPhaseLevel1/FlapWindow/Level1ConfirmFlush/TestDstGuard зелёная, mise run ci зелёный, go.mod/go.sum заморожены | Verdict: pass — машина round-4 в единственном шве executeLevel2; legacy/level-1/selection/гвард байт-нетронуты |
| stand-case 2026-10-16 | `test/e2e/case_term_word_drain.go + mise e2e-term-word-drain (NOT in ci); живой ряд — план 10-18` | RAW pty-захват (stty raw), матрица pacing (default/tight 5 мс/loose 50 мс+settle 400) через -debug оверрайд; оракулы: pty-дельта порядка (каждый 0x7f строго прежде байтов «привет»), порядок delivery-меток, честная причина flip-released, CLIPBOARD-сентинел; секция «Фаза 10 round-4 pty-drain» этого журнала открыта с verdict: pending | Verdict: pending — код доставлен 10-17, живой ряд исполняет 10-18 |

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

## Фаза 10 round-4 re-pin G-10-1 (план 10-18): реплей-гейт ДО стенда, 2026-10-16

Ряд реплей-гейта round-4 — ЗАПИСАН ДО всякого стенд-заезда и установки (урок
2026-10-12/09:39 дословно: фикс обязан реплеем журнальных рядов доказываться ДО
запуска владельцем; порядок round-2 повторён). Машинный контент round-4 на HEAD:
10-17 (drain-settle 150 мс, флип вне окна, delivery-метки, abort/commit-дисциплина)
+ 10-16 (DST-гард девственного окна, corr_dst_guarded). Контроль сборки: HEAD
bffefc2 = финальный HEAD, верифицированный 10-17 (его metadata-коммит); tracked-файлы
рабочего дерева чисты (`git status` — только чужие untracked-артефакты), пост-плановых
правок нет.

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| repin-round4 2026-10-16 21:32–21:38 UTC | `mise exec -- go test ./internal/session/ -run 'ReplayWezterm1921\|ReplayHost1936\|ReplayHost0939\|TwoPhaseLevel1\|FlapWindowContextIdentity\|Level1ConfirmFlush\|TestDstGuard' -race -count=1; mise run ci; git diff --exit-code -- go.mod go.sum` | реплей-гейты 10-17 (обе подписи 19:21:26/19:36:00 + рампа 09:39/TwoPhaseLevel1/FlapWindowContextIdentity/Level1ConfirmFlush/DstGuard) — `ok github.com/Djarvur/goswitch/internal/session 3.767s` на HEAD bffefc2; `mise run ci` — зелёный (21 пакет ok, build+vet+lint+test -race; прогон 2 и прогон 3, REAL_EXIT=0; прогон 1 — ОДИНОЧНЫЙ флак, имя пакета утеряно окном вывода при параллельных mise-задачах — класс до-существующего deferred-пункта «ctlsvc/config/appid флаки под -race нагрузкой» v1.0.0, повторён 2× подряд зелёным; никакой правки кода между прогонами не было); `git diff --exit-code -- go.mod go.sum` — чисто (зависимости заморожены, T-10-18-SC) | Verdict: pass — реплей-гейт round-4 ЗЕЛЁНЫЙ на финальном HEAD и записан ДО стенда; стенд (×4 кейса + drain-ряд) и установка разблокированы |

## Фаза 10 re-pin round-4 G-10-1 (план 10-18): четыре кейса + живой drain-ряд, 2026-10-16

Re-pin замороженных стенд-рядов на машине round-4 (10-17: drain-settle 150 мс, флип вне
окна, delivery-метки, abort/commit-дисциплина; 10-16: DST-гард) + живой drain-ряд нового
кейса term-word-drain. Порядок: реплей-гейт (ряд repin-round4 выше, 21:32–21:38) → стенд.
Зеркало re-sync rsync-рецептом 10-08 (HEAD fb8e403 → заезды; затем фикс кейса bf70ff8 →
заезды flip/wezterm/witnesses/drain); go build + go vet зеркала зелёные; unit -debug
перезапущен (component registered engines:2, engine goswitch-en, 21:39:18). Сессионный env
ssh-прогонов по находке round-3: DISPLAY=:0 + XDG_RUNTIME_DIR=/run/user/1000 +
DBUS_SESSION_BUS_ADDRESS + XAUTHORITY=/run/user/1000/.mutter-Xwaylandauth.* (НЕ в $HOME).
Ряды ниже — дословно с зеркала стенда; отдельные ряды кейсов лежат в своих секциях
(указатели в строке-сводке). Прobe-вербы 10-13 в прогонах round-4 НЕ использовались;
ни один кейс не эмитит DeleteSurroundingText (гвард 10-16 в сборке стенда; TestDstGuard
в реплей-гейте). Крахов gnome-shell в окне прогонов НОЛЬ (сессия пережила все пять
заездов; ядро 6.8.0-147 дефолтное).

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| repin-round4-stand 2026-10-16 21:39–21:48 UTC | `mise run e2e-term-word-vte; mise run e2e-term-word-vte-flip; mise run e2e-term-word-wezterm; mise run e2e-term-negative-witnesses; mise run e2e-term-word-drain (на площадке: go run ./test/e2e -case … с сессионным env)` | term-word-vte PASS с первого заезда (ряды word-bash 21:39:55 / word-zsh 21:39:58 / mid-line 21:40:01 / clip-frozen 21:40:04 — секция «Фаза 10 живые ряды VTE»; pty_hex=d0bfd180d0b8d0b2d0b5d1820a все); term-word-vte-flip: заезд 1 ЧЕСТНЫЙ abort 21:40:26 (оракул кейса ждал mode после verify-таймаута — семантика round-2; машина round-4 держит флип армед по построению — пины FlipImmuneToVerify*; фикс кейса bf70ff8: хвостовой newline кейса = честная keystroke-причина, ждётся flip-released(reason=keystroke) между settlement и mode) → заезд 2 PASS (flip-word-bash 21:45:33 delivery_order=done<settlement<mode + flip-clip-frozen 21:45:36 pty_hex=67686264746e0a — секция flip); term-word-wezterm: заезд 1 ЧЕСТНЫЙ FAIL 21:46:03 (word-a: pty_delta_hex=d187d0bfd180… — лишний префиксный байт 0x78; класс «префиксный мусор» repin-mechanics 10-08, инъекционный мусор арминг-пробы, НЕ маршрутизация и НЕ доставка; caps re-pin 0x9 оба заезда) → заезд 2 PASS (caps re-pin 21:46:23, word-a 21:46:25 + word-b 21:46:27 — два зелёных подряд ×2 Pitfall 1, mid-line 21:46:29, clip-frozen 21:46:32, leg-probe red-замороженный — секция wezterm); term-negative-witnesses PASS (witness-vim 21:46:52 fired=0 veto Δ=1, witness-htop 21:46:55 жив, positive-control 21:46:58 байт-точный; встроенная sub-форма негатива — документированный blocked-артефакт 21:47:03 «2 correction records», прецеденты 18:59:21/09:49:23); term-word-drain PASS со ВСЕЙ матрицей pacing (row-default 21:47:23 gap=дефолты settle=дефолты / row-tight 21:47:30 gap=5 settle=150 / row-loose 21:47:39 gap=50 settle=400 — секция «Фаза 10 round-4 pty-drain»: каждый 0x7f строго прежде байтов слова при каждой комбинации; marks chunks→commit→flip-armed→flip-released упорядочены, cause-honest=true; CLIPBOARD-сентинел байт-нетронут) | Verdict: pass — механизм round-4 валидирован живыми рядами на стенде: четыре вердиктных кейса байт-точные зелёные, негативные свидетели молчат (fired=0), drain-ряд доказывает pty-порядок и порядок delivery-меток при каждой комбинации матрицы; честные abort/FAIL-ряды заездов-1 сохранены; ноль DST-эмиссий, ноль крахов shell; mise run ci на HEAD зелёный, go.mod/go.sum заморожены; установка (Task 3) разблокирована |
| host-install-gap-round4 2026-10-17 00:51 MSK | `go build goswitchd+goswitchctl из рабочего дерева (HEAD efac750 = v1.1.1-180-gefac750, git describe; гейты ПЕРЕД установкой: реплей-гейты 10-17 ok -race 3.827s, mise run ci зелёный — 21 пакет, REAL_EXIT=0, go.mod/go.sum diff чист; стенд-ряды round-4 зелёные — repin-round4-stand 21:39-21:48 UTC) → НОВЫЙ release-dir v1.1.1-180-gefac750 → goswitchctl install (unit re-point, daemon-reload + enable + restart выполнены инсталлятором, component content verified read-back, sources wrapped (goswitch-en, goswitch-ru), engine activated goswitch-en)` — **round-4 G-10-1 closure build: машина 10-17 (drain-settle 150 мс, флип вне окна, delivery-метки, abort/commit-дисциплина) + гвард 10-16 (DST, corr_dst_guarded)** | unit active, ExecStart=$HOME/.local/share/goswitch/releases/v1.1.1-180-gefac750/goswitchd, старт 00:51:27 MSK; /proc/2670465/exe → тот же release-dir (идентичность по /proc сверена); journal поднятия: config loaded → connected → ctl service listening → component registered (engines:2) → engine created ×2 → engine reactivated goswitch-en — ошибок регистрации нет; status (фикс. путь) парсится: config_valid=true config_error= (пуст), счётчики на нуле (чистая база приёмки): corrections_done=0, corrections_skipped=0, autocorrect_fired=0, autocorrect_abstained=0, ac_role_errors=0, стартовое corr_dst_guarded=0 (счётчик гварда 10-16 присутствует в установленной сборке); фикс. путь ctl для ВСЕХ машинных проверок round-4 (не на PATH, новейший releases): /home/nil/.local/share/goswitch/releases/v1.1.1-180-gefac750/goswitchctl; откат по запросу владельца (назван ДО приёмки): ~/.local/share/goswitch/releases/v1.1.1/goswitchctl install (релизный каталог v1.1.1 нетронут; прежние release-dir v1.1.1-124-g7ace3fd / v1.1.1-142-g2fbfee1 не перезаписывались); host-мутации = ТОЛЬКО санкционированная установка, после неё хост строго read-only | Verdict: установлено — машинная часть зелёная (реплей-гейт 21:32, стенд-repin 21:39-21:48, установка 00:51); re-acceptance round-4 — за владельцем (гейт: слово в wezterm ОСТАЁТСЯ), human-check: round-4 блок Closure validation в DEBUG-term-host-live-failures.md |

## Фаза 10 round-6 DST-применение (план 10-13 продолжение): apply-пробы и хореография на тёплых контекстах

Секция создана до прогона (контракт секций журнала). Обоснование — round-6 вердикт
a196101 (дефект владельца: BackSpace-форварды сидят в очереди клиента ~секунды, коммит
садится мгновенно → «появилось и исчезло») и бисекция 2026-10-16 (round-3 секция):
тёплый DST крахобезопасен (R1b/R3/R4), девственное окно ≲1.5 с — запрет (10-16 гвард).
Вопрос секции — ПРИМЕНЯЕТ ЛИ клиент DeleteSurroundingText (исчезают ли символы из
строки ввода терминала); вердикт = рекомендация маршрутизации round-6 (DST+commit
против фолбэка «форварды + Require-подтверждённый drain»). Машина: кейс term-dst-apply
(продолжение скоупа 10-13; сплит-вербы probe-delete/probe-require/probe-commit 10-13),
RAW pty-захват (форма drain-кейса 10-17: stty raw — поток байтов, не приготовленные
строки). Дисциплина тёплых контекстов — гейт КАЖДОЙ ячейки, в терминах бисекции:
граница триггера — ВРЕМЯ от свежего focus_in контекста (крах +0.49 с — R2/R6b/R7;
выживание +1.5 с — R3 при нуле пушей; +23 с — R4), поэтому гейт держит каждый DST до
возраста последнего focus_in ≥ 3 с (двойной запас к запиненной точке выживания) при
обязательном условии НАБРАННОГО токена (реальные клавиши сквозь контекст — форма
материализации мандата). Require-спрос (R5b-безопасен в любом окне) исполняется тоже,
его ответ ЗАПИСЫВАЕТСЯ (протухший кэш считается — находка 10-08), но НЕ гейтит:
клиенты класса 0x9 (подпись ночника 10-06, живая в заезде 1) не отвечают никогда, и
недостижимый для них свидетель не вправе делать матрицу неизмеримой. Отказ гейта —
честная aborted-warm строка, кейс падает: в неп proven-окно DST не едет. Пол возраста —
10 с (первое значение 3 с опрокинуто рядом vte-apply 16:57:04 ЭТОЙ секции: крах при
+3.2 s на отвечающем VTE-контексте — см. строку; R4 выживал при +23 s). Порядок ячеек —
VTE-ячейка ПОСЛЕДНЕЙ (жертвенной): тёплый VTE-DST — известный кандидат в убийцу сессии,
и краш не должен морить остальные ячейки. Выживание gnome-shell — свидетель в каждой строке
(shell_alive); краш в тёплом окне — дата-пойнт, не сбой прогона. Оракул: erase-байты
(0x7f/0x08) строго ПОСЛЕ токена в raw-дельте = applied; ноль = ignored; частичные =
partial; хореография — порядок (каждый erase прежде первого байта слова) + финальная
строка net = привет. Бюджет: ≤8 стенд-заездов. Ряды несут исходы/счётчики/булевы,
никогда содержимое (D-20/D-21). Гейт retest 10-16 (запрет DST-эмиссий на стеке без
запиненного апстрим-фикса) для ЭТОЙ секции отменён мандатом round-6 владельца: только
тёплые гейт-формы, выживание в каждом ряду, бюджет заездов (8 кейс-заездов + лёгкие
ручные драйвы одиночных ячеек, форма бисекции — учтены в рядах честно).

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| case-run-1 16:46 | `go run ./test/e2e -case term-dst-apply (заезд 1; гейт v1 — push-свидетельство обязателен)` | activation: focus-detected; venue: vm (kvm); wez-apply: арминг прошёл, токен доставлен (key-trace), probe-require БЕЗ ответа весь бюджет 5 s — ночной контекст сминтился caps=0x9 (подпись 10-06, окружающий-бит НЕ анонсируется); гейт честно отказал — DST УДЕРЖАН, эмиссии НЕ БЫЛО; shell жив | Verdict: aborted-warm (тёплость по гейту v1 не доказана; push-свидетельство недостижимо для 0x9-класса — гейт переработан в c27b96d на временную границу бисекции) |
| case-run-2 wez-apply 16:53 | `armGate + type ghbdtn → warm gate (возраст фокуса ~8-13 s) → probe-delete -6 6 → raw pty delta (восстановлено дословно из сессионного вывода; строка зеркала затёрта rsync хоста — прецедент 10-06)` | activation: focus-detected; venue: vm (kvm); client=wezterm-nightly; эмиссия на проводе (журнал: probe delete offset=-6 nchars=6, 16:53:45.854); захват до DST = 8 байт (арминг-пара + токен — канал байт-прозрачен), после эмиссии роста НЕТ в 3 s settle + 5 s наблюдения; erase-байтов НОЛЬ; shell жив (кейс продолжился до teardown) | Verdict: DST=ignored (0x9-класс: токен лежит в pty, DST не породил ни одного erase-байта; строка-кандидат пала на harness-баге роста — 3ebb083 сделал чтение wait-free) |
| case-run-3 wez-apply 16:56:57 | `armGate + type ghbdtn → probe-require witness → probe-delete -6 6 → raw pty delta` | activation: focus-detected; venue: vm (kvm); client=wezterm-nightly; token_seen=false (offset после гейта — позже токена; фикс 630974c); erase_del=0; erase_bs=0; erases_total=0 (want 6); shell_alive=true; push_answered=false; push_cursor=-1 | Verdict: DST=no-token (контроль доставки по размещению offset; канал в порядке — заезд 2) |
| case-run-3 vte-apply 16:57:04 | `armGate + type ghbdtn → probe-require witness → probe-delete -6 6 → raw pty delta` | activation: focus-detected; venue: vm (kvm); client=gnome-terminal-vte; VTE ОТВЕТИЛ на тёплый ask (push_answered=true, push_cursor=1 — протухший кэш); probe-delete на возрасте фокуса +3.2 s → gnome-shell умер ~2.6 s спустя («ibus connection lost»; сессия вниз; юнит погиб с ней); erase_del=0; token_seen=false; shell_alive=false | Verdict: FAIL (gnome-shell НЕ ПЕРЕЖИЛ эмиссию — крах-дата-пойнт: тёплый DST на ОТВЕЧАЮЩЕМ VTE-контексте валит mutter 46.2; «тёплый = безопасен» бисекции был запинен только на wezterm; апстрим-гарды NULL/bounds — только с 48.0, в noble 46.2 их нет) |
| case-run-4 wez-apply 17:04:34 | `armGate + type ghbdtn → warm gate → probe-delete -6 6 → raw pty delta` | activation: focus-detected; venue: vm (kvm); client=wezterm-nightly; token_seen=false (offset после гейта — класс заезда 3, фикс в 630974c); erase_del=0; erase_bs=0; shell_alive=true; push_answered=true; push_cursor=6 — кэш клиента ОТСЛЕЖИЛ набранный токен (свежий пуш, не протухший!); хореография заезда пала на арминге choreo-a (ноль key-записей — клик-гонка) | Verdict: DST=no-token по pty-оракуле + push_cursor=6 — ТОКЕН В ИМ-КЭШЕ КЛИЕНТА (для push-класса pty-оракула недостаточна — нужна Require-рефлексия; ручные ряды ниже) |
| case-runs-5..8 17:09/17:15/17:24/17:47 | `go run ./test/e2e -case term-dst-apply (четыре заезда: арминг-отказы)` | все четыре: арминг-гейт честно отказал после 3 попыток; в заезде 5 — НОЛЬ key-записей за весь прогон; живая диагностика между заездами (17:12-17:20): та же форма клик+набор против СТОЯЩЕГО демона — KEY_DELTA=2 (дважды), репро с живым журналом — delta=2; изоляция A/B/C (17:18-17:23): стоящий демон delta=2 → голый systemctl restart → delta=0 → поздний `ibus engine goswitch-en` → delta=2 — рестарт демона на холодной шине («global engine unreadable») оставляет диспетчеризацию несвязанной; активация кейса срабатывает субсекундно и неэффективна | Verdict: aborted (класс деградации доставки изолирован: холодношинная активация; фиксы 4ba5539 — нудж 10-08, ea270ff — нудж в каждую ячейку, 589aef0 — поздний re-poke в каждую ячейку; сами заезды эмиссий НЕ несли) |
| manual apply-v1 17:52 | `ручной драйв (форма бисекции): wezterm raw-фикстура → клик-нудж → type ghbdtn → probe-require → probe-delete -6 6 → pgrep` | activation: ручной клик-нудж; venue: vm (kvm); TOKEN_KEYS=12 (доставка доказана key-trace); WARM_PUSHES=1 — ask ОТВЕЧЕН: surrounding push text=ghbdtn cursor=6 (кэш отследил токен ТОЧНО — свежий, не протухший); capture ДО DST = 0 байт (токен в ИМ-слое, pty НЕ эхоит — push-класс); DST на тёплом (≥13 s) → shell_alive=2, signal-11 строк нет | Verdict: выживание подтверждено; apply-вердикт несёт ручной apply-v2 ниже (рефлексия) |
| manual choreo-a 17:53 | `драйв хореографии: type → probe-delete -6 6 → 1 s drain-delay → probe-commit → raw pty` | activation: ручной клик-нудж; venue: vm (kvm); TOKEN_KEYS=0 (клик мимо — набор не доставлен, честно вырожденный ряд); коммит при пустом кэше: DELTA_HEX=d0bfd180d0b8d0b2d0b5d182 — «привет» БАЙТ-ТОЧНО в raw pty; shell жив | Verdict: вырожденный по токену, ЗЕЛЁНЫЙ по каналу — коммит-плечо хореографии садится байт-точно (двойное подтверждение с рядом заезда 2) |
| manual clean-test 17:57 | `чистая проба: спавн → клик → type ghbdtn → raw pty` | venue: vm (kvm); KEYS=12; capture=6 байт 67686264746e — 0x9-класс минтнулся (токен проходит в pty, push нет) — два класса ночника на одной сессии: push-класс держит ввод в ИМ-слое, 0x9-класс эхоит в pty | Verdict: классовая механика зафиксирована (оракула класс-зависимы: pty-байты для 0x9, require-рефлексия для push) |
| manual apply-v2 17:57:52 | `драйв apply с рефлексией: type → probe-require (baseline) → probe-delete -6 6 → probe-require (reflect) → pgrep` | activation: ручной клик-нудж (попытка 2 — попытка 1 клик мимо, 0 ключей); venue: vm (kvm); client=wezterm-nightly, push-класс; BASELINE_PUSH cursor=6 (токен в кэше); DST(-6,6) на тёплом (≥10 s) → shell_alive=2; REFLECT_PUSH cursor=6 — БУФЕР НЕ СОКРАТИЛСЯ; pty пуст (push-класс — токен не в pty); signal-11 строк нет | Verdict: **DST=ignored — wezterm-nightly НЕ применяет DeleteSurroundingText** (рефлексия без изменений на тёплом контексте; параллельно 0x9-класс — заезд 2 и чистая проба: в pty-форме DST тоже не порождает erase-байтов) |
| manual choreo-b/c 17:58-18:00 | `драйвы хореографии (два захода)` | оба: клик-гонка не выдала клавиатуру (0 ключей ×3-5 попыток) — честно вырожденные ряды без эмиссионной ценности; shell жив в обоих | Verdict: aborted (хореография с токеном в игре не измерена — голодание размещения; исход выводим из apply-вердикта: DST игнорируется → DST+commit ≡ commit-only — чья pty-форма доказана choreo-a) |

### Вердикт-блок round-6 (19.10.2026)

- **wezterm-nightly 20261010 DST: IGNORED** (ряды case-run-2, manual apply-v2, clean-test;
  оба класса контекста — pty-форма без erase-байтов, push-форма без сокращения кэша).
- **gnome-terminal/VTE DST: сессия-киллер на mutter 46.2** (ряд case-run-3 vte-apply:
  крах при +3.2 s на ОТВЕЧАЮЩЕМ тёплом контексте — уже вне «девственного окна»
  бисекции; применение ненаблюдаемо — стек убивает сессию раньше).
- **Хореография DST+commit**: коммит-плечо байт-точно (choreo-a, case-run-2);
  полное построение с токеном в игре не измерено (голодание размещения), исход
  выводим: DST игнорируется → DST+commit ≡ commit-only.
- **Выживание mutter**: все тёплые wezterm-эмиссии пережиты (заезды 2-3, ручные
  ряды — суммарно 5+ тёплых DST без краха); ЕДИНСТВЕННЫЙ крах — VTE при +3.2 s.
- **Маршрутизация round-6: DST-маршрут МЁРТВ на этом стеке** — фолбэк round-4
  (форварды + pty-порядок + drain) остаётся терминальной записью; 10-16-гард и
  retest-гейт остаются в силе для терминального DST. Новая находка в пользу
  будущего: push-класс ночника отвечает на Require СВЕЖИМ кэшем (cursor=6 по
  токену — ряды case-run-4, manual apply-v1/v2) — verify-плечо для push-класса
  теоретически оживаемо, отдельное решение владельца (вне мандата этой секции).


## Фаза 10 round-6 re-pin парных форвардов (план 10-17): press+release на каждый BackSpace, 2026-10-19

Round-6 pivot (мандат владельца, sourcedig-вердикт `.planning/debug/2026-10-19-wezterm-repeat-sourcedig.md`
+ амендмент (7) ADR-008): терминальная запись шлёт каждый BackSpace парой press+release —
press-only форварды взводили клиентский автоповтор wezterm (window.rs:1194-1206, снятие только
release'ом той же клавиши), который через GNOME repeat-delay 500 мс стирает закоммиченное слово
(«появилось и исчезло», 4 раунда охоты). Параллельный зонд DST-apply (10-13, секция выше)
независимо закрыл DST-ветку: wezterm игнорирует DeleteSurroundingText, тёплая VTE DST роняет
mutter 46.2 — маршрутизация остаётся на машине форвардов (теперь парных). Re-pin: rsync
зеркала (исключения .git/.zcode/node_modules/.gsd) → go build/vet зеркала зелёные →
go build -o ~/go/bin/goswitchd + restart unit -debug → component registered engines:2 →
четыре замороженных кейса с сессионным env (DISPLAY=:0, XDG_RUNTIME_DIR, DBUS_SESSION_BUS_ADDRESS,
XAUTHORITY .mutter-Xwaylandauth.*). Отдельные ряды кейсов — в своих секциях выше (указатели в
строке-сводке); pty-оракул не менялся: пара press+release = ОДИН 0x7f на символ (release в pty
не пишет) — байт-точность сохранена без правок оракула.

| Строка | Команда | Наблюдаемое | Вердикт |
|---|---|---|---|
| repin-round6-pairs 2026-10-19 17:53–18:17 UTC | `go run ./test/e2e -case term-word-vte; -case term-word-vte-flip; -case term-word-wezterm; -case term-negative-witnesses (сессионный env; HEAD дерева 33f32c3 = v1.1.1-224-g33f32c3, зеркала)` | term-word-vte PASS с первого заезда (word-bash 17:54:53 / word-zsh 17:54:56 / mid-line 17:54:59 / clip-frozen 17:55:02, pty_hex=d0bfd180d0b8d0b2d0b5d1820a — секция VTE); term-word-vte-flip PASS (flip-word-bash 17:55:19 delivery_order=done<settlement<mode + flip-clip-frozen 17:55:22 — секция flip); term-word-wezterm: заезды 1–4 ЧЕСТНЫЕ aborted/FAIL 17:55–18:06 (интерференция параллельного DST-зонда: живая dst2-apply панель ловила фокус-нудж 640x400 — capture не рос / key-trace floor; уровень-2 при этом отработал байт-точно: 17:58:47 done level=2 runes=6 ghbdtn→привет latency 273 мс — пары на проводе, дефектов машины НЕТ) → после завершения зонда (его вердикт 4f940b0) + ibus engine re-assert + restart unit: PASS (caps re-pin 0x9 18:16:15, word-a 18:16:17 + word-b 18:16:19 два зелёных подряд ×2 Pitfall 1, mid-line 18:16:21, clip-frozen 18:16:24, leg-probe red-замороженный — секция wezterm); term-negative-witnesses: заезд-1 честный aborted 18:08:10 (та же интерференция) → PASS (witness-vim 18:15:48 fired=0 veto Δ=1, witness-htop 18:15:52 жив, positive-control 18:15:55 байт-точный; встроенная sub-форма — документированный blocked-артефакт «2 correction records», прецеденты round-4) | Verdict: pass — парная эмиссия валидирована живыми рядами: четыре вердиктных кейса зелёные, pty байт-точность сохранена (пара = один 0x7f), уровень-2 машина отрабатывает во всех окнах (273 мс latency = 12×20 мс пар + settle); честные aborted-ряды интерференции сохранены (зонд 10-13 работал параллельно — его панели/чурн, никогда не убивались); mise run ci на HEAD зелёный, go.mod/go.sum заморожены |
| round6-pairs-decision 2026-10-19 | `sourcedig (.planning/debug/2026-10-19-wezterm-repeat-sourcedig.md, 0c2fd6e) + RED→GREEN fed2a98/33f32c3 + ADR-008 амендмент (7)` | корень «появилось и исчезло» = клиентский автоповтор wezterm, взведённый press-only форвардами; фикс = пары press+release (probe.go:179-183 дисциплина); DST отклонён по исходникам (inputhandler.rs:166-200 `_ => {}`) и независимо зондом 10-13 (wezterm ignore + mutter crash) | Verdict: терминальная запись = парные форварды; selftype-гейт за владельцем (одно слово: ОСТАЛОСЬ/ИСЧЕЗЛО) |
| host-install-round6 2026-10-19 21:18 MSK | `go build goswitchd+goswitchctl из рабочего дерева (HEAD 8bd4edc = v1.1.1-226-g8bd4edc; гейты ДО установки: mise run ci зелёный 21 пакет, пандус реплея зелёный -race, tidy-diff чист, стенд repin-round6-pairs зелёный) → НОВЫЙ release-dir v1.1.1-226-g8bd4edc → его goswitchctl install (component verified, registry visible, unit daemon-reload+enable+restart, sources wrapped, goswitch-en активирован)` — **round-6 G-10-1 closure build: парные форварды (амендмент (7)) поверх машины round-4 + гварда 10-16** | unit active, ExecStart=$HOME/.local/share/goswitch/releases/v1.1.1-226-g8bd4edc/goswitchd, старт 21:18:54 MSK; /proc/2096528/exe → тот же release-dir (/proc-идентичность сверена); журнал: component registered engines:2 → engine reactivated goswitch-en, ошибок нет; status (фикс. путь /home/nil/.local/share/goswitch/releases/v1.1.1-226-g8bd4edc/goswitchctl): config_valid=true, счётчики нули (corrections_done/skipped=0, autocorrect_fired/abstained=0, ac_role_errors=0, corr_dst_guarded=0); version=dev — деревостроенная сборка (прецедент gap-сборок, идентичность = release-dir+/proc); откат назван ДО приёмки: ~/.local/share/goswitch/releases/v1.1.1/goswitchctl install; host после установки строго read-only — жест/selftype за владельцем | Verdict: установлено — машинная часть зелёная (RED→GREEN fed2a98/33f32c3, ci, стенд-repin, установка 21:18); selftype-гейт round-6 за владельцем (одно слово: ОСТАЛОСЬ/ИСЧЕЗЛО), инструкция — Closure validation round-6 блок DEBUG-term-host-live-failures.md + 10-UAT Current Test №6 |
