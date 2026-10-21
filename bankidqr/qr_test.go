package bankidqr

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"

	qrcode "github.com/skip2/go-qrcode"
)

// what twofer sends in a frame: bankid.<qrStartToken>.<seconds>.<hmac>
const frameQR = "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.12.7b2410a2fdbae51a1f3c5c1e223752d3840ea4664444ceb81319f0707d219a3c"

// The logo may only replace modules in the middle: the finder, timing and
// alignment patterns a scanner locks onto have to stay as drawn.
func TestPNG_changes_only_the_centre(t *testing.T) {
	// a product's own logo, wider than tall
	wide := image.NewRGBA(image.Rect(0, 0, 120, 40))
	draw.Draw(wide, wide.Bounds(), image.NewUniform(color.RGBA{R: 200, A: 255}), image.Point{}, draw.Src)

	logos := map[string]image.Image{"modular finance": ModularFinanceLogo(), "own wide logo": wide}
	for name, logo := range logos {
		for _, size := range []int{256, 300, 400} {
			b, err := PNG(frameQR, size, logo)
			if err != nil {
				t.Fatalf("%s, size %d: %v", name, size, err)
			}
			got, err := png.Decode(bytes.NewReader(b))
			if err != nil {
				t.Fatalf("%s, size %d: decode png: %v", name, size, err)
			}

			changed := changedPixels(t, got, plain(t, size))
			if changed.Empty() {
				t.Fatalf("%s, size %d: no logo drawn", name, size)
			}
			bd := got.Bounds()
			centre := image.Rect(bd.Dx()*3/10, bd.Dy()*3/10, bd.Dx()*7/10, bd.Dy()*7/10)
			if !changed.In(centre) {
				t.Errorf("%s, size %d: logo reaches %v, outside the centre %v", name, size, changed, centre)
			}
		}
	}
}

func TestPNG_keeps_the_logo_aspect(t *testing.T) {
	red := color.RGBA{R: 200, A: 255}
	wide := image.NewRGBA(image.Rect(0, 0, 120, 40))
	draw.Draw(wide, wide.Bounds(), image.NewUniform(red), image.Point{}, draw.Src)

	img, err := Image(frameQR, 300, wide)
	if err != nil {
		t.Fatal(err)
	}
	// a QR has no red, so the red pixels are the logo
	logo := image.Rectangle{}
	bd := img.Bounds()
	for y := bd.Min.Y; y < bd.Max.Y; y++ {
		for x := bd.Min.X; x < bd.Max.X; x++ {
			if color.RGBAModel.Convert(img.At(x, y)) == red {
				logo = logo.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	if logo.Dx() != 3*logo.Dy() {
		t.Errorf("logo drawn %dx%d, want 3:1 like the source", logo.Dx(), logo.Dy())
	}
}

func TestImage_without_a_logo_is_the_plain_code(t *testing.T) {
	img, err := Image(frameQR, 300, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r := changedPixels(t, img, plain(t, 300)); !r.Empty() {
		t.Errorf("pixels %v differ from the plain code", r)
	}
}

func plain(t *testing.T, size int) image.Image {
	t.Helper()
	q, err := qrcode.New(frameQR, qrcode.Highest)
	if err != nil {
		t.Fatal(err)
	}
	return q.Image(size)
}

// changedPixels returns the smallest rectangle holding every pixel whose
// darkness differs between got and want.
func changedPixels(t *testing.T, got, want image.Image) image.Rectangle {
	t.Helper()
	if got.Bounds() != want.Bounds() {
		t.Fatalf("bounds %v, want %v", got.Bounds(), want.Bounds())
	}
	changed := image.Rectangle{}
	bd := want.Bounds()
	for y := bd.Min.Y; y < bd.Max.Y; y++ {
		for x := bd.Min.X; x < bd.Max.X; x++ {
			if dark(got.At(x, y)) != dark(want.At(x, y)) {
				changed = changed.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return changed
}

func dark(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r+g+b < 3*0x8000
}
