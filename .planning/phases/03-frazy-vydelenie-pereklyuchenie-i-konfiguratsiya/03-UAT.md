---
status: complete
phase: 03-Фразы, выделение, переключение и конфигурация
source: [03-VERIFICATION.md]
started: 2026-09-15T00:00:00Z
updated: 2026-09-15T18:55:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Повторный прогон матрицы v2 на HEAD
expected: `mise run e2e-matrix-v2` → 21/21 PASS на текущем HEAD; опционально повторный прогон green106 CI
result: pass
note: "owner-authorized orchestrator run 2026-09-15T21:41+03:00 on the live desktop: 21/21 PASS at HEAD 54ea460 (c23e98c code + 2 planning-docs commits), all 8 review fixes covered; report e2e-report.txt; preflight 7/7 green, exit 0"

### 2. Решение владельца — WINDOWS- ledger #2 (открыт): корпусное чтение D-24
expected: Подтвердить чтение D-24: однородный текст конвертируется целиком (уточнение D-22), поверхность `changed=false` зарезервирована для внешне-заякоренных диапазонов (выделение)
result: pass
note: "owner accepted 2026-09-15: уточнение D-22 resolution confirmed (homogeneous wholesale, matrix v1 word-ru-en stays green; changed=false reserved for externally anchored ranges; D-16→D-23 succession) — WINDOWS ledger #2 resolved"

### 3. Решение владельца — WINDOWS ledger #4 (открыт, unmet-truth): GTK4-Wayland не применяет IBus-forwarded события
expected: Демон-сторона MACR/level-2/Ctrl+V доказана на проводке (relay зафиксирован), но GTK4-виджеты не применяют forwarded-события. Принять как документированное платформенное ограничение (Chromium применяет) или расследовать; затем выровнять REQUIREMENTS.md (чекбокс MACR-01 всё ещё Pending — бухгалтерское расхождение, помечено Warning)
result: pass
note: "owner accepted 2026-09-15 as a documented GNOME platform limitation (wire relay = provable boundary; Chromium-family applies, GTK4 widgets ignore); REQUIREMENTS.md MACR-01 flipped Complete with the caveat, footer refreshed — WINDOWS ledger #4 resolved"

### 4. Глазами владельца — разворот вердикта по выделению zenity (03-07 D5)
expected: Выживающее выделение ctrl+a в zenity молочно поглощает commit (вердикт спайка 03-03 был загрязнён неработающими probe-кандидатами) — подтвердить наблюдаемое поведение вручную
result: pass
note: "owner acknowledged 2026-09-15; corrected actual pinned by matrix row select-all-zenity (expect_text ghbdtn, anchor step degraded); re-proven live in the 21/21 matrix run at HEAD 54ea460 the same day"

## Summary

total: 4
passed: 4
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps
