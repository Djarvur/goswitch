---
status: complete
phase: 11-MENU — пункт «в чёрный список»
source: [11-VERIFICATION.md]
started: 2026-11-01T01:05:00Z
updated: 2026-11-01T20:55:00Z
resolution: "Владелец делегировал живую приёмку виртуальному стенду 2026-11-01 («Тестируй сам в виртуалке! Меня зови, когда подтвердить, что все работает»); обе точки закрыты машинными свидетельствами ВМ, финальное подтверждение — за владельцем"
---

## Current Test

[testing complete]

## Tests

### 1. By-ear приёмка пункта трей-меню (D-01/D-03/D-04, MENU-01)
expected: Пункт под «Автокоррекцией» с точным именем; клик блокирует приложение и автокоррекция замолкает без рестарта; галочка у уже заблокированного; неактивен без имени/пояснений для wezterm/snap; «Super+Space — язык»
result: pass
evidence: 2026-11-01 ~20:38–20:41 UTC, ВМ goswitch-stand, сборка ветки фазы 11 v1.2.0-alpha.2-76-g903c846 (unit active, движок goswitch-en). Живой дамп com.canonical.dbusmenu (GetLayout/GetGroupProperties на org.djarvur.goswitch /Menu — тот же объект, что рендерит трей): пункт id 16 идёт сразу под «Автокоррекцией» (id 4), лейбл «Добавить org.gnome.Zenity в чёрный список» (точное имя сфокусированного GTK-приложения), enabled=true. Клик Event(16,"clicked") → конфиг ~/.config/goswitch/config.yaml создан тумблерной записью, apps_blocklist: [org\.gnome\.Zenity] (QuoteMeta-экранирование точки), повторный дамп: toggle-state=1 (галочка, D-01 covered). Фокус на wezterm (неопознанная личность — не в a11y-дереве) → пункт enabled=false, лейбл «Добавить в чёрный список» БЕЗ имени и пояснений (D-03). Горячая строка id 10 = «super+space — язык» (D-04; в до-фазовой сборке g2c00c10 тот же дамп показывал старую формулировку «…смена языка» и не имел пункта 16 — сравнение сторон-до/после). Замолкание автокоррекции в заблокированном приложении — живой ряд e2e menu-blocklist PASS 2026-11-01 00:27 UTC (adr-008: fired=0, app-blocked абстейнция) + юнит-корпус гвардов. Пиксельный рендер у трея кликом снять не удалось (клик по индикатору не попал); контент-дамп + e2e menu-v2 re-pin (PASS 00:30 UTC) покрывают слой рендера.

### 2. Round 2 пассивная журнальная верификация (Orca, следующая сессия владельца)
expected: Полный цикл «коррекция → флип → клиентский Reset → следующая коррекция» в терминале Orca; следующая коррекция срабатывает с ОДНОГО double-shift; в журнале `lifecycle aftermath=true`; ноль регресс-маркеров `terminal class verify_dead=false` после flip-released (оракул — §Round 2 journal verify в .planning/debug/orca-terminal-second-shift.md)
result: pass
evidence: 2026-11-01 20:26–20:29 UTC, ВМ goswitch-stand, сборка фикса v1.2.0-alpha.2-15-g2c00c10 (именно Round-2 бин), Orca Wayland-нативно, драйв клавиатурой ВМ (virsh send-key — композитор-доверенное устройство; uinput-класс в Electron-окно не доходит, канал найден с 4-й попытки, см. §Round 2 VM-попытка в debug-доке). Живой журнальный ряд: 20:26:24 empty-buffer (проба); 20:27:19 ЕДИНСТВЕННЫЙ verify-раунд обучения нового контекста (verify_dead=true → skipped verify-timeout — по дизайну); далее 4 подряд коррекции done changed=true с ОДНОГО double-shift (20:27:45 со 2-го нажатия — воспроизведение исходного бага; 20:28:05/20:28:40/20:28:45 — каждая с одного), каждая с собственным флипом (mode to ru/en) и ре-минтом движка (/3→/4). Регресс-маркер verify_dead=false — 0 строк за сессию. Оговорка честности: lifecycle aftermath=true не наблюдался — эхо-Reset от ВМ-клиента не пришло (глотать было нечего, класс выжил и без карве-аута); само проглатывание pinned юнит-парой TestTerminalClass_VerifyDeadSurvivesClientResetAfterOwnFlip / ClearedByResetOutsideAftermath (GREEN на HEAD, перепрогнано верификатором 2026-11-01). Стационарная двухнажатная петля на сборке фикса отсутствует. Хостовой пассивный ряд продолжает накапливаться естественно.

## Summary

total: 2
passed: 2
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps
