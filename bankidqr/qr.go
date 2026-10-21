// Package bankidqr draws a BankID QR code with a logo in its centre. Callers
// that draw their own code keep using the QR text of each frame.
package bankidqr

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"

	qrcode "github.com/skip2/go-qrcode"
)

//go:embed logo.png
var logoPNG []byte

var mfLogo = sync.OnceValue(func() image.Image {
	l, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		panic(fmt.Sprintf("bankidqr: embedded logo: %v", err))
	}
	return l
})

// ModularFinanceLogo returns the Modular Finance logo for Image and PNG. Every
// caller gets the same image, so do not draw on it.
func ModularFinanceLogo() image.Image {
	return mfLogo()
}

const (
	// logoShare is the logo's largest side as a share of the image. The highest
	// error correction recovers about 30% of the code, and the logo covers far
	// less.
	logoShare = 0.2
	// MaxLogoSide bounds the logo, which is scaled down for every code.
	MaxLogoSide = 1024
)

// encoder favours speed: callers draw a code for every frame, once a second
var encoder = png.Encoder{CompressionLevel: png.BestSpeed, BufferPool: &bufferPool{}}

type bufferPool struct{ p sync.Pool }

func (b *bufferPool) Get() *png.EncoderBuffer {
	e, _ := b.p.Get().(*png.EncoderBuffer)
	return e
}

func (b *bufferPool) Put(e *png.EncoderBuffer) {
	b.p.Put(e)
}

// Image draws text, the QR of a twofer frame, as a QR code with logo in its
// centre, keeping the logo's aspect ratio. A nil or empty logo draws the code
// alone. The code is size pixels wide, or wider when size is too small for it.
func Image(text string, size int, logo image.Image) (image.Image, error) {
	q, err := qrcode.New(text, qrcode.Highest)
	if err != nil {
		return nil, fmt.Errorf("err create qr code: %w", err)
	}

	code := q.Image(size)
	if logo == nil || logo.Bounds().Empty() {
		return code, nil
	}
	lb := logo.Bounds()
	if lb.Dx() > MaxLogoSide || lb.Dy() > MaxLogoSide {
		return nil, fmt.Errorf("logo is %dx%d, larger than %d px on a side", lb.Dx(), lb.Dy(), MaxLogoSide)
	}

	b := code.Bounds()
	side := int(float64(b.Dx()) * logoShare)
	w, h := side, max(side*lb.Dy()/lb.Dx(), 1)
	if lb.Dy() > lb.Dx() {
		w, h = max(side*lb.Dx()/lb.Dy(), 1), side
	}
	at := image.Pt(b.Min.X+(b.Dx()-w)/2, b.Min.Y+(b.Dy()-h)/2)
	tile := image.Rectangle{Min: at, Max: at.Add(image.Pt(w, h))}

	// the logo on a white margin, which keeps it apart from the modules around it
	area := image.NewRGBA(tile.Inset(-max(side/10, 2)).Intersect(b))
	draw.Draw(area, area.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(area, tile, scale(logo, w, h), image.Point{}, draw.Over)

	if p, ok := code.(*image.Paletted); ok {
		if img, ok := onPalette(p, area); ok {
			return img, nil
		}
	}
	img := image.NewRGBA(b)
	draw.Draw(img, b, code, b.Min, draw.Src)
	draw.Draw(img, area.Bounds(), area, area.Bounds().Min, draw.Src)
	return img, nil
}

// onPalette pastes area onto a copy of code with area's colours added to the
// palette, which keeps the PNG small. It fails when the colours do not fit.
func onPalette(code *image.Paletted, area *image.RGBA) (*image.Paletted, bool) {
	pal := append(color.Palette{}, code.Palette...)
	index := map[color.RGBA]uint8{}
	ab := area.Bounds()
	for y := ab.Min.Y; y < ab.Max.Y; y++ {
		for x := ab.Min.X; x < ab.Max.X; x++ {
			c := area.RGBAAt(x, y)
			if _, ok := index[c]; ok {
				continue
			}
			if len(pal) == 256 {
				return nil, false
			}
			index[c] = uint8(len(pal))
			pal = append(pal, c)
		}
	}

	img := image.NewPaletted(code.Bounds(), pal)
	if img.Stride == code.Stride {
		copy(img.Pix, code.Pix)
	} else {
		draw.Draw(img, img.Bounds(), code, code.Bounds().Min, draw.Src)
	}
	for y := ab.Min.Y; y < ab.Max.Y; y++ {
		for x := ab.Min.X; x < ab.Max.X; x++ {
			img.SetColorIndex(x, y, index[area.RGBAAt(x, y)])
		}
	}
	return img, true
}

// PNG is Image encoded as PNG.
func PNG(text string, size int, logo image.Image) ([]byte, error) {
	img, err := Image(text, size, logo)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = encoder.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("err encode qr png: %w", err)
	}
	return buf.Bytes(), nil
}

// scale resizes src to w×h, averaging the source pixels each target pixel
// covers so the logo's edges stay smooth.
func scale(src image.Image, w, h int) *image.RGBA {
	sb := src.Bounds()
	at := func(x, y int) color.RGBA64 {
		r, g, b, a := src.At(x, y).RGBA()
		return color.RGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: uint16(a)}
	}
	// the concrete image types read pixels without allocating
	if fast, ok := src.(image.RGBA64Image); ok {
		at = fast.RGBA64At
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		y0 := sb.Min.Y + y*sb.Dy()/h
		y1 := max(sb.Min.Y+(y+1)*sb.Dy()/h, y0+1)
		for x := range w {
			x0 := sb.Min.X + x*sb.Dx()/w
			x1 := max(sb.Min.X+(x+1)*sb.Dx()/w, x0+1)

			var r, g, bl, a, n uint64
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					c := at(sx, sy)
					r, g, bl, a, n = r+uint64(c.R), g+uint64(c.G), bl+uint64(c.B), a+uint64(c.A), n+1
				}
			}
			dst.SetRGBA64(x, y, color.RGBA64{R: uint16(r / n), G: uint16(g / n), B: uint16(bl / n), A: uint16(a / n)})
		}
	}
	return dst
}
