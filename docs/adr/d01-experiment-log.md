# D-01: журнал эксперимента — может ли демон приказать GNOME переключить активный источник ввода?

**Решение D-01 (CONTEXT фазы 01):** механизм переключения раскладки выбирается
экспериментом до любого кода коррекции/переключения. Проверочный вопрос:
«может ли демон приказать системе GNOME переключить активный источник ввода?»
Да → two-engine (два источника goswitch-en/goswitch-ru с XKB-раскладками us/ru;
нативный значок, Super+Space, MRU). Нет → внутренний флип — вариант B (один
движок с `<layout>us</layout>`, EN/RU — внутренний режим, паттерн работающего
прототипа владельца).

**Критерий доказательства — только НАБЛЮДАЕМОЕ (Pitfall 3):** пара
FocusOut(goswitch-en) → FocusIn(goswitch-ru) в логе демона плюс скрипт
последующего ввода (ghbdtn → привет = кириллица). Значение dconf само по себе
доказательством НЕ является: keyboard.js 46 не синхронизирует внешние изменения.

**Прогоны:** строки дописывает кейс `mise run e2e-d01` (test/e2e -case
d01-probe) — каждый прогон добавляет построчные вердикты. Колонка «Наблюдаемое»
содержит наблюдения, а не dconf-значения. Вердикты: `switched` /
`not-switched` / `unavailable`.

| Проба | Команда | Наблюдаемое | Вердикт (Verdict) |
|---|---|---|---|
| 1. gsettings 00:34:04 | `gsettings set org.gnome.desktop.input-sources sources/current → goswitch-ru` | focus_out(en)+0, focus_in(ru)+0; typed "ghbdtn" -> "ghbdtn"; no new focus_in(goswitch-ru) within 6s; | Verdict: not-switched |
| 1. gsettings 00:43:45 | `gsettings set org.gnome.desktop.input-sources sources/current → goswitch-ru` | focus_out(en)+0, focus_in(ru)+0; typed "ghbdtn" -> "ghbdtn"; no new focus_in(goswitch-ru) within 6s; | Verdict: not-switched |
| 2a. ydotool super+space 00:44:04 | `ydotool key super+space` | focus_out(en)+1, focus_in(ru)+0; typed "ghbdtn" -> ""; no new focus_in(goswitch-ru) within 6s; | Verdict: not-switched |
| 2b. ydotool alt+shift_l 00:44:20 | `ydotool key alt+Shift_L` | focus_out(en)+0, focus_in(ru)+0; typed "ghbdtn" -> "ghbdtn"; no new focus_in(goswitch-ru) within 6s; | Verdict: not-switched |
| 3. set-global-engine 00:44:31 | `ibus engine goswitch-ru` | focus_out(en)+1, focus_in(ru)+1; typed "ghbdtn" -> "ghbdtn" | Verdict: not-switched |
| 4. shell eval 00:44:47 | `gdbus call org.gnome.Shell.Eval 'Main.getInputSourceManager().inputSources[1].activate()'` | focus_out(en)+0, focus_in(ru)+0; typed "ghbdtn" -> "ghbdtn"; reply "(false, '')" err <nil> | Verdict: unavailable — Shell.Eval отклонил вычисление (unsafe mode выключен; включать запрещено — D-02/T-04-01) |
| 5. micro-extension 00:44:47 | `mkdir ~/.local/share/gnome-shell/extensions/goswitch-probe@localhost (~30 lines JS)` | extension written to /home/nil/.local/share/gnome-shell/extensions/goswitch-probe@localhost; running shell ignores new extensions without a restart (gnome-extensions list shows goswitch-probe: false); enabling requires a GNOME Shell restart = re-login (Wayland) | Verdict: unavailable — требует ре-логина сессии (не выполнен) |
| Result 00:44:47 | `—` | switched probes: false | DECISION: Option B — internal flip (the owner prototype pattern; see ADR-001) |
| 1. gsettings 00:45:38 | `gsettings set org.gnome.desktop.input-sources sources/current → goswitch-ru` | focus_out(en)+0, focus_in(ru)+0; typed "ghbdtn" -> "ghbdtn"; no new focus_in(goswitch-ru) within 6s; | Verdict: not-switched |
| 2a. ydotool super+space 00:45:58 | `ydotool key super+space` | focus_out(en)+1, focus_in(ru)+0; typed "ghbdtn" -> ""; no new focus_in(goswitch-ru) within 6s; | Verdict: not-switched |
| 2b. ydotool alt+shift_l 00:46:15 | `ydotool key alt+Shift_L` | focus_out(en)+0, focus_in(ru)+0; typed "ghbdtn" -> "ghbdtn"; no new focus_in(goswitch-ru) within 6s; | Verdict: not-switched |
| 3. set-global-engine 00:46:26 | `ibus engine goswitch-ru` | focus_out(en)+1, focus_in(ru)+1; typed "ghbdtn" -> "ghbdtn" | Verdict: not-switched |
| 4. shell eval 00:46:43 | `gdbus call org.gnome.Shell.Eval 'Main.getInputSourceManager().inputSources[1].activate()'` | focus_out(en)+0, focus_in(ru)+0; typed "ghbdtn" -> "ghbdtn"; reply "(false, '')" err <nil> | Verdict: unavailable — Shell.Eval отклонил вычисление (unsafe mode выключен; включать запрещено — D-02/T-04-01) |
| 5. micro-extension 00:46:43 | `mkdir ~/.local/share/gnome-shell/extensions/goswitch-probe@localhost (~30 lines JS)` | extension written to /home/nil/.local/share/gnome-shell/extensions/goswitch-probe@localhost; running shell ignores new extensions without a restart (gnome-extensions list shows goswitch-probe: false); enabling requires a GNOME Shell restart = re-login (Wayland) | Verdict: unavailable — требует ре-логина сессии (не выполнен) |
| Result 00:46:43 | `—` | switched probes: false | DECISION: Option B — internal flip (the owner prototype pattern; see ADR-001) |
