package bankid

import (
	"bytes"
	"context"
	"encoding/json"
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
	baseURL string
	client  *http.Client
}

func NewAPI(client *http.Client, baseURL string) *API {
	return &API{
		client:  client,
		baseURL: baseURL,
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
		case <-time.After(time.Second):
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

func (a *API) WatchForChange(ctx context.Context, orderRef string) (<-chan WatchChanges, error) {
	collectRequest := &CollectRequest{OrderRef: orderRef}
	if err := collectRequest.Validate(); err != nil {
		return nil, err
	}

	currentState, err := a.Collect(ctx, collectRequest)
	if err != nil {
		return nil, err
	}

	watch := make(chan WatchChanges, 1) // Make it a buffered channel so that we can post initial state before we return

	sendChange := func(change WatchChanges) {
		select {
		case watch <- change:
		case <-time.After(time.Second):
		}
	}
	updateStatus := func(state *CollectResponse) {
		sendChange(WatchChanges{
			Cancelled: state.Status == Failed,
			Status:    state.Status,
			Hint:      state.HintCode,
		})
	}

	go func(lastState *CollectResponse) {
		defer close(watch)
		for {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				sendChange(WatchChanges{Cancelled: true, Err: ctx.Err()})
				return
			case <-time.After(time.Second):
			}

			resp, err := a.Collect(ctx, collectRequest)
			if err != nil {
				sendChange(WatchChanges{Cancelled: true, Err: err})
				return
			}

			if resp.Status != lastState.Status || resp.HintCode != lastState.HintCode {
				updateStatus(resp)
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
		var bidError BankIdError
		err = json.Unmarshal(body, &bidError)
		if err == nil {
			err = bidError
		}

		return nil, err
	}

	var response Response
	err = json.Unmarshal(body, &response)

	return &response, err
}
