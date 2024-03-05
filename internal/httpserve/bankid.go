package httpserve

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/modfin/twofer/internal/bankid"
	"github.com/modfin/twofer/internal/sse"
	"io"
	"net/http"
	"strconv"
	"time"
)

func RegisterBankIDServer(e *echo.Echo, client *bankid.API) {
	e.POST("/bankid/v6/auth", func(e echo.Context) error {
		b, err := io.ReadAll(e.Request().Body)
		if err != nil {
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.AuthSignRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal auth request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		res, err := client.Auth(e.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating auth request against bankid: %s\n", err.Error())
			return e.JSON(500, "failed to initiate auth against BankId")
		}

		sender, err := sse.NewSender(e.Response().Writer)
		if err != nil {
			fmt.Printf("ERR: failed to setup auth response stream: %s\n", err.Error())
			return e.JSON(500, "failed to setup response stream")
		}

		sender.Prepare()

		interrupt := client.WatchForChange(e.Request().Context(), res.OrderRef)

		for i := 0; i < 30; i++ {
			select {
			case <-interrupt:
				return e.JSON(http.StatusOK, bankid.Empty{})
			default:
				break
			}

			msg := bankid.AuthSignAPIResponse{
				OrderRef: res.OrderRef,
				URI:      fmt.Sprintf("bankid:///?autostarttoken=%s", res.AutoStartToken),
				QR:       res.BuildQrCode(i),
			}

			var bytes []byte
			bytes, err = json.Marshal(msg)
			if err != nil {
				fmt.Printf("ERR: failed to build auth response message: %s\n", err.Error())
				return e.JSON(500, "failed to build response message")
			}

			event := sse.Event{
				Id:    strconv.Itoa(i),
				Event: "message",
				Data:  string(bytes),
			}

			err = sender.Send(event)
			if err != nil {
				fmt.Printf("ERR: failed to send auth response message: %s\n", err.Error())
				return e.JSON(500, "failed to send response message")
			}

			// Optimally subtract time that has elapsed, but no need to be that exact
			time.Sleep(time.Second)
		}

		return e.JSON(http.StatusOK, bankid.Empty{})
	})

	e.POST("/bankid/v6/sign", func(e echo.Context) error {
		b, err := io.ReadAll(e.Request().Body)
		if err != nil {
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.AuthSignRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal sign request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		res, err := client.Sign(e.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating sign request against bankid: %s\n", err.Error())
			return e.JSON(500, "failed to initiate sign against BankId")
		}

		sender, err := sse.NewSender(e.Response().Writer)
		if err != nil {
			fmt.Printf("ERR: failed to setup sign response stream: %s\n", err.Error())
			return e.JSON(500, "failed to setup response stream")
		}

		sender.Prepare()

		interrupt := client.WatchForChange(e.Request().Context(), res.OrderRef)

		for i := 0; i < 30; i++ {
			select {
			case <-interrupt:
				return e.JSON(http.StatusOK, bankid.Empty{})
			default:
				break
			}

			msg := bankid.AuthSignAPIResponse{
				OrderRef: res.OrderRef,
				URI:      fmt.Sprintf("bankid:///?autostarttoken=%s", res.AutoStartToken),
				QR:       res.BuildQrCode(i),
			}

			var bytes []byte
			bytes, err = json.Marshal(msg)
			if err != nil {
				fmt.Printf("ERR: failed to build sign response message: %s\n", err.Error())
				return e.JSON(500, "failed to build response message")
			}

			event := sse.Event{
				Id:    strconv.Itoa(i),
				Event: "message",
				Data:  string(bytes),
			}

			err = sender.Send(event)
			if err != nil {
				fmt.Printf("ERR: failed to send sign response message: %s\n", err.Error())
				return e.JSON(500, "failed to send response message")
			}

			// Optimally subtract time that has elapsed, but no need to be that exact
			time.Sleep(time.Second)
		}

		return e.JSON(http.StatusOK, bankid.Empty{})
	})

	e.POST("/bankid/v6/change", func(e echo.Context) error {
		b, err := io.ReadAll(e.Request().Body)
		if err != nil {
			fmt.Printf("ERR: failed to read change request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.ChangeRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal change request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		res, err := client.Change(e.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating change request against bankid: %s\n", err.Error())
			return e.JSON(500, "failed to start change request")
		}

		return e.JSON(http.StatusOK, res)
	})

	e.POST("/bankid/v6/collect", func(e echo.Context) error {
		b, err := io.ReadAll(e.Request().Body)
		if err != nil {
			fmt.Printf("ERR: failed to read collect request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.CollectRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal collect request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		res, err := client.Collect(e.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating collect request against bankid: %s\n", err.Error())
			return e.JSON(500, "failed to start collect against BankID")
		}

		return e.JSON(http.StatusOK, res)
	})

	e.POST("/bankid/v6/cancel", func(e echo.Context) error {
		b, err := io.ReadAll(e.Request().Body)
		if err != nil {
			fmt.Printf("ERR: failed to read cancel request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.CollectRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal cancel request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		_, err = client.Collect(e.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating cancel request against bankid: %s\n", err.Error())
			return e.JSON(500, "failed to start cancel against BankID")
		}

		return e.NoContent(http.StatusNoContent)
	})
}
