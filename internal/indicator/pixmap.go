// Package indicator is the daemon's tray presence (quick plan
// 260930-pf6): a deterministic EN/RU glyph renderer plus the
// org.kde.StatusNotifierItem object that publishes it on the session bus.
// The renderer is pure — no D-Bus, no dependencies — and the item rides
// the control service's EXISTING connection: one bus name
// (org.djarvur.goswitch), one object path more (/StatusNotifierItem),
// zero new module dependencies, zero icon files on disk.
package indicator

// The fixed tray canvas: a square the renderer stamps the glyph pair onto
// and never resizes at runtime — the SNI IconPixmap consumer scales as it
// pleases.
const (
	canvasWidth  = 24
	canvasHeight = 24
)

// Glyph geometry: a 5×7 cell scaled 2× — a chunky 10×14 glyph that
// survives the top bar's scaling — two glyphs side by side with a 4-pixel
// separator: 10+4+10 fills the canvas width exactly, 14 rows center
// vertically (5 blank rows above and below).
const (
	glyphWidth  = 5
	glyphHeight = 7
	glyphScale  = 2
	separator   = 4
)

// ARGB32 pixel geometry and the two-tone palette: glyph pixels fully
// opaque white (Ubuntu's top bar is dark — a transparent glyph would be
// invisible), background fully transparent.
const (
	pixelBytes = 4
	byteAlpha  = 0
	byteRed    = 1
	byteGreen  = 2
	byteBlue   = 3
	glyphByte  = 0xFF
)

// The 1-bit glyph font: exactly four cells — E, N, R, U — one byte per row
// of 5 bits, MSB = leftmost pixel. String constants keep the font
// compile-time constant: no map iteration, no randomness, a fully
// deterministic renderer.
const (
	fontE = "\x1f\x10\x10\x1e\x10\x10\x1f" // E — left glyph of "en"
	fontN = "\x11\x19\x19\x15\x13\x13\x11" // N — right glyph of "en"
	fontR = "\x1e\x11\x11\x1e\x14\x12\x11" // R — left glyph of "ru"
	fontU = "\x11\x11\x11\x11\x11\x11\x0e" // U — right glyph of "ru"

	// fontMSB is the leftmost pixel's mask of a font row.
	fontMSB = 1 << (glyphWidth - 1)

	// canvasMarginY is the transparent row band above and below the glyph
	// pair: (24 − 14) / 2 = 5 rows of vertical centering.
	canvasMarginY = (canvasHeight - glyphScale*glyphHeight) / 2
)

// The modeSymbol vocabulary of the session actor — PixmapFor's closed
// input set.
const (
	symbolEN = "en"
	symbolRU = "ru"
)

// Pixmap is one tray image: the fixed canvas in ARGB32, row-major
// top-to-bottom, 4 bytes per pixel in A, R, G, B byte order. The struct
// marshals over D-Bus as the (iiay) member of the SNI IconPixmap property
// — the exported field order IS the wire shape.
type Pixmap struct {
	Width  int32
	Height int32
	Pixels []byte
}

// PixmapFor renders the fixed two-glyph composition for the actor's
// modeSymbol value: "en" draws E+N, "ru" draws R+U. An unknown symbol
// reports ok=false with the zero Pixmap — the caller degrades, never
// panics.
func PixmapFor(symbol string) (Pixmap, bool) {
	var left, right string
	switch symbol {
	case symbolEN:
		left, right = fontE, fontN
	case symbolRU:
		left, right = fontR, fontU
	default:
		return Pixmap{}, false
	}

	pixels := make([]byte, canvasWidth*canvasHeight*pixelBytes)
	stampGlyph(pixels, left, 0)
	stampGlyph(pixels, right, glyphScale*glyphWidth+separator)

	return Pixmap{Width: canvasWidth, Height: canvasHeight, Pixels: pixels}, true
}

// stampGlyph stamps one 2×-scaled 5×7 glyph onto the canvas at the given
// x offset (the left glyph at 0, the right at glyph+separator), vertically
// centered. Out-of-range stamps are impossible by construction — the
// geometry constants sum to the canvas — so the loop carries no clamping.
func stampGlyph(pixels []byte, font string, x0 int) {
	y0 := canvasMarginY
	for row := range glyphHeight {
		for col := range glyphWidth {
			if font[row]&(fontMSB>>col) == 0 {
				continue
			}
			for dy := range glyphScale {
				for dx := range glyphScale {
					px := ((y0 + row*glyphScale + dy) * canvasWidth * pixelBytes) +
						(x0+col*glyphScale+dx)*pixelBytes
					pixels[px+byteAlpha] = glyphByte
					pixels[px+byteRed] = glyphByte
					pixels[px+byteGreen] = glyphByte
					pixels[px+byteBlue] = glyphByte
				}
			}
		}
	}
}
