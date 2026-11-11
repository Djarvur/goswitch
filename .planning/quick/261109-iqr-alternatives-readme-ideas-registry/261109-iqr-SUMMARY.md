---
phase: quick-261109-iqr
plan: 01
subsystem: docs
tags: [alternatives, readme, competitive-analysis, ideas-registry, wayswitch, lay]
requires:
  - README.md v1.2.2 (main 4511608)
provides:
  - README RU/EN раздел «Альтернативы»/«Alternatives» — таблица 6 открытых корректоров по исходникам + архитектурное различие IBus vs evdev+uinput + честные недостающие фичи
  - docs/IDEAS.md — реестр 32 идей (P1–P3) с донорами и file:line, «что не переносить», лицензионные оговорки, закрытые продукты
affects:
  - будущие вехи автокоррекции (IDEA-01…IDEA-10 — готовый вход /gsd-new-milestone)
tech-stack:
  added: []
  patterns:
    - сравнение альтернатив по исходникам с file:line-доказательствами в docs/IDEAS.md (публичный README несёт только проверенные утверждения)
key-files:
  created:
    - docs/IDEAS.md
  modified:
    - README.md
    - README.en.md
decisions:
  - в README включены только проекты с изученными исходниками; LinguaFix (закрытая сборка, исходников нет) упомянут только в docs/IDEAS.md «Закрытые продукты»
  - тон раздела — информационный канон владельца 2026-11-08: архитектурное различие и чужие само-признания (uaccess WaySwitch, «low-level keylogger» xswitcher) вместо оценочных превосходств; отдельный абзац «честно о том, чего у goswitch нет»
  - приоритеты P1 в реестре — гейты InputPurpose (IDEA-09), контракт пробела + prefetch (IDEA-06), focus receipt (IDEA-08), AT-SPI EditableText-фолбэк (IDEA-18), защищённые слова (IDEA-19), undo-обучение (IDEA-20)
metrics:
  duration: ~75 min
  completed: 2026-11-09
  tasks: 3
  commits: 3
status: complete
actuals:
  tokens: 61000
  tasks: 3
  commits: 3
plan_head_before: 4511608
---

# Quick Task 261109-iqr: Альтернативы в README + реестр идей — Summary

## What was done

Толчок — opennet №66588 (WaySwitch 1.0, 2026-11-02). Проведено исследование
с нулем готового раздела: статья прочитана; найдены и клонированы 7
репозиториев в /tmp/goswitch-alt-study (недоверенные данные — только чтение);
исходники изучены пятью параллельными агентами с цитатами file:line;
ключевые факты для README (лицензии, модели прав, чужие само-признания)
перепроверены напрямую по клонам.

1. **README.md** — новый раздел «Альтернативы» (после «Удаления», перед
   «Документацией»): проприетарные родоначальники; таблица шести открытых
   проектов (WaySwitch, lay, HumanType, xswitcher, WaylandSwitcher,
   layout-switcher) с языком/вводом/внесением правки/GNOME/лицензией;
   архитектурное различие (IBus committed-text против evdev+uinput и
   добора фокуса сторонними каналами; uaccess-широту признаёт README
   WaySwitch, «low-level keylogger» — README xswitcher); абзац «честно о
   том, чего у goswitch нет» (их коррекция не зависит от IME-фокуса;
   опечатки/транслит/числа прописью; граф правил lay; KDE/Sway/X11) и
   «что goswitch предлагает вместо» (права user-сессии, детерминизм,
   продукт-полировка, Go-аудит, MIT).
2. **README.en.md** — зеркальный «Alternatives», тот же порядок и факты.
3. **docs/IDEAS.md** (новый) — реестр 32 идей по группам (детектор;
   IBus-нативное внесение; жесты; цели правки; языки/данные;
   эксплуатация) с приоритетами P1–P3, донорами и file:line; разделы
   «Что не переносить», «Лицензионные оговорки» (lay некоммерческая,
   xswitcher AGPL, словарь CC-BY-SA отдельно), «Закрытые продукты».

## Verification

- Ворота зелёной итерации (Д-2): `go build`, `go vet`,
  `go test -race -count=1` — все пакеты ok; golangci-lint 2.15.0 (пин)
  — 0 issues; `go mod tidy` — diff пуст. (`mise run ci` падает на
  известном затенении go 1.27.3/1.23.12 — ворота гонялись по шагам из
  рецепта памяти gsd-stale-verification-recipe; код не менялся, только
  docs.)
- Раздел в README.md ровно один (grep), стоит между «Удалением» и
  «Документацией»; шесть репо-ссылок совпадают с клонами; зеркальность
  EN проверена структурой.
- Каждое утверждение о чужих проектах в README опирается на file:line в
  docs/IDEAS.md или на их собственные README.

## Deviations

Нет. (Дополнительный коммит — план до материалов — по образцу
261108-qwa.)

## Follow-ups

- P1-идеи реестра (IDEA-06…IDEA-09, 18–20) — кандидаты в следующую веху
  автокоррекции (`/gsd-new-milestone`); перед планированием сверить с
  кодом пункты «сверить» (SetContentType, контракт пробела, FSM жестов).
- Мониторинг WaySwitch/lay — оба активны ежедневными релизами; таблица
  README датирована «состояние на ноябрь 2026».
