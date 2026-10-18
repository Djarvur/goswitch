// Package uinject is the e2e stand's keyboard injector: a tiny in-repo
// uinput emitter that replaces the ydotool dependency of the host selfdrive
// driver. The 2026-10-18 pin: Ubuntu 24.04's ydotoold.service runs as root
// and binds /tmp/.ydotool_socket root:root 0600 — the user's client gets
// Permission denied — while /dev/uinput (crw--w---- root:input) is directly
// writable by the input-group user the stand runs as. This is the
// STACK.md-reserved fallback: ~150 LOC of stdlib syscall ioctls, no unsafe,
// no cgo, no new module dependency.
//
// The kernel contract (every number verified against the installed
// linux-libc-dev 6.8 headers, never from memory): Open whitelists EV_KEY
// and EV_SYN (UI_SET_EVBIT) plus exactly the keycode table's keys
// (UI_SET_KEYBIT) — the virtual device cannot emit anything else —
// registers the legacy uinput_user_dev payload through a 1116-byte write
// (the modern UI_DEV_SETUP ioctl takes a pointer argument, and unsafe is
// denied by the lint's depguard; the write path needs none), creates the
// device (UI_DEV_CREATE) and destroys it on Close (UI_DEV_DESTROY). The
// whitelist ioctls carry their argument BY VALUE: the kernel's uinput
// implementation reads them directly from the syscall's third parameter
// (uinput_set_bit treats arg as the bit number), the same proven form the
// production Go keyboards ride. Events are 24-byte input_events (timeval
// 2×int64 + type uint16 + code uint16 + value int32, little-endian) written
// through syscall.Write, each flush separated by an EV_SYN report.
//
// Failure is loud everywhere: a missing or unwritable control node, a
// rejected whitelist entry, a short write, a key outside the table — every
// one aborts the caller with the exact cause, never a guessed code or a
// partial typing. Nothing is injected at build, import or test time: Open
// belongs to the driver's battery start, and the unit corpus covers the
// pure logic (the table and the encoders) only.
package uinject
