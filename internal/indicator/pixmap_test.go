package indicator_test

import (
	"bytes"
	"encoding/binary"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/goswitch/internal/indicator"
)

// updateGolden is the standard -update switch of the golden corpus (SPEC
// §7.1): a registered flag needs a package global, and the switch is the
// documented regeneration contract of the pins.
//
//nolint:gochecknoglobals // the canonical -update golden switch
var updateGolden = flag.Bool("update", false, "rewrite the golden files with the renderer's current output")

// The pinned tray corpus constants (quick plan 260930-pf6): named once so
// the pins state intent instead of bare magic numbers (mnd) — the canvas
// is a fixed square the renderer never resizes at runtime, the palette is
// two-tone (opaque white or fully transparent, ARGB32 byte order).
const (
	testCanvasWidth  = 24
	testCanvasHeight = 24
	testPixelBytes   = 4 // ARGB32: 4 bytes per pixel
	testOpaque       = 0xFF
	testTransparent  = 0x00
	testSymbolEN     = "en"
	testSymbolRU     = "ru"
)

// goldenBody serializes the full (iiay) pixmap body: the little-endian
// width, the little-endian height, then the row-major ARGB32 pixel bytes —
// the exact byte stream the golden files pin.
func goldenBody(pm indicator.Pixmap) []byte {
	body := make([]byte, 8+len(pm.Pixels))
	binary.LittleEndian.PutUint32(body[0:4], uint32(pm.Width))
	binary.LittleEndian.PutUint32(body[4:8], uint32(pm.Height))
	copy(body[8:], pm.Pixels)

	return body
}

// goldenPath is the golden file of one mode symbol.
func goldenPath(symbol string) string {
	return filepath.Join("testdata", symbol+".golden")
}

// TestPixmapForGolden pins the renderer's determinism in the SPEC §7.1
// golden-file style: repeated calls are byte-stable, and the full (iiay)
// body of each symbol matches testdata/<symbol>.golden byte for byte. Run
// `go test ./internal/indicator -run TestPixmapForGolden -update` to
// regenerate the goldens after a deliberate renderer change; the default
// run compares bytes exactly.
func TestPixmapForGolden(t *testing.T) {
	update := *updateGolden
	for _, symbol := range []string{testSymbolEN, testSymbolRU} {
		t.Run(symbol, func(t *testing.T) {
			pm, ok := indicator.PixmapFor(symbol)
			if !ok {
				t.Fatalf("PixmapFor(%q) = ok=false, want the composition", symbol)
			}
			again, ok := indicator.PixmapFor(symbol)
			if !ok {
				t.Fatalf("second PixmapFor(%q) = ok=false, want the composition", symbol)
			}
			if pm.Width != again.Width || pm.Height != again.Height || !bytes.Equal(pm.Pixels, again.Pixels) {
				t.Error("repeated PixmapFor calls differ — the renderer must be deterministic")
			}

			body := goldenBody(pm)
			if update {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatalf("create testdata: %v", err)
				}
				if err := os.WriteFile(goldenPath(symbol), body, 0o600); err != nil {
					t.Fatalf("write golden: %v", err)
				}

				return
			}
			want, err := os.ReadFile(goldenPath(symbol))
			if err != nil {
				t.Fatalf("read golden %s (regenerate with go test -update): %v", goldenPath(symbol), err)
			}
			if !bytes.Equal(body, want) {
				t.Errorf("pixmap body drifted from %s — a deterministic renderer keeps the goldens byte-stable",
					goldenPath(symbol))
			}
		})
	}
}

// TestPixmapForUnknownSymbols pins the closed-input contract: the only
// valid inputs are the actor's modeSymbol values; an unknown symbol
// reports ok=false with the zero Pixmap — the caller degrades, never
// panics.
func TestPixmapForUnknownSymbols(t *testing.T) {
	for _, symbol := range []string{"", "EN", "RU", "xy", "en ru"} {
		pm, ok := indicator.PixmapFor(symbol)
		if ok {
			t.Errorf("PixmapFor(%q) = ok=true, want false — the input set is closed", symbol)
		}
		if pm.Width != 0 || pm.Height != 0 || len(pm.Pixels) != 0 {
			t.Errorf("PixmapFor(%q) = %+v, want the zero Pixmap", symbol, pm)
		}
	}
}

// TestPixmapCanvasIsFixed pins the fixed canvas: both compositions carry
// the same dimensions and the full pixel buffer — the renderer scales and
// centers the glyphs onto one canvas and never resizes at runtime.
func TestPixmapCanvasIsFixed(t *testing.T) {
	for _, symbol := range []string{testSymbolEN, testSymbolRU} {
		pm, ok := indicator.PixmapFor(symbol)
		if !ok {
			t.Fatalf("PixmapFor(%q) = ok=false, want the composition", symbol)
		}
		if pm.Width != testCanvasWidth || pm.Height != testCanvasHeight {
			t.Errorf("%s pixmap dims = %d×%d, want the fixed %d×%d canvas",
				symbol, pm.Width, pm.Height, testCanvasWidth, testCanvasHeight)
		}
		want := testCanvasWidth * testCanvasHeight * testPixelBytes
		if len(pm.Pixels) != want {
			t.Errorf("%s pixel bytes = %d, want %d (the full canvas)", symbol, len(pm.Pixels), want)
		}
	}
}

// TestPixmapTwoTonePalette pins the two-tone contract: every pixel is
// either fully opaque white or fully transparent — an accidental
// anti-alias or garbage byte fails — and every composition draws at least
// one opaque glyph pixel (Ubuntu's top bar is dark: transparent glyphs
// would be invisible).
func TestPixmapTwoTonePalette(t *testing.T) {
	for _, symbol := range []string{testSymbolEN, testSymbolRU} {
		pm, ok := indicator.PixmapFor(symbol)
		if !ok {
			t.Fatalf("PixmapFor(%q) = ok=false, want the composition", symbol)
		}
		opaque := 0
		for px := 0; px < len(pm.Pixels); px += testPixelBytes {
			a, r, g, b := pm.Pixels[px], pm.Pixels[px+1], pm.Pixels[px+2], pm.Pixels[px+3]
			switch {
			case a == testOpaque && r == testOpaque && g == testOpaque && b == testOpaque:
				opaque++
			case a == testTransparent && r == testTransparent && g == testTransparent && b == testTransparent:
			default:
				t.Fatalf("%s pixel at byte %d is (%#x,%#x,%#x,%#x), want opaque white or fully transparent",
					symbol, px, a, r, g, b)
			}
		}
		if opaque == 0 {
			t.Errorf("%s composition drew no opaque pixel — the glyph would be invisible on the dark top bar", symbol)
		}
	}
}

// TestPixmapSymbolsDiffer pins that the two compositions are distinct —
// a renderer collapsing en and ru would pass every palette and canvas pin.
func TestPixmapSymbolsDiffer(t *testing.T) {
	en, okEN := indicator.PixmapFor(testSymbolEN)
	ru, okRU := indicator.PixmapFor(testSymbolRU)
	if !okEN || !okRU {
		t.Fatalf("PixmapFor corpus symbols ok = %t/%t, want true/true", okEN, okRU)
	}
	if bytes.Equal(en.Pixels, ru.Pixels) {
		t.Error("en and ru compositions are identical — the tray icons must differ")
	}
}
