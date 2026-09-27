# Security Policy — goswitch

Реестр угроз и политика безопасности. Угрозы сводятся сюда из
threat-моделей планов по мере прохождения фаз; формат — построчный
реестр STRIDE (ASVS L1, `block_on: high`).

**Контакт:** Daniel Podolsky (владелец) — приватные отчёты через GitHub
Security Advisories репозитория `Djarvur/goswitch` (вкладка Security →
Report a vulnerability).

---

## Фаза 2 — Коррекция слова EN↔RU (планы 02-01..02-06)

Все угрозы реализованы соответствующими планами (диспозиция mitigate —
корпус тестов / живые кейсы), статус closed. Формулировки сжаты;
полные — в `<threat_model>` соответствующего плана.

| Threat ID | Категория | Компонент | Severity | Disposition | Mitigation (сжато) |
|-----------|-----------|-----------|----------|-------------|--------------------|
| T-02-01-01 | Tampering | арифметика диапазона (internal/correct/plan.go) | high | mitigate | счёт строго в рунах + хвостовая геометрия; корпус TestBuildPlan_* пинит до живого прогона |
| T-02-01-02 | Tampering | границы слова (internal/correct/buffer.go) | medium | mitigate | letterCapable из объединения EN+RU слоёв (D-14); корпус пинит ','/';' внутри, '/'/'-' — граница |
| T-02-02-01 | Elevation | стенд открывает chrome в профиле владельца | medium | mitigate | --user-data-dir строго во временный каталог; закрытие по PID своего процесса |
| T-02-02-02 | Information Disclosure | AT-SPI чтение чужих окон | low | mitigate | helper читает только запущенные стендом приложения; witness-гейт перед инъекцией |
| T-02-02-03 | DoS | зависший chrome/gnome-text-editor на живом столе | medium | mitigate | kill по PID + wait; reap-путь по образцу reapZenity; watchdog стенда |
| T-02-03-01 | Tampering | двухфазная коррекция (internal/session/actor.go) | high | mitigate | сверка ADR-004 до удаления: несовпадение/таймаут → тихий отказ, ноль эмиттов (TestActor_VerifyPaths) |
| T-02-03-02 | DoS | новые пути обработчиков | medium | mitigate | никаких ожиданий в ProcessKeyEvent (Pitfall 4); recover-шимы сохранены |
| T-02-03-03 | Information Disclosure | логи коррекции | high | mitigate | D-20/D-21: INFO — только reason/outcome без слов; исход→результат — -debug; тест лога |
| T-02-03-04 | Tampering | арифметика диапазона в проводке | high | mitigate | формулы поставлены корпусом 02-01; актёр вызывает BuildPlan — дублей арифметики нет |
| T-02-04-01 | Tampering | RU-потребление: неверный символ в поле | high | mitigate | коммит только по golden-таблицам layouts.ENToRU; немапленное — транзит |
| T-02-04-02 | DoS | флип-режим ломает ввод | high | mitigate | потребление только чистых печатных в RU; bare-модификаторы и Ctrl-комбо — транзит; кейс word-ru-en |
| T-02-04-03 | Tampering | буфер-desync на транзитных рунах RU | high | mitigate | инвариант script-true в обеих ветвях печати; пины TestActor_RUScriptTrueAllBranches |
| T-02-04-04 | Tampering | застрявший RU-режим после ошибки | medium | mitigate | INFO mode-запись наблюдаема; флип — идемпотентный toggle на каждый Single |
| T-02-05-01 | Tampering | бёрст Backspace×N стирает чужой текст (desync) | high | mitigate | счёт по рунам из буфера; сбросы CORR-09; verify-after; живые кейсы reset-* |
| T-02-05-02 | Tampering | caps-бит «врёт» → уровень 1 не применяется | high | mitigate | кейс ladder-chromium фиксирует фактический уровень; verify-after ловит дубль; матрица регрессионна |
| T-02-05-03 | DoS | бёрст форвардов перегружает клиента | low | accept | бёрст ≤ длине слова+хвоста (десятки); пейсинг — запасной ход |
| T-02-06-01 | Tampering | кейс-инъекция через YAML | medium | mitigate | строгий декодер KnownFields; кейсы в git (ревью); шаги только type/key/tap/focus — никакого exec |
| T-02-06-02 | Elevation | инъекция в чужие окна живого стола | high | mitigate | witness-гейт перед каждой инъекцией; поверхности только собственные; watchdog 180 с |
| T-02-06-03 | Information Disclosure | отчёт/логи содержат набранный текст | low | accept | отчёт и кейсы — открытые тестовые данные, не пользовательский ввод; лог демона следует D-20/D-21 |

## Фаза 2 — CI-раннер (план 02-07)

Раннер живёт в графической сессии владельца и исполняет код
репозитория на живом столе — граница доверия «GitHub → раннер в сессии».

| Threat ID | Категория | Компонент | Severity | Disposition | Mitigation |
|-----------|-----------|-----------|----------|-------------|------------|
| T-02-07-01 | Elevation | fork-PR запускает матрицу на столе владельца | high | mitigate | workflow_dispatch — единственный триггер; permissions contents: read; раннер подключён только к Djarvur/goswitch |
| T-02-07-02 | DoS | два стенда печатают одновременно (CI + локальный прогон) | high | mitigate | concurrency e2e-gnome без cancel-in-progress; свидетель-гейт стенда — последняя линия |
| T-02-07-02b | Information Disclosure | утечка registration-token | high | mitigate | токен одноразовый (минуты жизни), минтится в переменную без логирования, не коммитится; credentials живут в ~/actions-runner/.credentials вне репозитория |
| T-02-07-03 | Elevation | раннер исполняет PR-код из fork (вектор репо-уровня) | medium | mitigate | dispatch-only + PR-ревью (основные ветки защищены, PR-only поток); прогон — только вручную |
| T-02-07-04 | Tampering | bootstrap chore-PR меняет default branch | low | mitigate | дифф строго один файл workflow (без кода/шагов); ревью PR-потоком; сам workflow dispatch-only |

**Печать на живом столе владельца из CI — осознанное решение владельца
(D-19):** единственный триггер `workflow_dispatch` (право запуска —
роли с write), минимальные `permissions`, concurrency-группа, раннер
только этого репозитория. Инструкция раннера и политика —
`docs/ci-runner.md`.

## Принятые риски

| Risk ID | Threat Ref | Обоснование | Дата |
|---------|------------|-------------|------|
| R-02-01 | T-02-05-03 | Бёрст ForwardKeyEvent ограничен длиной слова+хвоста (порядка десятков нажатий) — перегрузка клиента нереалистична; пейсинг — запасной ход A4 | 2026-09-15 |
| R-02-02 | T-02-06-03 | Отчёт матрицы и кейсы содержат открытые тестовые слова (ghbdtn), не пользовательский ввод; лог демона следует контракту D-20/D-21 | 2026-09-15 |

---

*Фаза 1: реестр угроз фазы 1 вёлся в threat-моделях планов 01-01..01-05
(21 угроза, все закрыты); в дереве реестр Фазы 1 не переносился —
реестр SECURITY.md заведён планом 02-07.*
