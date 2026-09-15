---
status: testing
phase: 03-Фразы, выделение, переключение и конфигурация
source: [03-VERIFICATION.md]
started: 2026-09-15T00:00:00Z
updated: 2026-09-15T00:00:00Z
---

## Current Test

number: 1
name: Повторный прогон приёмочной матрицы v2 на HEAD (8 фикс-коммитов моложе последнего зелёного прогона)
expected: |
  `mise run e2e-matrix-v2` на живом столе — 21/21 PASS на текущем HEAD (d94a3ce+). Единственные зелёные прогоны (живой 21/21 и CI run 34984814995) сделаны на headSha 21df37c; после него 8 ревью-фиксов добавили 737 строк поведенческих изменений на путях, которые матрица проверяет (WR-05 verify-after debounce, WR-03 preflight UID auth, CR-01/CR-02 проводка конфиг-ручек). Headless-корпус на HEAD зелёный — это свежесть доказательства, не сигнал дефекта. Опционально: повторный dispatch green106.
awaiting: user response

## Tests

### 1. Повторный прогон матрицы v2 на HEAD
expected: `mise run e2e-matrix-v2` → 21/21 PASS на текущем HEAD; опционально повторный прогон green106 CI
result: [pending]

### 2. Решение владельца — WINDOWS- ledger #2 (открыт): корпусное чтение D-24
expected: Подтвердить чтение D-24: однородный текст конвертируется целиком (уточнение D-22), поверхность `changed=false` зарезервирована для внешне-заякоренных диапазонов (выделение)
result: [pending]

### 3. Решение владельца — WINDOWS ledger #4 (открыт, unmet-truth): GTK4-Wayland не применяет IBus-forwarded события
expected: Демон-сторона MACR/level-2/Ctrl+V доказана на проводке (relay зафиксирован), но GTK4-виджеты не применяют forwarded-события. Принять как документированное платформенное ограничение (Chromium применяет) или расследовать; затем выровнять REQUIREMENTS.md (чекбокс MACR-01 всё ещё Pending — бухгалтерское расхождение, помечено Warning)
result: [pending]

### 4. Глазами владельца — разворот вердикта по выделению zenity (03-07 D5)
expected: Выживающее выделение ctrl+a в zenity молочно поглощает commit (вердикт спайка 03-03 был загрязнён неработающими probe-кандидатами) — подтвердить наблюдаемое поведение вручную
result: [pending]

## Summary

total: 4
passed: 0
issues: 0
pending: 4
skipped: 0
blocked: 0

## Gaps
