---
phase: 13-a11y-retsepty-telegram-zcode
reviewed: 2026-11-06T14:54:50Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - scripts/a11y-selfcheck.py
  - test/e2e/spike_journal_test.go
  - docs/CONFIG.md
  - README.md
  - README.en.md
findings:
  critical: 0
  warning: 4
  info: 2
  total: 6
status: issues_found
---

# Phase 13: Code Review Report

**Reviewed:** 2026-11-06T14:54:50Z
**Depth:** standard
**Files Reviewed:** 5
**Status:** issues_found

## Summary

Scope: the phase-13 execution diff `8aa3882..HEAD` — the user-facing a11y
probe `scripts/a11y-selfcheck.py` (new, 224 lines), the journal-section pin
in `test/e2e/spike_journal_test.go` (+9), and the recipe/boundary
documentation in `docs/CONFIG.md`, `README.md`, `README.en.md`. Confirmed
daemon-forbidden boundary holds: zero changes under `internal/` and `cmd/`.

Verified correct (adversarial pass found these sound):

- **Role canon fidelity.** `TEXT_ROLES = {61, 79, 94}`, `ROLE_PASSWORD_TEXT
  = 40`, terminal 60 — matches `internal/appid/appid.go:90-96` exactly; the
  verdict logic gates on NUMBERS only, role-name strings are print-only
  cosmetics (lines 40-44, 151-160, 183). The known Pitfall-8
  (localized/divergent role names) is structurally avoided.
- **Exit-code contract.** 0 visible / 1 alive-but-not-visible / 2 usage
  error or no live process — implemented as documented (main, lines
  205-220); the argument gate rejects empty and extra args; live-proven on
  the VM stand (journal row «a11y-selfcheck контракт», all four matrix
  cells matched, script diff since = 0).
- **No content/secrets leakage.** The probe reads roles, state flags and
  pids only — never `getText`; `/proc` liveness reads `cmdline[0]` for an
  exe-basename substring compare, never prints argv or env, never spawns a
  subprocess or shell; the argv needle is data, not a command part.
- **Interpreter discipline.** Shebang `#!/usr/bin/python3`, mode 755,
  `py_compile` clean on /usr/bin/python3; every doc one-liner uses the full
  `/usr/bin/python3 scripts/a11y-selfcheck.py <app>` form — no collision
  with the existing `goswitchctl selfcheck` subcommand
  (cmd/goswitchctl/main.go:269), which the a11y recipes never reference.
- **Journal pin.** `secA11yProbe = "Фаза 13 A11Y-проба сборок (telegram
  snap/flatpak, zcode)"` matches the real header at
  docs/adr/adr-008-spike-log.md:721 byte-for-byte;
  `TestSpikeJournalSectionsExist` passes on host (`go test ./test/e2e/ -run
  TestSpikeJournalSectionsExist` → ok).
- **Docs vs journal facts.** ZCode: 2-node stub → exit 1, 43+ nodes → exit
  0 with focused role 79 [entry], API-key screen role 40, override
  self-rewrite 2/2, argv/env rewrite — all match journal rows r1/r2
  red/green. Telegram: layer-1 switching green in all builds, autocorrect
  red everywhere, role-75 one-node stub identical in the unconfined
  tarball build, screen-reader banner observed on flatpak/tarball and not
  on snap, owner verdict «Задокументировать» — all match the dated rows.
  RU/EN READMEs are structurally parallel (same blocks, same claims, same
  full-form one-liners).

Concerns: the probe's entry points (module import, a11y-bus access) are
unguarded, so environment failures produce a traceback that exits 1 — the
same code as the semantic "alive but not visible" verdict; the tree walk
can silently drop sibling subtrees, contradicting its own docstring; and
CONFIG.md's evidence quantifier "двумя раундами в каждой" is unsupported by
the journal row corpus for the snap rung (one dated round, not two).

## Warnings

### WR-01: Unguarded `gi` import — wrong interpreter yields a traceback and a dishonest exit 1

**File:** `scripts/a11y-selfcheck.py:35-38`
**Issue:** `import gi`, `gi.require_version("Atspi", "2.0")` and the
typelib import run unguarded. The project's own docstring (lines 21-22)
records that PATH `python3` is linuxbrew without gi — the natural user
invocation `python3 scripts/a11y-selfcheck.py zcode` (or a host missing
`gir1.2-atspi-2.0`) dies with a raw `ModuleNotFoundError` traceback and
exit code 1. Exit 1 is the contract's "process alive but tree/field not
there" verdict, so an environment failure is indistinguishable from the
semantic failure the tool exists to report. The doc one-liners mitigate by
mandating the full form, but the script is also executable directly and
nothing diagnoses the wrong-interpreter case.
**Fix:**
```python
try:
    import gi
    gi.require_version("Atspi", "2.0")
    from gi.repository import Atspi  # noqa: E402
except (ImportError, ValueError) as exc:  # no gi / no Atspi typelib
    print(f"a11y-selfcheck: python3 gi/Atspi unavailable ({exc}) — "
          "run via /usr/bin/python3", file=sys.stderr)
    sys.exit(2)  # environment/usage error, NOT the semantic failure code
```

### WR-02: Unguarded a11y-bus entry point — dead bus crashes instead of the promised clean exit 1

**File:** `scripts/a11y-selfcheck.py:49-51`
**Issue:** `applications()` calls `Atspi.get_desktop(0)` and
`desktop.get_child_count()` with no guard. The tool's documented usage is
over ssh with explicit session env (journal rows show exactly that); if the
invocation lacks the session environment or the at-spi bus is unavailable,
`get_desktop` can return None (→ `AttributeError` on `get_child_count`) or
the call raises `GLib.Error` — a traceback with exit 1, not the docstring's
promised clean "a11y not enabled" explanation (lines 27-29). Mid-walk bus
errors are handled (walk/role_number), but the entry point is not.
**Fix:**
```python
def applications():
    """Yield every AT-SPI application node on the desktop."""
    try:
        desktop = Atspi.get_desktop(0)
        if desktop is None:
            return
        count = desktop.get_child_count()
    except Exception:
        return  # no reachable a11y bus — falls through to the clean
                # "application NOT FOUND" -> exit 1 path
    for i in range(count):
        ...
```

### WR-03: `walk()` drops all siblings after the first dead child — verdict can flip 0 → 1

**File:** `scripts/a11y-selfcheck.py:70-74`
**Issue:** The child enumeration is a list comprehension inside one
try/except: if `get_child_at_index(k)` raises for the first dead node
(the documented bus-lag scenario, docstring lines 58-63), `continue`
skips children `k+1..N` of that parent entirely — including any text
field living there. This contradicts the docstring's own claim ("Every
step is individually best-effort — a transient node error skips that
node"), and for a verdict tool the sibling loss can turn a real
"VISIBLE / exit 0" into a false "no text-class field / exit 1".
**Fix:**
```python
    for k in range(node.get_child_count()):
        try:
            child = node.get_child_at_index(k)
        except Exception:
            continue  # one dead child hides only itself
        stack.append(child)
```
(`get_child_count()` itself may stay under the existing guard.)

### WR-04: CONFIG.md evidence quantifier «двумя раундами в каждой» unsupported for the snap rung

**File:** `docs/CONFIG.md:458`
**Issue:** The Telegram boundary section states the probe covered all three
builds «двумя раундами в каждой» (two rounds in each). The journal row
corpus (docs/adr/adr-008-spike-log.md, phase-13 section) contains
`a11y-проба telegram-snap r1`, `telegram-flatpak r1/r2`, `tarball r1/r2`
— there is NO dated snap r2 row, and no embedded second-round snap
evidence in any row. The ×2-for-snap claim exists only as an assertion
inside the summary row («a11y-проба итог»: snap ЗЕЛ/КРАСН ×2). So either
the snap r2 round ran and its row never landed (evidence loss — the exact
R1 silent-failure mode the journal pin guards against) or the quantifier
overstates. The substantive verdict (autocorrect unattainable; Qt6-bridge
ceiling, not confinement) does not hinge on snap ×2 and stands on the
remaining rows — but a shipped factual claim about evidence strength
should match the dated corpus.
**Fix:** Either land the missing dated snap r2 row in the journal before
shipping, or correct the quantifier, e.g.: «Проба трёх сборок — snap 7.3.3
(один датированный раунд), flatpak и официальный tarball (по два раунда)…»
and align the summary row's «×2» for snap accordingly.

## Info

### IN-01: Exit-code-2 gloss omits the usage-error arm of the contract

**File:** `docs/CONFIG.md:420`, `README.md:367-369`, `README.en.md:362-364`
**Issue:** The script's contract (docstring lines 24-30, main lines
209-210) returns 2 both for "no live process matches" AND for usage errors
(wrong argument count). All three doc locations gloss 2 as only the
no-process case. Harmless in practice (the usage error prints its own
usage line), but the docs' description of code 2 is incomplete relative to
the contract the phase live-proved.
**Fix:** Extend the gloss: «2 — нет живого процесса с таким именем (или
ошибка вызова)» / «2 — no process (or a usage error)».

### IN-02: f-string without placeholders in the password branch

**File:** `scripts/a11y-selfcheck.py:194`
**Issue:** `print(f"password field found (role 40) — видимо, но
неисправимо"` has an f-prefix with no interpolated expression (F541-class
lint), and the message mixes RU and EN mid-sentence. The RU phrase is
deliberate terminology (it quotes the CONFIG semantic for role 40), so
only the stray f-prefix is a defect; no Python linter runs in CI today,
so nothing catches it.
**Fix:** Drop the `f` prefix on the no-placeholder string (the f-strings
with real placeholders elsewhere stay).

---

_Reviewed: 2026-11-06T14:54:50Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
