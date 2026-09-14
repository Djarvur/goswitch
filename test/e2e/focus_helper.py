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
              App-name matching picks the FIRST matching application in the
              desktop tree, so it is only safe for uniquely named apps.
  focused-text
              print the text of the currently focused input surface — the
              readback for stand-spawned apps whose AT-SPI name is shared
              with the owner's own instances (e.g. "Google Chrome",
              live-verified 2026-09-12): app-name lookup would read the
              owner's window instead of the stand's.
  focused-text-pid <pid>
              focused-text scoped to the pid's application — the
              instance-exact readback pairing with focused-input-pid.
  focused-inputs
              print one app:role:chars=N line per focused input surface.
              AT-SPI focus can lag the compositor (a shell overlay may
              hold real focus while the case's node still reports focused,
              live-verified 2026-09-12) — the stand enumerates every
              candidate and refuses to inject while the shell's own entry
              is among them.
  focused-input-pid <pid>
              print role:chars=N when the focused input surface belongs
              to the application with the given process id, else (none).
              Instance-exact witness for stand-spawned apps whose AT-SPI
              name is shared with the owner's own instances.
  grab-input-pid <pid>
              grabFocus on the pid's first focusable input node. GTK4
              answers the call with an error and STILL services the
              widget grab as a fresh activation request whose timestamp
              outranks a focus-stealing denial (live-verified 2026-09-14)
              — the error is swallowed on purpose.
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
    """Read a node's whole (first-line) text.

    Two server-side quirks, both live-verified (plan 02-02):
    - GTK4's AT-SPI bridge only serves the modern GetStringAtOffset and
      raises on the deprecated GetTextAtOffset ("deprecated in favor of
      GetStringAtOffset"); Chromium answers the deprecated one. The
      modern granularity call goes first.
    - Atspi.Text.get_text's introspected signature is broken in atspi
      2.52 (takes exactly 1 argument) — kept as a last resort for
      servers that answer neither range call.
    """
    try:
        span = node.get_string_at_offset(0, Atspi.TextGranularity.LINE)
        return span.content
    except Exception:
        pass
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


def focused_inputs():
    """Yield (app_name, node) of every focused input-surface node.

    AT-SPI focus state can lag the compositor: a shell overlay may hold
    real keyboard focus while a freshly opened window still reports its
    node focused (live-verified 2026-09-12). Enumerating ALL candidates
    lets the stand refuse to inject while the shell's own entry is among
    them (plan 02-02).
    """
    for app in applications():
        for node in walk(app):
            try:
                if not node.get_state_set().contains(Atspi.StateType.FOCUSED):
                    continue
            except Exception:
                continue
            if role_name(node) in INPUT_ROLES:
                yield app.get_name(), node


def focused_input():
    """Return the first focused input-surface node, or None.

    Same node the witness reports. Readback keyed on this node cannot land
    in a same-named application of the owner (plan 02-02).
    """
    for _, node in focused_inputs():
        return node
    return None


def cmd_focused_text():
    """Print the text of the currently focused input surface."""
    node = focused_input()
    if node is None:
        return "no focused input surface"
    text = read_text(node)
    if text is None:
        return "cannot read the focused input surface"
    print(text)
    return None


def cmd_focused_text_pid(pid_str):
    """Print the focused input node's text, scoped to the pid's application."""
    try:
        pid = int(pid_str)
    except ValueError:
        return f"not a process id: {pid_str}"
    app = app_by_pid(pid)
    if app is None:
        return f"no application with pid {pid} in the AT-SPI tree"
    for node in walk(app):
        try:
            if not node.get_state_set().contains(Atspi.StateType.FOCUSED):
                continue
            if role_name(node) not in INPUT_ROLES:
                continue
        except Exception:
            continue
        text = read_text(node)
        if text is None:
            return "cannot read the focused input surface"
        print(text)
        return None
    return "no focused input surface in the pid's application"


def cmd_focused_inputs():
    """Print one app:role:chars=N line per focused input surface."""
    seen = False
    for app_name, node in focused_inputs():
        seen = True
        print(f"{app_name}:{role_name(node)}:chars={char_count(node)}")
    if not seen:
        print("(none)")
    return None


def app_by_pid(pid):
    """Return the application node whose process id is pid, or None."""
    for app in applications():
        try:
            if app.get_process_id() == pid:
                return app
        except Exception:
            continue
    return None


def first_input_node(app):
    """Return the app's first focusable input-surface node, or None."""
    for node in walk(app):
        try:
            if role_name(node) not in INPUT_ROLES:
                continue
            if not node.get_state_set().contains(Atspi.StateType.FOCUSABLE):
                continue
        except Exception:
            continue
        return node
    return None


def cmd_focused_input_pid(pid_str):
    """Print role:chars=N when the focused input surface has pid_str's pid."""
    try:
        pid = int(pid_str)
    except ValueError:
        return f"not a process id: {pid_str}"
    app = app_by_pid(pid)
    if app is not None:
        for node in walk(app):
            try:
                if not node.get_state_set().contains(Atspi.StateType.FOCUSED):
                    continue
                if role_name(node) not in INPUT_ROLES:
                    continue
            except Exception:
                continue
            print(f"{role_name(node)}:chars={char_count(node)}")
            return None
    print("(none)")
    return None


def cmd_grab_input_pid(pid_str):
    """GrabFocus the pid's first focusable input node, swallowing the error.

    GTK4 refuses the AT-SPI call ("This method is deprecated..."-class
    atspi_error) and still services the underlying widget grab as a fresh
    window-activation request — the refusal is expected, the grab works
    (live-verified 2026-09-14, plan 02-02).
    """
    try:
        pid = int(pid_str)
    except ValueError:
        return f"not a process id: {pid_str}"
    app = app_by_pid(pid)
    if app is None:
        return f"no application with pid {pid} in the AT-SPI tree (still starting?)"
    node = first_input_node(app)
    if node is None:
        return f"application {pid} has no focusable input node yet"
    try:
        node.grab_focus()
    except Exception:
        pass  # the refused call still triggers the fresh activation request
    return None


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
    elif len(argv) == 2 and argv[1] == "focused-text":
        problem = cmd_focused_text()
    elif len(argv) == 3 and argv[1] == "focused-text-pid":
        problem = cmd_focused_text_pid(argv[2])
    elif len(argv) == 2 and argv[1] == "focused-inputs":
        problem = cmd_focused_inputs()
    elif len(argv) == 3 and argv[1] == "focused-input-pid":
        problem = cmd_focused_input_pid(argv[2])
    elif len(argv) == 3 and argv[1] == "grab-input-pid":
        problem = cmd_grab_input_pid(argv[2])
    elif len(argv) == 3 and argv[1] == "focus":
        problem = cmd_focus(argv[2])
    else:
        print("usage: focus_helper.py witness | text <app> | focused-text | focused-text-pid <pid>"
              " | focused-inputs | focused-input-pid <pid> | grab-input-pid <pid>"
              " | focus <app-name>", file=sys.stderr)
        return 2
    if problem is not None:
        print(f"focus_helper: {problem}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
