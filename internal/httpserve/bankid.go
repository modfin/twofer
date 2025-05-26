package httpserve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/modfin/twofer/api"
	"github.com/modfin/twofer/internal/bankid"
	"github.com/modfin/twofer/internal/sse"
	"github.com/modfin/twofer/stream"
)

const qrCodeUpdatePeriod = time.Second

type NewStreamEncoder func(http.ResponseWriter) (stream.Encoder, error)

func RegisterBankIDServer(e *echo.Echo, client *bankid.API, newEncoder NewStreamEncoder) {
	e.POST("/bankid/v6/auth", auth(client))
	e.POST("/bankid/v6/authv2", authSign(client.Auth, client.WatchForChangeV2, qrCodeUpdatePeriod, newEncoder)) // Deprecated: Don't use
	e.POST("/bankid/v6/authv3", authSignV3(client.Auth, qrCodeUpdatePeriod, newEncoder))                        // Same as 'auth' except won't poll BankID collect API (since a completed/failed orderRef can only be collected once)
	e.POST("/bankid/v6/sign", sign(client))
	e.POST("/bankid/v6/signv2", authSign(client.Sign, client.WatchForChangeV2, qrCodeUpdatePeriod, newEncoder)) // Deprecated: Don't use
	e.POST("/bankid/v6/signv3", authSignV3(client.Sign, qrCodeUpdatePeriod, newEncoder))                        // Same as 'sign' except won't poll BankID collect API (since a completed/failed orderRef can only be collected once)
	e.POST("/bankid/v6/change", change(client))
	e.POST("/bankid/v6/collect", collect(client))
	e.POST("/bankid/v6/collectV3", collectV3(client))
	e.POST("/bankid/v6/cancel", cancel(client))
	e.POST("/bankid/v6/cancelV3", cancelV3(client))
}

func auth(client *bankid.API) func(echo.Context) error {
	// TODO: Refactor auth and sign into single function that can handle both since the code if pretty much identical?
	return func(e echo.Context) error {
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

		// In the case a client wants to initiate a new request every second instead of relying on SSE
		// we respond with the first entry and then close the connection
		response := e.QueryParam("type")
		if response == "once" {
			msg := bankid.AuthSignAPIResponse{
				OrderRef: res.OrderRef,
				URI:      fmt.Sprintf("bankid:///?autostarttoken=%s&redirect=null", res.AutoStartToken),
				QR:       res.BuildQrCode(0),
			}

			return e.JSON(200, msg)
		}

		sender, err := sse.NewSender(e.Response())
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
				URI:      fmt.Sprintf("bankid:///?autostarttoken=%s&redirect=null", res.AutoStartToken),
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
	}
}

func sign(client *bankid.API) func(echo.Context) error {
	return func(e echo.Context) error {
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

		// In the case a client wants to initiate a new request every second instead of relying on SSE
		// we respond with the first entry and then close the connection
		response := e.QueryParam("type")
		if response == "once" {
			msg := bankid.AuthSignAPIResponse{
				OrderRef: res.OrderRef,
				URI:      fmt.Sprintf("bankid:///?autostarttoken=%s&redirect=null", res.AutoStartToken),
				QR:       res.BuildQrCode(0),
			}

			return e.JSON(200, msg)
		}

		sender, err := sse.NewSender(e.Response())
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
				URI:      fmt.Sprintf("bankid:///?autostarttoken=%s&redirect=null", res.AutoStartToken),
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
	}
}

func change(client *bankid.API) func(echo.Context) error {
	return func(e echo.Context) error {
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
	}
}

func collect(client *bankid.API) func(echo.Context) error {
	return func(e echo.Context) error {
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
	}
}

func cancel(client *bankid.API) func(echo.Context) error {
	return func(e echo.Context) error {
		b, err := io.ReadAll(e.Request().Body)
		if err != nil {
			fmt.Printf("ERR: failed to read cancel request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload"})
		}

		var request bankid.CancelRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal cancel request message: %s\n", err.Error())
			return e.JSON(400, bankid.GenericResponse{Message: "invalid request payload content"})
		}

		err = client.Cancel(e.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating cancel request against bankid: %s\n", err.Error())
			return e.JSON(500, "failed to start cancel against BankID")
		}

		return e.NoContent(http.StatusNoContent)
	}
}

// Deprecated V2 API, stick to V1 or switch to V3
func createResponseFromAuthSign(r *bankid.AuthSignResponse, qrCodeTime int) api.BankIdV6Response {
	return api.BankIdV6Response{
		OrderRef: r.OrderRef,
		URI:      fmt.Sprintf("bankid:///?autostarttoken=%s&redirect=null", r.AutoStartToken),
		QR:       r.BuildQrCode(qrCodeTime),
	}
}

// Deprecated V2 API, stick to V1 or switch to V3
func createResponseFromCollect(change bankid.Change) api.BankIdV6Response {
	// Status == pending, only send status and hint updates
	if change.Status == bankid.Pending {
		return api.BankIdV6Response{
			OrderRef: change.OrderRef,
			Status:   string(change.Status),
			HintCode: string(change.HintCode),
		}
	}

	// Status == (complete | failed), send complete message
	return api.BankIdV6Response{
		OrderRef: change.OrderRef,
		Status:   string(change.Status),
		HintCode: string(change.HintCode),
		CompletionData: &api.BankIdV6CompletionData{
			User:            api.BankIdV6User(change.CompletionData.User),
			Device:          api.BankIdV6Device(change.CompletionData.Device),
			BankIdIssueDate: change.CompletionData.BankIdIssueDate,
			StepUp:          api.BankIdV6StepUp(change.CompletionData.StepUp),
			Signature:       change.CompletionData.Signature,
			OcspResponse:    change.CompletionData.OcspResponse,
		},
	}
}

// Deprecated V2 API, stick to V1 or switch to V3
func createResponseFromError(orderRef string, err error) api.BankIdV6Response {
	var bie bankid.BankIdError
	if errors.As(err, &bie) {
		return api.BankIdV6Response{
			OrderRef:  orderRef,
			Status:    api.StatusError,
			ErrorCode: bie.ErrorCode,
			ErrorText: bie.Details,
		}
	}
	return api.BankIdV6Response{
		OrderRef:  orderRef,
		Status:    api.StatusError,
		ErrorText: err.Error(),
	}
}

type (
	authSignFn func(context.Context, *bankid.AuthSignRequest) (*bankid.AuthSignResponse, error)
	watchFn    func(context.Context, string) (<-chan bankid.Change, error)
)

const (
	qrCodeEvent = "qrcode"
	statusEvent = "status"
	errorEvent  = "error"
)

// Deprecated: V2 API, use either /bankid/v6/auth or /bankid/v6/authv3 endpoints
func authSign(authSign authSignFn, watch watchFn, qrPeriod time.Duration, newStreamEncoder NewStreamEncoder) func(echo.Context) error {
	return func(c echo.Context) error {
		b, err := io.ReadAll(c.Request().Body)
		if err != nil {
			fmt.Printf("ERR: failed to read request body: %v\n", err)
			return c.JSON(400, createResponseFromError("", err))
		}

		var request bankid.AuthSignRequest
		err = json.Unmarshal(b, &request)
		if err != nil {
			fmt.Printf("ERR: failed to unmarshal auth request message: %v\n", err)
			return c.JSON(400, createResponseFromError("", err))
		}

		res, err := authSign(c.Request().Context(), &request)
		if err != nil {
			fmt.Printf("ERR: initiating auth request against bankid: %v\n", err)
			return c.JSON(400, createResponseFromError("", err))
		}

		// In the case a client wants to initiate a new request every second instead of relying on SSE
		// we respond with the first entry and then close the connection
		response := c.QueryParam("type")
		if response == "once" {
			return c.JSON(200, createResponseFromAuthSign(res, 0))
		}

		send, err := newStreamEncoder(c.Response())
		if err != nil {
			fmt.Printf("ERR: failed to setup auth response stream: %s\n", err.Error())
			return c.JSON(400, createResponseFromError(res.OrderRef, err))
		}

		err = send("", qrCodeEvent, createResponseFromAuthSign(res, 0))
		if err != nil {
			fmt.Printf("ERR: failed to write auth response to stream: %s\n", err.Error())
			return c.JSON(400, createResponseFromError(res.OrderRef, err))
		}

		changes, err := watch(c.Request().Context(), res.OrderRef)
		if err != nil {
			fmt.Printf("ERR: failed to setup response stream: %s\n", err.Error())
			return c.JSON(400, createResponseFromError(res.OrderRef, err))
		}

		// Stream new QR codes and status changes back to caller, while waiting
		// for status to become != pending. If hintCode is 'userSign', the barcode
		// have been read, and we'll stop sending new QR code strings to caller.
		updateQR := time.NewTicker(qrPeriod)
		qrCount := 1
		for {
			select {
			case <-updateQR.C:
				err = send("", qrCodeEvent, createResponseFromAuthSign(res, qrCount))
				if err != nil {
					fmt.Printf("ERR: failed to send updated QR code message: %v\n", err)
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to send updated QR code message")
				}
				qrCount++
			case state, ok := <-changes:
				if !ok {
					// channel were unexpectedly closed
					fmt.Println("ERR: change channel were unexpectedly closed")
					return echo.NewHTTPError(http.StatusInternalServerError, "change channel were unexpectedly closed")
				}

				if state.Err != nil {
					// Something failed, channel will close after this...
					fmt.Printf("ERR: collect returned error: %v\n", state.Err)
					err = send("", errorEvent, createResponseFromError(res.OrderRef, state.Err))
					if err != nil {
						fmt.Printf("ERR: failed to send status update: %v\n", err)
						return echo.NewHTTPError(http.StatusInternalServerError, "failed to send error message")
					}
					return nil
				}

				if state.HintCode == "userSign" {
					// Stop updateQR timer when QR-code have been scanned
					updateQR.Stop()
				}

				// Stream latest status to caller
				err = send("", statusEvent, createResponseFromCollect(state))
				if err != nil {
					fmt.Printf("ERR: failed to send status update: %v\n", err)
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to send status update")
				}

				// Check for completion
				if state.Status != "" && state.Status != bankid.Pending {
					fmt.Printf("reached status: %v\n", state.Status)
					return nil
				}
			}
		}
	}
}

func readBody[T any](body io.ReadCloser) (*T, error) {
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("invalid request payload: %w", err)
	}

	var request T
	err = json.Unmarshal(b, &request)
	if err != nil {
		return nil, fmt.Errorf("invalid request payload content: %w", err)
	}
	return &request, nil
}

func bankIdv6ErrorResponseV3(err error, detail string) api.BankIdv6ErrorResponseV3 {
	if err != nil {
		var bie bankid.BankIdError
		if errors.As(err, &bie) {
			return api.BankIdv6ErrorResponseV3{
				Origin:     api.ErrorOriginBankIDv6,
				StatusCode: bie.StatusCode,
				Code:       bie.ErrorCode,
				Detail:     bie.Details,
			}
		}
		return api.BankIdv6ErrorResponseV3{
			Origin: api.ErrorOriginTwofer,
			Code:   err.Error(),
			Detail: detail,
		}
	}
	return api.BankIdv6ErrorResponseV3{
		Origin: api.ErrorOriginTwofer,
		Detail: detail,
	}
}

// Similar to the /bankid/v6/auth and /bankid/v6/sign endpoints.
// The biggest difference is that it won't call the BankID collect API to watch for changes, since
// a completed/failed orderRef can only be collected once, so this endpoint will continue to send
// QR-codes for 30 seconds. All status updates must be handled outside twofer, by calling collect
// or change endpoints after the first QR-code has been returned. It also returns one or more
// api.BankIdV6AuthSignResponseV3 structs for a successful auth/sign request. For failed requests,
// a api.BankIdv6ErrorResponseV3 is returned instead.
func authSignV3(authOrSignFn authSignFn, qrPeriod time.Duration, newStreamEncoder NewStreamEncoder) func(echo.Context) error {
	return func(c echo.Context) error {
		request, err := readBody[api.BankIdv6AuthSignRequestV3](c.Request().Body)
		if err != nil {
			fmt.Printf("ERR: read request body error: %v\n", err)
			return c.JSON(http.StatusBadRequest, bankIdv6ErrorResponseV3(err, "read request body error"))
		}

		// Convert from public API to internal struct
		br := bankid.Requirement{
			PinCode:        request.PinCode,
			PersonalNumber: request.PersonalNumber,
		}
		r := &bankid.AuthSignRequest{
			EndUserIp:             request.EndUserIp,
			ReturnUrl:             request.ReturnUrl,
			Requirement:           br,
			UserVisibleData:       request.UserVisibleData,
			UserNonVisibleData:    request.UserNonVisibleData,
			UserVisibleDataFormat: request.UserVisibleDataFormat,
		}

		res, err := authOrSignFn(c.Request().Context(), r)
		if err != nil {
			fmt.Printf("ERR: auth/sign request error: %v\n", err)
			return c.JSON(http.StatusBadRequest, bankIdv6ErrorResponseV3(err, "auth/sign request error"))
		}

		bankIdV6AuthSignResponseV3 := func(r *bankid.AuthSignResponse, qrNo int) api.BankIdV6AuthSignResponseV3 {
			return api.BankIdV6AuthSignResponseV3{
				OrderRef: r.OrderRef,
				URI:      fmt.Sprintf("bankid:///?autostarttoken=%s&redirect=null", r.AutoStartToken),
				QR:       r.BuildQrCode(qrNo),
			}
		}

		// In the case a client wants to initiate a new request every second instead of relying on SSE
		// we respond with the first entry and then close the connection
		if request.Once {
			return c.JSON(http.StatusOK, bankIdV6AuthSignResponseV3(res, 0))
		}

		// Create SSE / NDJSON event stream
		send, err := newStreamEncoder(c.Response())
		if err != nil {
			fmt.Printf("ERR: failed to setup response stream: %v\n", err)
			return c.JSON(http.StatusInternalServerError, bankIdv6ErrorResponseV3(err, "failed to setup response stream"))
		}

		// Stream new QR codes for about 30 seconds
		for i := 0; i < 30; i++ {
			err = send(strconv.Itoa(i), "message", bankIdV6AuthSignResponseV3(res, i))
			if err != nil {
				fmt.Printf("ERR: failed to send QR-code message: %v\n", err)
				return c.JSON(http.StatusInternalServerError, bankIdv6ErrorResponseV3(err, "failed to send response message"))
			}
			time.Sleep(qrPeriod)
		}
		return nil
	}
}

// Pretty much the same as collect and change, except that it will return an api.BankIdV6CollectResponseV3 struct for
// successful requests, for failed requests, an api.BankIdv6ErrorResponseV3 is returned instead.
func collectV3(client *bankid.API) func(echo.Context) error {
	return func(c echo.Context) error {
		request, err := readBody[api.BankIdv6CollectRequestV3](c.Request().Body)
		if err != nil {
			fmt.Printf("ERR: read request body error: %v\n", err)
			return c.JSON(http.StatusBadRequest, bankIdv6ErrorResponseV3(err, "read request body error"))
		}

		var res *bankid.CollectResponse
		if request.WaitForChange || request.WaitUntilFinished {
			res, err = client.ChangeV3(c.Request().Context(), &bankid.ChangeRequest{
				OrderRef:          request.OrderRef,
				WaitUntilFinished: request.WaitUntilFinished,
			})
		} else {
			res, err = client.Collect(c.Request().Context(), &bankid.CollectRequest{OrderRef: request.OrderRef})
		}
		if err != nil {
			fmt.Printf("ERR: collect request error: %v\n", err)
			return c.JSON(http.StatusBadRequest, bankIdv6ErrorResponseV3(err, "collect request error"))
		}

		reply := api.BankIdV6CollectResponseV3{
			OrderRef: res.OrderRef,
			Status:   string(res.Status),
			HintCode: string(res.HintCode),
		}
		if res.Status == bankid.Complete {
			reply.CompletionData = &api.BankIdV6CompletionData{
				User:            api.BankIdV6User(res.CompletionData.User),
				Device:          api.BankIdV6Device(res.CompletionData.Device),
				BankIdIssueDate: res.CompletionData.BankIdIssueDate,
				StepUp:          api.BankIdV6StepUp(res.CompletionData.StepUp),
				Signature:       res.CompletionData.Signature,
				OcspResponse:    res.CompletionData.OcspResponse,
			}
		}
		return c.JSON(http.StatusOK, reply)
	}
}

// Pretty much the same as cancel, except that it will api.BankIdv6CancelResponseV3 struct for a successful request,
// for failed requests, an api.BankIdv6ErrorResponseV3 is returned instead.
func cancelV3(client *bankid.API) func(echo.Context) error {
	return func(c echo.Context) error {
		request, err := readBody[api.BankIdv6CancelRequestV3](c.Request().Body)
		if err != nil {
			fmt.Printf("ERR: read request body error: %v\n", err)
			return c.JSON(http.StatusBadRequest, bankIdv6ErrorResponseV3(err, "read request body error"))
		}

		err = client.Cancel(c.Request().Context(), &bankid.CancelRequest{OrderRef: request.OrderRef})
		if err != nil {
			fmt.Printf("ERR: cancel request error: %v\n", err)
			return c.JSON(http.StatusBadRequest, bankIdv6ErrorResponseV3(err, "cancel request error"))
		}

		return c.JSON(http.StatusOK, api.BankIdv6CancelResponseV3{Status: api.StatusComplete})
	}
}
