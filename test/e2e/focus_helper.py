#!/usr/bin/python3
"""AT-SPI helper for the goswitch e2e stand (TEST-03).

Subcommands:
  witness     print "<app>:<role>:chars=<n>" of the focused input surface.
              Focus-first traversal (G-4-1, plan 04-08): per application
              the FOCUSED state is checked on the window frames BEFORE
              descending, and only focused frames' subtrees are walked;
              the full-tree walk remains the fallback for when no frame is
              focused anywhere — fresh-session trees (gnome-shell 3673 +
              gjs 2437 nodes) made the bare full walk ~3.0s and crossed
              the stand's probe budget. Input-surface roles (TEXT,
              PASSWORD_TEXT, ENTRY) are preferred over window frames,
              which also carry the FOCUSED state in GNOME Shell's tree;
              when several surfaces report focused at once (bridge lag)
              the answer is the first focused-frame input node in
              desktop→frame order — exhaustive enumeration remains the
              focused-inputs contract. The shell's own entries — overview
              search and lock-screen prompt alike — surface as
              PASSWORD_TEXT (live-verified in phase 01-03), and a locked
              session still answers through the verbatim full-walk
              fallback.
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
  grab-input-pid <pid> [want-chars]
              grabFocus on the pid's first focusable input node — or, when
              want-chars is given, on the first focusable input node whose
              character count equals it (the e2e matrix re-focus: after a
              Tab, Chromium's omnibox (ENTRY, the URL's ~60 chars) is the
              app's first focusable input and would win every grab while
              the probe word sits in the page input). GTK4 answers the
              call with an error and STILL services the widget grab as a
              fresh activation request whose timestamp outranks a
              focus-stealing denial (live-verified 2026-09-14) — the error
              is swallowed on purpose.
  focus <app> best-effort Component.grabFocus on the app's newest window
              frame. Under GNOME Wayland focus-stealing prevention a
              background grab is refused (GTK4 errors, GTK3 returns false —
              live-verified in phase 01-03), which is why the stand drives
              its own input surface (the shell's search entry) instead.
  witness-events
              resident perf witness (plan 04-04, D-44): ONE
              Atspi.EventListener on object:text-changed, one startup cost,
              ONE LINE PER EVENT on stdout —
              "<RFC3339 UTC> <app> <pid> <text>". The timestamp is stamped
              at event arrival — the t1 of the measured ydotool→AT-SPI
              window; the text readback happens after the stamp, so its
              latency never lands inside the window. Events from
              non-input-surface roles are ignored. The process runs until
              killed by pid (the Go client owns the lifecycle; never
              pkill/pgrep -f — the pattern matches the caller's own
              command line, live hit in the phase research).
  witness-poll <pid>
              resident poller fallback for the same perf window (A2, used
              only if the event listener proves unstable on the desk):
              SAME line protocol, driven by a fixed in-process quantum —
              WITNESS_POLL_QUANTUM_S = 5 ms, the named constant D-44's
              methodology requires. Watches the pid's first input node and
              prints a line whenever its text CHANGES; a line per poll
              would flood the protocol without adding information. When
              the watched node dies the poller re-finds it inside the same
              pid's application; if the whole application is gone the
              poller idles and the Go await times out loudly (never
              silently).
  focused-app
              print the name of the application holding the focused window
              frame — the idle-desktop preflight primitive of the e2e
              matrix (G-4-2, plan 04-09): the matrix's first CI step asks
              "is the desktop busy" BEFORE paying for 31 cases, and a
              focused frame anywhere means a real window holds it. The
              same frames-before-descent discipline as the witness's
              focus-first pass (04-08), but with NO descent into frame
              subtrees and NO full walk — the busy/idle verdict is a
              frames-level question, exhaustive walks stay the witness
              modes' contract. "(none)" when no frame is focused
              anywhere; ANY focused frame answers with its application's
              name — shell/mutter surfaces are not classified
              modal/non-modal, interpreting the answer is the caller's
              (workflow preflight's) job. Live-proven mode: gi/AT-SPI
              needs a live session (precedent 04-04/04-08 — headless RED
              impossible).

Every caller must use /usr/bin/python3: PATH python3 is linuxbrew without
gi (01-PATTERNS § test/e2e).

Exit codes: 0 success, 1 named failure, 2 usage error.
"""
import re
import sys
import time
from datetime import datetime, timezone

import gi

gi.require_version("Atspi", "2.0")
from gi.repository import Atspi  # noqa: E402

# The fallback poller's fixed quantum (D-44): the methodology of the perf
# report names this constant as the observer's quantization when the
# witness-poll mode drives the measurement.
WITNESS_POLL_QUANTUM_S = 0.005  # 5 ms

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


def full_walk_witness():
    """Return the full-tree witness answer: a line, or "(none)".

    The verbatim body of the pre-04-08 cmd_witness — the fallback arm of
    the witness contract. Called only when the focus-first pass finds
    nothing, so every answer that does not come from a focused frame's
    subtree is bit-identical to the phase 01-03 semantics (the locked
    session's shell PASSWORD_TEXT included).
    """
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
                return f"{app.get_name()}:{role}:chars={char_count(node)}"
    if fallback is None:
        return "(none)"
    return f"{fallback[0]}:{fallback[1]}:chars=-1"


def focused_frame_witness():
    """Focus-first pass: answer from focused frames' subtrees only, or None.

    G-4-1 (04-UAT.md): the bare full walk touches EVERY node on the
    desktop — one D-Bus roundtrip each — and fresh-session trees
    (gnome-shell 3673 + gjs 2437 nodes) cost ~3.0s bare, crossing the
    stand's probe budget and failing healthy quiesce gates. Here the
    FOCUSED state is checked on each application's DIRECT children
    (frames, child order) BEFORE any descent, and only focused frames'
    subtrees are walked, with the per-node logic of full_walk_witness
    verbatim: per-node FOCUSED check, INPUT_ROLES preference, first
    focused node of any role as the frame-fallback. Returns the witness
    line of the first focused INPUT_ROLES node in desktop→frame order, or
    the frame-fallback line (chars=-1) when the focused subtrees hold
    only non-input nodes; None means nothing was focused anywhere and the
    caller falls back to the full walk.
    """
    fallback = None
    for app in applications():
        try:
            frames = [app.get_child_at_index(k) for k in range(app.get_child_count())]
        except Exception:
            continue
        for frame in frames:
            try:
                if not frame.get_state_set().contains(Atspi.StateType.FOCUSED):
                    continue
            except Exception:
                continue
            for node in walk(frame):
                try:
                    if not node.get_state_set().contains(Atspi.StateType.FOCUSED):
                        continue
                except Exception:
                    continue
                role = role_name(node)
                if fallback is None:
                    fallback = (app.get_name(), role)
                if role in INPUT_ROLES:
                    return f"{app.get_name()}:{role}:chars={char_count(node)}"
    if fallback is not None:
        return f"{fallback[0]}:{fallback[1]}:chars=-1"
    return None


def cmd_witness():
    """Print the focused input surface as app:role:chars=N, or (none).

    Focus-first first: the answer comes from focused frames' subtrees (an
    input line, or the frame-fallback chars=-1 arm). Only when that pass
    finds nothing does the verbatim full walk answer — every
    no-focused-frame scenario stays bit-identical to the pre-04-08
    semantics.
    """
    answer = focused_frame_witness()
    if answer is None:
        answer = full_walk_witness()
    print(answer)
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


def cmd_focused_app():
    """Print the application holding the focused frame, or (none).

    The idle-desktop preflight primitive (G-4-2, plan 04-09): the e2e
    matrix's first CI step asks "is the desktop busy" BEFORE paying for
    31 cases — one focused frame anywhere means a real window holds the
    desktop. Same frames-before-descent shape as focused_frame_witness
    (04-08): the FOCUSED state is checked on each application's DIRECT
    children (frames) only — NO descent into frame subtrees and NO full
    walk (full_walk_witness stays the witness modes' exclusive); the
    busy/idle question is decided at the frames level. Deliberate
    simplification per the plan: shell/mutter surfaces are NOT classified
    modal/non-modal — the name of ANY application with a focused frame
    is the answer ("(unnamed)" when the AT-SPI name is empty), "(none)"
    when no frame is focused anywhere; interpreting the answer is the
    caller's (workflow preflight's) job, not the helper's.
    """
    for app in applications():
        try:
            frames = [app.get_child_at_index(k) for k in range(app.get_child_count())]
        except Exception:
            continue
        for frame in frames:
            try:
                if not frame.get_state_set().contains(Atspi.StateType.FOCUSED):
                    continue
            except Exception:
                continue
            print(app.get_name() or "(unnamed)")
            return None
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


def cmd_grab_input_pid(pid_str, want_chars=None):
    """GrabFocus the pid's first focusable input node, swallowing the error.

    GTK4 refuses the AT-SPI call ("This method is deprecated..."-class
    atspi_error) and still services the underlying widget grab as a fresh
    window-activation request — the refusal is expected, the grab works
    (live-verified 2026-09-14, plan 02-02). With want_chars, only nodes
    holding exactly that character count qualify (see the usage note: the
    post-Tab omnibox must not shadow the page input).
    """
    try:
        pid = int(pid_str)
    except ValueError:
        return f"not a process id: {pid_str}"
    app = app_by_pid(pid)
    if app is None:
        return f"no application with pid {pid} in the AT-SPI tree (still starting?)"
    for node in walk(app):
        try:
            if role_name(node) not in INPUT_ROLES:
                continue
            if not node.get_state_set().contains(Atspi.StateType.FOCUSABLE):
                continue
            if want_chars is not None and char_count(node) != want_chars:
                continue
        except Exception:
            continue
        try:
            node.grab_focus()
        except Exception:
            pass  # the refused call still triggers the fresh activation request
        return None
    if want_chars is not None:
        return f"application {pid} has no focusable input node with {want_chars} chars yet"
    return f"application {pid} has no focusable input node yet"


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


def app_of(node):
    """Return the application node above node in the AT-SPI tree, or None.

    The application is the ancestor whose parent is the desktop root —
    the same instance-exact identity the pid-keyed readbacks use (plan
    02-02), here needed to name and pid-tag the witness lines.
    """
    try:
        cur = node
        while cur is not None:
            parent = cur.get_parent()
            if parent is None:
                return None
            if role_name(parent).startswith("DESKTOP"):
                return cur
            cur = parent
    except Exception:
        return None
    return None


def witness_line(node, text):
    """Format one witness protocol line: "<RFC3339 UTC> <app> <pid> <text>".

    The timestamp is stamped FIRST — before the app walk and the text
    readback — so the Go side's t1 is the observation arrival, never the
    readback completion (the honest upper bound of D-44 stays tight).
    """
    ts = datetime.now(timezone.utc).isoformat()
    app = app_of(node)
    name, pid = "(unknown)", "-"
    if app is not None:
        name = app.get_name() or name
        try:
            pid = str(app.get_process_id())
        except Exception:
            pid = "-"
    return f"{ts} {name} {pid} {text}"


def cmd_witness_events():
    """Resident object:text-changed listener — one line per input event.

    The listener hears every text change on the desktop; filtering to the
    input-surface roles drops window-title and shell noise before the line
    is printed (the Go client still matches by app name and text). Runs
    until the process is killed by pid.
    """

    def on_event(event):
        try:
            node = event.source
            if role_name(node) not in INPUT_ROLES:
                return None
            print(witness_line(node, read_text(node) or ""), flush=True)
        except Exception:
            return None  # a transient node error must not kill the witness
        return None

    listener = Atspi.EventListener.new(on_event)
    if not listener.register("object:text-changed"):
        return "could not register the object:text-changed listener"
    Atspi.event_main()
    return None


def cmd_witness_poll(pid_str):
    """Resident poller fallback — same line protocol, fixed 5 ms quantum.

    One line per TEXT CHANGE (not per poll: a line per poll would flood
    the protocol without adding information). The timestamp of a change
    line is the poll that observed it, so the honest quantization of this
    mode is WITNESS_POLL_QUANTUM_S — the methodology names it (D-44).
    """
    try:
        pid = int(pid_str)
    except ValueError:
        return f"not a process id: {pid_str}"
    node = None
    last = None
    while True:
        if node is not None and char_count(node) < 0:
            node = None  # the watched node died — re-find inside the same app
        if node is None:
            app = app_by_pid(pid)
            if app is not None:
                node = first_input_node(app)
        if node is not None:
            text = read_text(node) or ""
            if text != last:
                last = text
                print(witness_line(node, text), flush=True)
        time.sleep(WITNESS_POLL_QUANTUM_S)


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
    elif len(argv) == 2 and argv[1] == "focused-app":
        problem = cmd_focused_app()
    elif len(argv) == 3 and argv[1] == "focused-input-pid":
        problem = cmd_focused_input_pid(argv[2])
    elif len(argv) == 3 and argv[1] == "grab-input-pid":
        problem = cmd_grab_input_pid(argv[2])
    elif len(argv) == 4 and argv[1] == "grab-input-pid":
        try:
            want_chars = int(argv[3])
        except ValueError:
            return 2
        problem = cmd_grab_input_pid(argv[2], want_chars)
    elif len(argv) == 3 and argv[1] == "focus":
        problem = cmd_focus(argv[2])
    elif len(argv) == 2 and argv[1] == "witness-events":
        problem = cmd_witness_events()
    elif len(argv) == 3 and argv[1] == "witness-poll":
        problem = cmd_witness_poll(argv[2])
    else:
        print("usage: focus_helper.py witness | text <app> | focused-text | focused-text-pid <pid>"
              " | focused-inputs | focused-input-pid <pid> | grab-input-pid <pid> [want-chars]"
              " | focus <app-name> | witness-events | witness-poll <pid> | focused-app", file=sys.stderr)
        return 2
    if problem is not None:
        print(f"focus_helper: {problem}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
