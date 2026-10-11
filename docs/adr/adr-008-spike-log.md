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
