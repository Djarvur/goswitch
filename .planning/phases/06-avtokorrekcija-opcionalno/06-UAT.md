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
result: issue
reported: "Владелец предложил замерить сеттл-интервал после флипа. Замеры (пробы v1/v2 + журнал демона): границы времени нет; потеря = нажатие внутри окна переключения, которое путь демона растягивает сам — фабрика движка ждёт блокировку, удерживаемую flipTo, дедлайн 150мс рвёт узел, переключение доезжает с опозданием (WARN switch_engine context deadline exceeded на каждом флипе; engine created на +1мс после аборта). Прямой SetGlobalEngine 24/24 чисто, флип демона 5/5 съедает следующую букву."
severity: major
fixed_by: "gap-планы 06-09+06-10 (lock-free фабрика, ADR-006 amendment); живое доказательство 24/24 немедленных букв, RTT 6.5-9.2мс. Матрица v4 после фикса: 32/35 ×2 подряд идентично — все 9 дрейф-строк зелёные; остались word-mixed/phrase-mixed (замороженные WINDOWS #12, вердикт владельца) и super-space-alive (известная неинъектируемость Super+Space + фокус-транзиент)

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
issues: 1
pending: 4
skipped: 0
blocked: 0

## Deferred Follow-Ups

- test: 1
  idea: "РЕШЕНИЕ ВЛАДЕЛЬЦА 2026-10-03: смешанный текст конвертировать инверсией раскладки посимвольно (латиница→кириллица, кириллица→латиница) — целевая семантика для word-mixed/phrase-mixed; WINDOWS #12-фриз отклонён. Работа: отдельный план (флип-путь/контекст буфера), строки матрицы перезакрепить на новую семантику. Бэклог v1.1.x."
  deferred_at: 2026-10-03

## Gaps

- gap_id: G-6-1
  truth: "После флипа раскладки первая же буква доезжает до поля (матричные flip-строки зелёные в любой сессии)"
  status: resolved
  resolved_by: 06-09-PLAN.md + 06-10-PLAN.md
  resolved_at: 2026-10-03
  resolution: "lock-free фабрика (atomic emitterSlot), ADR-006 amendment; живое доказательство 24/24 немедленных букв (было 5/5 потерь), RTT 6.5-9.2мс, 8/8 флипов INFO-форма, two-source-flip регрессия зелёная (06-09/06-10 SUMMARYs)
