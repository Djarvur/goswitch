//nolint:testpackage // drives the unexported cache/synth seams — the sound corpus in-package precedent
package sound

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"os"
	"strings"
	"sync"
	"testing"
)

// The corpus's stream formats (the synth always targets a concrete format;
// zero values are refused loudly).
const (
	rate22050 = 22050
	rate44100 = 44100
	rate48000 = 48000

	channelsMono   = 1
	channelsStereo = 2
)

//nolint:gochecknoglobals // shared fixture formats (the detect_test corpus precedent)
var (
	formatMono      = streamFormat{Rate: rate22050, Channels: channelsMono}
	formatStereo    = streamFormat{Rate: rate44100, Channels: channelsStereo}
	format48kStereo = streamFormat{Rate: rate48000, Channels: channelsStereo}
)

// errDecodeRefused is the decode double's failure — the contained-error
// state of the cache cells (a failed pair answers miss, the caller
// synthesizes).
var errDecodeRefused = errors.New("decode refused by the corpus")

// TestPCMCache_DecodesOnceAndByteIdentical pins the cache contract: a
// (theme, event) pair decodes exactly ONCE at load, and repeat lookups
// answer the same entry byte-identically — lookups never re-decode.
func TestPCMCache_DecodesOnceAndByteIdentical(t *testing.T) {
	var mu sync.Mutex
	decodes := make(map[string]int)
	decode := func(theme, event string) (*pcm, error) {
		mu.Lock()
		decodes[theme+"/"+event]++
		mu.Unlock()

		return &pcm{format: formatMono, data: []byte("pcm-" + event)}, nil
	}
	c := newPCMCache(decode)
	// The duplicated list element must not buy a second decode.
	c.load("yaru", []string{fsEventBell, fsEventBell, fsEventMessage})

	first, ok := c.lookup("yaru", fsEventBell)
	if !ok {
		t.Fatal("bell answered miss after a successful load")
	}
	second, ok := c.lookup("yaru", fsEventBell)
	if !ok {
		t.Fatal("bell answered miss on the repeat lookup")
	}
	if !bytes.Equal(first.data, second.data) {
		t.Error("repeat lookups disagree — a decoded entry must come back byte-identically")
	}

	mu.Lock()
	defer mu.Unlock()
	if got := decodes["yaru/"+fsEventBell]; got != 1 {
		t.Errorf("bell decodes = %d, want exactly 1 — the pair decodes once", got)
	}
}

// TestPCMCache_MissingPairAnswersMiss pins the miss contract: a pair whose
// decode failed — and a pair never loaded — answer miss, so the caller
// plays the synthesized tone instead (never a wait, never an error).
func TestPCMCache_MissingPairAnswersMiss(t *testing.T) {
	c := newPCMCache(func(string, string) (*pcm, error) { return nil, errDecodeRefused })
	c.load("yaru", []string{fsEventBell})

	if _, ok := c.lookup("yaru", fsEventBell); ok {
		t.Error("a failed decode answered a hit — the pair must stay absent")
	}
	if _, ok := c.lookup("yaru", fsEventMessage); ok {
		t.Error("a never-loaded pair answered a hit")
	}
}

// TestPCMCache_AtomicSwapMidRebuild pins the atomic-swap contract: a
// theme-change rebuild publishes a new snapshot under the new theme key
// ATOMICALLY — mid-rebuild, the old theme keeps answering from the old
// snapshot, the new theme misses, and neither lookup ever waits on the
// load in flight (the never-waits contract).
func TestPCMCache_AtomicSwapMidRebuild(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	decode := func(theme, event string) (*pcm, error) {
		if theme == "new" {
			close(entered)
			<-release // the build wedges here — lookups must not care
		}

		return &pcm{format: formatMono, data: []byte("pcm-" + theme)}, nil
	}
	c := newPCMCache(decode)
	c.load("old", []string{fsEventBell})

	done := make(chan struct{})
	go func() {
		c.load("new", []string{fsEventBell})
		close(done)
	}()
	<-entered // the new-theme build is now blocked inside its decoder

	got, ok := c.lookup("old", fsEventBell)
	if !ok || !bytes.Equal(got.data, []byte("pcm-old")) {
		t.Errorf("old-theme lookup mid-rebuild = (%q, ok=%t), want the old snapshot's entry", got, ok)
	}
	if _, ok := c.lookup("new", fsEventBell); ok {
		t.Error("new-theme lookup mid-rebuild answered a hit — the swap must be atomic")
	}

	close(release)
	<-done
	got, ok = c.lookup("new", fsEventBell)
	if !ok || !bytes.Equal(got.data, []byte("pcm-new")) {
		t.Errorf("new-theme lookup after the swap = (%q, ok=%t), want the new snapshot's entry", got, ok)
	}
	if _, ok := c.lookup("old", fsEventBell); ok {
		t.Error("old-theme lookup after the swap answered a hit — the swap must replace the snapshot whole")
	}
}

// TestSynthTone_Deterministic pins the synthesized tones' determinism (the
// old WAV-corpus pin carried over): the synthesis is byte-stable across
// calls — pure math, no clocks, no rand — and the two pitches stay
// audibly distinct.
func TestSynthTone_Deterministic(t *testing.T) {
	first, err := synthTone(freqFlip, formatMono)
	if err != nil {
		t.Fatalf("synth flip tone: %v", err)
	}
	second, err := synthTone(freqFlip, formatMono)
	if err != nil {
		t.Fatalf("synth flip tone again: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Error("tone synthesis is not byte-stable across calls — no clocks, no rand")
	}

	autocorrect, err := synthTone(freqAutoCorrect, formatMono)
	if err != nil {
		t.Fatalf("synth autocorrect tone: %v", err)
	}
	if bytes.Equal(first, autocorrect) {
		t.Error("the two pitches are byte-identical — the tones must stay audibly distinct")
	}
}

// TestSynthTone_Shape pins the synth's shape: exactly synthSeconds worth
// of int16-LE frames at the GIVEN rate and channels — no resampling, the
// stream's own format — with the sine's leading zero crossing intact.
func TestSynthTone_Shape(t *testing.T) {
	for _, format := range []streamFormat{formatMono, formatStereo, format48kStereo} {
		data, err := synthTone(freqFlip, format)
		if err != nil {
			t.Fatalf("synth at %d Hz/%d ch: %v", format.Rate, format.Channels, err)
		}
		frames := int(toneSeconds * float64(format.Rate))
		want := 2 * frames * format.Channels
		if got := len(data); got != want {
			t.Errorf("synth bytes at %d Hz/%d ch = %d, want %d (int16 LE)",
				format.Rate, format.Channels, got, want)
		}
		if binary.LittleEndian.Uint16(data) != 0 {
			t.Errorf("first sample = %d, want 0 — the sine starts at its zero crossing",
				binary.LittleEndian.Uint16(data))
		}
	}
}

// TestSynthTone_HalfAmplitude pins the amplitude discipline: every sample
// stays within the half-scale ceiling — audible, comfortably free of
// clipping.
func TestSynthTone_HalfAmplitude(t *testing.T) {
	data, err := synthTone(freqFlip, formatStereo)
	if err != nil {
		t.Fatalf("synth flip tone: %v", err)
	}
	ceiling := toneAmplitude*maxSample + 1
	for i := 0; i+1 < len(data); i += 2 {
		v := int16(binary.LittleEndian.Uint16(data[i:]))
		if math.Abs(float64(v)) > ceiling {
			t.Fatalf("sample %d = %d exceeds the half-amplitude ceiling %v", i/2, v, ceiling)
		}
	}
}

// TestSynthTone_ZeroFormatRefused pins the loud refusal: a nil/zero format
// argument is an error — the synth always targets a concrete stream
// format, so a zero value is a caller bug, never a silently-empty tone.
func TestSynthTone_ZeroFormatRefused(t *testing.T) {
	if _, err := synthTone(freqFlip, streamFormat{}); !errors.Is(err, errSynthFormat) {
		t.Errorf("zero format err = %v, want errSynthFormat", err)
	}
	if _, err := synthTone(freqFlip, streamFormat{Rate: rate44100}); !errors.Is(err, errSynthFormat) {
		t.Errorf("zero channels err = %v, want errSynthFormat", err)
	}
	if _, err := synthTone(freqFlip, streamFormat{Channels: channelsStereo}); !errors.Is(err, errSynthFormat) {
		t.Errorf("zero rate err = %v, want errSynthFormat", err)
	}
}

// yaruBellPath is the system theme file the real-decoder cell reads —
// present on the owner's GNOME desktop, absent in headless CI (the 06-02
// SKIP precedent keeps CI independent of desktop state).
const yaruBellPath = "/usr/share/sounds/Yaru/stereo/bell.oga"

// TestDecodeOggVorbis_RealThemeFile pins the decoder adapter on REAL input
// where the desktop provides it: rate and channels are extracted from the
// stream and the float32 samples come out as int16 little-endian frames.
func TestDecodeOggVorbis_RealThemeFile(t *testing.T) {
	file, err := os.Open(yaruBellPath)
	if err != nil {
		t.Skipf("no system theme file here (%s) — CI runs the synth corpus only", yaruBellPath)
	}
	defer func() { _ = file.Close() }()

	decoded, err := decodeOggVorbis(file)
	if err != nil {
		t.Fatalf("decode %s: %v", yaruBellPath, err)
	}
	if decoded.format.Channels != channelsStereo {
		t.Errorf("channels = %d, want %d — Yaru theme PCM is stereo", decoded.format.Channels, channelsStereo)
	}
	if decoded.format.Rate != rate44100 {
		t.Errorf("rate = %d, want %d — Yaru theme PCM is 44.1 kHz", decoded.format.Rate, rate44100)
	}
	if len(decoded.data) == 0 || len(decoded.data)%(2*channelsStereo) != 0 {
		t.Errorf("data length = %d, want a non-empty multiple of stereo int16 frames", len(decoded.data))
	}
}

// TestDecodeOggVorbis_MalformedStaysContained pins T-SQU-01's containment:
// a malformed/hostile theme file decodes to an ERROR VALUE, never a panic.
func TestDecodeOggVorbis_MalformedStaysContained(t *testing.T) {
	if _, err := decodeOggVorbis(strings.NewReader("certainly not an ogg stream")); err == nil {
		t.Fatal("garbage decoded without an error — decode failures must be contained values, never panics")
	}
}
