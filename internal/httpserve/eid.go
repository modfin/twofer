package httpserve

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/modfin/twofer/internal/serveid"
	"io"
	"net/http"
)

func RegisterEIDServer(e *echo.Echo, s *serveid.Server) {
	e.GET("/v1/eid/providers", func(c echo.Context) error {
		providers, err := s.GetProviders()
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

		var req serveid.Req
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}

		request, err := serveid.FromReq(&req)
		if err != nil {
			return err
		}

		inter, err := s.AuthInit(c.Request().Context(), &request)
		if err != nil {
			return err
		}

		resp, err := serveid.ToInter(inter)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resp)
	})

	e.POST("/v1/eid/sign", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		var req serveid.Req
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}

		request, err := serveid.FromReq(&req)
		if err != nil {
			return err
		}

		inter, err := s.SignInit(c.Request().Context(), &request)
		if err != nil {
			return err
		}

		resp, err := serveid.ToInter(inter)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, resp)
	})

	e.POST("/v1/eid/change", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		var req serveid.Inter
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}

		inter, err := serveid.FromInter(&req)
		if err != nil {
			return err
		}

		resp, err := s.Change(c.Request().Context(), &inter)
		fmt.Println("CHANGE ", resp)
		if err != nil {
			return err
		}

		res, err := serveid.ToResp(resp)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
	})

	e.POST("/v1/eid/collect", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		var inter serveid.Inter
		err = json.Unmarshal(b, &inter)
		if err != nil {
			return err
		}

		req, err := serveid.FromInter(&inter)
		if err != nil {
			return err
		}

		resp, err := s.Collect(c.Request().Context(), &req)
		if err != nil {
			return err
		}

		res, err := serveid.ToResp(resp)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
	})

	e.POST("/v1/eid/peek", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		var req serveid.Inter
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}

		inter, err := serveid.FromInter(&req)
		if err != nil {
			return err
		}

		resp, err := s.Peek(c.Request().Context(), &inter)
		if err != nil {
			return err
		}

		res, err := serveid.ToResp(resp)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
	})

	e.POST("/v1/eid/cancel", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		var req serveid.Inter
		err = json.Unmarshal(b, &req)
		if err != nil {
			return err
		}

		inter, err := serveid.FromInter(&req)
		if err != nil {
			return err
		}

		_, err = s.Cancel(c.Request().Context(), &inter)
		if err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	})
}
