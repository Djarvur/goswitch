# Фаза 13: disposition код-ревью (13-REVIEW.md, 2026-11-06)

Источник: `13-REVIEW.md` (ревью 2026-11-06T14:54:50Z, 0 critical / 4 warning /
2 info). Все шесть находок — fixed; демон (`internal/`, `cmd/`) не тронут
(гейт `git diff --exit-code -- internal/ cmd/` чист).

| Находка | Статус | Коммит | Что сделано |
|---|---|---|---|
| WR-01 | fixed | `96a3705` | `import gi` / `require_version("Atspi","2.0")` / импорт типелибы закрыты try/except (ImportError/ValueError): проблема в stderr + `sys.exit(2)` вместо трейсинга с кодом 1. Проверено живьём linuxbrew-python3 (без gi): exit 2 + строка «run via /usr/bin/python3», трейсинга нет. |
| WR-02 | fixed | `96a3705` | Вход в a11y-шину в `applications()` (`Atspi.get_desktop(0)` + `get_child_count()`, включая None-десктоп как AttributeError) под guard-ом: проблема «a11y bus unreachable … the probe cannot answer» + exit 2 — не маскируется под семантический код 1. Глосс кода 2 в docstring расширен (ошибка вызова / нет процесса / сломанное окружение). |
| WR-03 | fixed | `bc1b29e` | Перечисление детей в `walk()` стало поштучным best-effort: `get_child_at_index(k)` в собственном try/except — мёртвый ребёнок скрывает только себя и своё поддерево, младшие братья остаются. `get_child_count()` под прежним общим guard-ом. Поведение проверено стабом: гибель индекса 1 из 3 не теряет индексы 0/2 и поддерево. |
| WR-04 | fixed | `a615ddc` | В adr-008-spike-log после ряда snap r1 добавлен датированный ряд восстановления «a11y-проба telegram-snap r2 (восстановлен 2026-11-06)» (провенанс: пассивное восстановление по `journalctl --user -u goswitchd`; слой 1 — пара 15:57:11.438/15:57:13.443; слои 2-3 пассивно НЕ восстановимы). Итог-ряд: snap переключение ЗЕЛ ×2 (теперь обоснован), видимость/автокоррекция snap — ×1 с пометкой о потерянной r2-строке и идентичном стабе paired-рунгов. CONFIG.md: «двумя раундами в каждой» → честная форма с парами раундов и восстановленным snap-слоем. |
| IN-01 | fixed | `8fbaf62` | Глосс кода 2 во всех трёх местах (docs/CONFIG.md, README.md, README.en.md) дополнен ветками ошибки вызова (ноль/лишние аргументы — live-факт 2026-11-03) и сломанного окружения зонда (синхронно с WR-01/WR-02). |
| IN-02 | fixed | `4f37110` | f-префикс снят у print-строки без подстановок в password-ветке (класс F541); текст, включая термин «видимо, но неисправимо», не тронут. |

## Гейты после фиксов (все зелёные)

- `/usr/bin/python3 -m py_compile scripts/a11y-selfcheck.py` — OK.
- `go test ./test/e2e/ -run TestSpikeJournalSectionsExist` — ok (заголовки
  секций журнала не сдвинуты; правки рядов пин не ломают).
- `mise run ci` (PATH без `/usr/local/go/bin` впереди пина mise —
  документированная ловушка; go 1.23.12) — build/vet/lint (0 issues)/test
  -race -count=1 ./... зелёные.
- `git diff --exit-code -- internal/ cmd/` — чисто (граница демона держит).

Живой smoke WR-01: `python3 scripts/a11y-selfcheck.py zcode` (linuxbrew,
без gi) → exit 2 + человекочитаемая проблема в stderr — контракт кодов
восстановлен (окружение ≠ «жив, но поля нет»).

---

_Disposition: 2026-11-06_
_Фиксер: ZCode (gsd-code-fixer)_
