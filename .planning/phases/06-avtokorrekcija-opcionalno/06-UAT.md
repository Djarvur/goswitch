---
status: testing
phase: 06-Система автокоррекции
source: [06-VERIFICATION.md]
started: 2026-10-02T20:20:00Z
updated: 2026-10-02T20:20:00Z
---

## Current Test

number: 1
name: WINDOWS #13 / критерий 5 — формальный двойной зелёный matrix-v4 на свежей сессии
expected: |
  Двойной прогон `mise run e2e-matrix-v4` на fresh_session (или два запуска подряд на стабильной сессии).
  Владелец выносит вердикт по 9 drift-строкам (первое нажатие после флипа теряется; контрольное
  доказательство: дерево до фазы 6 d96dcb2 падает байт-в-байт так же): принять с оговоркой о риске /
  чинить стол / ночной гейт. Автокоррекционные строки и ручные носители были зелёными ×2 (06-08).
  Плюс 2 замороженные строки WINDOWS #12 (word/phrase-mixed — семантика фазы 5).
awaiting: user response

## Tests

### 1. WINDOWS #13 — двойной зелёный matrix-v4 (свежая сессия) + вердикт по drift-строкам
expected: `mise run e2e-matrix-v4` ×2 зелёные на свежей сессии, ИЛИ владелец принимает drift-строки
  с задокументированной оговоркой (drift не регрессия фазы — доказано контрольным экспериментом)
result: [pending]

### 2. ADR-007 Proposed → Accepted
expected: Владелец переводит Status в Accepted на основании таблицы Acceptance Evidence (6 строк)
  в docs/adr/ADR-007-autocorrect-hybrid-detector-and-role-policy.md; строка 5 честно красная,
  пока не закрыт пункт 1
result: [pending]

### 3. CR-01 — живая проверка асинхронного окна подтверждения
expected: autocorrect включён для white-list приложения; набрать слово + разделитель и НЕМЕДЛЕННО
  продолжить набор (пересечение с role-RTT ≤25 мс). Поле: либо корректное исправление ровно
  в диапазоне слова, либо молчание со счётчиком `ac_skip_payload_stale`. Никакой порчи поля/зеркала.
result: [pending]

### 4. Живой UAT трёх e2e-кейсов на стабильной сессии
expected: `mise run e2e-autocorrect-fires` (исправление без хоткея + ручной Double Shift отменяет),
  `mise run e2e-autocorrect-password-silent` (пароль не тронут, witness роль 40, FINAL дословный,
  fired=0), `mise run e2e-autocorrect-terminal-silent` (fired=0 при любой IME-доставке)
result: [pending]

### 5. Judgment-запреты ADR (4 позиции)
expected: Владелец подтверждает вердикты, не имеющие авторитетного статуса: дословность audit-trail
  SPEC §11, GPL-чистота dictgen-данных, cache-not-basis (D-53), отсутствие утечки содержимого из
  фикстур (D-20/D-21)
result: [pending]

## Summary

total: 5
passed: 0
issues: 0
pending: 5
skipped: 0
blocked: 0

## Gaps
