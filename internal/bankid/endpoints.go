package bankid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"time"
)

const (
	AuthUrl    = "/rp/v6.0/auth"
	SignUrl    = "/rp/v6.0/sign"
	CollectUrl = "/rp/v6.0/collect"
	CancelUrl  = "/rp/v6.0/cancel"
)

type API struct {
	baseURL      string
	client       *http.Client
	pollInterval time.Duration
}

func NewAPI(client *http.Client, baseURL string, pollInterval time.Duration) *API {
	return &API{
		client:       client,
		baseURL:      baseURL,
		pollInterval: pollInterval,
	}
}

func (a *API) Ping() error {
	_, err := a.client.Get(a.baseURL)
	return err
}

func (a *API) Auth(ctx context.Context, r *AuthSignRequest) (*AuthSignResponse, error) {
	err := r.ValidateAuthRequest()
	if err != nil {
		return nil, err
	}

	res, err := post[AuthSignRequest, AuthSignResponse](ctx, a.client, r, a.baseURL+AuthUrl)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (a *API) Sign(ctx context.Context, r *AuthSignRequest) (*AuthSignResponse, error) {
	err := r.ValidateSignRequest()
	if err != nil {
		return nil, err
	}

	return post[AuthSignRequest, AuthSignResponse](ctx, a.client, r, a.baseURL+SignUrl)
}

func (a *API) Collect(ctx context.Context, r *CollectRequest) (*CollectResponse, error) {
	err := r.Validate()
	if err != nil {
		return nil, err
	}

	res, err := post[CollectRequest, CollectResponse](ctx, a.client, r, a.baseURL+CollectUrl)
	if err != nil {
		return nil, err
	}

	return res, err
}

func (a *API) Change(ctx context.Context, r *ChangeRequest) (*CollectResponse, error) {
	err := r.Validate()
	if err != nil {
		return nil, err
	}

	collectRequest := &CollectRequest{OrderRef: r.OrderRef}

	startState, err := a.Collect(ctx, collectRequest)
	if err != nil {
		return nil, err
	}

	if startState.Status == Complete || startState.Status == Failed {
		return startState, nil
	}

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		case <-time.After(a.pollInterval):
		}

		var resp *CollectResponse
		resp, err = a.Collect(ctx, collectRequest)
		if err != nil {
			return nil, err
		}

		if r.WaitUntilFinished && resp.Status != Pending {
			return resp, nil
		}

		if resp.HintCode != startState.HintCode {
			return resp, nil
		}
	}
}

func (a *API) ChangeV3(ctx context.Context, r *ChangeRequest) (*CollectResponse, error) {
	err := r.Validate()
	if err != nil {
		return nil, err
	}

	collectRequest := &CollectRequest{OrderRef: r.OrderRef}

	startState, err := a.Collect(ctx, collectRequest)
	if err != nil {
		return nil, err
	}

	if startState.Status == Complete || startState.Status == Failed {
		return startState, nil
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(a.pollInterval):
		}

		var resp *CollectResponse
		resp, err = a.Collect(ctx, collectRequest)
		if err != nil {
			return nil, err
		}

		if r.WaitUntilFinished {
			if resp.Status != Pending {
				return resp, nil
			}
			continue
		}

		if resp.HintCode != startState.HintCode {
			return resp, nil
		}
	}
}

func (a *API) WatchForChange(ctx context.Context, orderRef string) <-chan WatchResponse {
	watch := make(chan WatchResponse)

	go func() {
		changeRequest := &ChangeRequest{
			OrderRef:          orderRef,
			WaitUntilFinished: false,
		}

		check, err := a.Change(ctx, changeRequest)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				fmt.Println("ERR from change in bankid v6 auth/sign init: ", err)
				watch <- WatchResponse{
					Cancelled: true,
					Status:    "",
				}
				close(watch)
				return
			}

			watch <- WatchResponse{
				Cancelled: true,
				Status:    err.Error(),
			}
			close(watch)
			return
		}

		watch <- WatchResponse{
			Cancelled: false,
			Status:    string(check.Status),
		}
		close(watch)
		return
	}()

	return watch
}

type Change struct {
	CollectResponse
	Err error
}

func (a *API) WatchForChangeV2(ctx context.Context, orderRef string) (<-chan Change, error) {
	collectRequest := &CollectRequest{OrderRef: orderRef}

	currentState, err := a.Collect(ctx, collectRequest)
	if err != nil {
		return nil, err
	}

	watch := make(chan Change, 1) // Make it a buffered channel so that we can post initial state before we return

	sendError := func(err error) {
		select {
		case watch <- Change{Err: err}:
		case <-time.After(time.Second):
			fmt.Printf("WatchForChangeV2 send timeout, failed to send: %v", err)
		}
	}
	sendChange := func(change CollectResponse) {
		select {
		case watch <- Change{CollectResponse: change}:
		case <-time.After(time.Second):
			fmt.Printf("WatchForChangeV2 send timeout, failed to send: %v", change)
		}
	}

	sendChange(*currentState)

	go func(lastState *CollectResponse) {
		// Poll BankID every two seconds (according to their spec)
		pollTicker := time.NewTicker(time.Second * 2)
		defer pollTicker.Stop()
		defer close(watch)
		for {
			select {
			case <-ctx.Done():
				sendError(ctx.Err())
				return
			case <-pollTicker.C:
			}

			resp, err := a.Collect(ctx, collectRequest)
			if err != nil {
				sendError(err)
				return
			}

			if resp.Status != lastState.Status || resp.HintCode != lastState.HintCode {
				sendChange(*resp)
				lastState = resp
			}

			if resp.Status != Pending {
				return
			}
		}
	}(currentState)

	return watch, nil
}

func (a *API) Cancel(ctx context.Context, r *CancelRequest) error {
	err := r.Validate()
	if err != nil {
		return err
	}

	_, err = post[CancelRequest, Empty](ctx, a.client, r, a.baseURL+CancelUrl)
	return err
}

func post[Request any, Response any](ctx context.Context, client *http.Client, r *Request, url string) (*Response, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Add("content-type", "application/json")

	res, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		fmt.Printf("%s returned status code %d with data: %s\n", url, res.StatusCode, body)
		var bidError BankIdError
		if res.Header.Get(echo.HeaderContentType) == echo.MIMEApplicationJSON {
			err = json.Unmarshal(body, &bidError)
			if err != nil {
				fmt.Printf("failed to unmarshal BankIdError, error: %v\n", err)
			}
		}
		bidError.StatusCode = res.StatusCode
		return nil, bidError
	}

	var response Response
	err = json.Unmarshal(body, &response)

	return &response, err
}
