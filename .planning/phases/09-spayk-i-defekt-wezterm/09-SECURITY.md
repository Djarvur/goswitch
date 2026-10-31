---
phase: "9"
slug: "spayk-i-defekt-wezterm"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-10-31"
---

# Phase 9 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| зондовый глагол probe-forward → живой канал демона | любой процесс с доступом к session bus может звать ctl-метод — глагол выдаёт синтетические клавиши | синтетический keyspec; без -debug канал fail-closed |
| -debug трейс / pty-захваты → журналы | трассировка клавиш экспонирует ввод («пароли становятся видимыми») | только синтетический корпус (ghbdtn/SENTINEL); pty в журнал — только hex |
| пробы → рабочий стол владельца | инжекция и фокус идут в реальные окна; промах = набор в чужое окно | ключи-синтетика в окна стенда |
| SPEC/ADR-текст → будущий код Фаз 10–13 | ревизии — контракт, по которому будет писаться код | формулировки инвариантов (TERM-04) |
| SPEC/ADR → публичный MIT-репозиторий | документы публичны | механизмы, не секреты |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-09-01-01 | Tampering | формулировка вето §11 SPEC | high | mitigate | вето только в точке армирования (роль 60/identity); grep-гейты: «никогда не армируется в контексте терминала», «точке армирования», «в терминалах разрешена» — PASS (rev D-55); blocklist упомянут только как то, через что вето НЕ выражается | closed |
| T-09-01-02 | Tampering | потеря аудит-трейла при ревизии SPEC | medium | mitigate | дословное сохранение прежнего текста §5 в HTML-комментарии; diff-сверка: хунки только в 5 целевых секциях, прежний пункт встречается 2 раза (аудит-комментарий + ревизия) | closed |
| T-09-01-03 | Tampering | Evidence-ячейки ADR-008, предзаполненные веб-утверждениями | medium | mitigate | Proposed с пустыми плейсхолдерами журнала; `grep -c "https\?://" docs/adr/ADR-008-terminal-transport.md` = 0 | closed |
| T-09-01-04 | Information Disclosure | SPEC/ADR как публичные документы MIT-репозитория | low | accept | документы описывают механизмы, не секреты; ключи/пути владельца не упоминаются | closed |
| T-09-02-01 | Tampering/Elevation | probe-forward как поверхность злоупотребления | medium | mitigate | fail-closed nil-dep: `errProbeDisabled` = «probe disabled (daemon not under -debug)» (internal/ctlsvc/ctlsvc.go:47-100) — подтверждено живым прогоном 2026-10-29; строгая валидация keyspec — корпус internal/session/probe_test.go | closed |
| T-09-02-02 | Information Disclosure | -debug трейс и pty-захваты | high | mitigate | корпус только синтетический (ghbdtn — case_autocorrect.go:64); pty только hex в журнал (wezterm.go:452 hex.EncodeToString); захваты не коммитятся — flip2/t04-журналы untracked (живой ввод владельца, решение гигиены Pitfall 6; бесконтентный validate-daemon.log — единственный tracked прецедент) | closed |
| T-09-02-03 | DoS | промах инжекции в чужое окно владельца | high | mitigate | только поверхности стенда; focus-детектор по focus_in демона над spawn-baseline, без AT-SPI (wezterm.go:6-11); арминг-гейт pty-потока; слепая инжекция запрещена | closed |
| T-09-02-04 | DoS | юнит-демон владельца против демона стенда | medium | mitigate | preflight `systemctl --user is-active goswitchd` (case_install.go:323,373); юнит остановлен перед живым прогоном, teardown best-effort возвращает | closed |
| T-09-02-05 | Repudiation | вердикты из веб-источников вместо живых проб | high | mitigate | журнал дописывают только кейсы; Evidence-правило ADR-008; `grep -c "https\?://"` = 0 в ADR-008 и adr-008-spike-log.md | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

**Примечание по отставленным планам:** 09-03-PLAN.md и 09-04-PLAN.md имеют `status: superseded` (#2349, 2026-10-20) — их threat-реестры не входят в активный регистр фазы: содержимое доставлено Фазой 10 (ADR-008 Accepted планом 10-06) либо рассеянными вердиктами; риск-поверхность переходит под собственную security-проверку Фазы 10.

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-09-01 | T-09-01-04 | SPEC/ADR публичны по замыслу: механизмы, не секреты | владелец (план 09-01) | 2026-10-08 |
| AR-09-02 | T-09-02-02 | flip2/t04-журналы оставлены untracked на диске как свидетельства (несут живой ввод владельца); решение зафиксировано в 09-02-SUMMARY Self-Check | оркестратор, гигиена Pitfall 6 | 2026-10-20 |

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-31 | 9 | 9 | 0 | оркестратор /gsd-manager (L1 grep-depth, ASVS 1; register authored at plan time) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-10-31 (L1, ASVS 1 — threats_open: 0, block_on: high)
