package httpserve

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"twofer/grpc/gqr"
	"twofer/internal/servqr"
	"twofer/qr"

	"github.com/labstack/echo/v4"
)

func RegisterQRServer(e *echo.Echo, s *servqr.Server) {
	e.POST("/v1/qr", func(c echo.Context) error {
		b, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		var uriBody string
		err = json.Unmarshal(b, &uriBody)
		qrr, err := s.Generate(c.Request().Context(), &gqr.Data{
			RecoveryLevel: 0,
			Size:          0,
			Data:          uriBody,
		})
		if err != nil {
			return err
		}
		qrData := qr.QRData{
			Image:     qrr.Data,
			Reference: uriBody,
		}
		return c.JSON(http.StatusOK, qrData)
	})
}
