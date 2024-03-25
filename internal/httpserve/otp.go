package httpserve

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/modfin/twofer/internal/servotp"
	"io"
	"net/http"
)

func RegisterOTPServer(e *echo.Echo, s *servotp.Server) {
	e.POST("/v1/otp/enroll", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		var en servotp.Enrollment
		err = json.Unmarshal(b, &en)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		enrollResp, err := s.Enroll(c.Request().Context(), &en)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, enrollResp)
	})

	e.POST("/v1/otp/auth", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		var va servotp.Credentials
		err = json.Unmarshal(b, &va)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		authResp, err := s.Auth(c.Request().Context(), &va)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, authResp)
	})

	e.POST("/v1/otp/qr", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		var va servotp.Credentials
		err = json.Unmarshal(b, &va)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		qrImage, err := s.GetQRImage(c.Request().Context(), &va)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, qrImage)
	})
}
