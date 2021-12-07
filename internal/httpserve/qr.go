package httpserve

import (
	"net/http"

	"twofer/grpc/gqr"
	"twofer/internal/servqr"
	"twofer/qr"

	"github.com/labstack/echo/v4"
)

func RegisterQRServer(e *echo.Echo, s *servqr.Server) {
	e.GET("/v1/eid/qr/:ref", func(c echo.Context) error {
		ref := c.Param("ref")
		bankidDeeplink := "bankid:///?autostarttoken=" + ref
		qrr, err := s.Generate(c.Request().Context(), &gqr.Data{
			RecoveryLevel: 0,
			Size:          0,
			Data:          bankidDeeplink,
		})
		if err != nil {
			return err
		}
		qrData := qr.QRData{
			Image:     qrr.Data,
			Reference: ref,
		}
		return c.JSON(http.StatusOK, qrData)
	})
}
