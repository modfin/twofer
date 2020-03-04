package qrserver

import (
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/net/context"
	"twofer/twoferrpc/gqr"
)

func New() *Server {
	return &Server{}
}

type Server struct {
}

func (s Server) Generate(ctx context.Context, data *gqr.QRData) (*gqr.QRImage, error) {
	size := int(data.Size)

	if size < 10 {
		size = 256
	}

	level := qrcode.RecoveryLevel(data.RecoveryLevel)
	image, err := qrcode.Encode(data.Data, level, size)

	return &gqr.QRImage{
		Data:        image,
		ContentType: "image/png",
	}, err
}
