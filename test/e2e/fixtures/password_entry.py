#!/usr/bin/python3
"""GTK4 password-entry fixture of the goswitch e2e stand (plan 06-07).

A self-focusing GTK4 window (the startZenity mechanism: a freshly mapped
window takes keyboard focus without any grab) carrying BOTH input classes
the autocorrect policy must tell apart:

  GtkEntry         -- AT-SPI role TEXT (61): the correctable text input the
                      autocorrect-fires case types into;
  GtkPasswordEntry -- AT-SPI role PASSWORD_TEXT (40, live-verified on this
                      desktop 2026-09-27, 06-RESEARCH Q4): the forbidden
                      class the autocorrect-password-silent case proves
                      SILENT.

Usage: password_entry.py <timeout-seconds>

The process exits by itself after the timeout (GLib timeout -> main-loop
quit) and prints the case's stdout oracle lines — the zenity-spirit
delivery readback for a window with no CLI stdout contract:

  FINAL:<password entry text>  -- the "nothing happened" oracle of the
                                  password-silent case (the field must hold
                                  exactly what was injected, unchanged);
  PLAIN:<text entry text>      -- the field-content oracle of the
                                  autocorrect-fires case.

Privacy (T-06-07-03): the prints ARE the case's stdout oracle — the stand
never injects a real secret into this fixture, and the printed content is
the evidence the case itself asserts on (unchanged / corrected word).
Every caller must use /usr/bin/python3: PATH python3 is linuxbrew without
gi (01-PATTERNS § test/e2e).

The prgname/application-name pin gives the window its own AT-SPI bridge
identity: without it the GTK4 bridge derives the namespace from C argv[0]
("python3") — a namespace shared with any python GUI the owner runs, which
the autocorrect white list must never match.

Exit codes: 0 success, 2 usage error.
"""
import sys

import gi

gi.require_version("Gtk", "4.0")
from gi.repository import GLib, Gtk  # noqa: E402


def main(argv):
    if len(argv) != 2:
        print("usage: password_entry.py <timeout-seconds>", file=sys.stderr)
        return 2
    timeout = int(argv[1])

    GLib.set_prgname("gs-e2e-password")
    GLib.set_application_name("gs-e2e-password")
    win = Gtk.Window(title="gs-e2e-password")
    box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=6)
    plain = Gtk.Entry()
    secret = Gtk.PasswordEntry()  # AT-SPI role: password-text (40)
    box.append(plain)
    box.append(secret)
    win.set_child(box)
    win.present()

    loop = GLib.MainLoop()
    GLib.timeout_add_seconds(timeout, loop.quit)
    loop.run()
    print(f"FINAL:{secret.get_text()}", flush=True)
    print(f"PLAIN:{plain.get_text()}", flush=True)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
