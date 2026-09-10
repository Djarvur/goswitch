# Phase 1: ADR-пакет и каркас IBus-движка - Research

**Researched:** 2026-09-10
**Domain:** IBus input-method engine (Go/godbus) for GNOME Wayland — wire protocol, D-01 switching experiment, xkb table generation, hotkey FSM, e2e skeleton
**Confidence:** HIGH (live probes on the actual target machine this session + goibus source read verbatim + GNOME Shell 46 source read)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

### Механизм переключения раскладки (Decision #1)
- **D-01:** Механизм выбирается экспериментом в фазе 1, до кода коррекции/переключения. Проверочный вопрос: «может ли демон приказать системе GNOME переключить активный источник ввода?» Да → two-engine (два источника goswitch-en/goswitch-ru с XKB-раскладками us/ru; нативный значок, Super+Space, MRU). Нет → внутренний флип (один движок с `<layout>us</layout>`, EN/RU — внутренний режим, как в работающем прототипе владельца). Результат фиксируется ADR. — **Reversibility:** reversible — выбор происходит в самом начале фазы, до реализации зависимого кода
- **D-02:** Обходные пути (хаки) для программного переключения источника допустимы при условиях: без root, без правки системных файлов, работает на Ubuntu 24.04 GNOME Wayland, переживает перезагрузку. Пример кандидата: имитация нажатия Super+Space. — **Reversibility:** reversible
- **D-03:** Если победит внутренний флип (вариант B) — «один хозяин»: в источниках ввода остаётся только goswitch, системные us/ru убираются; Right Shift — единственный переключатель; Super+Space ничего не переключает. Двойное состояние исключено. — **Reversibility:** costly

### Тайминг тапов Right Shift
- **D-04:** Классическая схема с ожиданием: после тапа Right Shift демон ждёт окно различения, и только потом выполняет действие (одиночный → переключение, двойной → слово, тройной → фраза). Владелец осознанно отклонил Caramba-схему «мгновенный свитч на первый тап» и разные клавиши. — **Reversibility:** costly
- **D-05:** Окно различения тапов по умолчанию 300 мс, настраивается в YAML. NFR о скорости переформулируется (spec-delta к §5 спеки): «действия без ожидания — <50 мс; действия Right Shift — не дольше окна различения». — **Reversibility:** reversible

### Сброс буфера (CORR-09)
- **D-06:** «Клик мыши» убирается из триггеров сброса буфера (невыполнимо на уровне метода ввода — клик не доходит до IBus). Вместо него два механизма: (1) обязательная сверка перед коррекцией — буфер должен совпадать с текстом непосредственно перед курсором (surrounding text); не совпало → молча ничего не делать («abort, не мусорить»); (2) best-effort сброс буфера по скачку позиции курсора в surrounding text там, где приложение его полностью сообщает. Оформить как spec-delta к §4.3 спеки. — **Reversibility:** reversible

### Открытые решения (требуют владельца)
- **OPEN-01 (MACR-01):** Замена Super+Буква → Ctrl+Буква по приложениям — решение НЕ принимать без владельца. Рекомендация исполнителя (не решение): вынести в v2 через spec-delta, в документацию добавить рецепт настройки через keyd. Вопрос вернуть на утверждение ADR-пака (критерий фазы 1, гейт M0 «ревью владельца»). До решения планировщику MACR-01 не реализовывать и не планировать код.

### Claude's Discretion
- Формат и расположение файлов ADR (рекомендация: `docs/adr/ADR-00N-*.md`, нумерация сквозная) — на исполнителя, состав ADR-пака фиксирован критериями фазы.
- Порядок и структуру планов фазы (сначала чистые пакеты layouts/FSM, затем адаптер engine/, затем e2e-скелет — согласно выводам исследования).
- Технические детали capability ladder замены (DeleteSurroundingText → Backspace×N → clipboard, по поддержке клиента) — предписаны исследованием, реализации в фазе 2.

### Deferred Ideas (OUT OF SCOPE)
- Рецепт «Super+Буква → Ctrl+Буква через keyd» в документацию — кандидат в фазу 4 (документация), часть решения OPEN-01.
- OSD/уведомление при переключении раскладки — не запрашивалось, в объём не входит.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CORR-08 | Таблицы ЙЦУКЕН↔QWERTY incl. знаки `[ ] ; ' , . /`, генерируются `go:generate` из xkb symbols | §"Layout table generation pipeline": default ru variant IS winkeys on this machine (verified live); us `basic` section verified; full effective mapping table below; generator design + golden corpus |
| INTEG-01 | `goswitchd` как IBus engine (`_RegisterComponent`, `commit_text` injection) | §"IBus wire surface": complete verified wire map (address → Dial/Auth/Hello → RequestName → Factory → RegisterComponent); `commit_text` path lands via `CommitText(v)` signal (emitter included in adapter; correction itself is Phase 2) |
| INTEG-02 | Совместимость с keyd/xremap: транзит, без EVIOCGRAB | Engine returns `false` from `ProcessKeyEvent` for everything in Phase 1 (pure observer); no evdev/uinput import in product code (verifiable: `go vet`/grep — no `syscall.EVIOCGRAB`); live: keyd installed but service failed/disabled on target |
| INTEG-03 | Без root поверх дефолтного IM-стека Ubuntu 24.04 | Programmatic `RegisterComponent` needs no component XML and no root (prototype-proven on this machine); /dev/uinput 0660 root:input + user in `input` group (verified live) |
| INTEG-04 | Перерегистрация после рестарта ibus-daemon (ibus#2910) | §"Reconnect/re-register loop": address file re-discovery (new socket path per daemon generation — 15 stale sockets observed), closed-signal-channel detection, backoff, re-export Factory |
| INTEG-05 | recover-шим; паника не роняет движок/ввод | §"Pitfalls": godbus dispatches handlers on goroutines — one panic = process death; recover-wrapper pattern + variant-assertion guards |
| TEST-01 | Headless CI unit/golden | §"Validation Architecture": pure packages (layouts golden, FSM corpus) run with plain `go test`, zero D-Bus |
| TEST-02 | e2e-стенд: физическая инжекция (ydotool/uinput) → чтение результата | §"e2e skeleton": ydotool 0.1.8 verified installed, direct-uinput mode (no ydotoold); Phase 1 reads the *structured log* (AT-SPI text readback is Phase 2/TEST-04) |
| TEST-03 | e2e сам активирует целевое окно | §"e2e skeleton": AT-SPI `Component.grabFocus` via `/usr/bin/python3` gi helper (wmctrl NOT installed; X11-only anyway) |
| INST-04 | Структурные логи с уровнями; debug-режим с трассировкой клавиш | §"Structured logging design": `log/slog` JSON handler (API verified locally on go1.27.1), LevelVar, key-trace behind explicit opt-in (security) |
</phase_requirements>

## Summary

Phase 1's research gaps were all closable on this machine, because this machine **is** the target: the IBus 1.5.29-rc2 daemon is running, its address file and both owner prototypes are present, GNOME Shell is 46.0, and ydotool 0.1.8 is installed with the user already in the `input` group. The single highest-risk unknown — the exact GVariant wire surface for a pure-Go engine — is now fully mapped: I read the complete sarim/goibus source verbatim (component/engineDesc/text/factory/engine/bus/common .go), cross-checked every method and signal name against strings in the installed `libibus-1.0.so.5.0.529` and the installed headers `/usr/include/ibus-1.0/*.h`. One planning-relevant correction: the D-Bus method is **`RegisterComponent`** (no underscore; verified in libibus strings) — planning docs writing `_RegisterComponent` mean the same call.

The D-01 experiment (can a daemon command GNOME to switch the active input source?) now has a decisive protocol. I read GNOME Shell 46.0 `js/ui/status/keyboard.js` and `js/misc/ibusManager.js` directly: **keyboard.js 46 has NO handler for external engine changes** — `activateInputSource` (the only path that applies the XKB layout) fires solely from `InputSource.activate()` (Super+Space / MRU / popup / Alt+Shift), and ibusManager only closes the candidate popup on `global-engine-changed`. So calling IBus `SetGlobalEngine` alone can never satisfy two-engine; the experiment must test (in order): `gsettings set … current N`, synthetic Super+Space/Alt+Shift_L via ydotool (verified live bindings `['<Alt>Shift_L','<Super>space']`), `org.gnome.Shell.Eval` (dev-only: gated behind unsafe mode since GNOME 41), and a micro shell-extension as the clean fallback — with the owner-proven internal flip as Option B if all fail. A second corrective finding: the **default `ru` variant on Ubuntu 24.04 IS `winkeys`** (verified `default` directive in `/usr/share/X11/xkb/symbols/ru:9-10`), so the owner prototype's hand-written digit-row table was right for this machine — but generation from xkb still wins because it pins the table to evidence and covers punctuation levels in one pass.

The pure-Go side is conventional and low-risk: `go:generate` pipeline (parse `symbols/us` basic + `symbols/ru` default-with-includes + `ibuskeysyms.h` for keysym names→values, embed a small Cyrillic keysym→Unicode table, emit `layouts/tables.go`, pin with golden tests); a classic-with-waiting tap FSM (D-04/D-05: decisions fire at 300 ms window expiry after the last tap, modifier-use discrimination from the prototype's `saw_key` logic, fake-clock unit corpus); `log/slog` JSON logging with opt-in key trace. The environment audit found three practical snags for the plan: `wmctrl` is not installed (use AT-SPI grabFocus), `python3` resolves to linuxbrew in PATH (e2e helpers must call `/usr/bin/python3` where gi lives — verified importable), and `golangci-lint` is not installed (CI gap).

**Primary recommendation:** Sequence waves as pure packages → engine adapter + reconnect → D-01 experiment (uses the skeleton itself: register `goswitch-en`+`goswitch-ru`, observe FocusIn/FocusOut) → e2e skeleton → ADR pack (ADR-001 last, filled from the experiment log; MACR-01/OPEN-01 as an owner checkpoint, no code). Transience rule for Phase 1: the engine returns `false` for every key (pure observer) — transit by construction, desktop typing cannot break.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Key interception | IME layer (IBus engine via godbus, private socket) | — | Only the active input source's engine receives `ProcessKeyEvent`; evdev/compositor layers are off-limits (no-grab axiom) |
| Text injection (`commit_text`) | IME layer (engine signal `CommitText`) | — | The only injection primitive; uinput forbidden in product |
| Layout switch execution | GNOME session layer (gsettings / Shell extension / mutter keybinding) | IME layer (internal mode flip — Option B) | XKB layout is applied by mutter on `activateInputSource`; engine cannot apply XKB itself. D-01 decides which mechanism |
| Layout switch decision (tap FSM) | Daemon pure logic (`internal/hotkey`) | session actor | Must be headless-testable; no D-Bus knowledge |
| Layout tables | Pure data package (`layouts/`, generated) | — | Generated from system xkb files; golden-tested |
| Structured logging | Daemon process (`log/slog` → stderr → journal) | — | stdlib; zero deps; e2e parses the JSON stream |
| e2e key injection | Kernel/uinput layer (ydotool — test stand ONLY) | — | Physical path = same as live keyboard; compositor-independent |
| e2e window activation | Accessibility layer (AT-SPI `Component.grabFocus`) | — | TEST-03; wmctrl absent and X11-only |
| Autostart/supervision | systemd user manager | ibus-daemon (engine spawning not used — programmatic registration) | `org.freedesktop.IBus.session.GNOME.service` exists as ordering anchor (verified active) |
| IBus registration persistence | Daemon runtime (re-register loop) | — | ibus#2910: registrations die with the daemon by design |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/godbus/dbus/v5` | v5.2.2 (latest on proxy.golang.org — verified this session) | All D-Bus: IBus private socket (service side), later session-bus ctlsvc | Only maintained native Go D-Bus; proven against the IBus socket by ibus-bamboo in production [VERIFIED: go proxy + goibus source read] |
| `gopkg.in/yaml.v3` | v3.0.1 (latest — verified this session) | Config parsing (minimal use in Phase 1: FSM window default source) | Project-locked decision (STACK.md); zero deps |
| `github.com/fsnotify/fsnotify` | v1.10.1 (latest — verified this session) | Config hot reload | **Phase 3 concern** — do NOT add in Phase 1 unless config file watching lands early; keep go.mod deps to what Phase 1 uses |
| `log/slog` (stdlib) | go1.27.1 local toolchain; `go 1.23` in go.mod | Structured logging | INST-04; zero deps; JSON handler verified via `go doc log/slog.NewJSONHandler` |
| In-repo `engine/` adapter | ~1k LOC, zero external deps beyond godbus | IBus wire protocol | goibus has NO LICENSE (legal blocker); its source is the read-only behavioral reference — wire map reproduced below |

### Supporting (dev/e2e only)
| Tool | Version | Purpose | When to Use |
|------|---------|---------|-------------|
| ydotool | 0.1.8-3build1 (dpkg verified) | Physical-path key injection (e2e) | Direct `/dev/uinput` client — **no ydotoold daemon on 24.04**; prints a benign "ydotoold backend unavailable" notice; `/dev/uinput` writable by `nil` (group `input`, 0660 — verified live) |
| `/usr/bin/python3` + gi | 3.12.3, `gi` import OK (verified live) | AT-SPI helper: focus window, (Phase 2) text readback | MUST be the absolute path — `python3` in PATH resolves to linuxbrew, which lacks gi |
| `gdbus`/`busctl`/`dbus-monitor` | present (dbus-monitor/gdbus resolve to linuxbrew copies; `/usr/bin/busctl` exists) | Manual wire verification during development | Verify RegisterComponent traffic, NameOwnerChanged, etc. |
| `ibus` CLI (1.5.29) | `ibus engine` → `xkb:us::eng` (verified live); `ibus list-engine`, `ibus address`, `ibus restart` | Dev-loop state inspection and restart simulation | e2e preflight + resilience cases |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| sarim/goibus as dependency | in-repo adapter | goibus: no LICENSE — cannot ship in MIT project; source is legitimate reference (read verbatim this session) |
| gsettings subprocess for switching | direct dconf D-Bus writes | Not a stable contract; subprocess is dependency-free (project decision, STACK.md) |
| AT-SPI via python3-gi helper | pure-Go godbus AT-SPI client | AT-SPI is introspectable, but python helper is ~30 LOC and spec-prescribed; defer Go client |
| ydotool 0.1.8 | own ~150-LOC Go uinput injector | Only if ydotool timing proves flaky; contained fallback |

**Installation (Phase 1 scope):**
```bash
go get github.com/godbus/dbus/v5@v5.2.2
go get gopkg.in/yaml.v3@v3.0.1   # only when config parsing lands
# e2e stand packages: already installed on target (ydotool, python3-gi, gir1.2-atspi-2.0)
```

## Package Legitimacy Audit

> The seam's `package-legitimacy check` supports npm/pypi/crates only; for Go the authoritative registry is proxy.golang.org, queried directly this session.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| github.com/godbus/dbus/v5 | Go proxy (verified: latest = v5.2.2) | ~10 yrs | de-facto standard | github.com/godbus/dbus (BSD-2) | OK (registry-verified; also project-locked in STACK.md) | Approved |
| gopkg.in/yaml.v3 | Go proxy (verified: latest = v3.0.1) | ~10 yrs | ubiquitous | go-yaml org | OK (project-locked) | Approved |
| github.com/fsnotify/fsnotify | Go proxy (verified: latest = v1.10.1) | ~10 yrs | ubiquitous | fsnotify org | OK (project-locked) | Approved |

**Packages removed due to [SLOP] verdict:** none.
**Packages flagged as suspicious [SUS]:** none.
**No new packages beyond the three project-locked ones are recommended for Phase 1.**

## Architecture Patterns

### System Architecture Diagram (Phase 1 slice)

```
                    e2e stand (test/e2e, Go + python3-gi helper)
                    │ 1. preflight: address file, /dev/uinput, engine listed,
                    │    gi import, target window launched
                    │ 2. AT-SPI Component.grabFocus → target window focused (TEST-03)
                    │ 3. ydotool type/key (physical path: /dev/uinput → kernel)
                    ▼
 KERNEL → [keyd/xremap — currently disabled] → mutter (XKB translate)
                    ▼
 GNOME session: input source = goswitch-en/goswitch-ru (('ibus',..) in gsettings)
                    ▼ activateInputSource: keyboardManager.apply(layout) + setEngine
 IBus daemon (private socket ~/.cache/ibus/dbus-*) 
                    ▼ Factory.CreateEngine → /org/freedesktop/IBus/Engine/...
 goswitchd  engine/ adapter ── EngineEvent ──► session actor ──► internal/hotkey FSM
   │  ProcessKeyEvent → return false (transit, Phase 1 observer)        │
   │  recover-shim on every exported handler                            ▼
   │  reconnect+re-register loop ◄── connection closed / ibus restart   decisions → slog JSON (DEBUG key trace)
   ▼
 structured log (stderr → journal / e2e temp file)  ◄── e2e asserts here (M1 gate)

 D-01 experiment branch (same process): register 2 engines →
   [1] gsettings set current N ─┐
   [2] ydotool key Super+Space ─┼─► GNOME switches source? → FocusOut(en)+FocusIn(ru) in log
   [3] Shell.Eval activate()   ─┤     + typed-text script check (AT-SPI/living eye)
   [4] SetGlobalEngine probe   ─┘     → ADR-001: two-engine vs internal flip (Option B)
```

### Recommended Project Structure (Phase 1 additions to the empty skeleton)

```
goswitch/
├── go.mod                      # go 1.23; deps: godbus only at first
├── cmd/goswitchd/main.go       # thin main: run(ctx, cfg) error (go-ultimate pattern)
├── engine/                     # IBus wire adapter — only package importing godbus
│   ├── address.go              # $IBUS_ADDRESS → ~/.config/ibus/bus/<machine-id>-unix-wayland-0
│   ├── conn.go                 # Dial/Auth/Hello, reconnect + re-register loop
│   ├── types.go                # Component/EngineDesc/Text marshalling (wire map below)
│   ├── factory.go              # CreateEngine dispatch
│   ├── engine.go               # per-IC object: exported methods, recover-shim, event funnel
│   └── keys.go                 # keyval consts + state masks (verified values below)
├── internal/hotkey/            # pure tap FSM (D-04/D-05) + corpus tests
├── internal/logging/           # slog setup, key-trace mode
├── layouts/
│   ├── generator/main.go       # go:generate parser (dev-side; not needed in CI)
│   ├── tables.go               # GENERATED, committed
│   └── tables_test.go          # golden tests (run in CI)
├── test/e2e/                   # Go runner + preflight + python3 focus helper
└── docs/adr/ADR-001..005-*.md  # ADR pack
```

### Pattern 1: The IBus wire surface (verified — port these field orders verbatim)

**Address discovery** [VERIFIED: live `cat ~/.config/ibus/bus/4c8b3c0710fb4959927ecf0d99b6baec-unix-wayland-0`]:
```
IBUS_ADDRESS=unix:path=/home/nil/.cache/ibus/dbus-UKuQiVpe,guid=c3984db66ba7f080f4a2179b6aa26f0d
IBUS_DAEMON_PID=433493
```
Filename = `<machine-id>-unix-wayland-0` (machine-id `4c8b3c0710fb4959927ecf0d99b6baec`; sibling file `-unix-0` also exists). Precedence: `$IBUS_ADDRESS` env → newest matching bus file (select by `WAYLAND_DISPLAY`/`DISPLAY` suffix). The socket path changes on every ibus-daemon generation — **re-read on every reconnect, never cache** (15 stale `dbus-*` sockets in `~/.cache/ibus/` observed live).

**Connection** (goibus `bus.go`, read verbatim this session):
```go
conn, err := dbus.Dial(ibusAddress)          // IBUS_ADDRESS=unix:path=...
conn.Auth(GetUserAuth())                      // AuthExternal(uid), AuthCookieSha1 fallback
conn.Hello()
conn.RequestName(IBUS_SERVICE_IBUS /*"org.freedesktop.IBus"*/, dbus.NameFlagReplaceExisting)
```
`RequestName` doubles as the single-instance guard (AlreadyOwner → exit or take over per policy).

**Component registration** — D-Bus method `org.freedesktop.IBus.RegisterComponent` on `org.freedesktop.IBus` @ `/org/freedesktop/IBus`, one argument: `dbus.MakeVariant(component)`. The method name `RegisterComponent` (no underscore) is confirmed in installed `libibus-1.0.so.5.0.529` strings; the phase brief's `_RegisterComponent` refers to the same call. Wire struct field order (godbus marshals exported fields in declaration order; **the adapter must reproduce this order exactly**):

```
IBusComponent: (s, a{sv}, s×8, av, av)
  0 Name          "IBusComponent"            (string)
  1 Attachments   {}                          (map[string]dbus.Variant)
  2 ComponentName e.g. "org.freedesktop.IBus.goswitch"
  3 Description   e.g. "goswitch layout engine"
  4 Version       e.g. "0.1.0"
  5 License       "MIT"
  6 Author        "Djarvur"
  7 Homepage      "https://github.com/Djarvur/goswitch"
  8 Exec          "" (empty for programmatic-only registration; no spawn)
  9 Textdomain    ""
 10 ObservedPaths []                          (empty av)
 11 EngineList    []dbus.Variant              (av of EngineDesc structs)

IBusEngineDesc: (s, a{sv}, s×8, u, s×7)   [VERIFIED: goibus engineDesc.go field order; 17 fields total]
  0 Name          "IBusEngineDesc"
  1 Attachments   {}
  2 EngineName    "goswitch-en" / "goswitch-ru"
  3 LongName      "goswitch English (US)"
  4 Description   ...
  5 Language      "en" / "ru"
  6 License       "MIT"
  7 Author        "Djarvur"
  8 Icon          ""
  9 Layout        "us" / "ru"     ← drives mutter XKB for two-engine (gnome-shell reads engineDesc.layout)
 10 Rank          uint32 0
 11 Hotkeys       ""
 12 Symbol        "en" / "ru"     ← GNOME indicator label for two-engine
 13 Setup         ""
 14 LayoutVariant ""
 15 LayoutOption  ""
 16 Version       "0.1.0" → then Textdomain ""
```

**Factory** — export at path `/org/freedesktop/IBus/Factory`, interface `org.freedesktop.IBus.Factory`, method `CreateEngine(s engine_name) → (o object_path)`: mint `/org/freedesktop/IBus/Engine/goswitch/<n>`, export the engine object there, return the path.

**Engine object** — export on THREE interfaces simultaneously (goibus engine.go): `org.freedesktop.IBus.Engine`, `org.freedesktop.IBus.Service`, `org.freedesktop.DBus.Properties`. Methods (all verified as strings in libibus + goibus source):

| Method | Signature | Notes |
|--------|-----------|-------|
| `ProcessKeyEvent` | (u keyval, u keycode, u state) → b | synchronous; reply fast; `true` = consumed |
| `SetCursorLocation` | (i,i,i,i) | no-op in Phase 1 |
| `SetSurroundingText` | (v text, u cursor_pos, u anchor_pos) | log only in Phase 1 |
| `SetCapabilities` | (u caps) | store bitmap per input context |
| `FocusIn` / `FocusOut` / `Reset` / `Enable` / `Disable` | () | FocusOut/Reset cancel FSM + (Phase 2) buffers |
| `PageUp/PageDown/CursorUp/CursorDown` | () | no-op |
| `CandidateClicked` | (u,u,u) | no-op |
| `PropertyActivate` | (s,u), `PropertyShow`/`PropertyHide` | (s) — no-op Phase 1 |
| `Destroy` | () on `org.freedesktop.IBus.Service` | unexport object |
| `Get`/`GetAll`/`Set` | org.freedesktop.DBus.Properties | return empty |

**Engine → daemon signals** (emit on own object path):
- `CommitText (v IBusText)` — text injection primitive (Phase 2 wires it; adapter ships the emitter).
- `ForwardKeyEvent (u keyval, u keycode, u state)` — for the BackSpace fallback later.
- `UpdatePreeditText (v, u cursor_pos, b visible, u mode)` — goibus emits the **4-arg form** (mode = `IBUS_ENGINE_PREEDIT_CLEAR` = 0); installed header also shows classic 3-data-arg API + a separate `UpdatePreeditTextWithMode` string [VERIFIED: libibus strings; goibus engine.go]. Phase 1 does not need preedit at all (OSD deferred) — if/when first used, confirm with `dbus-monitor` on the socket (flagged in Assumptions).
- `DeleteSurroundingText (i offset_from_cursor, u nchars)` — **signal, no ack** on 1.5.29 (ibusengine.c main; fire-and-forget).
- `RequireSurroundingText ()` — ask clients for surrounding text.
- `HidePreeditText()`, `UpdateProperty(v)`.

**IBusText wire form** (goibus text.go): `(s "IBusText", a{sv} Attachments, s text, v AttrList)` where AttrList = `(s "IBusAttrList", a{sv}, au attributes)` — empty `au` is fine for plain commits.

### Pattern 2: Event decode constants (verified verbatim from installed headers)

From `/usr/include/ibus-1.0/ibustypes.h` [VERIFIED]:
```go
// ibustypes.h:70-77
IBUS_SHIFT_MASK   = 1 << 0
IBUS_CONTROL_MASK = 1 << 2
IBUS_MOD1_MASK    = 1 << 3   // Alt
IBUS_MOD4_MASK    = 1 << 6   // Super
IBUS_MOD5_MASK    = 1 << 7   // Level3/AltGr
// ibustypes.h:97
IBUS_RELEASE_MASK = 1 << 30
// ibustypes.h:119-126
IBUS_CAP_PREEDIT_TEXT     = 1 << 0
IBUS_CAP_SURROUNDING_TEXT = 1 << 5
IBUS_CAP_SYNC_PROCESS_KEY = 1 << 7
// ibustypes.h:138-139
IBUS_ENGINE_PREEDIT_CLEAR  = 0
IBUS_ENGINE_PREEDIT_COMMIT = 1
```
From `/usr/include/ibus-1.0/ibuskeysyms.h` [VERIFIED]:
```
IBUS_KEY_BackSpace 0xff08   IBUS_KEY_Tab 0xff09   IBUS_KEY_Return 0xff0d
IBUS_KEY_Escape 0xff1b      IBUS_KEY_Shift_L 0xffe1  IBUS_KEY_Shift_R 0xffe2
IBUS_KEY_Control_L 0xffe3   IBUS_KEY_space 0x020
```
The engine object feeds `EngineEvent{Keyval, Keycode, Release: state&(1<<30)!=0, Mods: state&0xff}` into the actor; the FSM never re-parses raw state. Engine lifecycle observed live in `/tmp/ibus_shift_test.log` (this session's timestamps): `bus connected → component registered → engine __init__ → focus_in` — assert this exact sequence in the e2e happy path.

### Pattern 3: Classic-with-waiting tap FSM (D-04/D-05)

Design (synthesis of D-04/D-05 + prototype `_handle_shift_release` logic, punto_engine.py:203-244):

**Contract:** pure function-object, no goroutines, no real clock — `Feed(ev KeyEvent, now time.Duration) (consumed bool, actions []Action)`. Timers are synthetic: the adapter schedules `time.AfterFunc(window)` and re-enters `Feed(TimerExpired{})` — deterministic unit corpus with injected timestamps.

**States:** `idle → taps1 → taps2 → taps3` (count = taps so far in current burst).

**Transitions:**
- `RShiftRelease` with no intervening key since its press, and gap-since-last-tap ≤ window → increment count, (re)arm timer `WindowMs` (default 300 ms, D-05; YAML-configurable later).
- `TimerExpired` in `tapsN` → emit `Action{N}` (1=switch, 2=word, 3=phrase), → idle. **All actions fire at window expiry after the last tap** — this is the owner-selected classic semantics; single-switch latency = window (D-05 NFR wording already accounts for it).
- Any non-RShift key event while RShift is held → modifier-use flag (prototype `rshift_saw_key`, punto_engine.py:89/219-224): the release is ignored, → idle, no action.
- Any non-RShift key between taps (after a release) → cancel pending sequence silently → idle (pins ambiguity; corpus must fix behavior).
- 4th tap within window → stay at `taps3`, re-arm (triple fires once) — corpus pins this.
- `FocusOut` / `Reset` / `Enable` re-entry → idle, disarm.
- Phase 1: FSM output is **logged only**; the adapter returns `false` (not consumed) for every event, so transit typing is untouched (INTEG-02).

**Corpus dimensions (unit, headless):** single/double/triple at window edges (299/300/301 ms), modifier-use (RShift+letter, RShift held through other key), intervening key between taps, 4th tap, FocusOut mid-burst, release-without-press, press held across Reset, ibus#2600-style 1 ms shift glitch (shift-down, key, shift-up → no action), burst of 10 RShift taps. The prototype's constants (DOUBLE_TAP_MS=450, SINGLE_TAP_MS=550, punto_engine.py:73-74) are prior art only — D-05 pins 300 ms default.

### Pattern 4: Layout table generation pipeline (CORR-08)

Inputs (all verified present on target):
- `/usr/share/X11/xkb/symbols/us` — section `xkb_symbols "basic"` at line 4 (default; self-contained) [VERIFIED: read this session]
- `/usr/share/X11/xkb/symbols/ru` — **`default` directive at line 9 selects `xkb_symbols "winkeys"`** (line 10), which `include "ru(common)"` (line 27, `hidden partial`) and overrides AE03–AE08 + AB10 + BKSL [VERIFIED: read this session]
- `/usr/include/ibus-1.0/ibuskeysyms.h` — keysym-name → value (104 Cyrillic defines, e.g. `IBUS_KEY_Cyrillic_io 0x6a3`, `IBUS_KEY_Cyrillic_a 0x6c1`, `IBUS_KEY_Cyrillic_shorti 0x6ca`) [VERIFIED]

Pipeline: `go:generate` runs `go run ./layouts/generator` → parse the two sections (handle `include` recursively within symbols/{us,ru}; skip `kpdl(comma)` — keypad-only) → join on xkb key name (TLDE, AE01–AE12, AD01–AD12, AC01–AC11, AB01–AB10, BKSL) → keysym name→value from ibuskeysyms.h → value→rune: Latin-1 direct for 0x20–0xff; Cyrillic via a small embedded keysym→Unicode table (generator embeds it; golden tests pin it) → emit `layouts/tables.go` with EN→RU and RU→EN maps at both shift levels. **Generated file is committed**; CI runs golden tests only (no xkb dependency in CI) — generation is dev-side.

Effective mapping the generator must produce (default ru = winkeys; derived from the two files read this session — the generator output is truth, this table is the review baseline):

| Position | us basic | ru (winkeys+common) | | Position | us basic | ru |
|---|---|---|---|---|---|---|
| TLDE | `` ` `` / `~` | `ё` / `Ё` | | AC10 | `;` / `:` | `ж` / `Ж` |
| AE02 | `2` / `@` | `2` / `"` | | AC11 | `'` / `"` | `э` / `Э` |
| AE03 | `3` / `#` | `3` / `№` | | BKSL | `\` / `\|` | `\` / `/` |
| AE04 | `4` / `$` | `4` / `;` | | AB01 | `z` / `Z` | `я` / `Я` |
| AE05 | `5` / `%` | `5` / `%` | | AB02 | `x` / `X` | `ч` / `Ч` |
| AE06 | `6` / `^` | `6` / `:` | | AB03 | `c` / `C` | `с` / `С` |
| AE07 | `7` / `&` | `7` / `?` | | AB04 | `v` / `V` | `м` / `М` |
| AE08 | `8` / `*` | `8` / `*` | | AB05 | `b` / `B` | `и` / `И` |
| AE11/12 | `-`/`_`, `=`/`+` | same | | AB06 | `n` / `N` | `т` / `Т` |
| AD11 | `[` / `{` | `х` / `Х` | | AB07 | `m` / `M` | `ь` / `Ь` |
| AD12 | `]` / `}` | `ъ` / `Ъ` | | AB08 | `,` / `<` | `б` / `Б` |
| letters | `q w e r t y…` | `й ц у к е н…` | | AB09 | `.` / `>` | `ю` / `Ю` |
| | | | | AB10 | `/` / `?` | `.` / `,` |

Golden corpus (SPEC examples + punctuation): `ghbdtn`→`привет`, `GHBDTN`→`ПРИВЕТ`, `Ghbdtn`→`Привет`; `[`↔`х`, `]`↔`ъ`, `;`↔`ж`, `'`↔`э`, `,`↔`б`, `.`↔`ю`, `/`↔`. , `?`↔`,`, `` ` ``↔`ё`, `~`↔`Ё`, `@`↔`"`, `#`↔`№`, `$`↔`;`, `^`↔`:`, `&`↔`?`, `|`↔`/`. Note the asymmetric ones (`&`↔`?` because ru `?` lives on AE07-shift while us `?` lives on AB10-shift — the join is by key position, not by character).

> Note vs project research: ARCHITECTURE.md cautioned the prototype used "winkeys-style digit-row maps while the session actually uses plain ru" — the live file shows plain `ru` **resolves to winkeys by default** on Ubuntu 24.04, so the prototype was right here; generation still removes the bug class and survives variant changes (parameterize by variant later).

### Pattern 5: Reconnect / re-register loop (INTEG-04) + recover-shim (INTEG-05)

- godbus has **no auto-reconnect**: on bus loss, `Signal()` channels close and calls return `ErrClosed`. Loop: detect closed channel / failed call → re-run address discovery (fresh socket path each daemon generation) → Dial/Auth/Hello → RequestName → re-export Factory → `RegisterComponent` → log re-registration. Backoff ~1–2 s with jitter; ibus#2910's validated workaround. GNOME re-activates the source engine after daemon restart (sources persist in gsettings; keyboard.js reloads on ibus `ready`).
- recover-shim: godbus dispatches each method call on its own goroutine; one unrecovered panic kills the process; abnormal engine death makes ibus-daemon suicide → desktop-wide input loss. Wrap every exported handler: `defer func(){ if r:=recover(); r!=nil { log error; return safe reply } }()`. Guard every variant type assertion with comma-ok (the #1 panic source). `kill -9` is uncatchable by design — that case is covered by systemd `Restart=on-failure` + the re-register loop, and the e2e asserts desktop typing survives it.

### Pattern 6: e2e skeleton (TEST-02/03, M1 gate)

Runner = Go program in `test/e2e/` (not `go test` — needs live session, own exit-code contract):
1. **Preflight, fail-fast with human-readable diagnostics** (INST-04 synergy): ibus address file exists+readable; `ibus list-engine` contains our engine; `/dev/uinput` writable; `/usr/bin/python3 -c "import gi"` OK; goswitchd log file growing (heartbeat log line).
2. **Source activation** (depends on D-01 outcome; during the experiment itself this is the independent variable): set `org.gnome.desktop.input-sources sources` to include `('ibus','goswitch-en')`/`('ibus','goswitch-ru')` — **save previous value, restore in teardown** (live desktop state — see Runtime State Inventory).
3. **Window activation** (TEST-03): launch `gnome-text-editor` (verified at `/usr/bin/gnome-text-editor`), focus via AT-SPI `Component.grabFocus` through a ~30-line `/usr/bin/python3` gi helper (absolute path mandatory — PATH `python3` is linuxbrew without gi). wmctrl is NOT installed and is X11-only anyway.
4. **Injection**: `ydotool type --key-delay <ms> 'ghbdtn'`; `ydotool key --key-delay <ms> Shift_R` ×N with pacing for tap windows (0.1.8 syntax: symbolic `Super+Space`-style sequences, verified via `ydotool key --help`).
5. **Assertion (M1 gate)**: parse goswitchd JSON log — key events present while goswitch is the GNOME-active source; FSM decisions logged (`single`/`double`/`triple`); lifecycle sequence `component registered → focus_in` observed. Phase 2 replaces log assertions with AT-SPI text readback (TEST-04 matrix).
6. **Resilience cases**: (a) `ibus restart` → wait-for-condition (not sleep) on re-registration log line → inject → events still logged; (b) `kill -9` goswitchd → type via a plain source → desktop input unaffected → daemon respawns (systemd unit or runner) → re-registers.
7. **Report**: per-case PASS/FAIL + observed log excerpt on failure; exit code ≠ 0 on any FAIL.

Timing-sensitive cases: make ydotool pacing configurable (slower on loaded machines) per PITFALLS; wait-for-condition helpers from day one.

### Pattern 7: Structured logging (INST-04)

`log/slog` JSON handler to stderr (API verified on local go1.27.1):
```go
level := new(slog.LevelVar)          // runtime-adjustable later via goswitchctl
level.Set(slog.LevelInfo)
h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
slog.SetDefault(slog.New(h))
```
- Levels: ERROR (adapter failures, re-register loop exhaustion), WARN (reconnects, caps anomalies), INFO (lifecycle: connected, registered, engine created, focus in/out, re-registered), DEBUG (key trace).
- **Key-trace debug mode** behind explicit `-debug` flag (default OFF — keystroke logs capture passwords, PITFALLS security table): each `ProcessKeyEvent` logs `key= keyval=0xffe2 keycode=62 state=0x1 release=false mods=SHIFT`; FSM logs state transitions and decisions. Document the warning in README.
- e2e runs goswitchd as a subprocess with `-debug` and stderr → temp file (no systemd dependency in the skeleton); the production systemd user unit (dev form this phase) ships stderr → journal.
- Never log buffer *contents* at INFO (lengths only) — carry the rule from day one.

### Pattern 8: ADR pack (success criterion 1)

Format per CONTEXT discretion: `docs/adr/ADR-00N-<slug>.md`, each with Status/Context/Decision/Consequences/Reversibility (quote CONTEXT wording). Composition fixed by the phase criteria:
- **ADR-001 Layout switching mechanism** — written LAST, from the D-01 experiment log (two-engine vs internal flip; include the kill criteria, the evidence, the losing candidates with observed failures). If internal flip wins → D-03 "один хозяин" consequences documented.
- **ADR-002 Tap semantics** — classic-with-waiting, 300 ms default, decision-at-window-expiry, modifier-use discrimination, 4th-tap rule (D-04/D-05; quote FSM corpus as the executable spec).
- **ADR-003 Replacement capability ladder** — DeleteSurroundingText (signal, no ack) → Backspace×N (runes; abort-not-garbage) → clipboard; per-client caps via SetCapabilities bitmap (implementation Phase 2, decision now).
- **ADR-004 Buffer reset triggers** — D-06 verbatim: no mouse click; mandatory surrounding-text check before correction; caret-jump best-effort reset; spec-delta to §4.3.
- **ADR-005 MACR-01 / OPEN-01** — owner checkpoint (гейт M0 «ревью владельца»); executor recommendation (v2 spec-delta + keyd recipe) is input, NOT a decision; no code planned either way.

### Anti-Patterns to Avoid
- **Replying late to `ProcessKeyEvent`** — synchronous daemon call; any waiting must happen after replying (timers re-enter the actor). Phase 1 replies `false` immediately.
- **Consuming bare modifier presses** — breaks Shift+letter app behavior; prototype consumed RShift always (punto_engine.py:355) and got away with it; Phase 1 must NOT (observer mode, return false).
- **Buffering both press and release** — every key arrives twice (`IBUS_RELEASE_MASK` bit 30); buffer paths must filter releases (FSM needs both; document the split).
- **Caching the IBus socket address** — it changes per daemon generation; 15 stale sockets observed live.
- **`python3` from PATH in e2e** — resolves to linuxbrew (no gi); use `/usr/bin/python3`.
- **Expecting `gsettings set current` to work without testing** — STACK verified writability, PITFALLS says runtime-ignored; only the D-01 experiment settles it (that's the point of D-01).
- **Hand-writing the layout table** — generate from xkb; the golden corpus pins it (CORR-08 is literally this requirement).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| D-Bus transport/auth/marshalling | custom socket protocol | godbus v5.2.2 | Auth handshake, variant marshalling, signal routing; proven against IBus socket by ibus-bamboo |
| IBus wire structs | guessing GVariant field order | port goibus field orders verbatim (map above) | Wrong order = silent registration failure or daemon confusion; goibus order is production-proven |
| Layout tables | transliteration-site copies | go:generate from /usr/share/X11/xkb/symbols | translit ≠ key-position; digit/punct rows differ per variant (winkeys proof) |
| JSON logging | custom logger | log/slog | stdlib, levels, structured attrs, zero deps |
| Key injection (e2e) | raw uinput ioctls (first) | ydotool 0.1.8 | installed, works without daemon on 24.04; own injector only as fallback |
| Window focus (e2e) | X11 tools / heuristics | AT-SPI Component.grabFocus | Wayland-correct; TEST-03; python3-gi installed |
| Tap timing | wall-clock sleeps in tests | injected clock + TimerExpired events | deterministic corpus; no flaky CI |

**Key insight:** everything risky in this phase is either (a) already proven on this machine by the prototypes (registration, key receipt, lifecycle) or (b) made testable by the wire-adapter/pure-logic split — the adapter is the only untestable-headlessly piece and it is ~1k LOC of field-order-faithful porting, not invention.

## Runtime State Inventory

> Phase 1 mutates live desktop state (input sources, IBus registrations, e2e injection). Verified live this session.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — daemon is stateless in Phase 1 (no config file, no persisted buffers) | None |
| Live service config | `gsettings org.gnome.desktop.input-sources sources` = `[('xkb', 'us'), ('xkb', 'ru')]`, `current` = `uint32 0` (verified live). D-01 experiment and e2e will overwrite `sources` temporarily | Experiment/e2e scripts MUST snapshot (`gsettings get`) and restore in teardown; ADR-001 may permanently change sources (owner's call) |
| OS-registered state (IBus) | Two prototype components registered system-wide: `/usr/share/ibus/component/punto-switcher.xml`, `test-shift.xml` (in `ibus list-engine`: `test-shift`, `punto-switcher` — verified live); goswitch will register **programmatically** (runtime only, no XML) → vanishes on ibus restart by design; 15 stale `~/.cache/ibus/dbus-*` sockets; registry cache at `~/.cache/ibus/bus/registry` | No migration needed; e2e preflight tolerates prototype engines; consider (owner consent) removing root-installed prototype XMLs when they start conflicting — separate manual step, NOT in phase code |
| OS-registered state (systemd) | `org.freedesktop.IBus.session.GNOME.service` active (ordering anchor for our unit); keyd.service **installed but failed/disabled** (verified live) — no interference during Phase 1 | Phase-1 dev unit: `PartOf=graphical-session.target`, `After=…IBus…service`; INTEG-02 "session with working keyd" cannot be auto-tested while keyd is dead — document as manual/optional check (starting keyd needs root) |
| Secrets/env vars | `IBUS_ADDRESS` not set in session env (address file is the source); no secrets involved | None |
| Build artifacts | Empty skeleton: no go.mod, no binaries, no `.github/`, no .gitignore at root; golangci-lint NOT installed | Wave 0: go.mod init, .gitignore, CI skeleton; `go vet` + `go test -race` suffice until golangci-lint lands (or `go install` it pinned) |

## Common Pitfalls

### Pitfall 1: Engine crash = desktop input death
**What goes wrong:** unrecovered panic in a handler kills goswitchd; abnormal engine death makes ibus-daemon kill itself; typing dies desktop-wide (TUX IM: 8 such bugs in 2 weeks).
**Why:** godbus dispatches handlers on arbitrary goroutines; variant type assertions are the #1 panic source; lifecycle coupling is undocumented.
**How to avoid:** recover-shim wrapping EVERY exported handler from the first D-Bus commit; comma-ok on every variant assertion; e2e case "kill -9 then type".
**Warning signs:** desktop typing freezes right after a goswitch action; journal shows ibus-daemon respawns.

### Pitfall 2: Engine sees keys only as the ACTIVE input source
**What goes wrong:** registered-but-not-active engine receives zero events; gate "hotkey in log" fails with no events at all.
**Why:** IBus routes `ProcessKeyEvent` only to the active source's engine; current sources are plain xkb (`[('xkb','us'),('xkb','ru')]`, verified live).
**How to avoid:** D-01/e2e setup must make goswitch the active source (gsettings sources write, one-time); the M1 gate explicitly runs "while goswitch is the GNOME-active source".
**Warning signs:** works standalone via `ibus engine goswitch-en`, dead after Super+Space away.

### Pitfall 3: `gsettings set current` writability ≠ runtime effect
**What goes wrong:** the write succeeds, layout doesn't change; test asserts dconf value and lies.
**Why:** Shell's in-process InputSourceManager holds live state; keyboard.js 46 has no external-change handler (verified in source this session).
**How to avoid:** D-01 experiment asserts the OBSERVED script of subsequently typed text (and FocusIn/FocusOut of our engines), never the dconf value.
**Warning signs:** e2e passes gsettings assertion, fails text assertion.

### Pitfall 4: Wire-struct field-order mismatch
**What goes wrong:** RegisterComponent "succeeds" but engines never appear / factory never called.
**Why:** hand-rolled GVariant ordering; no introspection on the IBus bus to catch it.
**How to avoid:** port goibus field orders verbatim (map above); verify with `ibus list-engine` + `dbus-monitor` on the socket during dev.
**Warning signs:** `ibus list-engine` lacks goswitch after registration with no error.

### Pitfall 5: Doubled keys / modifier corruption
**What goes wrong:** releases fed into buffers double every letter; consuming bare modifiers breaks Shift+letter.
**Why:** every key arrives twice (bit 30); the return contract is easy to miss.
**How to avoid:** decode release bit once in the adapter; FSM consumes press+release deliberately; Phase 1 returns false everywhere (observer).
**Warning signs:** doubled letters in any buffer-derived output; "typing feels dead" after a hotkey feature.

### Pitfall 6: e2e focus/permission ghost-failures
**What goes wrong:** injected keys go to the wrong window; ydotool "silently does nothing"; tests flake ~5%.
**Why:** notifications steal focus; /dev/uinput perms; Chromium a11y trees materialize late (Phase 2 concern).
**How to avoid:** preflight fail-fast checks; AT-SPI grabFocus before every case (TEST-03); wait-for-condition, never sleeps; per-case state reset.
**Warning signs:** passes interactively, fails under dispatch; failures concentrated in focus-dependent cases.

### Pitfall 7: Registry/cache staleness during the dev loop
**What goes wrong:** engine "disappears" after rebuilds; old component XMLs spawn-error forever.
**Why:** ibus registry cache is mtime-based; prototype XMLs in the root-owned dir reference fixed exec paths.
**How to avoid:** programmatic registration (no XML) for Phase 1 dev; `ibus list-engine` in preflight; if XML experiments happen — `ibus write-cache` + restart dance.
**Warning signs:** engine listed yesterday, gone today with no config change.

## Code Examples

### Engine adapter skeleton (wire-faithful)
```go
// Source: goibus source read verbatim this session (component.go/engineDesc.go/bus.go/engine.go);
// field order is the wire contract — do not reorder.
type Component struct {
	Name          string                  // "IBusComponent"
	Attachments   map[string]dbus.Variant // {}
	ComponentName string                  // "org.freedesktop.IBus.goswitch"
	Description   string
	Version       string
	License       string  // "MIT"
	Author        string
	Homepage      string
	Exec          string  // "" — programmatic registration only
	Textdomain    string
	ObservedPaths []dbus.Variant // {}
	EngineList    []dbus.Variant // MakeVariant of each EngineDesc (dereferenced struct value)
}

// registration call (goibus bus.go):
obj := conn.Object("org.freedesktop.IBus", "/org/freedesktop/IBus")
obj.Go("org.freedesktop.IBus.RegisterComponent", 0, ch, dbus.MakeVariant(component))
```

### FSM feed (deterministic, testable)
```go
type Event interface{} // KeyPress, KeyRelease, TimerExpired, Reset
type Action int        // Single, Double, Triple

type FSM struct { window time.Duration; taps int; sawKey bool; held bool /* + clock hook */ }

func (f *FSM) Feed(ev Event, now time.Duration) []Action {
	switch e := ev.(type) {
	case KeyPress:
		if e.Keyval == KeyvalShiftR { f.held = true; f.sawKey = false }
	case KeyRelease:
		if e.Keyval == KeyvalShiftR && f.held {
			f.held = false
			if f.sawKey { f.taps = 0; return nil } // modifier use — not a tap
			if f.taps < 3 { f.taps++ }              // 4th tap stays at 3
			f.deadline = now + f.window             // adapter arms AfterFunc(window) → TimerExpired
		}
	case TimerExpired:
		if f.taps > 0 { a := Action(f.taps); f.taps = 0; return []Action{a} }
	case Reset: f.taps, f.held, f.sawKey = 0, false, false
	}
	return nil
}
```

### slog key trace (opt-in debug)
```go
// DEBUG only, behind -debug flag; default OFF (keystroke logs capture passwords)
slog.Debug("key", "keyval", fmt.Sprintf("0x%x", ev.Keyval), "keycode", ev.Keycode,
	"release", ev.Release, "mods", fmt.Sprintf("0x%x", ev.Mods))
```

### e2e shell sketch (runner internals)
```bash
# preflight (each check → named diagnostic + exit 1)
test -r "$(ibus address >/dev/null 2>&1 && echo ok)" || fail "ibus CLI"
ibus list-engine | grep -q goswitch || fail "engine not registered"
test -w /dev/uinput || fail "/dev/uinput not writable (input group?)"
/usr/bin/python3 -c "import gi" || fail "python3-gi missing"
# case: M1 gate
/usr/bin/python3 focus_helper.py gnome-text-editor   # AT-SPI grabFocus
ydotool type --key-delay 40 "ghbdtn"
ydotool key --key-delay 120 Shift_R Shift_R          # paced taps for double
wait_for_log '"msg":"action","n":2' 5s               # wait-for-condition, not sleep
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| IBus engines via libibus C/GI bindings | pure-D-Bus wire adapters (godbus in Go) viable | proven 2018→2026 (goibus, ibus-bamboo) | in-repo `engine/` adapter instead of cgo |
| `gsettings set current` trusted to switch | GNOME 41+: Shell private APIs gated; keyboard.js has no external-change sync | GNOME 41–46 | D-01 experiment is mandatory; Eval needs unsafe mode |
| ydotool 1.0.x + ydotoold daemon | 24.04 ships 0.1.8, direct /dev/uinput | distro packaging | no daemon provisioning; "ydotoold unavailable" notice is benign |
| xkb tables hand-copied | generation from system xkb symbols + golden tests | standing practice | CORR-08 is this requirement |

**Deprecated/outdated:** xdotool (X11-only, dead on Wayland); goibus as dependency (no LICENSE); godbus v4 (EOL).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `UpdatePreeditText` 4-arg wire form `(v,u,b,u)` works on installed 1.5.29-rc2 (goibus emits it; header shows both classic and WithMode) | Wire surface | Phase 1 does not emit preedit (OSD deferred) — zero phase risk; verify with dbus-monitor when first used (Phase 3 OSD or indicator work) |
| A2 | IBus keycodes delivered by the GNOME IM stack are consistent evdev-range codes usable as buffer keys | Event decode | Buffer keys on keycode is a Phase 2 concern; M1 skeleton LOGS keycodes — the skeleton run itself resolves this empirically before Phase 2 planning |
| A3 | GNOME Shell re-activates the goswitch source after ibus-daemon restart (sources persist in gsettings; keyboard.js reload on ibus ready) | Reconnect loop | e2e resilience case (a) will observe; fallback: after re-register, daemon may need to poke activation (SetGlobalEngine or gsettings current) — record in ADR-001 |
| A4 | ydotool 0.1.8 injects reliably without ydotoold on this machine (binary + perms verified; actual injection not exercised this session) | e2e skeleton | First e2e preflight case is a self-test injection; fallback: ~150-LOC Go uinput injector (STACK) |
| A5 | gsettings `sources` restore in teardown is sufficient cleanup for live-desktop mutation | Runtime State | `current` index may shift if sources list changed length; snapshot both keys; verify in e2e teardown |
| A6 | Cyrillic keysym→Unicode table (~70 entries) embedded in generator, pinned by golden tests, is correct | Layout pipeline | Golden corpus (ghbdtn→привет + full rows) catches any wrong char at generation time, not runtime |

## Open Questions

1. **Does ANY programmatic switch mechanism satisfy D-02 on GNOME 46?**
   - What we know: keyboard.js 46 source rules out SetGlobalEngine as sufficient; gsettings writability verified but effect untested; ydotool Super+Space and micro-extension are untested on 46.
   - What's unclear: which candidate actually switches the source at runtime.
   - Recommendation: that's exactly the D-01 experiment (§ Pattern "D-01 branch"); run it with the skeleton, timeboxed, with Option B (owner-proven) as the no-drama fallback.
2. **Exact keycode semantics on the live IM path (A2)** — resolved empirically by the M1 skeleton's own key trace log; no extra work, just record the observation into Phase 2 planning.
3. **ADR-005 (MACR-01) owner decision** — cannot be researched; owner checkpoint at ADR review (OPEN-01 binds: no code, no plan for it).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| IBus daemon (private socket) | engine adapter | ✓ | 1.5.29-rc2 (`ibus version`; address file verified) | — |
| GNOME Shell (Wayland) | source switching, e2e | ✓ | 46.0 (`WAYLAND_DISPLAY=wayland-0`, `XDG_SESSION_TYPE=wayland`) | — |
| Go toolchain | everything | ✓ | go1.27.1 linux/amd64 | — |
| godbus/yaml.v3/fsnotify | stack | ✓ (proxy.golang.org: v5.2.2 / v3.0.1 / v1.10.1, all latest) | — |
| ydotool | e2e injection | ✓ | 0.1.8-3build1, `/dev/uinput` writable by user (input group) | Go uinput injector |
| /usr/bin/python3 + gi | e2e AT-SPI helper | ✓ | 3.12.3, `import gi` OK | pure-Go AT-SPI client later |
| gnome-text-editor | e2e target app | ✓ | /usr/bin/gnome-text-editor | gedit/Chrome (both present: google-chrome, chromium snap, firefox) |
| systemd user session | dev autostart unit | ✓ | graphical-session.target active; IBus user unit running | run daemon directly in dev loop |
| xkb symbols + ibuskeysyms.h | generator | ✓ | /usr/share/X11/xkb/symbols/{us,ru}; /usr/include/ibus-1.0/ibuskeysyms.h (104 Cyrillic defines) | committed generated file (CI never regenerates) |
| golangci-lint | CI lint gate | ✗ | — | `go vet` + `go test -race` in Wave 0; add pinned golangci-lint later (go-ultimate allows this staging) |
| wmctrl | — (not used) | ✗ | — | AT-SPI grabFocus (chosen mechanism anyway) |
| keyd (running) | INTEG-02 "session with keyd" | ✗ (installed; service failed/disabled — verified) | — | Architectural coexistence proof (no evdev in product) + optional manual check; starting keyd needs root, out of automated scope |

**Missing dependencies with no fallback:** none blocking.
**Missing dependencies with fallback:** golangci-lint (vet+race now), keyd-live-session (manual/optional).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | stdlib `testing` (go-ultimate mandate; no testify) |
| Config file | none yet — Wave 0 adds go.mod; CI workflow in `.github/workflows/` |
| Quick run command | `go test ./...` |
| Full suite command | `go vet ./... && go build ./... && go test -race ./...` (headless, no session needed) |

### Phase Requirements → Test Map (TEST-01/02/03 dimension split)
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CORR-08 / TEST-01 | Generated EN↔RU tables incl. punctuation `[ ] ; ' , . /` and digit-row shift forms | golden (headless) | `go test ./layouts/ -run Golden -v` | ❌ Wave 0 |
| TEST-01 | FSM: single/double/triple within/across 300 ms window; modifier-use; glitch corpus | unit, fake clock (headless) | `go test ./internal/hotkey/ -v` | ❌ Wave 0 |
| TEST-01 | Generator: xkb parse → identical tables on re-run; golden regeneration diff | golden (dev-side; CI tests the committed file) | `go generate ./layouts && git diff --exit-code layouts/` | ❌ Wave 0 |
| INST-04 | Structured log lines: levels, JSON shape, key-trace only in debug | unit (headless) | `go test ./internal/logging/ -v` | ❌ Wave 0 |
| INTEG-05 | recover-shim contains injected panics in every exported handler | unit with in-memory fake connection (headless) | `go test ./engine/ -run Recover -v` | ❌ Wave 0 |
| INTEG-01/02/03 | Register, receive keys as active source, transit (no consumption) | e2e (live session) | `go run ./test/e2e -case m1-gate` | ❌ later wave |
| INTEG-04 | `ibus restart` → re-register → keys flow again | e2e (live) | `go run ./test/e2e -case ibus-restart` | ❌ later wave |
| INTEG-05 | `kill -9` engine → desktop typing alive → respawn+re-register | e2e (live) | `go run ./test/e2e -case kill9-survive` | ❌ later wave |
| TEST-02/03 | Injection + self-activated window + log assertions (skeleton) | e2e (live) | `go run ./test/e2e -case skeleton` | ❌ later wave |
| D-01 | Switch-matrix probes (gsettings/ydotool/Eval/SetGlobalEngine) | e2e experiment script (live) | `go run ./test/e2e -case d01-probe` (evidence → ADR-001) | ❌ later wave |

### Sampling Rate
- **Per task commit:** `go vet ./... && go test ./...`
- **Per wave merge:** `go vet ./... && go build ./... && go test -race ./...`
- **Phase gate:** headless suite green + e2e skeleton green on the live session (M1: hotkey events in log while goswitch is active source; ibus restart and kill -9 cases pass) + ADR pack reviewed by owner (M0 gate for OPEN-01)

### Wave 0 Gaps
- [ ] `go.mod` (go 1.23, godbus dep), `.gitignore`, CI workflow (vet+build+test on push)
- [ ] `layouts/tables_test.go` — golden corpus (REQ CORR-08)
- [ ] `internal/hotkey/fsm_test.go` — event-stream corpus (TEST-01)
- [ ] `internal/logging/logging_test.go` — JSON shape + debug gating (INST-04)
- [ ] `engine/` fake-connection recover tests (INTEG-05)
- [ ] e2e runner skeleton + preflight + focus helper (TEST-02/03)

## Security Domain

> `security_enforcement: true`, ASVS L1. Phase 1 is a local, non-networked daemon — the applicable surface is small but real.

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | No auth surface; single-user session daemon (D-Bus socket membership is the boundary) |
| V3 Session Management | no | No sessions |
| V4 Access Control | marginal | Single-instance guard via `RequestName` ReplaceExisting (prevents double-spawn rogue engine) |
| V5 Input Validation | yes | D-Bus args are typed by godbus (u/i/s/v fixed signatures); every variant assertion comma-ok; YAML config validated when parsed (Phase 3) |
| V6 Cryptography | no | None; never hand-roll (nothing to encrypt) |

### Known Threat Patterns for a keystroke-logging daemon
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Key-trace logs capture passwords | Information disclosure | Debug trace behind explicit `-debug`, default off; never log buffer contents at INFO (lengths only); README warning |
| Malicious duplicate engine squatting our name | Spoofing | `RequestName` result check + exit policy on AlreadyOwner |
| D-Bus handler panic → desktop input DoS | Denial of service | recover-shim on every handler (INTEG-05) |
| Component XML exec path in world-writable dir (future install) | Tampering | Phase 1 uses programmatic registration (no XML); Phase 4 installer checks dir perms |

## Sources

### Primary (HIGH confidence — this session, on the target machine)
- Live probes: `ibus version/address/list-engine/engine`, `~/.config/ibus/bus/4c8b3c0710fb4959927ecf0d99b6baec-unix-wayland-0` (address + PID), `gsettings get` sources/current/wm keybindings, `gnome-shell --version`, `ydotool` 0.1.8 help+dpkg, `/dev/uinput` perms + `groups`, `go version`, `systemctl --user` (IBus unit, graphical-session, keyd failed), `pgrep ibus-daemon`, `/tmp/ibus_shift_test.log` lifecycle
- `/usr/include/ibus-1.0/` headers read: ibustypes.h (masks, caps — lines 70-97, 119-126, 138-139), ibusengine.h (process_key_event 103-107, update_preedit_text 215-233, forward_key_event 370-385, delete_surrounding_text 414-423), ibuskeysyms.h (keyvals; 104 Cyrillic defines), ibuscomponent.h; `strings libibus-1.0.so.5.0.529` (RegisterComponent, signal/method names, `(sa{sv}avav)` signature)
- sarim/goibus complete source read verbatim (component.go, engineDesc.go, bus.go, engine.go, factory.go, common.go, text.go) — wire field orders, method/signature call shapes
- /usr/share/X11/xkb/symbols/us (basic, lines 4-60) and symbols/ru (default→winkeys line 9-10, common 27-84, override keys) read
- Owner prototypes read in full: `/home/nil/.local/share/ibus-test/test_shift.py` (77 lines), `/home/nil/.local/share/punto-switcher/punto_engine.py` (438 lines — tap logic, EN2RU table, fallbacks)
- GNOME Shell 46.0 sources fetched and read: `js/ui/status/keyboard.js` (no external-engine sync; activateInputSource), `js/misc/ibusManager.js` (global-engine-changed handling, set_global_engine_async)
- Go module proxy: godbus v5.2.2 / yaml.v3 v3.0.1 / fsnotify v1.10.1 (all latest — verified)

### Secondary (MEDIUM confidence)
- [GNOME GitLab #5502 — getInputSourceManager().inputSources[N].activate() via Eval](https://gitlab.gnome.org/GNOME/gnome-shell/-/issues/5502)
- [GNOME Discourse — Shell private D-Bus restrictions](https://discourse.gnome.org/t/unable-to-call-the-remote-getwindows-gnome-method-via-dbus/21201), [AskUbuntu — dbus calls to gnome-shell restricted](https://askubuntu.com/questions/1412130/dbus-calls-to-gnome-shell-dont-work-under-ubuntu-22-04), [Eval GJS extension](https://extensions.gnome.org/extension/5952/eval-gjs/), [unsafe-mode-menu](https://github.com/linushdot/unsafe-mode-menu)
- Project research (verified there): .planning/research/{SUMMARY,ARCHITECTURE,STACK,PITFALLS}.md — ibus#2910, #2354, #2600; two-engine pattern; goibus license check; ydotool/AT-SPI facts

### Tertiary (LOW confidence)
- None load-bearing; UpdatePreeditText 4-arg wire form flagged (A1) pending dbus-monitor confirmation when first used

## Metadata

**Confidence breakdown:**
- Wire surface: HIGH — three independent sources agree (goibus source, installed libibus strings, installed headers) + prototype-proven lifecycle on this machine
- D-01 protocol: HIGH for the negative finding (keyboard.js/ibusManager source read); MEDIUM for candidate outcomes until the experiment runs (by design — that is D-01's purpose)
- Layout generation: HIGH — both xkb files and keysym header read; winkeys-default corrected against project research
- FSM/e2e/logging: HIGH — conventional patterns, all constants verified, environment audited

**Research date:** 2026-09-10
**Valid until:** 2026-10-10 (stable domain; IBus/GNOME versions pinned on this machine)
