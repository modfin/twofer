package bankid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/modfin/twofer/eid/bankid/bankidm"
	"io"
	"net/http"
)

type API struct {
	parent *Client
}

func (a *API) Ping() error {
	_, err := a.parent.client.Get(a.parent.baseURL)
	return err
}

func (a *API) Auth(ctx context.Context, request bankidm.AuthRequest) (r *bankidm.AuthResponse, err error) {

	data, err := request.Marshal()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", a.parent.baseURL+authURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "application/json")
	res, err := a.parent.client.Do(req)
	if err != nil {
		return nil, err
	}

	resdata, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		var err1 bankidm.BankIdError
		err = json.Unmarshal(resdata, &err1)
		if err == nil {
			err = err1
		}
		return nil, err
	}

	var resp bankidm.AuthResponse
	err = json.Unmarshal(resdata, &resp)
	return &resp, err

}

func (a *API) Sign(ctx context.Context, request bankidm.SignRequest) (r *bankidm.SignResponse, err error) {

	data, err := request.Marshal()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", a.parent.baseURL+signURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "application/json")
	res, err := a.parent.client.Do(req)
	if err != nil {
		return nil, err
	}

	resdata, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		var err1 bankidm.BankIdError
		err = json.Unmarshal(resdata, &err1)
		if err == nil {
			err = err1
		}
		return nil, err
	}

	var resp bankidm.SignResponse
	err = json.Unmarshal(resdata, &resp)
	return &resp, err
}

func (a *API) Collect(ctx context.Context, orderRef string) (r *bankidm.CollectResponse, err error) {

	data := []byte(fmt.Sprintf(`{"orderRef":"%s"}`, orderRef))

	req, err := http.NewRequest("POST", a.parent.baseURL+collectURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "application/json")
	res, err := a.parent.client.Do(req)
	if err != nil {
		return nil, err
	}

	resdata, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(resdata))
	if res.StatusCode != 200 {
		var err1 bankidm.BankIdError
		err = json.Unmarshal(resdata, &err1)
		if err == nil {
			err = err1
		}
		return nil, err
	}

	var resp bankidm.CollectResponse
	err = json.Unmarshal(resdata, &resp)
	return &resp, err
}

func (a *API) Cancel(ctx context.Context, orderRef string) (err error) {
	data := []byte(fmt.Sprintf(`{"orderRef":"%s"}`, orderRef))

	req, err := http.NewRequest("POST", a.parent.baseURL+cancelURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "application/json")
	res, err := a.parent.client.Do(req)
	if err != nil {
		return err
	}

	resdata, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		var err1 bankidm.BankIdError
		err = json.Unmarshal(resdata, &err1)
		if err == nil {
			err = err1
		}
		return err
	}

	return nil
}
