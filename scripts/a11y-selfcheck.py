#!/usr/bin/python3
"""AT-SPI self-check for goswitch: is an application's text field visible? (D-06)

Usage: /usr/bin/python3 scripts/a11y-selfcheck.py <app-name-substring>

An external AT-SPI probe of INPUT-FIELD VISIBILITY — the daemon is NOT
touched. The probe walks the live AT-SPI tree, finds the application by a
case-insensitive name substring, and reports the role NUMBERS of the text
fields it sees. Roles compare by NUMBER against the daemon's canon
(internal/appid/appid.go): 61 text, 79 entry, 94 document text are the
correctable text class; 40 password text is visible but NEVER corrected;
every other number is printed without assumption — which role a Qt/Electron
field actually carries is decided by the probe, never by this script. Role
names appear in the output for humans only, never in a condition: name
strings are localized and diverge between toolkits for the same enum value.

The process-liveness check reads /proc directly (os.listdir + cmdline) —
never a subprocess or a shell: the application name from argv is DATA for
the comparison, never assembled into a command.

Every caller must use /usr/bin/python3: PATH python3 is linuxbrew without
gi (01-PATTERNS § test/e2e).

Exit codes: 0 success, 1 named failure, 2 usage error.
  0 — the application is in the live AT-SPI tree AND a text-class field is
      visible (role 61/79/94): autocorrect can engage;
  1 — the application process is alive, but the tree/text field is not
      there: a11y not enabled, the app not restarted after enabling it, or
      incoming role calls denied (registered on the bus without servicing);
  2 — usage error, no live process matches the name, or the environment
      is broken (gi/Atspi unavailable, a11y bus unreachable): the probe
      cannot answer.
"""
import os
import sys

try:
    import gi

    gi.require_version("Atspi", "2.0")
    from gi.repository import Atspi  # noqa: E402
except (ImportError, ValueError) as exc:  # no gi / no Atspi typelib
    print(f"a11y-selfcheck: python3 gi/Atspi unavailable ({exc}) — "
          "run via /usr/bin/python3 (PATH python3 may lack gi)",
          file=sys.stderr)
    sys.exit(2)  # environment failure, NOT the semantic "not visible" exit 1

# The daemon's role canon (internal/appid/appid.go:85-96): the autocorrect
# policy gates on role NUMBERS. The probe prints the same numbers the daemon
# gates on; the human-readable names next to them are cosmetics only.
TEXT_ROLES = {61, 79, 94}  # RoleText / RoleEntry / RoleDocumentText
ROLE_PASSWORD_TEXT = 40  # RolePasswordText — visible but never corrected


def applications():
    """Yield every AT-SPI application node on the desktop."""
    try:
        desktop = Atspi.get_desktop(0)
        count = desktop.get_child_count()
    except Exception as exc:  # a11y bus unreachable (e.g. ssh without session env)
        print(f"a11y-selfcheck: a11y bus unreachable ({exc}) — the probe "
              "cannot answer", file=sys.stderr)
        sys.exit(2)  # environment failure, NOT the semantic "not visible" exit 1
    for i in range(count):
        app = desktop.get_child_at_index(i)
        if app is not None:
            yield app


def walk(app):
    """Breadth-first yield every node of one application tree.

    The a11y bus lags the compositor: nodes vanish while we walk (windows
    open/close during the stand's own surface switching) and any access to
    a defunct node raises. Every step is individually best-effort — a
    transient node error skips that node instead of killing the witness.
    """
    stack = [app]
    while stack:
        node = stack.pop(0)
        if node is None:
            continue
        yield node
        try:
            children = [node.get_child_at_index(k) for k in range(node.get_child_count())]
        except Exception:
            continue
        stack.extend(children)


def process_pids(needle):
    """Return pids whose executable contains needle (case-insensitive).

    Liveness is answered by READING /proc — no subprocess, no shell: the
    application name is data for the comparison, never a command part.
    """
    low = needle.lower()
    pids = []
    for entry in os.listdir("/proc"):
        if not entry.isdigit():
            continue
        try:
            with open(os.path.join("/proc", entry, "cmdline"), "rb") as fh:
                cmdline = fh.read().split(b"\x00")
        except OSError:
            continue
        if not cmdline or not cmdline[0]:
            continue
        exe = os.path.basename(cmdline[0].decode("utf-8", "replace"))
        if low in exe.lower():
            pids.append(int(entry))
    return sorted(pids)


def role_number(node):
    """Return the node's role NUMBER (the daemon's comparison canon) or None."""
    try:
        return int(node.get_role())
    except Exception:
        return None


def find_application(needle):
    """Return (app, name) of the first desktop application matching needle."""
    low = needle.lower()
    for app in applications():
        try:
            name = app.get_name() or ""
        except Exception:
            continue
        if low in name.lower():
            return app, name
    return None, None


def probe_tree(app):
    """Walk the app's tree collecting role evidence.

    Returns (role_counts, focused_roles, role_call_failures, nodes_seen):
    role_counts maps role number -> node count (numbers only — the verdict
    never reads a name); focused_roles lists the role numbers of nodes
    reporting the FOCUSED state; role_call_failures counts nodes whose
    GetRole call raised — the "registered on the bus but incoming calls
    denied" signature (the AppArmor wall of the 2026-10-08 snap verdict).
    """
    role_counts = {}
    focused_roles = []
    role_call_failures = 0
    nodes_seen = 0
    for node in walk(app):
        nodes_seen += 1
        num = role_number(node)
        if num is None:
            role_call_failures += 1
            continue
        role_counts[num] = role_counts.get(num, 0) + 1
        try:
            if node.get_state_set().contains(Atspi.StateType.FOCUSED):
                focused_roles.append(num)
        except Exception:
            pass  # a transient node error must not kill the probe
    return role_counts, focused_roles, role_call_failures, nodes_seen


def describe_role(num):
    """Human-readable label for a role number — cosmetics, never a verdict."""
    labels = {
        40: "password text (видимо, но неисправимо — автокоррекция не правит)",
        60: "terminal (never corrected)",
        61: "text",
        79: "entry",
        94: "document text",
    }
    return labels.get(num, "role (no assumption)")


def cmd_selfcheck(needle, pids):
    """Print the visibility verdict for the live application, or a problem."""
    print(f"app: {needle}")
    print(f"process: alive (pids: {' '.join(str(p) for p in pids)})")
    app, tree_name = find_application(needle)
    if app is None:
        print("at-spi tree: application NOT FOUND in desktop children")
        return ("process alive but application is not in the AT-SPI tree"
                " (a11y not enabled, or the app was not restarted after"
                " enabling it)")
    print(f"at-spi tree: found (name: {tree_name})")
    role_counts, focused_roles, failures, nodes_seen = probe_tree(app)
    seen = ", ".join(
        f"{num} x{role_counts[num]} [{describe_role(num)}]"
        for num in sorted(role_counts)
    )
    print(f"roles seen (numbers, no names in the verdict): {seen or '(none)'}")
    if failures:
        print(f"role calls failed on {failures}/{nodes_seen} nodes"
              " (incoming GetRole denied or dead nodes)")
    text_found = sorted(num for num in role_counts if num in TEXT_ROLES)
    if focused_roles:
        print("focused node roles: "
              + ", ".join(f"{num} [{describe_role(num)}]"
                          for num in sorted(set(focused_roles))))
    if text_found:
        print(f"text-class fields (61/79/94): {text_found}")
        print(f"verdict: text field VISIBLE (role {text_found[0]} in the"
              " text class) — the daemon's autocorrect gate can engage")
        return None
    if ROLE_PASSWORD_TEXT in role_counts:
        print(f"password field found (role 40) — видимо, но неисправимо"
              " (password): the autocorrect policy never corrects it")
        return ("only a password field (role 40) is visible — видимо, но"
                " неисправимо; no correctable text field (61/79/94)")
    return ("no text-class field (61/79/94) in the application tree"
            + (f"; roles seen: {sorted(role_counts)}" if role_counts
               else "; the tree answered nothing")
            + (f"; role calls failed on {failures}/{nodes_seen} nodes"
               " (registered without servicing?)" if failures else ""))


def main(argv):
    if len(argv) == 2 and argv[1]:
        needle = argv[1]
    else:
        print("usage: a11y-selfcheck.py <app-name-substring>", file=sys.stderr)
        return 2
    pids = process_pids(needle)
    if not pids:
        print(f"a11y-selfcheck: no live process matches: {needle}",
              file=sys.stderr)
        return 2
    problem = cmd_selfcheck(needle, pids)
    if problem is not None:
        print(f"a11y-selfcheck: {problem}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
