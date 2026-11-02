# Phase 13: A11Y — рецепты Telegram/ZCode - Context

**Gathered:** 2026-11-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Рецепты включения доступности в самих приложениях (ZCode: user-local `.desktop` override с `--force-renderer-accessibility`; Telegram: живая проба и выбор сборки), документация рецептов и self-check одной командой. **Демон goswitchd НЕ меняется** (решение владельца 2026-10-08): ни нового кода демона, ни изменений его поведения. Весь код фазы — внешний self-check-скрипт и документация.

</domain>

<decisions>
## Implementation Decisions

### Telegram: проба и выбор сборки
- **D-01:** Probe-first — первый план фазы живая ВМ-проба Telegram ДО любых решений о рецепте/сборке. Основание — сообщение владельца (дословно, без интерпретации): «сейчас у меня переключенив телеграме работает, проверь в виртуалке» (2026-11-02, про хост владельца). Проба устанавливает машинную истину и перепроверяет вердикт AppArmor из A11Y-02 (2026-10-08). Что именно работает (переключение раскладки / слово / строка) — решает проба, не формулировка сообщения.
- **D-02:** На ВМ ставятся ОБЕ установки одной сессией: snap (реплика хоста владельца — проверяет и его сообщение, и старый AppArmor-вердикт) и flatpak (кандидат-рунг для рецепта). Полная сравнительная таблица.
- **D-03:** Если unconfined-проба не подтвердит a11y-достижимость поля — исполнитель автоперебирает форматы (tarball → deb → flatpak), стоп на первом зелёном, все ряды в журнал.
- **D-04:** Рунг лестницы при выборе новой сборки — **Flatpak** (выбор владельца). — **Reversibility:** reversible — смена формата сборки не ломает контракт.

### Плац приёмки
- **D-05:** Все проверки и зачёт — ТОЛЬКО на ВМ goswitch-stand (директива 6 + выбор владельца «Только ВМ» для ZCode): ZCode ставится на ВМ, стол владельца не трогаем вообще.
- Стенд сейчас запаркован (демон inactive, конфиг удалён): план обязан включать подъём — gap-сборка текущей ветки, конфиг стенда, свежая Wayland-сессия (recipe 10-08: `ssh stand@192.168.122.150`, `/tmp/stand-env.sh`, reboot-прецедент при мёртвой сессии).

### Self-check
- **D-06:** Готовый скрипт в репо (`scripts/`) с человекочитаемым вердиктом и кодом возврата; в документации — одна строка запуска. Демон не меняется: скрипт — внешний инструмент (AT-SPI-зонд видимости поля ввода).

### Документация
- **D-07:** Краткие рецепты в `README.md` и `README.en.md`; детали, побочные эффекты и границы — `docs/CONFIG.md`; residual-risk заметка (ручная коррекция на скрытых полях копирует в PRIMARY) — в README.

### Claude's Discretion
- Точный текст рецептов и порядок шагов; внутренности self-check-скрипта; порядок snap→flatpak внутри пробной сессии; форма сравнительной таблицы пробы.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Требования и границы
- `.planning/REQUIREMENTS.md` — A11Y-01, A11Y-02, A11Y-03 (traceability фазы)
- `.planning/ROADMAP.md` §Phase 13 — Goal и 4 критерия успеха
- `docs/SPEC.md` §4.3/§4.4 — двухразрешительная модель (role-60 AT-SPI, blocklist): почему автокоррекция молчит без a11y-дерева

### Прошлые вердикты
- `docs/adr/adr-008-spike-log.md` — AppArmor-вердикт snap-Telegram (живая проба 2026-10-08), venue-ряды стенда, session-env механика (XAUTHORITY-находка)

### Стенд ВМ
- `docs/ci-runner.md` — машина стенда goswitch-stand и её каналы
- `.planning/phases/10-term-slovo/` — SUMMARY планов 10-06/10-08: recipe подъёма стенда, gap-сборка, rsync-зеркало

### Доки-получатели
- `README.md`, `README.en.md`, `docs/CONFIG.md` — куда пишутся рецепты (D-07)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/appid/` — наблюдение identity приложения (что видит демон)
- `goswitchctl status` — роль-диагностика (`ac_role_errors`) — машинный оракул «видит ли демон поле»
- Каналы драйва ВМ: `virsh send-key` — единственный канал клавиатуры до IBus в любые окна (Electron тоже); AT-SPI+CDP для фокуса; скриншоты `virsh screenshot`; меню демона — D-Bus-дамп

### Established Patterns
- Директива 6: живые проверки только на ВМ, без исключений
- Probe-first (прецедент Фазы 12): живая проба до фиксации рецепта
- Два зелёных раунда подряд (SPEC §7.2) для вердиктов площадки

### Integration Points
- Существующие разделы доступности в README/CONFIG (рецепты toolkit-accessibility с Фазы 7) — новые рецепты дополняют их

</code_context>

<specifics>
## Specific Ideas

- Владелец про Telegram дословно: «сейчас у меня переключенив телеграме работает, проверь в виртуалке». НЕ интерпретировать как «автокоррекция работает» — состав работающего (переключение/слово/строка) устанавливает ВМ-проба.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 13-A11Y — рецепты Telegram/ZCode*
*Context gathered: 2026-11-03*
