package uinject

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"syscall"
	"time"
)

// The named errors of the injector — every refusal errors.Is-checkable
// (the house ASVS V5 idiom).
var (
	// ErrKeyUnknown is the fail-closed miss of the keycode vocabulary: the
	// requested key has no physical mapping and the injector never guesses
	// a code. Every lookup miss (Keycode callers, resolve, the Type/Chord
	// paths) wraps it.
	ErrKeyUnknown = errors.New("key is outside the injector's keycode table")

	// ErrShortWrite is the control node's partial write: an event (or the
	// registration payload) left the injector incomplete — loud, never a
	// silently partial injection.
	ErrShortWrite = errors.New("short write on the uinput control node")

	// ErrClosed is the lifecycle refusal: Close (or any inject) on an
	// already-destroyed injector.
	ErrClosed = errors.New("injector already closed")
)

// The uinput control requests and event-type codes, transcribed from the
// installed headers (linux/uinput.h, linux/input-event-codes.h — each
// number verified on this machine via gcc, never from memory).
const (
	uiSetEvbit   = 0x40045564 // UI_SET_EVBIT _IOW('U', 100, int)
	uiSetKeybit  = 0x40045565 // UI_SET_KEYBIT _IOW('U', 101, int)
	uiDevCreate  = 0x5501     // UI_DEV_CREATE _IO('U', 1)
	uiDevDestroy = 0x5502     // UI_DEV_DESTROY _IO('U', 2)

	evSyn = 0x00 // EV_SYN: the report separator
	evKey = 0x01 // EV_KEY: a key press/release

	// The input_event.value states of an EV_KEY event.
	keyRelease = 0
	keyPress   = 1
)

// The input_event wire layout on this ABI (sizeof = 24, verified): timeval
// (tv_sec int64, tv_usec int64) + type uint16 + code uint16 + value int32,
// little-endian. The timeval is stamped explicitly (time.Now()) — honest
// timestamps, never a zeroed struct.
const (
	evSize     = 24
	evSecOff   = 0
	evUsecOff  = 8
	evTypeOff  = 16
	evCodeOff  = 18
	evValueOff = 20

	nsecPerUsec = 1000
)

// DevicePath is the kernel's virtual-input control node — root:input 0660,
// directly writable by the input-group user the stand runs as.
const DevicePath = "/dev/uinput"

// The virtual keyboard's registration payload (the legacy uinput_user_dev
// write — the modern UI_DEV_SETUP ioctl takes a pointer argument and unsafe
// is denied, while the write path needs none): name[80] + input_id (4×
// uint16) + ff_effects_max (uint32) + four zeroed abs arrays (ABS_CNT=64 ×
// int32 each — a keyboard has no absolute axes). The kernel sizes the
// registration by the struct's exact size (verified: 1116 bytes).
const (
	devNameSize  = 80
	devIDOff     = 80
	devVendorOff = 82
	devProductOf = 84

	// devVersionOff is the input_id version field's offset.
	devVersionOff = 86
	devUserDevSiz = 1116

	// The device identity (visible as the input device's name) and the USB
	// id triple of the registration payload (BUS_USB).
	devName    = "goswitch-e2e-uinject"
	devBusUSB  = 0x03
	devVendor  = 0x1
	devProduct = 0x1
	devVersion = 0x1

	// ioctlArgNone is the unused by-value argument of the _IO requests.
	ioctlArgNone = 0
)

// uinjectKeycodes is the injector's whole vocabulary: exactly the keys the
// host selfdrive driver types — the digits (the layout-independent marker),
// the correction corpus's letters (ghbdtn plus the Ctrl-U u) and the named
// chord keys. Every value is transcribed verbatim from
// /usr/include/linux/input-event-codes.h (the positional keyboard rows,
// NOT alphabetical — the transcription rule of internal/session/probe.go;
// KEY_G=34, KEY_H=35: the work order's g=35/h=36 would be KEY_H/KEY_J).
// A name outside the table is a loud ErrKeyUnknown, never a guessed code.
//
//nolint:gochecknoglobals,mnd // the transcribed header table, cited verbatim (read-only)
var uinjectKeycodes = map[string]uint16{
	"1": 2, "2": 3, "3": 4, "4": 5, "5": 6, "6": 7, "7": 8, "8": 9, "9": 10, "0": 11,
	"g": 34, "h": 35, "b": 48, "d": 32, "t": 20, "n": 49, "u": 22,
	"q": 16, "w": 17, "e": 18, "r": 19, "y": 21, "i": 23, "o": 24, "p": 25,
	"a": 30, "s": 31, "f": 33, "j": 36, "k": 37, "l": 38,
	"z": 44, "x": 45, "c": 46, "v": 47, "m": 50,
	"-": 12, ",": 51, ".": 52,
	" ": 57,
	"BackSpace": 14, "ShiftRight": 54, "LeftCtrl": 29, "Enter": 28, "Space": 57,
}

// Keycode resolves one key name to its evdev keycode: the second result is
// false for any name outside the table — callers fail loudly with
// ErrKeyUnknown (never a guessed code).
func Keycode(name string) (uint16, bool) {
	code, ok := uinjectKeycodes[name]

	return code, ok
}

// EncodePress builds the 24-byte input_event of a key press (value 1)
// stamped with t.
func EncodePress(code uint16, t time.Time) [evSize]byte {
	return encodeEvent(evKey, code, keyPress, t)
}

// EncodeRelease builds the 24-byte input_event of a key release (value 0)
// stamped with t.
func EncodeRelease(code uint16, t time.Time) [evSize]byte {
	return encodeEvent(evKey, code, keyRelease, t)
}

// EncodeSyn builds the report separator: EV_SYN type 0, code 0, value 0 —
// the flush that hands the pending events to the kernel's readers.
func EncodeSyn(t time.Time) [evSize]byte {
	return encodeEvent(evSyn, 0, 0, t)
}

// encodeEvent lays out one input_event: the explicit timeval (honest
// wall-clock stamps), the type/code pair and the value, all little-endian
// (the kernel ABI of this platform).
func encodeEvent(typ, code uint16, value uint32, t time.Time) [evSize]byte {
	var ev [evSize]byte
	sec, usec := epochParts(t)
	binary.LittleEndian.PutUint64(ev[evSecOff:], sec)
	binary.LittleEndian.PutUint64(ev[evUsecOff:], usec)
	binary.LittleEndian.PutUint16(ev[evTypeOff:], typ)
	binary.LittleEndian.PutUint16(ev[evCodeOff:], code)
	binary.LittleEndian.PutUint32(ev[evValueOff:], value)

	return ev
}

// epochParts splits t into the timeval's unsigned (sec, usec): the
// pre-epoch clamps keep the conversions honest for any wall clock.
func epochParts(t time.Time) (sec, usec uint64) {
	s := max(t.Unix(), 0)
	u := max(t.Nanosecond()/nsecPerUsec, 0)

	return uint64(s), uint64(u)
}

// resolve checks every rune of text against the keycode table BEFORE any
// key leaves: the first miss is a loud ErrKeyUnknown naming the rune and
// its code point (never a guessed code, never a partial typing).
func resolve(text string) error {
	for _, r := range text {
		if _, ok := Keycode(string(r)); !ok {
			return fmt.Errorf("%w: rune %q (U+%04X)", ErrKeyUnknown, r, r)
		}
	}

	return nil
}

// resolveKey resolves one named key, loud on a table miss.
func resolveKey(name string) (uint16, error) {
	code, ok := Keycode(name)
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrKeyUnknown, name)
	}

	return code, nil
}

// Injector is one open uinput virtual keyboard: Open whitelists its keys
// and creates the device, the methods inject through the kernel's evdev
// path, Close destroys it. The zero value is unusable — always Open.
type Injector struct {
	fd    int // the /dev/uinput control fd; the -1 sentinel once closed
	sleep func(ctx context.Context, d time.Duration) error
}

// Option configures an Injector at Open time.
type Option func(*Injector)

// WithClock replaces the injector's pacing clock — the tests' seam: the
// unit corpus exercises the pure table and encoders only and never opens
// the device, so nothing ever sleeps on a real timer.
func WithClock(fn func(ctx context.Context, d time.Duration) error) Option {
	return func(inj *Injector) {
		inj.sleep = fn
	}
}

// Open creates the virtual keyboard: it opens the control node
// (fail-closed — the stand user must be in the input group), whitelists
// EV_KEY/EV_SYN and exactly the table's key codes, registers the legacy
// uinput_user_dev payload and creates the device. Nothing is injected
// until a method call — and the whitelist means nothing outside the table
// CAN be.
func Open(opts ...Option) (*Injector, error) {
	inj := &Injector{fd: -1, sleep: realSleep}
	for _, opt := range opts {
		opt(inj)
	}
	fd, err := syscall.Open(DevicePath, syscall.O_WRONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w (the stand user must be in the input group)", DevicePath, err)
	}
	inj.fd = fd
	if err := inj.setup(); err != nil {
		_ = syscall.Close(fd)
		inj.fd = -1

		return nil, err
	}

	return inj, nil
}

// Tap presses and releases one key: the press, the pacing hold, the
// release (the marker digit, the gesture taps, Type's rune loop).
func (inj *Injector) Tap(ctx context.Context, name string, delayMs int) error {
	code, err := resolveKey(name)
	if err != nil {
		return err
	}
	if err := inj.writeKey(code, keyPress); err != nil {
		return fmt.Errorf("%s press: %w", name, err)
	}
	if err := inj.pause(ctx, delayMs); err != nil {
		return fmt.Errorf("%s hold: %w", name, err)
	}
	if err := inj.writeKey(code, keyRelease); err != nil {
		return fmt.Errorf("%s release: %w", name, err)
	}

	return nil
}

// Press holds one key down (a chord's step; the caller owns the pacing
// between the chord's steps).
func (inj *Injector) Press(name string) error {
	code, err := resolveKey(name)
	if err != nil {
		return err
	}
	if err := inj.writeKey(code, keyPress); err != nil {
		return fmt.Errorf("%s press: %w", name, err)
	}

	return nil
}

// Release lifts one held key (a chord's step).
func (inj *Injector) Release(name string) error {
	code, err := resolveKey(name)
	if err != nil {
		return err
	}
	if err := inj.writeKey(code, keyRelease); err != nil {
		return fmt.Errorf("%s release: %w", name, err)
	}

	return nil
}

// Type types text rune by rune with the pacing hold per key (the marker
// digit, the correction corpus): every rune resolves BEFORE the first key
// leaves — a miss is a loud ErrKeyUnknown naming the rune, never a partial
// typing.
func (inj *Injector) Type(ctx context.Context, text string, delayMs int) error {
	if err := resolve(text); err != nil {
		return err
	}
	for _, r := range text {
		if err := inj.Tap(ctx, string(r), delayMs); err != nil {
			return err
		}
	}

	return nil
}

// Chord taps key names sequentially with the pacing hold — the driver's
// ssKeys form (ydotool's space-separated key list), NOT a simultaneous
// press: each key completes before the next begins.
func (inj *Injector) Chord(ctx context.Context, names []string, delayMs int) error {
	for _, name := range names {
		if err := inj.Tap(ctx, name, delayMs); err != nil {
			return err
		}
	}

	return nil
}

// Close destroys the virtual device and closes the control node. The
// closed sentinel is set FIRST, so a double Close is a loud ErrClosed
// instead of a second destroy on a stale fd.
func (inj *Injector) Close() error {
	if inj.fd < 0 {
		return fmt.Errorf("close %s: %w", DevicePath, ErrClosed)
	}
	fd := inj.fd
	inj.fd = -1
	destroyErr := inj.ioctl(uiDevDestroy, ioctlArgNone)
	if err := syscall.Close(fd); err != nil && destroyErr == nil {
		return fmt.Errorf("close %s: %w", DevicePath, err)
	}
	if destroyErr != nil {
		return fmt.Errorf("UI_DEV_DESTROY: %w", destroyErr)
	}

	return nil
}

// setup whitelists the event types and the key codes, registers the device
// payload and creates the virtual keyboard.
func (inj *Injector) setup() error {
	for _, bit := range []uint16{evSyn, evKey} {
		if err := inj.ioctl(uiSetEvbit, bit); err != nil {
			return fmt.Errorf("UI_SET_EVBIT(%d): %w", bit, err)
		}
	}
	seen := make(map[uint16]struct{}, len(uinjectKeycodes))
	for _, code := range uinjectKeycodes {
		if _, dup := seen[code]; dup {
			continue
		}
		seen[code] = struct{}{}
		if err := inj.ioctl(uiSetKeybit, code); err != nil {
			return fmt.Errorf("UI_SET_KEYBIT(%d): %w", code, err)
		}
	}
	if err := inj.registerDevice(); err != nil {
		return err
	}
	if err := inj.ioctl(uiDevCreate, ioctlArgNone); err != nil {
		return fmt.Errorf("UI_DEV_CREATE: %w", err)
	}

	return nil
}

// registerDevice writes the legacy uinput_user_dev payload: the device
// name, the USB id triple and ff_effects_max — the abs arrays ride zeroed
// (a keyboard has no absolute axes).
func (inj *Injector) registerDevice() error {
	dev := make([]byte, devUserDevSiz)
	copy(dev[:devNameSize], devName)
	binary.LittleEndian.PutUint16(dev[devIDOff:], devBusUSB)
	binary.LittleEndian.PutUint16(dev[devVendorOff:], devVendor)
	binary.LittleEndian.PutUint16(dev[devProductOf:], devProduct)
	binary.LittleEndian.PutUint16(dev[devVersionOff:], devVersion)
	if n, err := syscall.Write(inj.fd, dev); err != nil {
		return fmt.Errorf("register %s: %w", devName, err)
	} else if n != devUserDevSiz {
		return fmt.Errorf("register %s: %w (%d of %d bytes)", devName, ErrShortWrite, n, devUserDevSiz)
	}

	return nil
}

// ioctl runs one uinput control request with its BY-VALUE argument: the
// kernel's uinput implementation reads UI_SET_EVBIT/UI_SET_KEYBIT's
// argument directly from the syscall's third parameter (uinput_set_bit
// treats arg as the bit number), so no pointer indirection — and no
// unsafe — is involved (the by-value form the production Go keyboards
// ride).
func (inj *Injector) ioctl(req uint32, arg uint16) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(inj.fd), uintptr(req), uintptr(arg)); errno != 0 {
		return fmt.Errorf("ioctl %#x arg %#x: %w", req, arg, errno)
	}

	return nil
}

// emit writes one 24-byte input_event followed by its EV_SYN flush — the
// kernel hands evdev packets to readers on the separator only, so every
// step is delivered the moment it is written.
func (inj *Injector) emit(ev [evSize]byte) error {
	if err := inj.write(ev[:], "event"); err != nil {
		return err
	}
	syn := EncodeSyn(time.Now())

	return inj.write(syn[:], "syn")
}

// write is one bounded syscall.Write: a short write on the control node
// wraps ErrShortWrite, never a silently partial event.
func (inj *Injector) write(buf []byte, what string) error {
	n, err := syscall.Write(inj.fd, buf)
	if err != nil {
		return fmt.Errorf("write %s: %w", what, err)
	}
	if n != len(buf) {
		return fmt.Errorf("write %s: %w (%d of %d bytes)", what, ErrShortWrite, n, len(buf))
	}

	return nil
}

// writeKey emits one EV_KEY event: the press or release state of code.
func (inj *Injector) writeKey(code uint16, value uint32) error {
	return inj.emit(encodeEvent(evKey, code, value, time.Now()))
}

// pause holds the pacing delay through the injector's clock —
// context-aware: an aborted battery interrupts a hold, never wedges it.
func (inj *Injector) pause(ctx context.Context, delayMs int) error {
	if delayMs <= 0 {
		return nil
	}

	return inj.sleep(ctx, time.Duration(delayMs)*time.Millisecond)
}

// realSleep is the production clock: an abortable timer wait — the
// driver's sleepCtx semantics (Ctrl-C interrupts a pacing hold).
func realSleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("sleep aborted: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}
