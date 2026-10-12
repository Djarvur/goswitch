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





