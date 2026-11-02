# Phase 13: A11Y — рецепты Telegram/ZCode - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-11-03
**Phase:** 13-A11Y — рецепты Telegram/ZCode
**Areas discussed:** Telegram-решение, Плац приёмки, Self-check, Доки рецептов

---

## Telegram-решение

**Q1: Рунг лестницы Telegram — что пробуем и документируем?**

| Option | Description | Selected |
|--------|-------------|----------|
| Unconfined tarball | Бинарник с telegram.org в ~/Applications + .desktop; лёгкий откат на snap | |
| Deb (сторонний репо) | apt/ppa-сборка без snap-песочницы | |
| Flatpak | org.telegram.desktop; своя песочница — a11y-достижимость доказывать пробой | ✓ |
| Демон-fallback | Остаемся на snap, документируем ручной жест + отдельная приёмка | |

**User's choice:** Flatpak

**Q2: Если unconfined-проба не подтвердит a11y-достижимость — что делает исполнитель?**

| Option | Description | Selected |
|--------|-------------|----------|
| Следующий формат | Автоперебор форматов: tarball → deb → flatpak; стоп на первом зелёном; все ряды в журнал | ✓ |
| Сразу fallback | Первая неудача → пишем fallback-рецепт | |
| Стоп-эскалация | Неудача → стоп и эскалация владельцу | |

**User's choice:** Следующий формат

**Q3: Telegram остаётся внутри Фазы 13 или выделяется?**

**User's choice (free text):** «сейчас у меня переключенив телеграме работает, проверь в виртуалке»

**Notes:** Дословно зафиксировано, не интерпретировалось. Следствие: probe-first D-01 (живая ВМ-проба до решения о сборке). Telegram остаётся в фазе. Telegram на ВМ не установлен (проверено в сессии: нет snap/flatpak/процесса) — проба включает установку.

**Q4: Что ставим на ВМ для живой пробы?**

| Option | Description | Selected |
|--------|-------------|----------|
| Сначала snap-реплика | Реплицирует хост владельца, перепроверяет AppArmor-вердикт | |
| Сразу flatpak | Пропускаем snap | |
| Snap + flatpak | Обе установки одной сессией — реплика + кандидат-рецепт, полная таблица | ✓ |

**User's choice:** Snap + flatpak

---

## Плац приёмки

**Q: Плац приёмки ZCode-рецепта?**

| Option | Description | Selected |
|--------|-------------|----------|
| ВМ + ваш by-ear | ZCode на ВМ для машинной проверки; финальный зачёт — стол владельца (прецедент G-10-1) | |
| Только ВМ | Все проверки и зачёт только на ВМ; стол владельца не трогаем | ✓ |

**User's choice:** Только ВМ

**Notes:** Телеграм-часть — та же логика (поручение «проверь в виртуалке»). Стенд запаркован — план включает подъём.

---

## Self-check

**Q: Форма self-check одной командой?**

| Option | Description | Selected |
|--------|-------------|----------|
| Скрипт в репо | scripts/ + человекочитаемый вердикт + код возврата; в доке одна строка запуска | ✓ |
| Строка в доке | Копипаст-строка (python3-gi AT-SPI-зонд), без нового файла | |
| Скрипт + строка | Оба варианта | |

**User's choice:** Скрипт в репо

---

## Доки рецептов

**Q: Где живут рецепты?**

| Option | Description | Selected |
|--------|-------------|----------|
| README + CONFIG | Кратко в README.md/README.en.md, детали в docs/CONFIG.md, residual-risk в README | ✓ |
| Отдельная страница | docs/RECIPES.md, из README ссылка | |
| Только CONFIG | Всё в docs/CONFIG.md, README не трогаем | |

**User's choice:** README + CONFIG

---

## Claude's Discretion

- Текст рецептов, внутренности self-check-скрипта, порядок snap→flatpak в пробной сессии, форма таблицы пробы (владелец не выбирал — зоны из «остальное возьму на усмотрение»)

## Deferred Ideas

- Не возникло
