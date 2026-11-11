---
phase: quick-261111-i9r
plan: 01
subsystem: docs
tags: [readme, distro-support, tested-on, requirements]
requires:
  - README.md v1.2.2 (main 4511608)
provides:
  - README RU/EN: подразделы «Протестировано»/«Tested on» и «Где предположительно заработает»/«Expected to work» в разделе «Требования» — проверенная среда (Ubuntu 24.04/GNOME 46/Wayland/IBus 1.5.29/systemd 255) и критерий предполагаемой переносимости (GNOME на Wayland с IBus), с явными границами (arm64 не проверен, GNOME>46 не проверен, без-IBus и X11 не заявлены)
affects:
  - ожидания пользователей до установки; фидбек-петля «отчёты о работе/не-работе приветствуются»
tech-stack:
  added: []
  patterns:
    - граница «проверено/предположительно» в доках проводится явно, с датой сверки фактов
key-files:
  created: []
  modified:
    - README.md
    - README.en.md
decisions:
  - ветка от main (4511608), не от ветки PR #32 — правки в «Требованиях» (верх файла) не пересекаются с разделом «Альтернативы» (#32, низ файла), PR сливаются в любом порядке
  - arm64-статус сформулирован честно («бинарники собираются и публикуются, живой проверки не было») вместо маркетингового «поддерживается arm64»
  - для KDE/Sway/Hyprland — отсылка к разделу «Альтернативы» (PR #32), а не дублирование
metrics:
  duration: ~25 min
  completed: 2026-11-11
  tasks: 2
  commits: 2
status: complete
actuals:
  tokens: 12000
  tasks: 2
  commits: 2
plan_head_before: 4511608
---

# Quick Task 261111-i9r: Дистрибутивы в README — Summary

## What was done

Запрос владельца 2026-11-11: в README — на каком дистрибутиве тестировался
и где предположительно заработает. Факты сверены живьём на хосте
(Ubuntu 24.04.5 LTS, gnome-shell 46.0, ibus 1.5.29-rc2, systemd 255,
amd64/T2-MacBook) и по docs/ci-runner.md (ВМ goswitch-stand — Ubuntu
24.04/GNOME 46; e2e-матрица — workflow_dispatch).

- **README.md** — в «Требованиях» два подраздела:
  «### Протестировано» (Ubuntu 24.04: GNOME 46/Wayland/IBus 1.5.29/
  systemd 255 — ежедневная рабочая сессия автора + ВМ стенда; arm64 —
  бинарники публикуются, живой проверки не было) и
  «### Где предположительно заработает» (критерий: GNOME на Wayland с
  IBus как дефолтным фреймворком; Fedora Workstation/Debian/Arch/
  openSUSE; только штатные интерфейсы сессии; GNOME>46 — не проверен,
  отчёты приветствуются; не заявлены без-IBus сеансы (отсылка к
  «Альтернативам») и X11).
- **README.en.md** — зеркальные «### Tested on» / «### Expected to work».

## Verification

- Ворота зелёной итерации: build/vet/`test -race -count=1` ok;
  golangci-lint 2.15.0 (пин) — 0 issues; tidy-diff пуст.
- Подразделы по одному разу, внутри «Требований» до «Установки»;
  маркированный список не тронут.

## Deviations

Нет.

## Follow-ups

- Живой прогон на arm64 (или хотя бы ВМ aarch64) — кандидат в будущую
  веху; после него оговорку в «Протестировано» снять.
- Если появятся отчёты пользователей о Fedora/Debian — зафиксировать
  проверенные конфигурации в этом же подразделе.
