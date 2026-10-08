// The persistent sound-server connection (todo «Звук», owner «делай»
// 2026-10-05): one lazily dialed jfreymuth/pulse client carrying ONE
// int16-LE playback stream whose format follows the decoded theme PCM.
// The stream's reader serves SILENCE between tones through a select's
// default arm so the library's blocking Start always returns — the
// prototype's SIGQUIT lesson — and the play write is the prototype's
// 133 µs handoff. The dial lives OFF the gesture path (the worker calls
// it); a down server answers its own fast error, never a retry loop
// (T-SQU-03).

package sound

import (
	"fmt"
	"time"

	"github.com/jfreymuth/pulse"
	"github.com/jfreymuth/pulse/proto"
)

// audioStream is one live playback stream — the write/drain/close shape
// both the production adapter and the corpus's recording fake speak.
type audioStream interface {
	play(data []byte) error
	drain()
	close()
}

// audioConn is one live sound-server connection: a stream factory and the
// close the mute lifecycle needs (requirement 3: muted means NO
// connection at all).
type audioConn interface {
	openStream(format streamFormat) (audioStream, error)
	close()
}

// dialer dials the sound server — the instant-failure seam: a down server
// is the dial's own fast error, no deadline machinery on the play path.
type dialer func() (audioConn, error)

// silenceChunk is the reader's silence quantum — the buffer that feeds
// the stream's Start and every quiet drain. Even-sized: int16 pairs.
const silenceChunk = 1024

// toneReader is the persistent stream's sample source: the pending tone's
// bytes, silence between tones.
type toneReader struct {
	pending chan []byte // the worker's tone handoff (buffered 1)
	rest    []byte      // the in-flight tone's unread remainder
}

// newToneReader builds one stream source.
func newToneReader() *toneReader {
	return &toneReader{pending: make(chan []byte, 1)}
}

// Read serves audio to the stream: a pending tone's remaining bytes
// first, silence through the select's DEFAULT arm otherwise — Start must
// return without a pending tone (the corpus's blocking contract), and one
// Read call is exactly one wakeup: no internal loop, no spin — the
// server-paced drain sets the quiet callback rate.
func (r *toneReader) Read(p []byte) (int, error) {
	if len(r.rest) == 0 {
		select {
		case data := <-r.pending:
			r.rest = data
		default:
			clear(p)

			return len(p), nil
		}
	}
	n := copy(p, r.rest)
	r.rest = r.rest[n:]

	return n, nil
}

// feed hands one tone's PCM to the reader — the play write (the
// prototype's 133 µs). A still-pending previous tone leaves this one
// dropped: the enqueue discipline keeps one tone in flight, and real-time
// playback drains a 60 ms tone long before the next gesture.
func (r *toneReader) feed(data []byte) {
	select {
	case r.pending <- data:
	default:
	}
}

// appName is the client name the sound server's mixers show.
const appName = "goswitch"

// toneLatency is the tone stream's server-side buffer — the prototype's
// stable ~30 ms quantum.
const toneLatency = 30 * time.Millisecond

// latencyOption caps the stream's server-side buffer at toneLatency. The
// default negotiation let pipewire-pulse choose 8192 samples (~186 ms),
// and the persistent stream queues every tone BEHIND the buffered
// silence — the owner's by-ear gate (261006-squ HV-1) heard the old
// spawn delay again. Reads the rate and channel count the preceding
// options pinned — openStream wraps it in PlaybackRawOption after them
// (the library's option-order contract); int16-LE is the stream's only
// wire format (2 bytes/sample).
func latencyOption() func(*proto.CreatePlaybackStream) {
	return func(o *proto.CreatePlaybackStream) {
		// int16-LE is the stream's only wire format.
		const (
			bytesPerSample = 2
			maxToTarget    = 2 // the library's double-buffer shape (its own PlaybackLatency form)
		)
		// #nosec G115 -- a sub-second fraction times a wire rate fits uint32
		frames := uint32(toneLatency.Seconds() * float64(o.Rate))
		o.BufferTargetLength = frames * uint32(o.Channels) * bytesPerSample
		o.BufferMaxLength = maxToTarget * o.BufferTargetLength
		o.AdjustLatency = true
	}
}

// pulseConn is the production audioConn over jfreymuth/pulse: one client
// (the dial), one stream at a time (the worker recreates on format
// change), int16-LE wire format.
type pulseConn struct {
	client *pulse.Client
}

// dialPulse opens the persistent native-protocol connection — the
// prototype's 3 ms dial; a down server answers its own fast error.
//
//nolint:ireturn // the seam hands the connection interface back (the emitter() precedent)
func dialPulse() (audioConn, error) {
	client, err := pulse.NewClient(pulse.ClientApplicationName(appName))
	if err != nil {
		return nil, fmt.Errorf("dial sound server: %w", err)
	}

	return &pulseConn{client: client}, nil
}

// openStream creates and starts one int16-LE playback stream at the given
// format. Start blocks until the first buffer reaches the server — the
// reader's silence default answers it immediately (the corpus pins the
// contract; the prototype's SIGQUIT hangs without it).
//
//nolint:ireturn // the seam hands the connection-stream interface back (the emitter() precedent)
func (c *pulseConn) openStream(format streamFormat) (audioStream, error) {
	reader := newToneReader()
	stream, err := c.client.NewPlayback(
		pulse.NewReader(reader, proto.FormatInt16LE),
		pulse.PlaybackSampleRate(format.Rate),
		pulse.PlaybackChannels(channelMap(format.Channels)),
		pulse.PlaybackRawOption(latencyOption()), // after rate/channels — the option-order contract
	)
	if err != nil {
		return nil, fmt.Errorf("open playback stream: %w", err)
	}
	stream.Start()

	return &pulseStream{stream: stream, reader: reader}, nil
}

func (c *pulseConn) close() {
	c.client.Close()
}

// channelMap builds the wire channel map for a channel count — mono and
// stereo are the theme reality (Yaru is stereo); anything beyond is a
// best-effort L/R mirroring.
func channelMap(channels int) proto.ChannelMap {
	switch channels {
	case synthChannels:
		return proto.ChannelMap{proto.ChannelMono}
	case stereoChannels:
		return proto.ChannelMap{proto.ChannelFrontLeft, proto.ChannelFrontRight}
	}
	m := make(proto.ChannelMap, channels)
	for i := range m {
		if i%2 == 0 {
			m[i] = proto.ChannelFrontLeft

			continue
		}
		m[i] = proto.ChannelFrontRight
	}

	return m
}

// pulseStream is the production audioStream: the library stream plus the
// reader the worker feeds.
type pulseStream struct {
	stream *pulse.PlaybackStream
	reader *toneReader
}

// play hands one tone to the stream — the one write: the reader serves it
// as the server drains. A dead stream answers its error so the worker
// tears the connection down and the next tone starts fresh.
func (s *pulseStream) play(data []byte) error {
	if err := s.stream.Error(); err != nil {
		return fmt.Errorf("playback stream: %w", err)
	}
	s.reader.feed(data)

	return nil
}

func (s *pulseStream) drain() {
	s.stream.Drain()
}

func (s *pulseStream) close() {
	s.stream.Close()
}
