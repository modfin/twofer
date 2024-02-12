package httpserve

import (
	"encoding/json"
	"net/http"

	"github.com/modfin/twofer/grpc/geid"
	"github.com/modfin/twofer/internal/serveid"

	"github.com/labstack/echo/v4"
)

func RegisterEIDServer(e *echo.Echo, s *serveid.Server) {

	e.GET("/v1/eid/providers", func(c echo.Context) error {
		providers, err := s.GetProviders(c.Request().Context(), &geid.Empty{})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, providers)
	})

	e.POST("/v1/eid/auth", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		var req geid.Req
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}
		inter, err := s.AuthInit(c.Request().Context(), &req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, inter)
	})

	e.POST("/v1/eid/sign", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		var req geid.Req
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}
		inter, err := s.SignInit(c.Request().Context(), &req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, inter)
	})

	e.POST("/v1/eid/change", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		var req geid.Inter
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}
		resp, err := s.Change(c.Request().Context(), &req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, resp)
	})

	e.POST("/v1/eid/collect", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		var req geid.Inter
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}
		resp, err := s.Collect(c.Request().Context(), &req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, resp)
	})

	e.POST("/v1/eid/peek", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		var req geid.Inter
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}
		resp, err := s.Peek(c.Request().Context(), &req)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, resp)
	})

	e.POST("/v1/eid/cancel", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		var req geid.Inter
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}
		_, err = s.Cancel(c.Request().Context(), &req)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})
}
