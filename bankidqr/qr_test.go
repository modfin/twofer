package bankidqr

import (
	"bytes"
	"image"
	"image/png"
	"testing"

	qrcode "github.com/skip2/go-qrcode"
)

// what twofer sends in a frame: bankid.<qrStartToken>.<seconds>.<hmac>
const frameQR = "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.12.7b2410a2fdbae51a1f3c5c1e223752d3840ea4664444ceb81319f0707d219a3c"

// The logo may only replace modules in the middle: the finder, timing and
// alignment patterns a scanner locks onto have to stay as drawn.
func TestPNG_changes_only_the_centre(t *testing.T) {
	for _, size := range []int{256, 300, 400} {
		b, err := PNG(frameQR, size)
		if err != nil {
			t.Fatalf("size %d: %v", size, err)
		}
		got, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatalf("size %d: decode png: %v", size, err)
		}
		q, err := qrcode.New(frameQR, qrcode.Highest)
		if err != nil {
			t.Fatal(err)
		}
		plain := q.Image(size)
		if got.Bounds() != plain.Bounds() {
			t.Fatalf("size %d: bounds %v, want %v", size, got.Bounds(), plain.Bounds())
		}

		bd := plain.Bounds()
		centre := image.Rect(bd.Dx()*3/10, bd.Dy()*3/10, bd.Dx()*7/10, bd.Dy()*7/10)
		changed := 0
		for y := bd.Min.Y; y < bd.Max.Y; y++ {
			for x := bd.Min.X; x < bd.Max.X; x++ {
				if dark(got.At(x, y)) == dark(plain.At(x, y)) {
					continue
				}
				if !image.Pt(x, y).In(centre) {
					t.Fatalf("size %d: pixel %d,%d outside the centre changed", size, x, y)
				}
				changed++
			}
		}
		if changed == 0 {
			t.Errorf("size %d: no logo drawn", size)
		}
	}
}

func dark(c interface{ RGBA() (r, g, b, a uint32) }) bool {
	r, g, b, _ := c.RGBA()
	return r+g+b < 3*0x8000
}
