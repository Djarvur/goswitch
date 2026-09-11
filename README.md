# goswitch

Keyboard layout fixer & text corrector for GNOME Wayland. Go + IBus engine.

Punto-Switcher-style **hotkey correction** (`ghbdtn` → `привет`) that:
- fixes last word / whole phrase / selection,
- works via an IBus input method engine (no evdev grab — coexists with keyd/xremap),
- commits text directly instead of emulating backspaces,
- ships with a full end-to-end test harness (uinput injection + AT-SPI verification).

Status: **specification phase** — see [docs/SPEC.md](docs/SPEC.md).

## Режим отладки и конфиденциальность

Флаг `-debug` включает трассировку нажатий клавиш: каждое событие (keyval,
keycode, модификаторы, press/release) пишется в JSON-лог уровня DEBUG.

**Предупреждение:** такие логи могут содержать пароли и другой вводимый
секретный текст — трассируются ВСЕ нажатия, включая набранные в полях ввода
паролей. В обычном (production) режиме stderr демона уходит в systemd journal
без трассировки клавиш. Не оставляйте `-debug` включённым постоянно; включайте
его только на время отладки и удаляйте логи после разбора.

## Dev-запуск (без root)

Демон работает в пользовательской сессии как systemd user unit, привязанный к
graphical-session.target (см. [dist/systemd/user/goswitchd.service](dist/systemd/user/goswitchd.service)):

```sh
go install ./cmd/goswitchd
cp dist/systemd/user/goswitchd.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user start goswitchd
```

Для трассировки клавиш добавьте `-debug` к `ExecStart` (см. предупреждение
выше). Логи: `journalctl --user -u goswitchd -f`.

## License

MIT
