package httpserve

import (
	"encoding/json"
	"fmt"
	"github.com/modfin/twofer/internal/sse"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/modfin/twofer/internal/bankid"
)

type (
	statusUpdate struct {
		Status bankid.Status
		Hint   bankid.HintCode
	}
)

func RegisterBankIDServer(e *echo.Echo, client *bankid.API) {
	e.POST("/bankid/v6/auth", auth(client))
	e.POST("/bankid/v6/sign", sign(client))
	e.POST("/bankid/v6/change", change(client))
	e.POST("/bankid/v6/collect", collect(client))
	e.POST("/bankid/v6/cancel", cancel(client))
}

func auth(client *bankid.API) func(c echo.Context) error {
	// TODO: Refactor auth and sign into single function that can handle both since the code if pretty much identical?
	return func(c echo.Context) error {
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

		sender, err := sse.NewSender(c.Response())
		if err != nil {
			fmt.Printf("ERR: failed to setup auth response stream: %s\n", err.Error())
			return err
		}

		sender.Prepare()

		changes, err := client.WatchForChange(c.Request().Context(), res.OrderRef)
		if err != nil {
			return c.JSON(500, "failed to setup response stream")
		}

		updateQR := time.NewTicker(time.Second)
		qrCount := 1
		for {
			select {
			case <-updateQR.C:
				err = sender.Send("message", res.APIResponse(qrCount)) // TODO: Should we change "message" to something else?
				if err != nil {
					fmt.Printf("ERR: failed to send auth response message: %v\n", err)
					return c.JSON(500, "failed to send response message")
				}
			case state, ok := <-changes:
				if !ok {
					return c.JSON(http.StatusOK, bankid.Empty{})
				}
				// TODO: Stop updateQR timer when QR-code have been scanned
				err = sender.Send("status", statusUpdate{Status: state.Status, Hint: state.Hint})
				if err != nil {
					fmt.Printf("ERR: failed to send status update: %v\n", err)
					return c.JSON(500, "failed to send response message")
				}
			}
		}
	}
}

func sign(client *bankid.API) func(c echo.Context) error {
	// TODO: Refactor auth and sign into single function that can handle both since the code if pretty much identical?
	return func(c echo.Context) error {
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

		sender, err := sse.NewSender(c.Response())
		if err != nil {
			fmt.Printf("ERR: failed to setup auth response stream: %s\n", err.Error())
			return err
		}

		sender.Prepare()

		changes, err := client.WatchForChange(c.Request().Context(), res.OrderRef)
		if err != nil {
			return c.JSON(500, "failed to setup response stream")
		}

		updateQR := time.NewTicker(time.Second)
		qrCount := 1
		for {
			select {
			case <-updateQR.C:
				err = sender.Send("message", res.APIResponse(qrCount)) // TODO: Should we change "message" to something else?
				if err != nil {
					fmt.Printf("ERR: failed to send sign response message: %v\n", err)
					return c.JSON(500, "failed to send response message")
				}
			case state, ok := <-changes:
				if !ok {
					return c.JSON(http.StatusOK, bankid.Empty{})
				}
				// TODO: Stop updateQR timer when QR-code have been scanned
				err = sender.Send("status", statusUpdate{Status: state.Status, Hint: state.Hint})
				if err != nil {
					fmt.Printf("ERR: failed to send status update: %v\n", err)
					return c.JSON(500, "failed to send response message")
				}
			}
		}
	}
}

func change(client *bankid.API) func(c echo.Context) error {
	return func(c echo.Context) error {
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
	}
}

func collect(client *bankid.API) func(c echo.Context) error {
	return func(c echo.Context) error {
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
	}
}

func cancel(client *bankid.API) func(c echo.Context) error {
	return func(c echo.Context) error {
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
	}
}
