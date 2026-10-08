# Phase 8: a11y-магия приложений, документация и конфиг-skill - Research

**Researched:** 2026-10-05
**Domain:** GNOME 46 accessibility mechanics (Chromium/Electron/GTK/snap), goswitch config-schema extension patterns, docs restructure, Agent Skills SKILL.md generation
**Confidence:** HIGH (a11y mechanics live-probed on the target desktop + Chromium source read; code patterns read in-repo; docs facts read in-repo)

## Summary

Phase 8 carries three backlog streams in dependency order: (1) a config section listing apps for which goswitch automatically enables accessibility («a11y-магия», D-8-1/D-8-2), (2) docs accuracy + russian-first restructure (D-8-7..D-8-9), (3) a SKILL.md for AI assistants generated from docs/CONFIG.md (D-8-10/D-8-11). No phase requirement IDs exist (REQUIREMENTS.md v1 is all-Complete) — the three CONTEXT.md decisions are the requirements source.

The load-bearing discovery of this research: **`org.gnome.desktop.interface toolkit-accessibility` is read by Chromium/Electron ONCE at app startup and never watched live** — verified both by reading Chromium's `ui/accessibility/platform/atk_util_auralinux.cc` (main branch) and by a live experiment on the target desktop (flipping the key while ZCode was running did not grow its a11y tree within seconds). The owner's manual ZCode precedent therefore equals «set the key, then the app picks it up at its next start». Consequences for design: the daemon's reconcile sets a *persistent* switch and guarantees trees only for apps *started after* enabling; docs must state the restart semantics. A second lever exists — `org.a11y.Status.IsEnabled` on the a11y bus — and was **live-verified settable without root** via a plain D-Bus Properties.Set, but on this GNOME 46 it is *decoupled* from the gsettings key (flipping the key did not move the property; the running `at-spi-bus-launcher` synced neither way).

Stream-2 facts: README.ru.md (217 lines) is materially stale against README.md (410 lines) — the entire v1.1.0 Autocorrect block (7 subsections incl. «Where autocorrect stays silent», «Switch sounds», GTK3 limitation) and «Input sources and the mode indicator» exist only in English; CONFIG.md is currently fully English (D-8-9 makes it russian-first). Stream-3 mechanics: CONFIG.md's key table is a strictly formatted markdown table (the single source D-8-10 names), so the generator parses that table and renders a marked region of `skills/goswitch-config/SKILL.md`, gated by the project's existing `tidy-diff`-style discipline.

**Primary recommendation:** Stream 1 ships the minimal magic of D-8-6 — a new `a11y` config section (regex app list, zero-value-off like `autocorrect`), a daemon-side reconciler that idempotently sets the global gsettings key (read-verify-then-set, no-pipes subprocess, outside the actor mutex) plus optionally the session-scoped IsEnabled D-Bus belt; the gsettings key joins `install-state.json` for the uninstall revert. Per-app `.desktop` overrides are documented but recommended OUT of the v1 diff (the global key covers the owner's ZCode case after app restart). Spec-delta lands before code (D-55).

<user_constraints>
## User Constraints (from CONTEXT.md)

### A11y-магия: что решено владельцем дословно

- **D-8-1 (владелец, todo a11y-magic-apps-list, 2026-10-05):** нужен НЕ перечень в
  документах, а **секция в конфиге** со списком приложений, которым goswitch автоматически
  включает accessibility-магию «как сделали вручную для ZCode» (прецедент подтверждён).
  Форма — «секция типа `a11y.apps` / расширение `autocorrect`» (точное имя — discretion).
- **D-8-2 (владелец, todo a11y-magic-apps-list):** механику применения (gsettings toolkit-accessibility,
  snap-оверрайды, флаги запуска) **исследовать при планировании** — research определяет
  фактические механизмы per-app на GNOME 46 до кода.

### A11y-магия: рекомендации агента (не подтверждены владельцем)

- **D-8-3:** Момент применения — **демон-reconcile**: при старте и на каждом hot reload
  списка демон идемпотентно приводит систему к состоянию из конфига. Install лишь создаёт
  секцию в конфиге.
- **D-8-4:** Откат — **revert при uninstall** (ASVS-прецедент: чорды wm.keybindings
  восстанавливаются, снапшот-дисциплина sources). Удаление приложения из списка НЕ
  откатывает применённое (невозможно отличить от включённого владельцем вручную до
  goswitch). Reversibility: reversible — снапшот + revert-шаг.
- **D-8-5:** Матчинг списка приложений — **regex-подстрока** (RE2, анкеровка явная) —
  консистентность с `autocorrect.apps_blocklist` (D-53); строгий decode, валидация
  компиляции на Load.
- **D-8-6:** Набор магии — минимально достаточный: глобальный
  `gsettings org.gnome.desktop.interface toolkit-accessibility true` (ровно то, что
  включалось вручную для ZCode) + опциональные per-app user-level правки запуска
  (`~/.local/share/applications` override с флагами типа
  `--force-renderer-accessibility`). Всё без root, всё в $HOME. Snap-оверрайды —
  research решает по необходимости; глобальный ключ НЕ снимается динамически
  (нельзя выключать «для одного приложения» — ключ общий).

### Документация: русский-первый

- **D-8-7 (владелец, дословно):** «README.ru как главная, README как перевод» →
  **README.md = русский**, английский перевод — README.en.md; синхронность поддерживается.
- **D-8-8:** README/CONFIG.md/SPEC сверяются с фактическим поведением v1.1.0 (меню,
  тумблеры, звуки, blocklist, adopt+watch, «где молчит»); расхождения в ДОКАХ правятся;
  если расхождение = дефект поведения — новый todo, код в этой фазе не правится.
- **D-8-9:** CONFIG.md — русский-первый (пользовательский док); ACCEPTANCE.md /
  ci-runner.md / SECURITY.md остаются EN. SPEC.md уже русский.

### Skill «Конфигурация goswitch»

- **D-8-10:** Синхрон с доками — **single source**: конфиг-раздел SKILL.md генерируется
  из docs/CONFIG.md mise-задачей (рассинхрон невозможен); разделы процедур и диагностики
  — ручной каркас, CI lint-гейт сверяет ключи/дефолты.
- **D-8-11:** Расположение — `skills/goswitch-config/SKILL.md` (project skill).

### Claude's Discretion

- Имя секции конфига (`a11y.apps` vs иное) и схема ключей — по канонам схемы
  (ACTIVE-условия, потолки, strict decode, complete-document — прецедент 06-04/07-02).
- Разбиение на планы/волны (порядок потоков 1→2→3).
- Объём регресса: `mise run ci` обязателен; e2e-матрица не расширяется; юнит/golden на
  новый код по директивам TDD.
- Состав документации-сверки: полный обход README/CONFIG/SPEC против кода.

### Deferred Ideas (OUT OF SCOPE)

- Snap-оверрайды как механизм магии — если research одобрит (D-8-6 оставил на его усмотрение).
- Показ наблюдаемой a11y-идентичности в `goswitchctl status` — backlog 06-UAT Deferred.
- Новые возможности коррекции/переключения, GUI, терминалы, OSD — v2.

### Project Constraints (from AGENTS.md / CONVENTIONS.md — no ./CLAUDE.md exists)

- Go ≥ 1.23, только stdlib + `godbus/dbus` + минимальные зависимости — никаких новых Go-модулей.
- Строгий TDD (red → green → refactor); зелёная итерация: `go build ./...`, `go test -race ./...`, `golangci-lint run`.
- Проверочные сценарии — mise-задачи, не make; golangci-lint максимально строгий с Фазы 1.
- GitHub Actions актуальны: pr-sanity + dependabot + scheduled govulncheck.
- Без root, всё в $HOME; conventional commits; изменения через PR.
- Русская проза владельческих текстов: «сочетание клавиш» (не «чорда»), без жаргона.
- Role-гейт фаз 6/7 (fail-closed) не трогается.
</user_constraints>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| a11y-конфиг-секция (schema/validate/hot reload) | Daemon (`internal/config`) | — | The config package owns every schema section; strict decode + Validate are there already |
| a11y-магия применение (gsettings/IsEnabled) | Daemon (reconciler component) | — | D-8-3: daemon-reconcile at start + on reload; subprocess side-effects live outside the actor mutex (WR-01 precedent) |
| a11y revert при uninstall | Installer (`internal/install`) | Daemon | D-8-4: goswitchctl uninstall restores the snapshot; snapshot source = install-state.json |
| Документация (README/CONFIG/SPEC) | Docs tree | — | Pure content work; spec-delta before code (D-55) |
| SKILL.md генерация | Dev tooling (mise task + Go test) | CI | D-8-10: generated from CONFIG.md; gate rides the existing test/tidy-diff discipline |

## Standard Stack

### Core

No new dependencies — the phase runs entirely on the existing stack plus OS tools already used in-repo.

| Component | Version | Purpose | Why Standard |
|-----------|---------|---------|--------------|
| Go toolchain | 1.27.x local, module `go 1.23` [VERIFIED: go.mod is unchanged; STACK.md] | All stream-1/-3 code | Project constraint |
| `gopkg.in/yaml.v3` | v3.0.1 (already in go.mod) | Schema section + (existing) Node round-trip writer | In-repo |
| `github.com/godbus/dbus/v5` | v5.2.2 (already in go.mod) | Optional `org.a11y.Status.IsEnabled` belt (session bus → a11y bus address) | In-repo; the ctlsvc/indicator packages already speak session-bus godbus |
| `gsettings` CLI (os/exec) | GNOME 46, ships with OS [VERIFIED: live target 2026-10-05] | The a11y key read/verify/set — same no-pipes subprocess shape as `internal/sound`/`activate.go` | Zero deps; the STACK decision against direct dconf D-Bus |
| `busctl`/godbus for org.a11y.Bus | at-spi2-core 2.52 [VERIFIED: live, `at-spi-bus-launcher` running] | IsEnabled belt (optional) | Live-verified settable without root |
| python3-gi + Atspi | 3.48.2 / 2.52.0 [VERIFIED: live dpkg per STACK.md] | Optional manual/executor a11y-tree probes (not product code) | e2e stand convention (`test/e2e/focus_helper.py` uses `Atspi.get_desktop(0)`) |
| mise tasks | mise.toml | `skill-gen` task + gates | Directive 3: mise, not make |
| Agent Skills format | SKILL.md + YAML frontmatter | `skills/goswitch-config/SKILL.md` | [CITED: agentskills.io/specification] + in-repo precedent |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| gsettings subprocess for the key | Direct `ca.desrt.dconf` D-Bus write via godbus | Rejected by STACK (subprocess overhead irrelevant on user-paced actions; CLI is robust) |
| gsettings read of current key value | dconf read | Same — stay with the CLI |
| IsEnabled D-Bus belt | nothing (key only) | Belt costs one godbus call and needs no revert (session-scoped); recommend shipping it only if the planner wants snap-Electron coverage belt |
| Parse CONFIG.md table for skill-gen | Structured YAML/JSON source rendered into BOTH docs | Two renderers = two sources of drift; CONFIG.md IS the single source (D-8-10) — parse the table it already has |
| Rust/other for the generator | Go dev-tool under `cmd/` or `internal/` | Repo is Go-only; lint applies |

**Installation:** none — zero new packages. Package Legitimacy Audit below.

## Package Legitimacy Audit

No external packages are installed by this phase (constraint: stdlib + existing go.mod only). Nothing to audit.

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram (stream 1: the a11y reconcile)

```
                         config.yaml (a11y section)
                                   │ edit
                                   ▼
                     fsnotify (dir watch, debounce 200 ms)
                                   │ reparse (strict decode + Validate)
                                   ▼
                        Watcher.Snapshot()  ◄──── one read per fold (CONF-02)
                                   │
        ┌──────────────────────────┼─────────────────────────────┐
        ▼                          ▼                             ▼
  Actor.applySnapshot()     A11yReconciler.diff()          pushMenuSync()
  (existing fold —           desired = section ACTIVE?      (existing)
   new field only)           changed vs last-applied?
        │                          │ yes
        │                          ▼ (async, serialized, outside actor mutex)
        │                   ┌─ apply() ─┐
        │                   │ gsettings get  org.gnome.desktop.interface toolkit-accessibility
        │                   │   → already true?  no-op (idempotent, no dconf churn)
        │                   │ gsettings set  ... true           (persistent lever)
        │                   │ busctl-equivalent Properties.Set  org.a11y.Status.IsEnabled=true
        │                   │   (optional belt, session-scoped, no snapshot needed)
        │                   └───────────┘
        ▼                          │
  hot path unaffected   ◄──────────┘ (no-pipes subprocess; failures = ONE WARN per episode)
                                   │
                                   ▼
            apps started FROM NOW ON read the key at their own startup
            (Chromium: one-shot ShouldEnableAccessibility → tree ON)
            running apps: NO live change — restart required (documented)

  goswitchctl install ──► saveState(): snapshot pre-install key value into install-state.json
  goswitchctl uninstall ─► restore from install-state.json (if the field exists; else WARN + skip)
```

Entry point: config edit or daemon start. Decision points: section ACTIVE (enabled + non-empty list), effective-state diff. External dependencies: gsettings CLI, a11y bus. The role-gate (phase 6/7 fail-closed) is untouched — this phase only widens *which apps HAVE trees*.

### Recommended Project Structure

```
internal/config/       # + a11y section (schema, validate) — the 07-02 pattern
internal/a11y/         # NEW: reconciler — config seam, Runner seam, apply/revert logic
internal/install/      # + snapshot field in installState, restore step in Uninstall
internal/session/      # + one folded field / async trigger in applySnapshot (or wiring in cmd/goswitchd)
cmd/goswitchd/         # wire reconciler beside activate.IfOwned (self-reactivation precedent)
cmd/skillgen/          # NEW dev tool: CONFIG.md table → SKILL.md marked region (or internal/skillgen)
skills/goswitch-config/SKILL.md   # frame (hand) + generated region (marked)
.github/workflows/     # unchanged — the gate rides existing `mise run test`/tidy-diff shape
```

### Pattern 1: Config section with zero-value-off + ACTIVE-conditional validation

**What:** the new section decodes with the strict decoder for free; zero value = off; ceilings and regex compilation bite where they can matter.
**When to use:** stream 1's schema.
**Example** (the pinned in-repo pattern to mirror — `internal/config/config.go:117-123, 356-385`):

```go
// Source: internal/config/config.go (verbatim field names/ceilings)
type Autocorrect struct {
    Enabled       bool     `yaml:"enabled"`
    AppsBlocklist []string `yaml:"apps_blocklist"`
    MinWordLen    int      `yaml:"min_word_len"`
    TrigramMargin float64  `yaml:"trigram_margin"`
    TrigramFloor  float64  `yaml:"trigram_floor"`
}
// ceilings: maxAutocorrectBlocklist = 64 (config.go:28)
// validate: TrimSpace-empty pattern refused BEFORE regexp.Compile (config.go:363-370);
// errors name field AND index: "autocorrect.apps_blocklist[%d] = %q: %w"
```

For `a11y`, the analogous shape (planner's discretion on the exact name per D-8-1): `enabled` + `apps []string` (RE2 substring, explicit anchoring), possibly nothing else — the magic set is fixed by D-8-6, not per-app configurable. Ceiling 64 (the `maxMACRApps`/`maxAutocorrectBlocklist` figure, [VERIFIED: config.go:26-31]).

### Pattern 2: no-pipes fork-shaped subprocess (the gsettings call)

**What:** every subprocess in the daemon is either deadline-bounded with captured output (reads) or the no-pipes fire-and-forget shape (fire-and-forget writes).
**Example** ([VERIFIED: internal/sound/sound.go:84-93 and internal/clipboard/clipboard.go:151-171] — quotes: `newPlayProc` builds `exec.Command(name, args...)` with "Stdin/Stdout/Stderr stay nil — no pipes"; `execRunner` uses `exec.CommandContext` with the deadline-bounded ctx):

```go
// Read-verify-set form (reads want output + deadline; sets are no-pipes):
out, err := runner(ctx, "gsettings", "get", "org.gnome.desktop.interface", "toolkit-accessibility")
// out == "true\n" → no-op; else fire-and-forget set, no pipes, reaped goroutine (sound.go:192)
```

### Pattern 3: install snapshot / uninstall restore

**What:** pre-state captured verbatim into a 0600 JSON before any mutation; uninstall restores only-if-present.
**Example** ([VERIFIED: internal/install/install.go:47, 61-62, 218-222] — verbatim: `permPrivate = 0o600 // install-state.json`, `stateDirRel = ".local/share/goswitch"`, `stateFile = "install-state.json"`; `installState` currently carries `Sources`, `SwitchInputSource`, `SwitchInputSourceBackward`):

```go
type installState struct {
    Sources             string `json:"sources"`
    SwitchInputSource   string `json:"switch_input_source"`
    SwitchInputSourceBw string `json:"switch_input_source_backward"`
    // + the new pre-install toolkit-accessibility value (e.g. ToolKitAccessibility string `json:"toolkit_accessibility"`)
}
```

### Pattern 4: async side-effect trigger from the fold (sound-sink precedent)

**What:** the actor's fold (under mutex) must not run subprocesses. The sound sink shows the sanctioned shape: a field-diff inside the fold, the side-effect behind its own mutex, one-way lock ordering (`actor.mu → sink.mu`, actor.go:1288-1309).
**Example** ([VERIFIED: internal/session/actor.go:1199-1277] — `applySnapshot` folds ~15 document fields into actor state; `refreshACBlocklist` recompiles the blocklist only when the document's pattern list changed — "one compile per edit, not per boundary"):

```go
// Source: actor.go fold shape (paraphrase of the pinned discipline)
a.opts.AutoCorrectBlocklist = snap.Autocorrect.AppsBlocklist
a.refreshACBlocklist(snap.Autocorrect.AppsBlocklist) // diff-gated work
// a11y equivalent: a.a11ySink.Apply(snap.A11y) — sink diffs against last-applied,
// runs the subprocess chain on its own goroutine/serialization
```

### Pattern 5: golden-style generation gate (stream 3)

**What:** regenerate-and-compare, the `tidy-diff` shape, so a stale SKILL.md fails the ordinary test gate.
**Precedents in-repo:** [VERIFIED: mise.toml:35-38 (`[tasks.ci]` aggregates mise tasks), mise.toml:214-216 (`[tasks.dictgen-regen]` = `go generate ./layouts`, dev-only generation precedent), pr-sanity.yml runs `mise run build/vet/lint/test/tidy-diff`]. The gate can be (a) a Go test that runs the generator over CONFIG.md and compares against the committed SKILL.md region, or (b) a mise task + `git diff --exit-code` CI step. Option (a) needs no workflow change and is lint-covered — recommended.

### Anti-Patterns to Avoid

- **Setting the key on every fold/reload:** dconf write churn. Read-verify-then-set, act only on effective diff (Pattern 4).
- **Interpolating config list entries into argv:** the app list selects WHICH fixed actions run; it never becomes an argument (the T-05-02-01/ASVS V5 argv-render discipline, [CITED: internal/install/install.go:121-123 comment — "user line never reaches `gsettings set`"]).
- **Reverting the key when an app leaves the list:** D-8-4 forbids it — manual vs goswitch-enabled is indistinguishable.
- **Writing .desktop Exec lines from user config:** if per-app overrides ever ship, the Exec template must be fixed — a user-controlled `apps` entry must never compose a command line.
- **Putting the reconciler under the actor mutex:** subprocess side-effects on the keystroke hot path violate the WR-01 discipline (subprocess sequences run outside the mutex).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Setting/reading GNOME settings | dconf D-Bus writer | `gsettings` CLI via os/exec | STACK decision; CLI handles dconf backend details, zero deps |
| Setting a11y-bus IsEnabled | raw a11y-socket protocol | godbus `Properties.Set` on `org.a11y.Bus` (session-bus address lookup like the indicator does) | Live-verified one call; godbus already a dep |
| YAML section decode/strictness | custom defaults overlay | existing `Load` (`KnownFields(true)`) + zero value off | The 06-04 RED-stub precedent: sections' strict decode and last-good reload need ZERO new code in load.go/watch.go [VERIFIED: STATE.md Phase-6 note + load.go:27-39] |
| Regex safety | custom pattern sanitiser | stdlib `regexp` (RE2) + Compile-at-Load + ceiling | RE2 is linear-time by construction |
| Config persist (menu/installer writes) | new writer | `internal/config/writer.go` Node round-trip, temp+rename, 0600 | Already proven for the two toggles |
| SKILL.md sync | manual proofreading | generator + diff gate | D-8-10: рассинхрон невозможен |
| a11y-tree probes for manual verification | new tooling | `test/e2e/focus_helper.py` subcommands / 10-line python3-gi walk | The stand's oracle exists |

**Key insight:** every mechanism this phase needs already has an in-repo precedent — the only genuinely new knowledge is the *desktop semantics* (what the switches actually do), which this research establishes live.

## Runtime State Inventory

Rename/refactor applies to the docs restructure (README.md language swap) and, more importantly, this phase *mutates desktop state*. Explicit inventory:

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | No databases. Config docs live in git. | none |
| Live service config | **Desktop runtime state on the target (live-verified 2026-10-05):** `toolkit-accessibility` = **false**; `org.a11y.Status.IsEnabled` = **false**; running ZCode (Electron 41) a11y tree = 2 nodes, 0 text inputs (renderer a11y OFF — consistent with «ZCode молчит»); Telegram snap not presenting a usable tree. | The daemon reconcile has real work to do on first enable; docs must describe restart semantics |
| OS-registered state | `~/.local/share/applications/` already carries user desktop files (`dev.zed.Zed.desktop`, `hermes.desktop`, `vk-messenger.desktop`, `org.telegram.desktop._3e485da34fc040f9218e3891ecde1e6c.desktop`, mimeapps) — a name-collision surface if per-app overrides ever ship. `install-state.json` exists on installed machines WITHOUT the new a11y field. | Uninstall restore must treat a missing a11y field as "nothing to revert" (WARN + skip), never guess |
| Secrets/env vars | none new | none |
| Build artifacts | none (no generated binaries change; `skillgen` is a dev tool) | none |
| Docs-file rename surface | README.md becomes Russian, README.en.md becomes English (D-8-7). GitHub renders README.md as the landing page — links from SPEC/CONFIG to «README» sections must be re-pointed; dependabot/SECURITY badges unaffected. | Section-anchored links audit in stream 2 |

## Common Pitfalls

### Pitfall 1: expecting live effect on running apps
**What goes wrong:** flipping the key (or IsEnabled) while Chromium/Electron/GTK apps run changes nothing in their trees; the feature looks broken.
**Why it happens:** Chromium's `ShouldEnableAccessibility()` is a one-shot startup check guarded by a static flag — no PropertiesChanged watch [VERIFIED: chromium.googlesource.com …/atk_util_auralinux.cc (main), fetched 2026-10-05]. Live confirmation: ZCode's subtree stayed at 2 nodes after both flips.
**How to avoid:** reconcile sets persistent state; docs and any apply-time logging say «приложения подхватят при следующем запуске».
**Warning signs:** manual testing claims «не работает» right after enabling.

### Pitfall 2: conflating the two switches (key vs IsEnabled)
**What goes wrong:** assuming `toolkit-accessibility` drives `org.a11y.Status.IsEnabled` (or vice versa) and verifying the wrong one.
**Why it happens:** historically coupled via at-spi2's bus launcher; on this GNOME 46 the coupling did not fire on a mid-session flip (launcher running since session start) [VERIFIED: live, 2026-10-05 — key=true left IsEnabled=false].
**How to avoid:** verify the key with `gsettings get`; verify IsEnabled with `busctl --user call org.a11y.Bus /org/a11y/bus org.freedesktop.DBus.Properties Get ss org.a11y.Status IsEnabled`; treat them as independent levers.
**Warning signs:** a green key-check with still-silent apps (Chromium consults BOTH at startup, exact composition unquoted — see Assumptions A1).

### Pitfall 3: probe context lies (AppArmor label)
**What goes wrong:** a11y-tree probes launched from inside ZCode (this research session's context) see partial trees: snap apps (telegram-desktop, chromium, snap-store) deny `org.a11y.atspi.Cache.GetItems` to the zcode-labeled sender, and even zenity's entry widget was missing from a dialog walk.
**Why it happens:** AppArmor mediation on the a11y bus discriminates by peer label; snap profiles restrict foreign senders.
**How to avoid:** run tree probes from the normal unconfined user context (the e2e stand runner does); never from a zcode-labeled shell.
**Warning signs:** dbind-WARNING lines about AppArmor policy in probe output.

### Pitfall 4: dconf churn from the reconcile
**What goes wrong:** `gsettings set` fired on every config fold (keystroke-paced) spams dconf/journald.
**How to avoid:** diff-gate in the fold (last-applied desired state), apply asynchronously, read-verify-then-set (Pitfall-2's `gsettings get` doubles as the churn guard).
**Warning signs:** journal flooded with apply records after a typing burst.

### Pitfall 5: the snapshot/revert seam mismatch (daemon applies, installer reverts)
**What goes wrong:** D-8-3 makes the *daemon* apply the key, D-8-4 makes *goswitchctl uninstall* revert it — but install-state.json was written by an older install (no a11y field), so uninstall has nothing to restore and the temptation is to "restore false".
**How to avoid:** install extends `saveState` with the pre-install key value (the canonical flow runs install before the daemon ever applies); uninstall restores only if the field exists, else WARN + skip. Never fabricate a default.
**Warning signs:** uninstall on an older install resetting a manually-enabled key.

### Pitfall 6: .desktop override staleness (if per-app overrides ever ship)
**What goes wrong:** a frozen user-override `Exec=` masks upstream changes (binary path/flags move on app update) and the app silently stops launching correctly.
**Why it happens:** the override shadows the upstream desktop file permanently.
**How to avoid:** fixed templates, a marker comment identifying goswitch-owned files, and documented removal; prefer NOT shipping overrides in v1 (research recommendation — the global key covers the owner's case).
**Warning signs:** an app failing to launch after its update.

### Pitfall 7: backwards-compat of the config contract
**What goes wrong:** assuming existing configs need edits.
**How to avoid:** the section must decode zero-value-off (the Autocorrect D-54 precedent): documents without it load unchanged, disabled. Note CONFIG.md's «exactly six sections» statement (CONFIG.md:14-15) must become seven; SPEC has no count statement (grep-verified), only §4's contract — spec-delta territory (D-55).
**Warning signs:** a user config suddenly refusing to load after upgrade — would be a defect, not a migration.

## Code Examples

### The Chromium startup check (authoritative, quoted from fetched source)
```cpp
// Source: chromium.googlesource.com/chromium/src/+/main/ui/accessibility/platform/atk_util_auralinux.cc
// env vars checked ("1" → true, "0" → false, short-circuit):
//   "ACCESSIBILITY_ENABLED", "GNOME_ACCESSIBILITY", "QT_ACCESSIBILITY"  (+ CHROME_HEADLESS skip)
// then a one-shot D-Bus Get: org.a11y.Bus /org/a11y/bus org.freedesktop.DBus.Properties
//   Get "org.a11y.Status" "IsEnabled"
// then a direct GSettings read:
//   ui::GSettingsNew("org.gnome.desktop.interface"); g_settings_get_boolean(..., "toolkit-accessibility")
// NO PropertiesChanged subscription: `static bool initialized = false; if (initialized || !ShouldEnableAccessibility()) return;`
```

### The IsEnabled belt (live-verified command form)
```sh
# Source: live probe on the target, 2026-10-05 (readback confirmed; no root)
busctl --user call org.a11y.Bus /org/a11y/bus org.freedesktop.DBus.Properties \
  Set ssv org.a11y.Status.IsEnabled b true
# godbus equivalent: a Properties.Set on the well-known name org.a11y.Bus (session bus)
```

### Desktop-file override mechanics (if ever needed)
```ini
# Source: /usr/share/applications/zcode.desktop (verbatim Exec line, live)
#   Exec=/opt/ZCode/zcode %U
# user override at ~/.local/share/applications/zcode.desktop (same desktop-file-id wins):
[Desktop Entry]
Type=Application
Name=ZCode
Exec=env ACCESSIBILITY_ENABLED=1 /opt/ZCode/zcode --force-renderer-accessibility %U
Icon=zcode
StartupWMClass=ZCode
```
[CITED: wiki.archlinux.org/title/Desktop_entries — XDG user-dir precedence; forum.snapcraft.io/t/overriding-desktop-files-on-ubuntu-snaps — same-basename override of /var/lib/snapd/desktop/applications entries, with reported DE quirks. Live: XDG_DATA_DIRS puts `/var/lib/snapd/desktop` last; `~/.local/share/applications` already in active use.]

### SKILL.md shape
```markdown
---
name: goswitch-config
description: Настройка goswitch — все секции config.yaml, goswitchctl, трей-меню; правь конфиг по этому справочнику
---
<!-- frame: процедуры (автокоррекция, blocklist, хоткеи, звуки, a11y-магия), диагностика -->
<!-- goswitch-config:generated BEGIN — from docs/CONFIG.md key table, do not edit -->
…generated…
<!-- goswitch-config:generated END -->
```
[CITED: agentskills.io/specification — frontmatter required fields `name` (lowercase, ≤64 chars) + `description`; progressive loading = only frontmatter preloads. In-repo precedent: `.zcode/skills/gsd-ns-context/SKILL.md` frontmatter with `allowed-tools`.]

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| toolkit-accessibility ⟶ org.a11y.Status.IsEnabled considered one mechanism | Decoupled on GNOME 46: key is persistent + read directly by Chromium at startup; IsEnabled is session-scoped, settable via D-Bus | Verified live 2026-10-05 | The daemon sets the key; the belt is optional and needs no revert discipline |
| GTK apps gated on the key | GTK4 (GTE) and GTK3 (zenity) register trees with the key false on GNOME 46 [VERIFIED: live walk — GTE 22 nodes, zenity dialog tree, both with key=false] | current 46 packages | The magic matters for Chromium-family (and possibly Qt), not for GTK on this desktop |
| `app.setAccessibilitySupportEnabled` as the Electron lever | macOS/Windows only per docs; Linux uses Chromium's startup checks + `--force-renderer-accessibility` | Electron docs today | No in-app API path for the daemon — external levers only |
| «enable globally (then restart the application)» docs remedy | Confirmed correct, with the refinement that the lever Chromium reads includes the gsettings key directly | this research | The existing remedy text survives the docs sweep; add the restart emphasis |

**Deprecated/outdated:**
- The AskUbuntu-era claim that GNOME's a11y key flips `org.a11y.Status.IsEnabled` live — did not hold on this machine.
- Anything expecting `GNOME_ACCESSIBILITY` env to update mid-session (session-start semantics only, [ASSUMED]).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Chromium's boolean composition of the three startup checks (env / IsEnabled Get / GSettings key) is OR-like — setting the key alone suffices at next start | Standard Stack / Pitfall 2 | If AND-like, the daemon must also set IsEnabled (belt becomes mandatory) — cheap; the owner's manual key+restart precedent suggests OR |
| A2 | `ACCESSIBILITY_ENABLED`/`GNOME_ACCESSIBILITY` env vars are set by the session at start when a11y is on | State of the Art | None for design (we do not rely on env) |
| A3 | Snap-confined Chromium may be unable to read user dconf (needs the desktop/dconf plug); the IsEnabled D-Bus belt would still reach it | Alternatives Considered | Only affects snap-Electron coverage; belt mitigates |
| A4 | zenity's missing entry widget in the tree walk was a probe-context artifact (zcode label), not a real effect of the key | Pitfall 3 | Documentation-level; GTK registration is key-independent either way |
| A5 | `org.a11y.Status.IsEnabled` is session-scoped (resets at session restart) | Pitfall 5 | None — the key stays the persistent truth; belt needs no snapshot |
| A6 | `QT_ACCESSIBILITY=1` env enables Qt a11y (Telegram) at its start | Code Examples | Only if snap/Qt overrides get implemented; live check belongs to that work |
| A7 | ZCode's manual precedent = «key set + ZCode restarted» (the exact manual steps are the owner's memory; todo says «как сделали вручную») | Summary | If the owner did something else (e.g. flag), the mechanism map still holds — the key is source-verified as a Chromium startup input regardless |

## Open Questions (RESOLVED — dispositions locked by plan adoption, 2026-10-05)

1. **Who snapshots the pre-apply key value — install or the daemon?**
   - What we know: install-state.json is the restore contract; daemon applies (D-8-3); uninstall reverts (D-8-4).
   - What's unclear: the canonical flow (install before daemon) favors install; a daemon-maintained state file covers installs older than this field.
   - Recommendation: extend `saveState` (Pattern 3) AND treat a missing field as WARN+skip in uninstall. Planner decides the exact split.
   - RESOLVED (adopted by plan 08-04): install extends `saveState` — the pre-apply key value is snapshotted into install-state.json at install; uninstall restores only-if-present (missing field = WARN + skip, never "restore false"); the daemon keeps no state file of its own.
2. **Does the v1 diff include per-app `.desktop` overrides?**
   - What we know: D-8-6 says «опциональные»; the global key covers the ZCode case after app restart; overrides carry the staleness pitfall.
   - Recommendation: document the mechanism (Code Examples), keep the v1 diff to the global key + belt. Owner confirms at verify gate.
   - RESOLVED (adopted by plans 08-01 + 08-06): the v1 diff carries the global key + belt only; the §4 revision (08-01) fixes "mechanism documented, outside the v1 diff"; docs/CONFIG.md (08-06) carries the manual override recipe with staleness warnings — no automation in v1.
3. **Ship the IsEnabled belt?**
   - What we know: one godbus call, no root, live-verified; session-scoped so no revert needed; helps IsEnabled-only consumers and (assumed) snap clients.
   - Recommendation: yes — it is one idempotent call with no revert debt.
   - RESOLVED (adopted by plan 08-03): the belt ships — exported NewDBusStatusSetter() godbus adapter sets org.a11y.Status.IsEnabled=true on every activation episode; session-scoped (A5), no revert debt.
4. **SKILL.md language.**
   - What we know: D-8-9 makes CONFIG.md russian-first; the generated section inherits its source language.
   - Recommendation: russian-first, consistent with its generator source.
   - RESOLVED (adopted by plan 08-08): SKILL.md is russian-first; the generated region inherits its source language (docs/CONFIG.md after 08-06); recorded in the plan's assumptions.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| gsettings CLI | stream 1 apply/verify | ✓ | GNOME 46.0 | direct dconf D-Bus (rejected by STACK) |
| busctl (systemd) | IsEnabled belt + manual probes | ✓ | systemd 255 | godbus call |
| godbus/dbus v5 | belt (product code) | ✓ | v5.2.2 in go.mod | — |
| python3-gi + Atspi | manual/executor tree probes | ✓ (system `/usr/bin/python3` ONLY — `linuxbrew` python3 lacks gi) | 3.48.2 / 2.52.0 | focus_helper.py subcommands |
| mise | task/gates | ✓ | per mise.toml | — |
| ZCode (Electron 41) | the precedent app | ✓ running, `/opt/ZCode/zcode`, `/usr/share/applications/zcode.desktop` | 3.14.4 / Electron 41.0.3 | — |
| Telegram | snap candidate app | ✓ snap 7.2.9 | 7.2.9 | — |
| google-chrome | chromium-family reference | ✓ (phase-2 verified on this desktop) | 153 | chromium snap 153.0.8010.47 |
| flatpak | override path variant | ✗ no apps installed | — | not needed |
| go / golangci-lint | all streams | ✓ via mise | 1.27.1 / pinned | — |

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** flatpak (no apps — nothing to cover).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `-race`; golden files (CONVENTIONS) |
| Config file | none needed — package tests |
| Quick run command | `mise run test` (= `go test -race -count=1 ./...`) |
| Full suite command | `mise run ci` (build, vet, lint, test, tidy-diff) |

### Phase Requirements → Test Map
(No formal REQ IDs — backlog-derived; mapped to the three streams.)

| Req (stream) | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| S1 schema | `a11y` section: zero value off; strict decode; regex compile-at-Load; empty/blank pattern refused; ceiling 64; errors name field+index | unit | `go test ./internal/config/ -run TestA11y` | ❌ Wave 0 |
| S1 reconcile | idempotent apply (read-verify-then-set), diff-gating, one-WARN episodes, no-op when section inactive | unit (Runner seam) | `go test ./internal/a11y/` | ❌ Wave 0 |
| S1 revert | installState carries the key value; uninstall restores only-if-present; missing field → WARN skip | unit | `go test ./internal/install/ -run TestA11y` | ❌ Wave 0 |
| S2 docs | bilingual section-header parity (optional gate), links resolve | unit or mise task | `mise run docs-gate` (planner's call) | ❌ Wave 0 (optional) |
| S3 skill | generated SKILL.md region matches regeneration from CONFIG.md | unit (golden) | `go test ./internal/skillgen/` (or cmd/skillgen) | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `mise run test` (plus `mise run lint`)
- **Per wave merge:** `mise run ci`
- **Phase gate:** full `mise run ci` green; e2e matrix NOT extended (owner discretion; the a11y path is outside the correction path)

### Wave 0 Gaps
- [ ] `internal/config/a11y_test.go` (or section additions to config_test.go) — schema corpus
- [ ] `internal/a11y/` reconciler corpus with the Runner seam
- [ ] `internal/install` snapshot/restore extension tests
- [ ] skillgen golden test
- Framework install: none

## Security Domain

security_enforcement enabled (config), ASVS L1.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — (no auth surface) |
| V3 Session Management | no | — |
| V4 Access Control | yes (light) | user-scoped state only ($HOME, user gsettings/dconf, user bus); no root anywhere |
| V5 Input Validation | yes | strict decode (`KnownFields`), RE2 compile-at-Load, empty-pattern refusal, ceiling 64, range errors naming field+index (the D-33/D-53 canon) |
| V6 Cryptography | no | — |
| V14 File Operations | yes (if overrides ship) | 0600, atomic temp+rename (writer.go precedent), fixed templates only |
| V1 (surface) — command construction | yes | argv-render discipline: config list never enters argv (T-05-02-01 precedent) |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Config content reaching `gsettings` argv | Tampering | fixed argv shapes; list selects actions, never arguments |
| a11y-магия broadens what a11y clients can observe (typed text in enabled apps) | Information Disclosure | opt-in via config section (user's explicit choice), docs state the tradeoff; role-gate/privacy contract untouched |
| Giant app list / regex catastrophic backtracking | DoS | ceiling 64; RE2 linear-time |
| Stale goswitch-owned desktop override hijacking app launches | Tampering/Elevation (local) | marker comment + fixed Exec template + removal path; recommend not shipping overrides in v1 |
| Uninstall reverting a manually-set key | DoS (user trust) | snapshot only-if-present restore (Pitfall 5) |

## Sources

### Primary (HIGH confidence)
- Live target system, 2026-10-05: `gsettings get/range`, `busctl` org.a11y.Status reads/sets, AT-SPI tree walks (zcode 2 nodes; GTE 22 nodes; zenity dialog; baseline 18 apps), process inventory (`/opt/ZCode/zcode` Electron 41.0.3, `at-spi-bus-launcher --launch-immediately`, `at-spi2-registryd --use-gnome-session`), desktop-file inventory (`/usr/share/applications/zcode.desktop` Exec verbatim; `/var/lib/snapd/desktop/applications/`), `snap list`, XDG_DATA_DIRS, AppArmor denial transcripts
- Chromium source, `ui/accessibility/platform/atk_util_auralinux.cc` (main branch, fetched via WebFetch 2026-10-05) — startup-only check, env var list, IsEnabled Get, direct GSettings key read
- Electron official docs, `electronjs.org/docs/latest/api/app` (fetched) — `setAccessibilitySupportEnabled` macOS/Windows-only; AT-detection semantics
- In-repo code (Read this session): internal/config/config.go (schema/ceilings/validate — quotes verbatim above), load.go, watch.go, writer.go, internal/session/actor.go (applySnapshot fold), internal/sound/sound.go, internal/clipboard/clipboard.go, internal/install/install.go (state discipline), internal/activate/activate.go
- In-repo docs (Read): docs/CONFIG.md (full), README.md/README.ru.md (structure + heads), mise.toml, .github/workflows/pr-sanity.yml, .planning/todos/pending/a11y-magic-apps-list.md, 07-UAT.md (remedy record), 08-DISCUSSION-LOG.md

### Secondary (MEDIUM confidence)
- [agentskills.io/specification](https://agentskills.io/specification) + [platform.claude.com Agent Skills overview](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview) — SKILL.md format
- [wiki.archlinux.org/title/Desktop_entries](https://wiki.archlinux.org/title/Desktop_entries), [forum.snapcraft.io/t/overriding-desktop-files-on-ubuntu-snaps](https://forum.snapcraft.io/t/overriding-desktop-files-on-ubuntu-snaps) — override precedence, snap quirks
- freedesktop AT-SPI2 docs (via search) — org.a11y.Status semantics («apps should not use IsEnabled to gate support»)
- Debian at-spi2 ships qt-a11y env snippets (via search) — Qt a11y lever context

### Tertiary (LOW confidence)
- Search-level claims about at-spi-bus-launcher↔gsettings sync behavior (the live probe on THIS machine is the controlling evidence)

## Metadata

**Confidence breakdown:**
- a11y mechanics: HIGH — live-probed on the exact target desktop + authoritative Chromium source read; the two switches' behaviors are direct observations
- Code/config patterns: HIGH — read in-repo with line citations
- Docs facts: HIGH — read in-repo (sizes, missing sections, language)
- Snap/Qt and env-var periphery: MEDIUM/LOW — flagged in Assumptions A2/A3/A6

**Research date:** 2026-10-05
**Valid until:** ~2026-11-04 (stable: desktop semantics could shift only with distro updates)
