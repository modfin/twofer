package servqr

import (
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/net/context"
	"twofer/grpc/gqr"
)

func New() *Server {
	return &Server{}
}

type Server struct {
}

func (s Server) Generate(ctx context.Context, data *gqr.Data) (*gqr.Image, error) {
	size := int(data.Size)

	if size < 10 {
		size = 256
	}

	level := qrcode.RecoveryLevel(data.RecoveryLevel)
	image, err := qrcode.Encode(data.Data, level, size)

	return &gqr.Image{
		Data:        image,
		ContentType: "image/png",
	}, err
}
