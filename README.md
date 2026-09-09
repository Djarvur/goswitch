# goswitch

Keyboard layout fixer & text corrector for GNOME Wayland. Go + IBus engine.

Punto-Switcher-style **hotkey correction** (`ghbdtn` → `привет`) that:
- fixes last word / whole phrase / selection,
- works via an IBus input method engine (no evdev grab — coexists with keyd/xremap),
- commits text directly instead of emulating backspaces,
- ships with a full end-to-end test harness (uinput injection + AT-SPI verification).

Status: **specification phase** — see [docs/SPEC.md](docs/SPEC.md).

## License

MIT
