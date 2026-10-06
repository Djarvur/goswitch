// The PCM layer of the sound redesign (todo «Звук», owner «делай»
// 2026-10-05): theme events decode ONCE per (theme, event) pair into
// int16-LE samples through jfreymuth/oggvorbis, live in an atomically
// swapped snapshot cache keyed by theme, and every miss — daemon start,
// theme switch, unknown event — answers the deterministic synthesized
// sine at the stream's own rate and channels. The play path NEVER waits
// for a decode: lookups are lock-free against the atomic snapshot and a
// missing pair is a miss, not a load.
package sound

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"sync/atomic"

	"github.com/jfreymuth/oggvorbis"
)

// errSynthFormat is the synth's loud refusal: a zero/nil stream format is
// a caller bug — the synthesized tone always targets a concrete stream
// format, never a silently-empty run.
var errSynthFormat = errors.New("synthesized tone needs a concrete stream format")

// errDecodeFormat is the decoder's defensive refusal: a stream without a
// positive rate and channel count cannot follow the PCM rule.
var errDecodeFormat = errors.New("decoded theme stream carries no usable format")

// putInt16LE writes one int16 sample's little-endian wire form — the
// shared conversion of the decoder and the synth.
func putInt16LE(dst []byte, v int16) {
	// #nosec G115 -- the int16 bit pattern IS the PCM sample encoding:
	// a negative sample's two's-complement bytes are the wire form.
	binary.LittleEndian.PutUint16(dst, uint16(v))
}

// streamFormat is the concrete wire shape of one PCM run — int16
// little-endian frames at the given rate and channel count (the stream's
// format follows the decoded theme PCM; a change recreates the stream).
type streamFormat struct {
	Rate     int
	Channels int
}

// pcm is one decoded theme event: its stream format and the interleaved
// int16-LE sample bytes.
type pcm struct {
	format streamFormat
	data   []byte
}

// decodeOggVorbis decodes one opened theme file: the decoder's
// interleaved float32 samples come out clamped to half-less-than-full
// scale int16 little-endian frames carrying the stream's rate and
// channels. A malformed or hostile file is a contained error value,
// never a panic (T-SQU-01).
func decodeOggVorbis(r io.Reader) (*pcm, error) {
	samples, format, err := oggvorbis.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("decode theme ogg vorbis: %w", err)
	}
	if format.Channels <= 0 || format.SampleRate <= 0 {
		return nil, fmt.Errorf("decode theme ogg vorbis: %w", errDecodeFormat)
	}
	out := make([]byte, 2*len(samples))
	for i, s := range samples {
		v := int16(min(max(float64(s), -1), 1) * maxSample)
		putInt16LE(out[2*i:], v)
	}

	return &pcm{
		format: streamFormat{Rate: format.SampleRate, Channels: format.Channels},
		data:   out,
	}, nil
}

// synthTone synthesizes one fallback tone: a pure sine at freq Hz over
// synthSeconds at the GIVEN stream rate and channels (no resampling —
// the stream's own format), half amplitude, int16 little-endian. The
// synthesis is byte-deterministic (pure math: no clocks, no rand — the
// old WAV-corpus pin carried over). A zero/nil format is refused loudly:
// the synth always targets a concrete stream format.
func synthTone(freq float64, format streamFormat) ([]byte, error) {
	if format.Rate <= 0 || format.Channels <= 0 {
		return nil, fmt.Errorf("synth tone %.0f Hz: %w", freq, errSynthFormat)
	}
	frames := int(toneSeconds * float64(format.Rate))
	out := make([]byte, 2*frames*format.Channels)
	for i := range frames {
		v := int16(math.Sin(twoPi*freq*float64(i)/float64(format.Rate)) * toneAmplitude * maxSample)
		for ch := range format.Channels {
			putInt16LE(out[2*(i*format.Channels+ch):], v)
		}
	}

	return out, nil
}

// pcmSnapshot is one immutable decode of a whole theme: the snapshot key
// and the per-event PCM entries (a failed pair is simply absent — the
// caller synthesizes).
type pcmSnapshot struct {
	theme   string
	entries map[string]*pcm
}

// pcmCache is the (theme, event) PCM cache: lookups read the current
// atomic snapshot and NEVER block on a load in flight (the never-waits
// contract), a theme-change rebuild publishes its snapshot whole through
// one atomic swap.
type pcmCache struct {
	snap   atomic.Pointer[pcmSnapshot]
	decode func(theme, event string) (*pcm, error) // the corpus seam; production composes the resolver with decodeOggVorbis
}

// newPCMCache builds one cache over the given decode seam — the corpus
// counts decodes through it, production resolves and decodes.
func newPCMCache(decode func(theme, event string) (*pcm, error)) *pcmCache {
	return &pcmCache{decode: decode}
}

// lookup answers the (theme, event) pair from the current snapshot: a
// foreign theme key or an absent pair is a miss — the caller plays the
// synthesized tone. Lock-free against any load in flight.
func (c *pcmCache) lookup(theme, event string) (*pcm, bool) {
	snap := c.snap.Load()
	if snap == nil || snap.theme != theme {
		return nil, false
	}
	p, ok := snap.entries[event]

	return p, ok
}

// load builds a fresh snapshot for one theme — decoding each pair ONCE
// (a duplicated event list decodes once), failed pairs simply absent —
// and swaps it in atomically. The caller serializes loads (the single
// playback worker) and runs them strictly OFF the play path.
func (c *pcmCache) load(theme string, events []string) {
	entries := make(map[string]*pcm, len(events))
	for _, event := range events {
		if _, done := entries[event]; done {
			continue
		}
		p, err := c.decode(theme, event)
		if err != nil {
			continue // a failed pair answers miss — the synthesized tone covers it
		}
		entries[event] = p
	}
	c.snap.Store(&pcmSnapshot{theme: theme, entries: entries})
}
