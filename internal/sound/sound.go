// Package sound plays the daemon's acoustic feedback (plan 07-08, owner
// decision «Звуки при переключении», 07-CONTEXT): a short tone on EVERY
// script flip and a distinct tone on every autocorrect fire, played by a
// best-effort subprocess — the desktop's canberra-gtk-play preferred, a
// paplay fallback over bundled tones — that can never block, error or
// otherwise alter the gesture that triggered it: every failure is ONE WARN
// per episode and a quiet no-op. The tones know nothing about text — no
// word content ever reaches this package or its logs (D-20/D-21).
package sound

import (
	"bytes"
	"encoding/binary"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// The pinned player binaries (T-07-08-01: fixed names from PATH, arguments
// are flags, event names and daemon-owned cache paths only — never user
// content; no shell anywhere).
const (
	binCanberra = "canberra-gtk-play"
	binPaplay   = "paplay"
)

// The degradation reasons of the warn-once episodes (the closed vocabulary
// the corpus pins — reasons and binary names only, never user context,
// D-20/D-21).
const (
	reasonPlayerMissing = "player-missing"    // neither canberra nor paplay installed
	reasonStartFailed   = "start-failed"      // the player child could not be launched
	reasonToneWrite     = "tone-write-failed" // a bundled tone could not be materialized
)

// The tone-cache shape of the paplay fallback: the bundled tones live in
// the user's cache dir — an owner-private directory carrying owner-only
// files (T-07-08-05).
const (
	cacheSubdir  = "goswitch"
	toneDirPerm  = 0o700
	toneFilePerm = 0o600
)

// The bundled tones of the paplay fallback (the plan's flagged design
// decision): short deterministic sines with audibly distinct pitches,
// synthesized at first need and cached under the user's cache dir — no
// binary assets in git, no go:generate step, pure stdlib (T-07-08-SC: zero
// dependencies, the auditability constraint intact).
const (
	toneRate        = 22050.0 // samples per second — the canonical low-fidelity PCM rate
	toneSeconds     = 0.06    // a short confirmation click — tens of milliseconds
	toneAmplitude   = 0.5     // half full scale — audible, comfortably free of clipping
	maxSample       = 32767.0 // the 16-bit full scale the amplitude scales
	freqFlip        = 880.0   // the flip tone's pitch (A5)
	freqAutoCorrect = 1320.0  // the autocorrect tone's pitch (E6) — audibly distinct
	twoPi           = 2 * math.Pi
	wavFlip         = "flip.wav"
	wavAutoCorrect  = "autocorrect.wav"
)

// The PCM constants of the bundled tones: linear 16-bit mono at the fixed
// rate — the canonical plain-PCM RIFF form paplay plays as-is.
const (
	pcmFormat         = 1           // linear PCM
	pcmMono           = 1           // one channel
	pcmBits           = 16          // bits per sample
	pcmBytesPerSample = pcmBits / 8 // two bytes per 16-bit sample
	pcmSubFmtLen      = 16          // the fmt chunk's byte length for PCM
	riffHeaderLen     = 44          // the canonical RIFF header length
	riffPad           = 36          // the RIFF rule: ChunkSize = riffPad + data length
)

// playProc is one spawned player child — the corpus seam (the daemon
// launcher's procStarter form): Start brings the player up, Wait reaps it.
type playProc interface {
	Start() error
	Wait() error
}

// newPlayProc builds one player child: the pinned binary with NOTHING
// attached to Stdin/Stdout/Stderr — the no-pipes fork shape (the wl-copy
// precedent, clipboard.go:141-150): a piped descriptor would make a reaper
// wait on a fork-shaped grandchild's write-ends (T-07-08-03).
//
//nolint:ireturn // the seam hands the spawned-child interface back (the emitter() precedent)
func newPlayProc(name string, args []string) playProc {
	//nolint:noctx // deadline-free by design — Start plus a reaped Wait; nothing kills a playing tone (research Q7)
	return exec.Command(name, args...) // Stdin/Stdout/Stderr stay nil — no pipes
}

// Player plays the two tones of the acoustic feedback (plan 07-08): Flip
// for every script flip, AutoCorrect for every fired correction. The
// backend decision is made at the FIRST tone and cached — never a LookPath
// per flip (T-07-08-02); every failure is one WARN per episode and a quiet
// no-op, never an error to the gesture that triggered the tone.
type Player struct {
	flipEvent string // the flip tone's sound-theme event (the schema constant)
	acEvent   string // the autocorrect tone's event (the document's effective value; hot — WR-02)
	lookPath  func(string) (string, error)
	newProc   func(string, []string) playProc
	cacheDir  func() (string, error)

	mu      sync.Mutex      // serializes the decision cache and the warn episodes
	decided bool            // the backend decision exists (made at the first tone)
	backend string          // the decided binary; "" when neither player is installed
	warned  map[string]bool // one WARN per episode per reason — a healthy launch reopens the budget
}

// New returns the production player for the two sound-theme events: the
// flip event is the schema constant (config.DefaultSoundFlipEvent — NOT
// configurable, the owner asked for the switch and the distinct
// autocorrect event only), the autocorrect event the document's effective
// value.
func New(flipEvent, autocorrectEvent string) *Player {
	return &Player{
		flipEvent: flipEvent,
		acEvent:   autocorrectEvent,
		lookPath:  exec.LookPath,
		newProc:   newPlayProc,
		cacheDir:  os.UserCacheDir,
		warned:    make(map[string]bool),
	}
}

// Flip plays the flip tone: fire-and-forget — the synchronous cost is the
// argv choice and the child's Start; the playback itself lives in the
// child, reaped by its own goroutine. The hot path of the gesture never
// waits for a sound.
func (p *Player) Flip() {
	p.play(p.flipEvent, wavFlip, freqFlip)
}

// AutoCorrect plays the autocorrect tone — the same fire-and-forget path
// with the document's effective event and its own bundled pitch. The event
// reads under the mutex: the fold's SetAutocorrectEvent may land on any
// goroutine (WR-02).
func (p *Player) AutoCorrect() {
	p.mu.Lock()
	event := p.acEvent
	p.mu.Unlock()
	p.play(event, wavAutoCorrect, freqAutoCorrect)
}

// SetAutocorrectEvent re-pins the autocorrect tone's document event (WR-02):
// the actor's fold pushes every changed effective value, so the
// sound.autocorrect_event key is HOT — the D-32 contract holds without a
// restart. The mutex guard keeps the AutoCorrect read race-free.
func (p *Player) SetAutocorrectEvent(event string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.acEvent = event
}

// play launches one tone's player child. The canberra path pins the argv
// [canberra-gtk-play -i <event>] — the event is ONE argv element, no shell
// (T-07-08-01); the paplay fallback spawns over the bundled tone file the
// cache carries, the event name irrelevant to the fork. A failed launch is
// one WARN per episode and a quiet no-op; a successful launch hands the
// child to its reaper goroutine — a later non-zero exit (an event the
// desktop's sound theme lacks) stays silent, it is a theme gap, not a
// daemon failure.
func (p *Player) play(event, wavName string, freq float64) {
	backend := p.backendOf()
	if backend == "" {
		return // no player on this desktop — the decision's one WARN already fired
	}
	var argv []string
	switch backend {
	case binCanberra:
		argv = []string{binCanberra, "-i", event}
	case binPaplay:
		wav, ok := p.ensureTone(wavName, freq)
		if !ok {
			return // the cache write failed — the episode's one WARN already fired
		}
		argv = []string{binPaplay, wav}
	default:
		return // an undecided backend cannot reach here — the guard keeps the switch honest
	}
	cmd := p.newProc(argv[0], argv[1:])
	if err := cmd.Start(); err != nil {
		p.warn(reasonStartFailed, err)

		return
	}
	p.closeEpisode()               // a healthy launch proves the sound stack — a later failure warns again
	go func() { _ = cmd.Wait() }() // reap without a deadline — nothing kills a playing tone
}

// backendOf resolves the playback backend in force: decided at the FIRST
// tone and cached — a LookPath per flip would put a filesystem probe on
// every gesture (T-07-08-02). Neither player installed is the sound-off
// state: one WARN for the episode, quiet afterwards. The caller reads the
// answer without holding the mutex.
func (p *Player) backendOf() string {
	p.mu.Lock()
	if !p.decided {
		p.decided = true
		switch {
		case p.installed(binCanberra):
			p.backend = binCanberra
		case p.installed(binPaplay):
			p.backend = binPaplay
		}
	}
	backend := p.backend
	p.mu.Unlock()
	if backend == "" {
		p.warn(reasonPlayerMissing, nil)
	}

	return backend
}

// installed reports whether the binary name resolves on PATH. The caller
// holds the mutex.
func (p *Player) installed(bin string) bool {
	_, err := p.lookPath(bin)

	return err == nil
}

// ensureTone materializes one bundled tone in the user's cache dir and
// returns its path: a size-matching file from an earlier run is REUSED
// (the synthesis is deterministic — same bytes every time — and a rewrite
// per flip is pure churn), anything else is regenerated. A failed cache
// write is one WARN per episode and no tone.
func (p *Player) ensureTone(name string, freq float64) (string, bool) {
	data := toneWAV(freq)
	root, err := p.cacheDir()
	if err != nil {
		p.warn(reasonToneWrite, err)

		return "", false
	}
	dir := filepath.Join(root, cacheSubdir)
	path := filepath.Join(dir, name)
	if info, err := os.Stat(path); err == nil && info.Size() == int64(len(data)) {
		return path, true // the cached tone from an earlier run — reuse, never rewrite
	}
	if err := os.MkdirAll(dir, toneDirPerm); err != nil {
		p.warn(reasonToneWrite, err)

		return "", false
	}
	if err := os.WriteFile(path, data, toneFilePerm); err != nil {
		p.warn(reasonToneWrite, err)

		return "", false
	}

	return path, true
}

// toneWAV synthesizes one bundled tone: a sine at freq Hz, 16-bit PCM mono
// at the fixed rate — byte-deterministic (pure math: no clocks, no rand),
// canonical RIFF header via encoding/binary.
func toneWAV(freq float64) []byte {
	samples := int(toneSeconds * toneRate)
	dataLen := samples * pcmBytesPerSample
	wav := make([]byte, riffHeaderLen+dataLen)
	header := wavHeader{
		ChunkID:       [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize:     uint32(riffPad + dataLen),
		Format:        [4]byte{'W', 'A', 'V', 'E'},
		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: pcmSubFmtLen,
		AudioFormat:   pcmFormat,
		NumChannels:   pcmMono,
		SampleRate:    toneRate,
		ByteRate:      uint32(toneRate) * uint32(pcmMono) * uint32(pcmBytesPerSample),
		BlockAlign:    pcmMono * pcmBytesPerSample,
		BitsPerSample: pcmBits,
		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: uint32(dataLen),
	}
	var head bytes.Buffer
	// binary.Write into a memory buffer cannot fail — the guard keeps the
	// unchecked-error discipline without a panic path.
	if err := binary.Write(&head, binary.LittleEndian, header); err != nil {
		return nil
	}
	copy(wav, head.Bytes())
	for i := range samples {
		v := int16(math.Sin(twoPi*freq*float64(i)/toneRate) * toneAmplitude * maxSample)
		// #nosec G115 -- the int16 bit pattern IS the PCM sample encoding:
		// a negative sample's two's-complement bytes are the wire form.
		binary.LittleEndian.PutUint16(wav[riffHeaderLen+i*pcmBytesPerSample:], uint16(v))
	}

	return wav
}

// wavHeader is the canonical 44-byte little-endian PCM RIFF header —
// packed by encoding/binary, the form paplay's PCM reader plays as-is.
type wavHeader struct {
	ChunkID       [4]byte // "RIFF"
	ChunkSize     uint32  // riffPad + the data length
	Format        [4]byte // "WAVE"
	Subchunk1ID   [4]byte // "fmt "
	Subchunk1Size uint32  // the fmt chunk's byte length (PCM)
	AudioFormat   uint16  // linear PCM
	NumChannels   uint16  // mono
	SampleRate    uint32
	ByteRate      uint32  // SampleRate × BlockAlign
	BlockAlign    uint16  // channels × bytes per sample
	BitsPerSample uint16  // 16
	Subchunk2ID   [4]byte // "data"
	Subchunk2Size uint32  // the data length
}

// closeEpisode reopens the warn budget: a successful launch proves the
// sound stack healthy — a LATER failure warns again (the appidWarned
// reset form: a healthy answer closes the episode).
func (p *Player) closeEpisode() {
	p.mu.Lock()
	defer p.mu.Unlock()

	clear(p.warned)
}

// warn records one degradation WARN for the reason — ONE per episode (the
// actor's warnAppid form): the first failure of an episode warns, the rest
// stay quiet. The record names the reason, the transport error and the
// pinned binary names only — never user context (D-20/D-21).
func (p *Player) warn(reason string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.warned[reason] {
		return
	}
	p.warned[reason] = true
	if err == nil {
		slog.Warn("sound unavailable", "reason", reason)

		return
	}
	slog.Warn("sound unavailable", "reason", reason, "error", err)
}
