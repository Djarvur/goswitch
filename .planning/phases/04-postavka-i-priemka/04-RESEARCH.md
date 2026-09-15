# Phase 4: Поставка и приёмка - Research

**Researched:** 2026-09-16
**Domain:** User-space packaging/install (systemd user unit + IBus component XML), release engineering (goreleaser + GitHub Releases + `go install`), performance measurement (ydotool→AT-SPI latency, VmHWM), e2e matrix v3 (gedit GTK3 surface, Chromium window modes), owner acceptance
**Confidence:** HIGH (install mechanics and perf constraints verified live on the target machine this session)

## Summary

Phase 4 is a delivery phase: no new daemon behavior, everything is packaging, measurement, matrix widening, and acceptance. The three load-bearing technical questions were answered by live probes on the target desktop (this machine = the green106 e2e runner, per `docs/ci-runner.md`):

**IBus install mechanics (INST-01).** The user registry cache is `~/.cache/ibus/bus/registry`; `ibus write-cache` (no flags) writes it but scans **only** `/usr/share/ibus/component` — `~/.config/ibus/component/` is NOT a default scan path on ibus 1.5.29. With `IBUS_COMPONENT_PATH=$HOME/.config/ibus/component ibus write-cache` the user XML enters the cache — but a subsequent plain `ibus write-cache` (no env) **evicts** it, and the running ibus-daemon does not see cache-written user components until a daemon restart. Consequences: every write-cache invocation in install/selfcheck must carry the env var; selfcheck should treat "component missing from registry" as a repairable state (re-run env-cache); and same-session visibility needs `ibus restart` (desktop survival proven by the stand's ibus-restart case) — the GNOME input-source picker/list only refresh after the daemon re-reads the registry.

**Perf measurement (INST-03).** The current e2e witness spawns `/usr/bin/python3 focus_helper.py` per poll: measured **60–80 ms per call**, with `witnessPoll = 100 * time.Millisecond` — the existing poll loop physically cannot resolve a p95 < 50 ms; its quantization floor is ~60 ms even with zero sleep. The perf run (D-43/D-44) therefore needs an upgraded witness: recommended is an event-driven AT-SPI `object:text-changed` listener (resident helper process, one startup cost, per-event timestamps) or at minimum a resident line-protocol helper polled at 5 ms. Memory: `VmHWM` in `/proc/<pid>/status` is the kernel-tracked peak RSS — a single read at run end, zero dependencies (D-45).

**Matrix v3 surfaces.** gedit 46 on Ubuntu 24.04 is **GTK3** (libgedit-gtksourceview-300, a fork of GtkSourceView for GTK3) — genuinely a different IM/AT-SPI generation from the GTK4 gnome-text-editor already covered; it is **not installed** on this machine/runner and needs `sudo apt install gedit` (human prerequisite). Chrome 153 on this desktop defaults to **native Wayland** (verified: wayland socket fds, zero X11 fds), so "оконные режимы Chromium" = adding the `--ozone-platform=x11` (XWayland) mode explicitly — the IM path differs (zwp_text_input_v3 vs IBus D-Bus input context), so capability tiers may differ and must be spike-pinned per the "pin the actual" discipline.

Release engineering: goreleaser **v2.18.1** (2026-09-05, verified via proxy.golang.org + GitHub API) with its **default ldflags already stamping `main.version`/`main.commit`/`main.date`** — D-37 needs only an uninitialized `var version string` (with "dev" fallback) in each main package; both channels (`go install github.com/Djarvur/goswitch/cmd/goswitchd@vX.Y.Z` + GitHub Releases assets) are compatible with the current `go.mod` (`go 1.23`, 3 direct deps, no changes needed).

**Primary recommendation:** Build `goswitchctl install/uninstall/selfcheck` as Go code (D-39) around a command-runner seam for TDD; always set `IBUS_COMPONENT_PATH` for every `ibus write-cache`; add an event-driven (or resident) AT-SPI witness mode for the perf run before trusting any latency number; install gedit (owner prerequisite) and spike gedit + Chromium-x11 surfaces before composing matrix v3.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
Verbatim from `.planning/phases/04-postavka-i-priemka/04-CONTEXT.md` `## Implementation Decisions`:

- **D-37:** Версия прошивается в бинарник на release-сборках (`-ldflags -X`: тег + commit); `goswitchd --version` и статус в `goswitchctl` показывают точную сборку — база разбора баг-репортов публичного инструмента. Сборки через `go install` идут без прошивки и показывают фолбэк (dev/unknown).
- **D-38:** Сборку и публикацию релиза ведёт goreleaser (`.goreleaser.yaml`; CI-инструмент — НЕ зависимость демона, рантайм-принцип «stdlib + godbus + минимум» не затронут). Каналы: GitHub Releases-артефакты + `go install github.com/Djarvur/goswitch/cmd/goswitchd@vX.Y.Z`.
- **D-39:** Установку выполняют подкоманды `goswitchctl install / uninstall / selfcheck` — не bash-скрипт и не README-шаги. install идемпотентен: пишет component XML в `~/.config/ibus/component/` (регистрация через `IBUS_COMPONENT_PATH`), systemd user unit (`~/.config/systemd/user/`), гоняет `ibus write-cache`, `systemctl --user daemon-reload` + enable --now. Тестируется Go-корпусом и e2e (принцип «полный цикл без человека»).
- **D-40:** install автоматически приводит стол к «одному хозяину» (D-03/ADR-001): gsettings input-sources = только goswitch, прежние источники сохраняются для восстановления при uninstall. После установки всё работает сразу, без ручных шагов.
- **D-41:** selfcheck — сквозной живой: версия бинарника → компонент виден ibus (`ibus list-engine`/registry) → user unit active → движок реально зарегистрирован (ListActiveEngines) → конфиг валиден → источник ввода goswitch. Каждый пункт зелёный/красный с подсказкой фикса — fail-fast стиль e2e-preflight.
- **D-42:** uninstall — полный откат: stop/disable юнита, удаление юнита + component XML + ibus write-cache, восстановление сохранённых источников ввода. Пользовательский `~/.config/goswitch/` НЕ трогается по умолчанию (наработки пользователя); флаг `--purge` удаляет и его.
- **D-43:** Статистика латентности — выделенный perf-прогон (mise-задача/e2e-режим): N повторов горячего кейса (коррекция слова + переключение раскладки), отчёт p50/p95/p99. Выборка однородная, не размазана по разнородным кейсам матрицы.
- **D-44:** «Реакция на горячую клавишу» = только e2e-окно ydotool→AT-SPI (дословно критерий №3 роадмапа). Интервал опроса AT-SPI-свидетеля фиксируется методикой — число честной верхней оценкой, воспроизводимо. Внутренний tap→commit-лог в приёмку НЕ входит.
- **D-45:** Память: оракул — `/proc/<pid>/status` VmHWM (пик RSS) против бюджета < 50 МБ; VmRSS снимается в контрольных точках (после старта, конец прогона). Ноль средовых зависимостей (cgroup-учёт user-сессий не нужен).
- **D-46:** Числа perf-измерений публикуются таблицей в README (p50/p95/p99 + RSS peak); актуализация таблицы — часть релизного цикла.
- **D-47:** Расширение покрытия — matrix v3 (новый YAML по образцу v1→v2): v2 остаётся замороженным принятым эталоном Фазы 3; v3 добавляет поверхность gedit и оконные режимы Chromium. Состав кейсов — планировщик по результатам исследования поверхностей.
- **D-48:** «Матрица зелёная дважды подряд на нетронутой сессии» = свежеподнятая GNOME-сессия раннера → прогон v3 №1 → прогон v3 №2 сразу, без пересоздания сессии и ручного вмешательства между прогонами (проверяет, что первый прогон не портит состояние для второго); любое падение — exit ≠ 0.
- **D-49:** Ручная приёмка владельца — письменный чек-лист в repo (установка с нуля по README на живом столе → живой набор всех жестов → uninstall), проходится как UAT-гейт verify-work фазы — та же механика, что приняла Фазы 1–3 (одна сессия по §7.2 спеки).
- **D-50:** README и инструкция установки двуязычные: `README.md` (EN) + `README.ru.md` (RU), синхронизация обоих — часть релизного цикла.

### Claude's Discretion
Verbatim from CONTEXT.md:
- Формат релизных ассетов (tar.gz vs голые бинарники + checksums), список архитектур (amd64-only vs +arm64), схема тегов и release notes — планировщик в связке с goreleaser-конфигом (решение владельца: «на усмотрение»)
- Точный N повторов и состав горячего кейса perf-прогона — пинируется методикой при реализации D-43
- Число и состав кейсов matrix v3 — после исследования gedit-поверхности (её IM-путь/класс виджета) и оконных режимов Chromium
- Дефолтный YAML-конфиг при первой установке (генерировать ли `~/.config/goswitch/config.yaml` при install, или демон работает на вшитых дефолтах до первого конфига) — в рамках принципов D-31..D-33
- Расположение чек-листа приёмки (отдельный docs/файл vs раздел README) и его точная структура
- Порядок задач, разбиение на планы, детали goreleaser/workflow-интеграции

### Deferred Ideas (OUT OF SCOPE)
Verbatim from CONTEXT.md:
- Терминальный класс (clipboard-уровень для терминалов, SPEC §6 Q3 / TERM-01) — остаётся во v2 (перенос из Фаз 2–3, во Фазу 4 не входит)
- OSD-уведомление при переключении раскладки (компенсация индикатора ADR-001) — вне объёма, остаётся во v2
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| INST-01 | Установка без root: systemd user unit, регистрация engine в IBus, `go install` + бинарник из GitHub releases | Live-verified IBus user-component mechanics (env-cache, eviction, daemon-restart visibility); goreleaser v2.18.1 default version stamping; `go install pkg@tag` subdirectory-main semantics; existing dev-unit template and `goswitchctl` CLI skeleton to extend |
| INST-03 | Реакция на горячую клавишу < 50 мс; потребление памяти < 50 МБ | Measured witness cost (60–80 ms/call) proves a poll-based e2e window cannot resolve <50 ms → event-driven/resident witness design; VmHWM semantics verified; p50/p95/p99 methodology for a homogeneous hot-case sample |
</phase_requirements>

## Project Constraints (from AGENTS.md / CONVENTIONS.md — no CLAUDE.md exists)

Checked: neither `./CLAUDE.md` nor `./.zcode/CLAUDE.md` exists. The binding directives live in the workspace `AGENTS.md` (GSD blocks from PROJECT/CONVENTIONS) and `.planning/STATE.md`:

1. **Strict TDD** — every behavior task is red → green → refactor; verify blocks run tests at each step (D-07).
2. **Green iteration** — an iteration ends only with `go build ./...`, `go test -race ./...`, `golangci-lint run` all clean; no "fix lint later" (D-08).
3. **mise, not make** — build/test/lint/e2e tasks are mise tasks in `mise.toml`; mise also pins tools (`[tools] go = "1.23"`, `golangci-lint = "2"`) — a goreleaser pin belongs there too if used locally (verified: `mise registry goreleaser` → `aqua:goreleaser/goreleaser`).
4. **golangci-lint maximal-strict config from Phase 1** (`.golangci.yml`, v2, `default: all`, depguard deny `unsafe`, `local-prefixes: github.com/Djarvur/goswitch`).
5. **GitHub Actions kept current**: pr-sanity on push/PR; dependabot weekly; scheduled govulncheck (D-11a/b/c). The release workflow is a new addition to this set.
6. **Runtime dependency principle**: `go install`-able daemon, stdlib + godbus + fsnotify + yaml.v3 only — goreleaser is a CI/dev tool and must NOT enter `go.mod` (D-38 locks this).
7. **All Go work follows the go-ultimate skill** (project skill at `.zcode/skills/go-ultimate/` — STATE.md Blockers/Concerns).
8. **Git workflow (user-level AGENTS.md)**: feature branches cut from up-to-date default branch; the phase branch `gsd/phase-04-postavka-i-priemka` already exists and carries the phase docs.
9. **Log privacy**: INFO = counters/reasons, details only under `-debug` (D-20/D-21) — applies to install/selfcheck output too.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Component XML + registry cache install | Client tool (`goswitchctl install`) | — | Pure user-space file write + `ibus write-cache` subprocess with env var; no root, no daemon involvement |
| systemd user unit lifecycle | Client tool | systemd user manager | `systemctl --user daemon-reload/enable --now/restart` subprocesses; unit template rendered from the binary's own identity |
| Input-sources takeover ("один хозяин", D-40) | Client tool | gsettings/GNOME | Save + set `org.gnome.desktop.input-sources sources` via `gsettings` subprocess (snapshot/restore pattern proven by the e2e stand since Phase 1); immediate engine activation is ibus SetGlobalEngine, not gsettings `current` (live-verified Phase 1: shell ignores runtime `sources/current` writes) |
| Engine registration (runtime) | Daemon (`goswitchd`) | — | Unchanged from Phases 1–3: `_RegisterComponent` on the private IBus socket; the XML is for cold-start/picker visibility, not for runtime registration |
| Selfcheck | Client tool | daemon (`ctlsvc.Status`), ibus bus (ListActiveEngines), systemd, gsettings | Orchestrates existing liveness surfaces; each check fail-fast with a fix hint (preflight pattern) |
| Version stamping | Build (goreleaser ldflags) | binaries (`--version`, status), component payload | Single `var version string` per main package; `engine.NewComponent` currently hardcodes `Version: "0.1.0"` — wire-through decision for the planner |
| Release build/publish | CI (GitHub Actions) | goreleaser, GitHub Releases | Tag-triggered workflow, `contents: write` scoped to that workflow only |
| Latency measurement | e2e stand (perf mode) | resident/event AT-SPI witness | t0 at ydotool injection, t1 at AT-SPI observation; honest upper bound per D-44 |
| Memory measurement | e2e stand (perf mode) | kernel /proc | VmHWM single read; VmRSS checkpoints |
| Matrix v3 breadth | e2e stand | surfaces (gedit, chromium x11 mode) | Same YAML schema/runner; new surface drivers + spike-pinned vocab entries |

## Standard Stack

### Core

| Library/Tool | Version | Purpose | Why Standard |
|--------------|---------|---------|--------------|
| goreleaser (dev/CI tool — NOT a Go dependency) | **v2.18.1** (2026-09-05) `[VERIFIED: proxy.golang.org @latest + GitHub API gh api repos/goreleaser/goreleaser/releases/latest]` | Release builds, archives, checksums, GitHub Releases publish (D-38) | De-facto standard Go release tool; default ldflags already stamp `main.version` (see Code Examples); binary distribution via aqua (mise-installable: `mise registry goreleaser` → `aqua:goreleaser/goreleaser`) |
| GitHub Actions release workflow | actions-provided | Tag `v*` → goreleaser | Extends the existing D-11 workflow set; needs `permissions: contents: write` scoped to this workflow only (pr-sanity precedent is `contents: read`) |
| `go install github.com/Djarvur/goswitch/cmd/{goswitchd,goswitchctl}@vX.Y.Z` | Go toolchain ≥1.23 `[CITED: go.dev/ref/mod]` | Second install channel (INST-01) | Subdirectory main packages supported (`go install golang.org/x/tools/gopls@latest` is the docs' own example); installs to `$HOME/go/bin` — exactly where the existing dev unit points (`ExecStart=%h/go/bin/goswitchd`) |
| Existing runtime deps (unchanged) | godbus v5.2.2, fsnotify v1.10.1, yaml.v3 v3.0.1 `[VERIFIED: go.mod:5-9 read this session]` | Daemon runtime | D-38 explicitly keeps the daemon's minimal-dep principle; `go install` channel works with the module as-is |

### Supporting

| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| `ibus write-cache` + `IBUS_COMPONENT_PATH` | ibus 1.5.29-2 (target) `[VERIFIED: dpkg + live probes]` | Registry cache refresh after XML install/uninstall (D-39/D-42) | EVERY invocation must carry the env var (see Pitfall 1) |
| `systemctl --user` | systemd 255 (Ubuntu 24.04) | Unit enable/restart/stop lifecycle | Dev unit template already proven on target (`dist/systemd/user/goswitchd.service`) |
| `gsettings` subprocess | GNOME 46 | Input-sources save/set/restore (D-40/D-42) | Stand's snapshot/restore pattern since Phase 1; writes go to `sources`, never runtime `current` |
| AT-SPI event listener (`object:text-changed`) | at-spi2-core 2.52.0, python3-gi 3.48.2 `[VERIFIED: installed per STACK research]` | Perf witness for sub-50 ms window (D-44) | New resident/event mode in `focus_helper.py`; standard AT-SPI2 event class `[ASSUMED — not yet implemented; standard AT-SPI2 event set]` |
| `/proc/<pid>/status` VmHWM | kernel | Memory oracle (D-45) | Single read at run end; kB units; kernel running max `[CITED: man7.org proc_pid_status(5)]` |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| goreleaser | hand-rolled `go build` + `gh release upload` script | Rejected: checksums/changelog/archive matrix are goreleaser's job; D-38 locks goreleaser |
| Event-driven AT-SPI witness | resident poller at 5 ms | Poller keeps the current helper shape (line protocol) but each poll still costs an in-process AT-SPI call; event listener has no quantum at all — implement the event listener, keep a bounded poll fallback for stability |
| VmHWM | periodic VmRSS sampling | Rejected: sampling can miss spikes; VmHWM is the kernel's own high-water mark, one syscall, D-45 names it the oracle |
| `gedit` as third surface | more GTK4 apps | Rejected: roadmap criterion 2 names gedit explicitly; its value IS the GTK3 IM/AT-SPI generation difference |

**Installation:**
```bash
# No new Go module dependencies. Dev/CI tool pin (fits D-09/D-11 mise convention):
mise use -t goreleaser@2.18.1   # or add goreleaser = "2.18.1" under [tools] in mise.toml
```

**Version verification (this session):**
```bash
curl -s https://proxy.golang.org/github.com/goreleaser/goreleaser/v2/@latest
# {"Version":"v2.18.1","Time":"2026-09-05T21:04:09Z",...,"URL":"https://github.com/goreleaser/goreleaser"}
gh api repos/goreleaser/goreleaser/releases/latest --jq .tag_name   # v2.18.1
mise registry goreleaser                                          # aqua:goreleaser/goreleaser
```

## Package Legitimacy Audit

> The seam's `package-legitimacy check` supports npm/pypi/crates only (live attempt: `Error: Usage: ... --ecosystem <npm|pypi|crates>`) — Go-ecosystem verification done against the two authoritative registries instead.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| github.com/goreleaser/goreleaser/v2 | proxy.golang.org + GitHub Releases API | first release 2016+; v2 line since 2024; v2.18.1 2026-09-05 | de-facto standard Go release tool (npm-style download stats N/A for Go modules) | github.com/goreleaser/goreleaser (VCS origin confirmed by proxy.golang.org `Origin.URL`) | OK (dual-source verified; seam gate not runnable for Go) | Approved |
| github.com/Djarvur/goswitch (self) | proxy (once tagged) + git remote | n/a | n/a | this repo (origin `git@github.com:Djarvur/goswitch.git`, public) | OK | Approved |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

No package enters `go.mod`: goreleaser is a build-time tool invoked via mise/CI, never imported by the daemon (D-38). The only new executable installed on user machines is goswitch's own build output.

## Architecture Patterns

### System Architecture Diagram

```text
INSTALL (goswitchctl install, user-space only)
  ├─ render component XML (identity mirrors engine wire component)
  │    → ~/.config/ibus/component/goswitch.xml
  ├─ IBUS_COMPONENT_PATH=~/.config/ibus/component ibus write-cache
  │    → ~/.cache/ibus/bus/registry now lists goswitch-en/goswitch-ru
  ├─ render systemd user unit (absolute binary path)
  │    → ~/.config/systemd/user/goswitchd.service
  ├─ systemctl --user daemon-reload && systemctl --user enable --now goswitchd
  │    → goswitchd registers at runtime (_RegisterComponent) + ctlsvc on session bus
  ├─ save prior gsettings input-sources → state file OUTSIDE ~/.config/goswitch (Pitfall 8)
  └─ gsettings set org.gnome.desktop.input-sources sources "[('ibus','goswitch-en')]"
       (immediate activation: ibus SetGlobalEngine; `current` writes are ignored by shell — Phase 1 live finding)

RUN (unchanged from Phases 1–3)
  keypress → ydotool/uinput → compositor → ibus-daemon → goswitchd engine (ProcessKeyEvent)
    → correction plan → CommitText/DeleteSurroundingText → focused field
    → AT-SPI bridge (witness)

PERF RUN (D-43..D-46, dedicated mise task)
  N × [t0=pre-ydotool timestamp → injection → AT-SPI text-changed event (or 5ms resident poll) → t1]
    → latency sample t1−t0 (honest upper bound; includes ydotool spawn, <10 ms measured)
  end of run: read /proc/<pid>/status VmHWM  →  report p50/p95/p99 + RSS peak → README table

RELEASE (tag v* → GitHub Actions)
  goreleaser build (ldflags stamp main.version/commit/date) → archives+checksums → GitHub Releases
  second channel: go install github.com/Djarvur/goswitch/cmd/goswitchd@vX.Y.Z (unstamped "dev" fallback)

ACCEPTANCE (D-48/D-49)
  fresh runner GNOME session → matrix v3 run #1 → run #2 (no session reset between; exit ≠ 0 on any FAIL)
  owner UAT checklist: install from README → live gestures → uninstall
```

### Recommended Project Structure (additions only)

```
cmd/goswitchctl/     # + install / uninstall / selfcheck subcommands (D-39/D-41/D-42)
internal/install/    # pure install logic: XML/unit rendering, gsettings save/restore,
                     #   write-cache/systemctl/gsettings via a command-runner seam (TDD-able)
dist/                # existing unit template stays the reference; install renders its own copy
.goreleaser.yaml     # NEW: two builds (goswitchd, goswitchctl), linux/amd64 (+arm64 at discretion)
.github/workflows/release.yml   # NEW: tag v* trigger, contents:write, mise installs goreleaser
test/e2e/cases/matrix-v3.yaml   # NEW: v2 frozen + gedit surface + chromium x11-mode cases
test/e2e/perf.go     # NEW: perf mode (D-43) — hot-case loop + report
test/e2e/focus_helper.py  # + resident/event witness mode (perf oracle, D-44)
docs/ACCEPTANCE.md   # or README section — owner UAT checklist (D-49, discretion)
README.md / README.ru.md  # bilingual rewrite (D-50) + perf table (D-46)
```

### Pattern 1: Every write-cache carries the env var (and selfcheck repairs)

**What:** `ibus write-cache` scans only `/usr/share/ibus/component` by default. The user dir enters the cache ONLY when `IBUS_COMPONENT_PATH` is set on the exact invocation — and a later plain invocation by any other tool EVICTS it (both live-verified, see Pitfalls 1–2).

**When to use:** install, uninstall, selfcheck — selfcheck's "component visible" step should re-run the env-cache as a repair action before failing (idempotent by design, D-39).

```text
DATA_xR7qL2vW_START (verbatim from /usr/include/ibus-1.0/ibusregistry.h:102-107, read this session)
 * Read all XML files in a IBus component directory (typically
 * /usr/share/ibus/component/ *.xml) and update the registry object.
 * IBUS_COMPONENT_PATH environment valuable is also available for
 * the custom component directories, whose delimiter is ':'.
DATA_xR7qL2vW_END
```

### Pattern 2: Component XML mirrors the wire identity (single source)

**What:** The XML's `<name>`/engine names/layout/language must byte-match what goswitchd registers at runtime, or the registry component and the live component fork (registry keys components by name — live-verified: a duplicate-named copy did not double-register).

Verbatim in-repo identity (this session):
- `[VERIFIED: engine/types.go:96-101]` — `ComponentName: "org.freedesktop.IBus.goswitch"`, `Description: "goswitch layout engine"`, `Version: "0.1.0"`, `License: "MIT"`, `Author: "Djarvur"`, `Homepage: "https://github.com/Djarvur/goswitch"`, `Exec: ""`
- `[VERIFIED: cmd/goswitchd/main.go:135-136]` — `engine.NewEngineDesc("goswitch-en", "goswitch English (US)", "en", "us", "en")`, `engine.NewEngineDesc("goswitch-ru", "goswitch Русская", "ru", "ru", "ru")`

Working XML shape (from the owner's prototype on this machine — `[VERIFIED: /usr/share/ibus/component/test-shift.xml read this session]`, shown with a random fence per boundary rules):

```xml
DATA_kM3nW9pQ_START
<?xml version="1.0" encoding="UTF-8"?>
<component>
  <name>org.freedesktop.IBus.TestShift</name>
  <description>Test shift capture</description>
  <exec>/home/nil/.local/share/ibus-test/test_shift.py</exec>
  <version>0.0.1</version>
  <author>nil</author>
  <license>MIT</license>
  <homepage>https://example.invalid</homepage>
  <textdomain></textdomain>
  <engines>
    <engine>
      <name>test-shift</name>
      <language>xx</language>
      <symbol>?</symbol>
      <setup>/bin/true</setup>
      <layout>us</layout>
      <icon_prop_key></icon_prop_key>
      <icon>/usr/share/ibus/icon/ibus-keyboard.svg</icon>
      <longname>Test Shift</longname>
      <description>Test</description>
      <rank>0</rank>
    </engine>
  </engines>
</component>
DATA_kM3nW9pQ_END
```

goswitch's XML keeps `<exec>` EMPTY (empty string in the wire component today `[VERIFIED: engine/types.go:20]` — "programmatic registration only, never spawned") so systemd stays the sole supervisor; an `<exec>` pointing at goswitchd would create a second, unsupervised spawn path colliding with the single-instance guard (ctlsvc `RequestName` ReplaceExisting semantics, Phase 3). Final call is the planner's; research recommends empty.

### Pattern 3: Version stamping rides goreleaser's defaults (D-37)

goreleaser's Go builder injects by default `[CITED: goreleaser.com/customization/builds/builds/go]`:

```text
DATA_zQ8fT4bR_START (verbatim default ldflags, goreleaser v2 docs)
'-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=goreleaser'
DATA_zQ8fT4bR_END
```

So each main package needs only:

```go
// Source: goreleaser.com/customization/builds/builds/go/ (default -X main.version injection)
package main

// version is stamped by goreleaser ldflags on release builds; `go install`
// builds it without stamping — the fallback is the honest dev marker (D-37).
var version = "dev"
```

Note the chain-through decisions: `goswitchd --version` (new flag), version inside `goswitchctl status` (via ctlsvc — a new status field), and optionally `engine.NewComponent`'s hardcoded `Version: "0.1.0"` `[VERIFIED: engine/types.go:98,128]` — either wire the stamped value through `engine.Config` or consciously leave the wire component on a scheme version.

### Pattern 4: Perf run = homogeneous hot case + event-driven witness (D-43/D-44)

- Sample: N repeats of the SAME hot case (word correction; layout flip measured via the follow-up rune's text event, since a flip alone changes no text) — one distribution, not a mix.
- Oracle window: t0 stamped immediately before the ydotool call; t1 = arrival of the AT-SPI observation (event line read from a resident helper). Includes ydotool spawn (measured <10 ms `[VERIFIED: live timing]`) — an honest upper bound, exactly what D-44 demands.
- Report p50/p95/p99 + VmHWM; anything the witness cannot resolve must be stated in the methodology (fixed poll interval named).
- Memory: VmHWM read once at end (kernel running max, kB) + VmRSS at the two checkpoints (after start, end of run) `[CITED: man7.org proc_pid_status(5)]`:

```go
// Source: man7.org/linux/man-pages/man5/proc_pid_status.5.html — VmHWM is the
// kernel-tracked peak RSS; parse the single "VmHWM:\t<num> kB" line.
func readVmHWM(pid int) (int64, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if v, ok := strings.CutPrefix(line, "VmHWM:"); ok {
			n, _ := strconv.ParseInt(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(v), "kB")), 10, 64)
			return n, nil // kB
		}
	}
	return 0, fmt.Errorf("no VmHWM line for pid %d", pid)
}
```

### Pattern 5: Matrix v3 = v2 discipline, new surfaces spike-pinned first

- v2 stays frozen (`matrix-v2.yaml` untouched, the accepted Phase 3 baseline); v3 is a new YAML `[VERIFIED: mise.toml:93-99 — the v1→v2 precedent]`.
- Surface vocabulary is closed and validated (`surface` must be one of `zenity|chromium|gnome-text-editor` `[VERIFIED: test/e2e/matrix.go:42-44,51-53]`) — adding `gedit` and an x11-mode chromium surface means extending the vocabulary AND the drivers, with spike-proven key/chord names only ("a name outside this set fails case validation, never the live run" `[VERIFIED: test/e2e/matrix.go:96-99]`).
- gedit driver needs the GTE-style isolation treatment (single-instance avoidance via XDG redirection — the 02-02 GTE lesson: `--standalone + XDG_DATA_HOME in temp`); gedit has no `--standalone` — spike required (Discretion: "после исследования gedit-поверхности").
- Chromium x11 mode: spawn flag `--ozone-platform=x11`; current cases run native Wayland by default (verified), so the x11 surface is the NEW tier; pin its actual ladder level/selection behavior, never assume parity with the Wayland mode.

### Anti-Patterns to Avoid

- **Writing `~/.cache/ibus/bus/registry` directly** — it is a binary GVariant-serialized IBusRegistry `[VERIFIED: live inspection]`; only `ibus write-cache` / the daemon may write it.
- **Running `ibus write-cache` without `IBUS_COMPONENT_PATH` after install** — evicts the user component (live-verified). Also never document bare `ibus write-cache` in README as a user step.
- **Saving the pre-install input sources inside `~/.config/goswitch/`** — D-42 preserves that dir on default uninstall; the backup would survive the uninstall that needs it. Use e.g. `~/.local/share/goswitch/`.
- **Per-poll python spawns in the perf loop** — 60–80 ms each `[VERIFIED: live timing]`; a <50 ms p95 becomes unmeasurable, not just inaccurate.
- **pkill/pgrep `-f` patterns from inside scripts on the owner's desktop** — the pattern matches the script's own shell cmdline (hit live during this research); use pid-based process identity.
- **Assuming `gsettings set ... current` activates anything** — GNOME 46 shell runtime-ignores it (Phase 1 live finding); activation is ibus SetGlobalEngine.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|------|
| Release packaging/publishing | tar+scp/gh scripts | goreleaser | checksums, archives, changelog, tag templating; D-38 locks it |
| Version stamping | custom build scripts | goreleaser default ldflags (`-X main.version=...`) | already the default; `go install` fallback needs zero code |
| IBus registry updates | GVariant cache writer | `IBUS_COMPONENT_PATH=... ibus write-cache` | binary cache format is ibus-private; subprocess is the only supported writer |
| systemd lifecycle | manual spawn/restart in Go | `systemctl --user` subprocesses | enablement, dependency ordering, restart policy are systemd's |
| Percentiles | streaming stats lib | sort + index (stdlib) | p50/p95/p99 over N≤1000 samples is `s[int(float64(len(s)-1)*p)]`; a stats dependency would violate the dep principle |
| Memory peak | RSS sampling loop | `/proc/<pid>/status` VmHWM | kernel's own high-water mark; no sampling misses (D-45) |
| gsettings access | D-Bus dconf client | `gsettings` subprocess | STACK.md verdict: robust against dconf backend details, zero new deps |

**Key insight:** every hand-roll candidate above already has a tool that owns the edge cases (cache format evolution, release signing conventions, kernel accounting). The only genuinely new engineering in this phase is the install/selfcheck orchestration and the perf witness — that is where plan effort belongs.

## Runtime State Inventory

> Phase 4 is a delivery phase (not a rename/refactor), but install/uninstall deliberately manages OS-registered runtime state, and the owner's machine (also the e2e runner) carries dev-era artifacts — inventoried so the plan can reconcile them.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data (user config) | No `~/.config/goswitch/` exists yet `[VERIFIED: ls this session]` | install's default-config decision (Discretion) defines whether it creates one |
| Live service config (ibus registry cache) | `~/.cache/ibus/bus/registry` exists, contains owner prototypes (`test-shift.xml`, `punto-switcher.xml` in `/usr/share/ibus/component/` — system-level, root-installed) `[VERIFIED: live strings of the cache + dir listing]` | install/uninstall must leave system components untouched; e2e assertions must target the goswitch component name only |
| OS-registered state (systemd user units) | `goswitchd.service` NOT installed on this machine right now (only `goswitch-ci-runner.service` — the runner's own unit) `[VERIFIED: systemctl --user is-active → inactive; ls ~/.config/systemd/user/]` | install e2e starts near-clean here; still snapshot any pre-existing goswitchd unit for idempotency tests (install-over-install) |
| Secrets/env vars | No goswitch env vars in play; daemon takes flags only (`-debug`, `-config`) `[VERIFIED: cmd/goswitchd/main.go:26-27]` | none |
| Build artifacts / release state | 0 git tags, no `.goreleaser.yaml`, no GitHub Releases yet `[VERIFIED: git tag empty; ls this session]` | first tag creates the v1 release; tag scheme is planner discretion |

**Canonical question check:** after all repo files are updated, the systems still holding goswitch state are: the ibus registry cache (written by write-cache), systemd user manager (unit files + enablement symlinks), gsettings input-sources (set by install), and GitHub Releases (published artifacts). All four are exactly what install/uninstall/selfcheck must own — none found in any other category (verified this session).

## Common Pitfalls

### Pitfall 1: Plain `ibus write-cache` evicts user-dir components
**What goes wrong:** After a correct env-var install, any later bare `ibus write-cache` (run by another engine's installer, or by a user following stale docs) rewrites the cache WITHOUT the user component — the engine vanishes from `ibus list-engine` and the GNOME picker after the next daemon restart.
**Why:** Default scan path is only `/usr/share/ibus/component`; the user dir was known to the cache only via the env-var observation.
**Live proof (this session):** user XML in cache = 1 occurrence after env-cache → **0 after plain cache**.
**How to avoid:** every goswitch-driven write-cache sets `IBUS_COMPONENT_PATH=$HOME/.config/ibus/component`; selfcheck re-runs it as a repair; README documents it as one command, never split.
**Warning signs:** selfcheck step "component visible" red after unrelated IBus changes.

### Pitfall 2: The running daemon does not reload the cache — visibility needs a daemon restart
**What goes wrong:** XML + env-cache succeed, but `ibus list-engine` (the live daemon) still doesn't list goswitch — a unique-name probe was invisible even while present in the cache file `[VERIFIED: live probe]`.
**Why:** the daemon reads the registry at startup; it does not watch the cache file for new component dirs it never observed.
**How to avoid:** install (or selfcheck-repair) ends with `ibus restart` to refresh the daemon's registry — desktop survival is proven by the stand's own ibus-restart case `[VERIFIED: test/e2e/matrix.go:588-601 — "The remedy is the ibus-native `ibus restart` (the stand's own ibus-restart case proves the desktop survives it)"]`; goswitchd's reconnect loop re-registers (INTEG-04). Sequencing (restart before/after unit start) is a planner call; runtime registration works regardless.
**Warning signs:** cache contains goswitch, list-engine doesn't.

### Pitfall 3: Poll-quantized witness cannot measure <50 ms
**What goes wrong:** reusing the matrix's witness loop for perf yields numbers whose floor is the observer, not the daemon (60–80 ms python spawn + 100 ms poll `[VERIFIED: live timing + test/e2e/case_m1.go:19-21]`) — every p95 would read ≥100 ms even if the daemon answered in 5 ms.
**How to avoid:** dedicated witness for the perf run: resident event listener (`object:text-changed`) or resident line-protocol poller at ~5 ms; methodology (D-44) names the mechanism and its quantum.
**Warning signs:** perf report where min latency ≈ poll interval.

### Pitfall 4: Write-cache same-second flake
**What goes wrong:** one observed failure this session — env-cache run immediately after XML creation did not observe the file (next identical run did).
**Why (ASSUMED):** mtime granularity race in the observed-path cache.
**How to avoid:** install writes the XML, then verifies via the cache/list-engine and re-runs write-cache once before failing.
**Warning signs:** intermittent "component not in registry" right after install.

### Pitfall 5: Layout-flip latency has no text event
**What goes wrong:** a flip (single Right Shift) produces no AT-SPI text change; an event witness never fires.
**How to avoid:** methodology defines the flip case as flip-tap + one probe rune; the text event of the probe rune closes the window (still ydotool→AT-SPI, D-44-compliant).

### Pitfall 6: gedit absent + GTK3 isolation unknown
**What goes wrong:** matrix v3 gedit cases fail wholesale — gedit is NOT installed on this machine/runner `[VERIFIED: which gedit → empty]` (needs `sudo apt install gedit`, universe repo — owner prerequisite); and gedit is GTK3 with its own single-instance/session semantics — the GTE `--standalone` trick does not transfer as-is.
**How to avoid:** owner installs gedit once on desktop+runner (document in ci-runner.md); a gedit driver spike (XDG redirection, window spawn, AT-SPI app/widget naming) precedes case composition — exactly the Discretion item "после исследования gedit-поверхности".
**Warning signs:** e2e preflight missing-binary check (add gedit to the matrix preflight table).

### Pitfall 7: Chromium "window modes" — the default is already native Wayland
**What goes wrong:** assuming current chromium cases run X11/XWayland; they run native Wayland (Chrome 153 on this desktop: wayland socket fds present, zero X11 fds `[VERIFIED: /proc/<pid>/fd live probe]`), so the "additional window mode" is the explicit `--ozone-platform=x11` XWayland surface with a DIFFERENT IM path (zwp_text_input_v3 vs IBus input context).
**How to avoid:** treat x11-mode as a new surface tier: spike first, pin actual ladder/selection behavior per the "pin the actual, never the assumption" discipline (02-05 precedent).
**Warning signs:** identical expect_level pins for both modes (assumption, not observation).

### Pitfall 8: Saved input-sources backup placed where uninstall won't look
**What goes wrong:** install saves prior sources into `~/.config/goswitch/`; default uninstall preserves that dir (D-42) — restore works, but `--purge` deletes the backup before... or worse, backup saved somewhere uninstall removes while D-42 semantics say preserve user data — either way the pair (save location × uninstall reachability) must be designed as one contract.
**How to avoid:** state file under `~/.local/share/goswitch/` (install-managed, always removed/restored by uninstall); `~/.config/goswitch/` stays user-owned (D-42).
**Warning signs:** uninstall tests that only check unit/XML removal, not sources restoration.

### Pitfall 9: Install/uninstall e2e on the live owner desktop
**What goes wrong:** an e2e install test that fails mid-way leaves the owner with goswitch-only input sources or no unit.
**How to avoid:** wrap install/uninstall e2e in the stand's gsettings snapshot/restore pattern (Phase 1 precedent); cases must be idempotent-safe and teardown-verified (`verifyRestored` precedent in matrix.go).

### Pitfall 10: Release workflow permissions and the D-11 convention
**What goes wrong:** a release workflow inheriting broad tokens, or goreleaser version drift between local and CI.
**How to avoid:** `permissions: contents: write` on the release workflow ONLY (pr-sanity stays `contents: read`); goreleaser pinned via mise.toml `[tools]` and installed by `mise install` in CI (the 01-05 precedent: CI and local run the same tasks).

### Pitfall 11: WINDOWS ledger discrepancy blocks /gsd-ship
**What goes wrong:** `.planning/WINDOWS.md` frontmatter reports `open_count: 2` (items #2 and #4 marked open in table and JSON) while STATE.md/REQUIREMENTS.md record both closed by owner decisions on 2026-09-15 `[VERIFIED: files read this session]`.
**How to avoid:** reconcile the ledger early in the phase (housekeeping task) — with `windows_enforce`, ship gates on `open_count > 0`.

## Code Examples

### Minimal .goreleaser.yaml (two binaries, linux, default stamping)

```yaml
# Source: goreleaser.com/customization/builds/builds/go/ (defaults quoted in Pattern 3).
# Version stamping needs NO custom ldflags — the Go builder default already
# injects -X main.version={{.Version}} -X main.commit -X main.date (D-37).
version: 2
builds:
  - id: goswitchd
    main: ./cmd/goswitchd
    goos: [linux]          # default is [darwin, linux, windows] — override (discretion: arch list)
    goarch: [amd64]        # default is [386, amd64, arm64] — override; +arm64 is owner discretion
  - id: goswitchctl
    main: ./cmd/goswitchctl
    goos: [linux]
    goarch: [amd64]
archives:
  - format: tar.gz         # asset format is discretion (tar.gz vs bare binaries)
checksum:
  name_template: 'checksums.txt'
changelog:
  sort: asc
  filters:
    exclude: ['^docs:', '^test:', '^chore:']
release:
  # GitHub Releases publish is goreleaser's default target under GITHUB_TOKEN.
```

### Install sequence (Go, seam-friendly sketch)

```go
// Source: live-verified mechanics on ibus 1.5.29 (this research); identity from
// engine/types.go:96-101 and cmd/goswitchd/main.go:135-136 (read this session).
const userComponentDir = ".config/ibus/component" // under $HOME

// EVERY write-cache carries the env var — Pitfall 1 (plain runs evict).
cmd := exec.CommandContext(ctx, "ibus", "write-cache")
cmd.Env = append(os.Environ(), "IBUS_COMPONENT_PATH="+filepath.Join(home, userComponentDir))
// …then: systemctl --user daemon-reload; enable --now goswitchd;
// save prior sources to ~/.local/share/goswitch/install-state.json (Pitfall 8);
// gsettings set org.gnome.desktop.input-sources sources "[('ibus', 'goswitch-en')]".
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Hand-installed dev unit (README `cp` steps) `[VERIFIED: README.md:29-34]` | `goswitchctl install` subcommands (D-39) | this phase | install becomes testable, idempotent, self-checking |
| goreleaser v1 config (`version: 1` omitted) | goreleaser v2 with explicit `version: 2` | v2 GA 2024 | config must carry `version: 2`; v2.18.1 current `[VERIFIED: proxy.golang.org]` |
| xdotool-style X11 e2e | ydotool/AT-SPI on Wayland | Phases 1–3 (settled) | unchanged; matrix v3 adds surfaces, not mechanisms |
| RSS sampling for memory budgets | `/proc/<pid>/status` VmHWM single read | D-45 locks it | kernel-accounted peak; no sampling misses |

**Deprecated/outdated:**
- AskUbuntu lore "gsettings `current` is read-only / CLI-switchable": superseded by live findings (writable but runtime-ignored by the GNOME 46 shell; SetGlobalEngine is the activation path).
- ibus#2354 lore (chromium lacks DeleteSurroundingText): already disproven live in Phase 2 (caps 0x29, level 1 applies) — matrix v3 pins actual behavior per mode, not issue-tracker age.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Chromium `--ozone-platform=x11` launches and exposes an XWayland/X11 IM path usable as a matrix surface (flag is standard; spawn verified this session, socket-level verification incomplete due to probe cleanup) | Matrix v3 / Pitfall 7 | x11 tier needs different pinning or is dropped (criterion 2 wording covers Chromium window modes — needs a live spike) |
| A2 | AT-SPI `object:text-changed` events (or a resident line-protocol helper) can serve as the perf witness with sub-10 ms quantum | Pattern 4 / Pitfall 3 | fallback: resident poller at named quantum; methodology must then state a coarser honest bound |
| A3 | The one write-cache same-second miss was an mtime-granularity race | Pitfall 4 | if systematic, install needs a different verify loop (retry on cache content, not list-engine) |
| A4 | gedit single-instance/session isolation is solvable with XDG redirection (GTE 02-02 analogue) | Pitfall 6 | gedit driver needs a different isolation mechanism; spike decides |
| A5 | `gsettings set org.gnome.desktop.input-sources sources "[('ibus','goswitch-en')]"` is accepted with ibus-type entries (format follows the verified xkb pair shape; not live-tested this session) | Pattern 1/install | install's final activation step may need SetGlobalEngine orchestration first — trivially testable in install e2e |
| A6 | `ibus restart` refreshes the daemon registry for cache-written user components (restart re-reads cache; stand proves desktop survives it; the exact fresh-XML→restart→list-engine chain was not run as one probe this session) | Pitfall 2 | if restart does not pick user components, install needs daemon-env investigation — early spike in plan |
| A7 | Tag scheme / asset format details (v1.0.0 vs 0.x; tar.gz vs bare; +arm64) — owner discretion per CONTEXT | Standard Stack | none (explicitly delegated) |
| A8 | "Свежеподнятая GNOME-сессия раннера" (D-48) maps to a documented runner-session restart procedure (runner lives in owner's graphical session — green106) | Open Questions | double-run definition may need an owner-agreed operational step (relogin before the gate run) |

## Open Questions

1. **WINDOWS.md ledger vs closure notes** — frontmatter `open_count: 2` while STATE/REQUIREMENTS record #2 and #4 closed by owner 2026-09-15. What we know: both files read this session, discrepancy is factual. What's unclear: whether closure was recorded only in docs, not via the ledger verb. Recommendation: reconcile as an early housekeeping task (ship gate reads open_count).
2. **D-48 session-freshness operationally** — what we know: runner is inside the owner's graphical session. What's unclear: who/what resets it and how the workflow asserts freshness. Recommendation: preflight check (session start time vs run start) + documented owner relogin step before the gate run.
3. **Component XML `<exec>`** — empty (recommended: systemd sole supervisor) vs binary path (ibus can spawn on demand, double-supervision conflicts). Recommendation: empty; planner confirms.
4. **Matrix v3 composition** — explicitly planner-owned after gedit/x11 spikes (CONTEXT Discretion); this research provides the surface facts, not the case list.
5. **Default config generation on install** (CONTEXT Discretion) — generate `~/.config/goswitch/config.yaml` at install vs built-in defaults until first user config; note Phase 3's strict rule: empty/missing EXPLICIT config = visible start refusal, defaults apply only without `-config` — install must not write a config that accidentally trips the strict decode.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| ibus (+ write-cache, list-engine) | INST-01 install | ✓ | 1.5.29-2 `[VERIFIED: dpkg]` | — |
| systemctl --user | INST-01 | ✓ | systemd 255 (dev unit proven) | — |
| gsettings | INST-01 (D-40) | ✓ | GNOME 46 | — |
| **gedit** | matrix v3 surface | **✗ NOT INSTALLED** `[VERIFIED: which gedit empty]` | — (46.2 in universe) | **none — criterion 2 requires it; owner runs `sudo apt install gedit` on desktop+runner** |
| google-chrome | matrix v3 (Wayland default verified) | ✓ | 153.0.8010.36 `[VERIFIED: --version]` | snap chromium exists as stand fallback (surface.go research A3) |
| ydotool | e2e stand | ✓ | 0.1.8 | own uinput injector (~150 LOC, STACK fallback) |
| python3-gi + AT-SPI | e2e stand/perf witness | ✓ | 3.48.2 / 2.52.0 | pure-Go godbus AT-SPI reader (later option) |
| goreleaser | release builds | ✗ local (not needed locally) | v2.18.1 pinned via mise `aqua:goreleaser/goreleaser` `[VERIFIED: mise registry]` | CI-only installation via `mise install` |
| gh CLI (authenticated, repo admin) | release/runner admin | ✓ | 2.100.0, logged in `[VERIFIED: gh auth status]` | — |
| Go toolchain | build | ✓ | go1.27.1 (module go 1.23) `[VERIFIED: go version]` | — |
| green106 self-hosted runner (gnome label) | double-run gate (D-48) | ✓ | per docs/ci-runner.md (this machine) | — |
| Git tags / GitHub Releases | release channel | ✗ none yet (0 tags) `[VERIFIED: git tag empty]` | — | created by the phase itself |

**Missing dependencies with no fallback:** gedit (blocking for matrix v3 — human prerequisite: `sudo apt install gedit` on the target desktop, which is also the runner; add to docs/ci-runner.md machine requirements).
**Missing dependencies with fallback:** goreleaser local install (CI-only usage acceptable; local `mise use` when dogfooding releases).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (+ golden files) — project convention; strict TDD mode (config `tdd_mode: true`) |
| Config file | `mise.toml` (tasks) + `.golangci.yml` (strict lint) |
| Quick run command | `mise run ci` (build+vet+lint+test -race) |
| Full suite command | `mise run ci` + live e2e tasks (owner desktop / green106 dispatch) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| INST-01 (install) | XML/unit rendering, idempotency, sources save/set | unit (fake `$HOME`, command-runner seam for systemctl/gsettings/ibus) | `go test -race ./internal/install/ -run TestInstall` | ❌ Wave 0 |
| INST-01 (uninstall) | full rollback incl. sources restore, `--purge` semantics | unit + live e2e | `go test -race ./internal/install/ -run TestUninstall` | ❌ Wave 0 |
| INST-01 (selfcheck) | each check green/red with fix hint; repair path re-runs env-cache | unit (fake runners) + live e2e | `go test -race ./internal/install/ -run TestSelfcheck` | ❌ Wave 0 |
| INST-01 (e2e cycle) | install → selfcheck → live use → uninstall on real session, snapshot-protected | e2e (live) | `go run ./test/e2e -case install-cycle` (new; mise task) | ❌ Wave 0 |
| INST-03 (latency) | N-sample p50/p95/p99 report via upgraded witness, exit ≠ 0 over budget | e2e perf run (live) | `mise run e2e-perf` (new) | ❌ Wave 0 |
| INST-03 (memory) | VmHWM < 50 MB oracle + VmRSS checkpoints in perf report | e2e perf run (live) | `mise run e2e-perf` | ❌ Wave 0 |
| Criterion 2 (matrix v3 + double run) | v3 green twice consecutively on fresh session | e2e (live + CI dispatch) | `mise run e2e-matrix-v3` ×2 / workflow dispatch | ❌ Wave 0 (`matrix-v3.yaml`) |
| Criterion 4 (owner UAT) | written checklist walked by owner in one session | manual-only (justified: SPEC §7.2 owner acceptance is by definition a human gate — the same mechanic that closed Phases 1–3) | checklist doc (`docs/ACCEPTANCE.md` or README section) | ❌ Wave 0 |
| D-37 (version) | `--version`/status show stamped vs dev fallback | unit (flag parse) + release smoke | `go test ./cmd/goswitchd -run TestVersion` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `mise run ci` (headless, fast)
- **Per wave merge:** `mise run ci` + relevant live e2e task on the desktop
- **Phase gate:** full suite green + double matrix-v3 run on fresh runner session + perf report + owner UAT (verify-work)

### Wave 0 Gaps
- [ ] `internal/install/` package + tests (command-runner seam, fake `$HOME`) — INST-01 unit corpus
- [ ] `test/e2e/cases/matrix-v3.yaml` + gedit/x11 surface drivers + vocabulary extension — criterion 2
- [ ] perf mode: `test/e2e/perf.go`, `focus_helper.py` witness mode, `e2e-perf` mise task — INST-03
- [ ] install-cycle e2e case + preflight gedit check — INST-01 live
- [ ] `.goreleaser.yaml` + release workflow + mise goreleaser pin — D-38
- [ ] `README.md` rewrite + `README.ru.md` + perf table + install instructions — D-50/D-46
- [ ] acceptance checklist doc — D-49
- [ ] WINDOWS.md ledger reconciliation — ship-gate hygiene

## Security Domain

`security_enforcement: true`, ASVS L1 (`.planning/config.json`).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | no auth surfaces in phase scope |
| V3 Session Management | no | — |
| V4 Access Control | partial | install writes ONLY under `$HOME` (XML, unit, state file); no root anywhere — enforced by refusing to write outside XDG paths; `systemctl --user` scope is inherent |
| V5 Input Validation | yes | uninstall restores input-sources from the saved state file: validate the parsed gvariant shape BEFORE `gsettings set` (a corrupt/injected state file must never brick the keyboard: restore falls back to a safe default `"[('xkb', 'us')]"` and reports); config validation unchanged (Phase 3 strict decode) |
| V6 Cryptography | no | checksums produced by goreleaser (sha256) — consumed, not hand-rolled |
| V14 Config | yes | unit/XML/state files written with explicit 0644/0600 perms, atomic write-then-rename; install resolves the daemon binary via absolute path (no `$PATH` lookup at unit start); unit template carries fixed `ExecStart`, no shell interpolation |

### Known Threat Patterns for user-space installer on GNOME/IBus

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Corrupted/tampered saved-sources state → keyboard bricked on uninstall | Tampering/DoS | shape-validate before restore; safe xkb fallback; state file 0600 under `~/.local/share/goswitch/` |
| Registry eviction by third-party write-cache (Pitfall 1) | DoS (availability of the IME) | selfcheck detects + repairs (re-runs env-cache); documented as known behavior |
| Unit hijack (another process pre-creates `~/.config/systemd/user/goswitchd.service`) | Elevation within user session | install overwrites atomically and verifies content hash after write; `systemctl --user` shows the unit's fragment path |
| Release artifact substitution | Tampering | goreleaser checksums.txt published with assets; `go install` channel integrity via proxy.golang.org checksum DB |
| Release workflow token overreach | Elevation (CI) | `permissions: contents: write` scoped to the release workflow only; no secrets needed beyond GITHUB_TOKEN |
| Log privacy in selfcheck/install output | Information disclosure | INFO level carries paths/verdicts only; key-trace detail stays behind `-debug` (D-20/D-21) |

## Sources

### Primary (HIGH confidence — live probes on the target machine, this session)
- `ibus write-cache` behavior matrix: user cache `~/.cache/ibus/bus/registry`; default scan `/usr/share/ibus/component` only; `IBUS_COMPONENT_PATH` pickup; plain-cache eviction; live-daemon non-visibility of cache-written unique-name component; one same-second flake — all reproduced by direct probes (create XML → cache → inspect → list-engine → cleanup; machine state restored after each probe)
- `/usr/include/ibus-1.0/ibusregistry.h:102-107` — IBUS_COMPONENT_PATH semantics (verbatim quoted in Pattern 1)
- `/usr/share/ibus/component/test-shift.xml`, `punto-switcher.xml` — working component XML shapes read in full
- `/proc/<ibus-daemon>/environ|cmdline` — daemon has no IBUS_COMPONENT_PATH; `/usr/bin/ibus-daemon --panel disable`
- Witness cost timing: `python3 focus_helper.py witness` ≈ 60–80 ms ×3; `ydotool --help` < 10 ms; `witnessPoll = 100 ms` from `test/e2e/case_m1.go:19-21`
- Chrome 153 default ozone = native Wayland (`/proc/<pid>/fd` socket probe, 8-proc instance)
- In-repo files read this session: `cmd/goswitchctl/main.go`, `cmd/goswitchd/main.go`, `engine/types.go`, `engine/factory.go`, `engine/conn.go` (grep), `test/e2e/matrix.go`, `test/e2e/surface.go` (grep), `mise.toml`, `go.mod`, `README.md`, `dist/systemd/user/goswitchd.service`, `.github/workflows/e2e-matrix.yml`, `docs/SPEC.md` §5–§9, `docs/ci-runner.md`, `.planning/{REQUIREMENTS,STATE,ROADMAP,WINDOWS}.md`, phase CONTEXT
- Environment facts: gedit absent; 0 git tags; no `.goreleaser.yaml`; no `goswitchd` user unit installed; no `~/.config/goswitch`; gh 2.100.0 authenticated; go1.27.1; `mise registry goreleaser` → `aqua:goreleaser/goreleaser`

### Secondary (MEDIUM confidence)
- [goreleaser Go builder docs](https://goreleaser.com/customization/builds/builds/go/) — default ldflags/builds matrix (verbatim quote in Pattern 3)
- goreleaser v2.18.1 — [proxy.golang.org](https://proxy.golang.org/github.com/goreleaser/goreleaser/v2/@latest) + GitHub Releases API (dual-source; the seam's legitimacy verb does not support the Go ecosystem — documented in the audit)
- [go.dev/ref/mod](https://go.dev/ref/mod) — `go install pkg@version` subdirectory main packages, `@latest`/pseudo-versions, GOBIN
- [man7.org proc_pid_status(5)](https://man7.org/linux/man-pages/man5/proc_pid_status.5.html) — VmHWM semantics
- [Launchpad gedit 46.1-3 noble](https://code.launchpad.net/ubuntu/noble/amd64/gedit/46.1-3), [pkgs.org gedit 46.2-2](https://ubuntu.pkgs.org/24.04/ubuntu-universe-amd64/gedit_46.2-2_amd64.deb.html), [Debian libgedit-gtksourceview](https://packages.debian.org/trixie/libgedit-gtksourceview-300-3), [LFS gedit](https://www.linuxfromscratch.org/blfs/view/stable/postlfs/gedit.html) — gedit is GTK3 + libgedit-gtksourceview-300, no GTK4 port

### Tertiary (LOW confidence)
- AT-SPI `object:text-changed` event availability/latency as perf witness — standard event set per AT-SPI2 docs, not yet implemented/measured here (A2)
- ibus daemon cold-start vs cache interplay for user components (A6) — restart-refresh chain assumed from observed cache behavior + stand precedent

## Metadata

**Confidence breakdown:**
- Install mechanics (IBus component path): HIGH — every claim reproduced live on the target ibus 1.5.29, including the two failure modes (eviction, daemon non-visibility)
- Release engineering: HIGH for facts (v2.18.1, default stamping, go install semantics via official docs/registries); MEDIUM for exact yaml details until the config exists
- Perf methodology: HIGH on the constraint (witness quantum measured), MEDIUM on the event-witness design (A2) until implemented
- Matrix v3 surfaces: HIGH on environment facts (gedit absent, GTK3; Chrome Wayland default), LOW→MEDIUM on behavior pins (spikes pending, per CONTEXT discretion)

**Research date:** 2026-09-16
**Valid until:** 2026-10-16 (stable domain; goreleaser version may bump — re-check tag before pinning)
