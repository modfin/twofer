// Package bankidqr draws a BankID QR code with the Modular Finance logo in its
// centre. Callers that draw their own code keep using the QR text of each frame.
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

var logo = sync.OnceValues(func() (image.Image, error) {
	return png.Decode(bytes.NewReader(logoPNG))
})

// logoShare is the logo's width as a share of the image. The highest error
// correction recovers about 30% of the code, and the logo covers far less.
const logoShare = 0.2

// PNG draws text, the QR of a twofer frame, as a QR code size pixels wide with
// the logo in its centre.
func PNG(text string, size int) ([]byte, error) {
	q, err := qrcode.New(text, qrcode.Highest)
	if err != nil {
		return nil, fmt.Errorf("err create qr code: %w", err)
	}
	l, err := logo()
	if err != nil {
		return nil, fmt.Errorf("err decode logo: %w", err)
	}

	code := q.Image(size)
	b := code.Bounds()
	img := image.NewRGBA(b)
	draw.Draw(img, b, code, b.Min, draw.Src)

	w := int(float64(b.Dx()) * logoShare)
	at := image.Pt(b.Min.X+(b.Dx()-w)/2, b.Min.Y+(b.Dy()-w)/2)
	tile := image.Rectangle{Min: at, Max: at.Add(image.Pt(w, w))}
	// a white margin keeps the logo apart from the modules around it
	draw.Draw(img, tile.Inset(-max(w/10, 2)), image.White, image.Point{}, draw.Src)
	draw.Draw(img, tile, scale(l, w), image.Point{}, draw.Over)

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return nil, fmt.Errorf("err encode qr png: %w", err)
	}
	return buf.Bytes(), nil
}

// scale resizes src to a w×w square, averaging the source pixels each target
// pixel covers so the logo's edges stay smooth.
func scale(src image.Image, w int) *image.RGBA {
	sb := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, w, w))
	for y := range w {
		y0 := sb.Min.Y + y*sb.Dy()/w
		y1 := max(sb.Min.Y+(y+1)*sb.Dy()/w, y0+1)
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
