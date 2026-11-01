---
status: testing
phase: 11-MENU — пункт «в чёрный список»
source: [11-VERIFICATION.md]
started: 2026-11-01T01:05:00Z
updated: 2026-11-01T01:05:00Z
---

## Current Test

number: 1
name: By-ear приёмка владельцем пункта трей-меню (живой стол)
expected: |
  Пункт «Добавить ИмяПриложения в чёрный список» виден под тумблером «Автокоррекция» с точным
  наблюдаемым именем; клик добавляет приложение в блок-лист, автокоррекция в нём замолкает;
  уже заблокированное приложение — с галочкой; неизвестная личность — пункт неактивен без имени
  и без пояснений; строка «Super+Space — язык» вместо «смена режима».
awaiting: user response

## Tests

### 1. By-ear приёмка пункта трей-меню (D-01/D-03/D-04, MENU-01)
expected: Пункт под «Автокоррекцией» с точным именем; клик блокирует приложение и автокоррекция замолкает без рестарта; галочка у уже заблокированного; неактивен без имени/пояснений для wezterm/snap; «Super+Space — язык»
result: [pending]

### 2. Round 2 пассивная журнальная верификация (Orca, следующая сессия владельца)
expected: Полный цикл «коррекция → флип → клиентский Reset → следующая коррекция» в терминале Orca; следующая коррекция срабатывает с ОДНОГО double-shift; в журнале `lifecycle aftermath=true`; ноль регресс-маркеров `terminal class verify_dead=false` после flip-released (оракул — §Round 2 journal verify в .planning/debug/orca-terminal-second-shift.md)
result: [pending]

## Summary

total: 2
passed: 0
issues: 0
pending: 2
skipped: 0
blocked: 0

## Gaps
