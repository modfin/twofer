package httpserve

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"twofer/grpc/gpwd"
	"twofer/internal/servpwd"
)

func RegisterPWDServer(e *echo.Echo, s *servpwd.Server) {
	e.POST("/v1/pwd/enroll", func(c echo.Context) error {
		b, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		var enrollReq gpwd.EnrollReq
		err = json.Unmarshal(b, &enrollReq)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		enrollResp, err := s.Enroll(c.Request().Context(), &enrollReq)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, enrollResp)
	})

	e.POST("/v1/pwd/auth", func(c echo.Context) error {
		b, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		var authReq gpwd.AuthReq
		err = json.Unmarshal(b, &authReq)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		authResp, err := s.Auth(c.Request().Context(), &authReq)
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}
		return c.JSON(http.StatusOK, authResp)
	})
}