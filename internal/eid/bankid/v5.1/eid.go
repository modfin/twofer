package bankid_v51

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/modfin/twofer/internal/eid"
	bankidm "github.com/modfin/twofer/internal/eid/bankid/v5.1/bankidm"
	"net"
	"net/http"
	"time"
)

type Eeid struct {
	api *API
}

func NewEid(client *http.Client, baseURL string) *Eeid {
	return &Eeid{
		api: NewAPI(client, baseURL),
	}
}

func (e *Eeid) Name() string {
	return Name()
}

func (e *Eeid) AuthInit(ctx context.Context, req *eid.Req) (in *eid.Inter, err error) {
	if req.Who == nil {
		req.Inferred()
	}
	if req.Who.IP == nil {
		req.IP("127.0.0.1")
	}

	r := bankidm.AuthRequest{
		EndUserIP: req.Who.IP.String(),
	}

	if req.Who.SSN != "" {
		r.SSN = req.Who.SSN
	}
	resp, err := e.api.Auth(ctx, r)
	if err != nil {
		return
	}
	in = &eid.Inter{
		Req:      req,
		Mode:     eid.AUTH,
		Ref:      resp.OrderRef.OrderRef,
		Inferred: resp.AutoStartToken,
		URI:      fmt.Sprintf("bankid:///?autostarttoken=%s", resp.AutoStartToken),
	}

	return in, nil
}

func (e *Eeid) SignInit(ctx context.Context, req *eid.Req) (in *eid.Inter, err error) {
	if req.Who == nil {
		req.Inferred()
	}
	if req.Who.IP == nil {
		req.IP("127.0.0.1")
	}

	r := bankidm.SignRequest{
		EndUserIP: req.Who.IP.String(),
		SSN:       req.Who.SSN,
	}

	if req.Payload != nil {
		r.UserVisibleData = req.Payload.Text
		r.UserNonVisibleData = req.Payload.Data
	}

	resp, err := e.api.Sign(ctx, r)
	if err != nil {
		return
	}

	in = &eid.Inter{
		Req:      req,
		Mode:     eid.SIGN,
		Ref:      resp.OrderRef.OrderRef,
		Inferred: resp.AutoStartToken,
		URI:      fmt.Sprintf("bankid:///?autostarttoken=%s", resp.AutoStartToken),
	}
	return
}

func (e *Eeid) Change(ctx context.Context, req *eid.Inter, cancelOnErr bool) (resp *eid.Resp, err error) {
	res, err := e.change(ctx, req.Ref, cancelOnErr)
	if err != nil {
		return
	}
	resp, err = bidResToEidRes(res)
	if err != nil {
		return
	}
	resp.Inter = req
	return
}

func (e *Eeid) change(ctx context.Context, orderRef string, cancelOnErr bool) (resp *bankidm.CollectResponse, err error) {
	defer func() {
		if err != nil && cancelOnErr {
			go func() {
				fmt.Println("Canceling order,", err)
				err = e.api.Cancel(ctx, orderRef)
				if err != nil {
					fmt.Println("could not cancel order", err)
				}
			}()
		}
	}()

	startState, err := e.api.Collect(ctx, orderRef)
	if err != nil {
		return nil, err
	}
	switch startState.Status {
	case bankidm.STATUS_FAILED, bankidm.STATUS_COMPLETE:
		return resp, nil
	}
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		case <-time.After(time.Second):
		}

		resp, err = e.api.Collect(ctx, orderRef)
		if err != nil {
			return nil, err
		}

		if resp.HintCode != startState.HintCode {
			return resp, nil
		}
	}
}

func (e *Eeid) Peek(ctx context.Context, in *eid.Inter) (resp *eid.Resp, err error) {
	res, err := e.api.Collect(ctx, in.Ref)
	if err != nil {
		return
	}
	resp, err = bidResToEidRes(res)
	if err != nil {
		return
	}
	resp.Inter = in
	return
}

func (e *Eeid) Collect(ctx context.Context, in *eid.Inter, cancelOnErr bool) (resp *eid.Resp, err error) {
	res, err := e.collect(ctx, in.Ref, cancelOnErr)
	if err != nil {
		return
	}
	resp, err = bidResToEidRes(res)
	if err != nil {
		return
	}
	resp.Inter = in
	return
}

func (e *Eeid) collect(ctx context.Context, orderRef string, cancelOnErr bool) (resp *bankidm.CollectResponse, err error) {
	defer func() {
		if err != nil && cancelOnErr {
			go func() {
				fmt.Println("Canceling order,", err)
				err = e.api.Cancel(ctx, orderRef)
				if err != nil {
					fmt.Println("could not cancel order", err)
				}
			}()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		case <-time.After(time.Second):
		}

		resp, err = e.api.Collect(ctx, orderRef)
		if err != nil {
			return nil, err
		}

		switch resp.Status {
		case bankidm.STATUS_FAILED, bankidm.STATUS_COMPLETE:
			return resp, nil
		}
	}
}

func (e *Eeid) Cancel(ctx context.Context, intermediate *eid.Inter) error {
	return e.api.Cancel(ctx, intermediate.Ref)
}

func (e *Eeid) Ping() error {
	return e.api.Ping()
}

func bidResToEidRes(res *bankidm.CollectResponse) (resp *eid.Resp, err error) {
	resp = &eid.Resp{}

	switch res.Status {
	case bankidm.STATUS_PENDING:
		resp.Status = eid.STATUS_PENDING
		switch res.HintCode {
		case bankidm.HINT_STARTED, bankidm.HINT_USER_SIGN:
			resp.Status = eid.STATUS_ONGOING
		}
	case bankidm.STATUS_FAILED:
		switch res.HintCode {
		case bankidm.HINT_EXPIRED:
			resp.Status = eid.STATUS_EXPIRED
		case bankidm.HINT_USER_CANCEL:
			resp.Status = eid.STATUS_CANCELED
		case bankidm.HINT_CANCELED:
			resp.Status = eid.STATUS_RP_CANCELED
		case bankidm.HINT_START_FAILED:
			resp.Status = eid.STATUS_START_FAILED
		default:
			resp.Status = eid.STATUS_FAILED
		}
	case bankidm.STATUS_COMPLETE:
		resp.Status = eid.STATUS_APPROVED

		resp.Info.SSN = res.CompletionData.User.PersonalNumber
		resp.Info.SSNCountry = "SE"
		resp.Info.Name = res.CompletionData.User.Name
		resp.Info.Surname = res.CompletionData.User.Surname
		resp.Info.IP = net.ParseIP(res.CompletionData.Device.IPAddress)
		resp.Info.DateOfBirth, _ = time.Parse("20060102", res.CompletionData.User.PersonalNumber[:8])

		resp.Extra = map[string]interface{}{
			"cert":     res.CompletionData.Cert,
			"fullName": res.CompletionData.User.Name,
		}
		resp.Signature, err = json.Marshal(struct {
			Signature    string `json:"signature"`
			OCSPResponse string `json:"ocspResponse"`
		}{
			Signature:    res.CompletionData.Signature,
			OCSPResponse: res.CompletionData.OCSPResponse,
		})
	default:
		resp.Status = eid.STATUS_UNKNOWN
	}
	return
}
