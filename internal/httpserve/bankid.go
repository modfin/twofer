package httpserve

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/modfin/twofer/internal/bankid"
)

func RegisterBankIDServer(e *echo.Echo, client *bankid.API) {
	e.POST("/bankid/v6/auth", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.AuthSignRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal auth request message: %s\n", err.Error())
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		res, err := client.Auth(c.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating auth request against bankid: %s\n", err.Error())
			return c.JSON(500, "failed to initiate auth against BankId")
		}

		// In the case a client wants to initiate a new request every second instead of relying on SSE
		// we respond with the first entry and then close the connection
		response := c.QueryParam("type")
		if response == "once" {
			return c.JSON(200, res.APIResponse(0))
		}

		sender, interrupt, err := client.WatchForChange(c, res.OrderRef)
		if err != nil {
			return c.JSON(500, "failed to setup response stream")
		}

		for i := 0; i < 30; i++ {
			err = sender.Send("message", res.APIResponse(i)) // TODO: Should we change "message" to something else?
			if err != nil {
				fmt.Printf("ERR: failed to send auth response message: %s\n", err.Error())
				return c.JSON(500, "failed to send response message")
			}

			// Optimally subtract time that has elapsed, but no need to be that exact
			select {
			case <-interrupt:
				return c.JSON(http.StatusOK, bankid.Empty{})
			case <-time.After(time.Second):
			}
		}

		return c.JSON(http.StatusOK, bankid.Empty{})
	})

	e.POST("/bankid/v6/sign", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.AuthSignRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal sign request message: %s\n", err.Error())
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		res, err := client.Sign(c.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating sign request against bankid: %s\n", err.Error())
			return c.JSON(500, "failed to initiate sign against BankId")
		}

		// In the case a client wants to initiate a new request every second instead of relying on SSE
		// we respond with the first entry and then close the connection
		response := c.QueryParam("type")
		if response == "once" {
			return c.JSON(200, res.APIResponse(0))
		}

		sender, interrupt, err := client.WatchForChange(c, res.OrderRef)
		if err != nil {
			return c.JSON(500, "failed to setup response stream")
		}

		for i := 0; i < 30; i++ {
			err = sender.Send("message", res.APIResponse(i)) // TODO: Should we change "message" to something else?
			if err != nil {
				fmt.Printf("ERR: failed to send sign response message: %s\n", err.Error())
				return c.JSON(500, "failed to send response message")
			}

			// Optimally subtract time that has elapsed, but no need to be that exact
			select {
			case <-interrupt:
				return c.JSON(http.StatusOK, bankid.Empty{})
			case <-time.After(time.Second):
			}
		}
		return c.JSON(http.StatusOK, bankid.Empty{})
	})

	e.POST("/bankid/v6/change", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			fmt.Printf("ERR: failed to read change request message: %s\n", err.Error())
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.ChangeRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal change request message: %s\n", err.Error())
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		res, err := client.Change(c.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating change request against bankid: %s\n", err.Error())
			return c.JSON(500, "failed to start change request")
		}

		return c.JSON(http.StatusOK, res)
	})

	e.POST("/bankid/v6/collect", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			fmt.Printf("ERR: failed to read collect request message: %s\n", err.Error())
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.CollectRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal collect request message: %s\n", err.Error())
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		res, err := client.Collect(c.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating collect request against bankid: %s\n", err.Error())
			return c.JSON(500, "failed to start collect against BankID")
		}

		return c.JSON(http.StatusOK, res)
	})

	e.POST("/bankid/v6/cancel", func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			fmt.Printf("ERR: failed to read cancel request message: %s\n", err.Error())
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.CancelRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal cancel request message: %s\n", err.Error())
			return c.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		err = client.Cancel(c.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating cancel request against bankid: %s\n", err.Error())
			return c.JSON(500, "failed to start cancel against BankID")
		}

		return c.NoContent(http.StatusNoContent)
	})
}
