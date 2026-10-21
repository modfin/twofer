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

// ModularFinanceLogo returns the Modular Finance logo for Image and PNG.
func ModularFinanceLogo() image.Image {
	return mfLogo()
}

// logoShare is the logo's largest side as a share of the image. The highest
// error correction recovers about 30% of the code, and the logo covers far
// less.
const logoShare = 0.2

// Image draws text, the QR of a twofer frame, as a QR code size pixels wide
// with logo in its centre, keeping the logo's aspect ratio. A nil logo draws
// the code alone.
func Image(text string, size int, logo image.Image) (image.Image, error) {
	q, err := qrcode.New(text, qrcode.Highest)
	if err != nil {
		return nil, fmt.Errorf("err create qr code: %w", err)
	}

	code := q.Image(size)
	if logo == nil {
		return code, nil
	}
	b := code.Bounds()
	img := image.NewRGBA(b)
	draw.Draw(img, b, code, b.Min, draw.Src)

	side := int(float64(b.Dx()) * logoShare)
	lb := logo.Bounds()
	w, h := side, max(side*lb.Dy()/lb.Dx(), 1)
	if lb.Dy() > lb.Dx() {
		w, h = max(side*lb.Dx()/lb.Dy(), 1), side
	}
	at := image.Pt(b.Min.X+(b.Dx()-w)/2, b.Min.Y+(b.Dy()-h)/2)
	tile := image.Rectangle{Min: at, Max: at.Add(image.Pt(w, h))}
	// a white margin keeps the logo apart from the modules around it
	draw.Draw(img, tile.Inset(-max(side/10, 2)), image.White, image.Point{}, draw.Src)
	draw.Draw(img, tile, scale(logo, w, h), image.Point{}, draw.Over)
	return img, nil
}

// PNG is Image encoded as PNG.
func PNG(text string, size int, logo image.Image) ([]byte, error) {
	img, err := Image(text, size, logo)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("err encode qr png: %w", err)
	}
	return buf.Bytes(), nil
}

// scale resizes src to w×h, averaging the source pixels each target pixel
// covers so the logo's edges stay smooth.
func scale(src image.Image, w, h int) *image.RGBA {
	sb := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		y0 := sb.Min.Y + y*sb.Dy()/h
		y1 := max(sb.Min.Y+(y+1)*sb.Dy()/h, y0+1)
		for x := range w {
			x0 := sb.Min.X + x*sb.Dx()/w
			x1 := max(sb.Min.X+(x+1)*sb.Dx()/w, x0+1)

			var r, g, bl, a, n uint32
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					cr, cg, cb, ca := src.At(sx, sy).RGBA()
					r, g, bl, a, n = r+cr, g+cg, bl+cb, a+ca, n+1
				}
			}
			dst.Set(x, y, color.RGBA64{R: uint16(r / n), G: uint16(g / n), B: uint16(bl / n), A: uint16(a / n)})
		}
	}
	return dst
}
