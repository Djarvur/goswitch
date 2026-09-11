#!/usr/bin/python3
"""AT-SPI helper for the goswitch e2e stand (TEST-03).

Subcommands:
  witness     print "<app>:<role>:chars=<n>" of the focused input surface.
              Input-surface roles (TEXT, PASSWORD_TEXT, ENTRY) are preferred
              over window frames, which also carry the FOCUSED state in
              GNOME Shell's tree. The shell's own entries — overview search
              and lock-screen prompt alike — surface as PASSWORD_TEXT
              (live-verified in phase 01-03).
  text <app>  print the text of the app's first text-bearing node — the
              desktop-input readback primitive (bridge to Phase 2 TEST-04).
  focus <app> best-effort Component.grabFocus on the app's newest window
              frame. Under GNOME Wayland focus-stealing prevention a
              background grab is refused (GTK4 errors, GTK3 returns false —
              live-verified in phase 01-03), which is why the stand drives
              its own input surface (the shell's search entry) instead.

Every caller must use /usr/bin/python3: PATH python3 is linuxbrew without
gi (01-PATTERNS § test/e2e).

Exit codes: 0 success, 1 named failure, 2 usage error.
"""
import re
import sys

import gi

gi.require_version("Atspi", "2.0")
from gi.repository import Atspi  # noqa: E402

INPUT_ROLES = ("TEXT", "PASSWORD_TEXT", "ENTRY")


def role_name(node):
    """Return the stable role label (TEXT, PASSWORD_TEXT, ...)."""
    match = re.match(r".*ATSPI_ROLE_(\w+)", str(node.get_role()))
    return match.group(1) if match else str(node.get_role())


def applications():
    """Yield every AT-SPI application node on the desktop."""
    desktop = Atspi.get_desktop(0)
    for i in range(desktop.get_child_count()):
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


def char_count(node):
    """Return the node's text length, or -1 when it has no Text iface."""
    try:
        return node.get_character_count()
    except Exception:
        return -1


def cmd_witness():
    """Print the focused input surface as app:role:chars=N, or (none)."""
    fallback = None
    for app in applications():
        for node in walk(app):
            try:
                if not node.get_state_set().contains(Atspi.StateType.FOCUSED):
                    continue
            except Exception:
                continue
            role = role_name(node)
            if fallback is None:
                fallback = (app.get_name(), role)
            if role in INPUT_ROLES:
                print(f"{app.get_name()}:{role}:chars={char_count(node)}")
                return None
    if fallback is None:
        print("(none)")
    else:
        print(f"{fallback[0]}:{fallback[1]}:chars=-1")
    return None


def read_text(node):
    """Read a node's whole text.

    atspi 2.52 quirk (live-verified): Atspi.Text.get_text's introspected
    signature is broken (takes exactly 1 argument) — the deprecated
    get_text_at_offset works and returns a TextRange whose content is the
    text.
    """
    try:
        span = node.get_text_at_offset(0, Atspi.TextBoundaryType.LINE_START)
        return span.content
    except Exception:
        pass
    try:
        return node.get_text(0, node.get_character_count())
    except Exception as exc:
        print(f"focus_helper: cannot read text node: {exc}", file=sys.stderr)
        return None


def cmd_text(app_name):
    """Print the text of the app's first text-bearing node."""
    for app in applications():
        if app_name.lower() not in (app.get_name() or "").lower():
            continue
        for node in walk(app):
            if role_name(node) not in INPUT_ROLES:
                continue
            if char_count(node) < 0:
                continue
            text = read_text(node)
            if text is not None:
                print(text)
                return None
    return f"no text-bearing node found in application: {app_name}"


def cmd_focus(app_name):
    """Grab focus for the newest window frame of the named application."""
    target = None
    for app in applications():
        if app_name.lower() in (app.get_name() or "").lower():
            target = app
    if target is None:
        return f"application not found in the AT-SPI tree: {app_name}"
    count = target.get_child_count()
    if count == 0:
        return f"application has no window yet: {app_name}"
    window = target.get_child_at_index(count - 1)
    try:
        if not window.grab_focus():
            return f"grabFocus refused for {app_name} (Wayland focus-stealing prevention)"
    except Exception as exc:
        return f"grabFocus failed for {app_name}: {exc}"
    return None


def main(argv):
    if len(argv) == 2 and argv[1] == "witness":
        problem = cmd_witness()
    elif len(argv) == 3 and argv[1] == "text":
        problem = cmd_text(argv[2])
    elif len(argv) == 3 and argv[1] == "focus":
        problem = cmd_focus(argv[2])
    else:
        print("usage: focus_helper.py witness | text <app> | focus <app-name>", file=sys.stderr)
        return 2
    if problem is not None:
        print(f"focus_helper: {problem}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
