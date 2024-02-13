package bankid

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
	"github.com/modfin/twofer/eid"
	"github.com/modfin/twofer/eid/bankid/bankidm"
)

type eeid struct {
	parent *Client
}

func (e eeid) Name() string {
	return Name()
}

func (e eeid) Change(ctx context.Context, req *eid.Inter, cancelOnErr bool) (resp *eid.Resp, err error) {
	res, err := e.parent.Change(ctx, req.Ref, cancelOnErr)
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

func (e eeid) AuthInit(ctx context.Context, req *eid.Req) (in *eid.Inter, err error) {
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
	resp, err := e.parent.API().Auth(ctx, r)
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

func (e eeid) SignInit(ctx context.Context, req *eid.Req) (in *eid.Inter, err error) {
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

	resp, err := e.parent.API().Sign(ctx, r)
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

func (e eeid) Peek(ctx context.Context, in *eid.Inter) (resp *eid.Resp, err error) {
	res, err := e.parent.api.Collect(ctx, in.Ref)
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

func (e eeid) Collect(ctx context.Context, in *eid.Inter, cancelOnErr bool) (resp *eid.Resp, err error) {
	res, err := e.parent.Collect(ctx, in.Ref, cancelOnErr)
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

func (e eeid) Cancel(ctx context.Context, intermediate *eid.Inter) error {
	return e.parent.api.Cancel(ctx, intermediate.Ref)
}

func (e eeid) Ping() error {
	return e.parent.API().Ping()
}
