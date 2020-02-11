package bankid

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
	"twofer/eid"
	"twofer/eid/bankid/bankidm"
)




type eeid struct {
	parent *Client	
}

func (e eeid) Name() string {
	return Name()
}

func (e eeid) AuthInit(ctx context.Context, req *eid.Req) (in *eid.Inter, err error) {
	if req.Who == nil{
		req.Inferred()
	}
	if req.Who.IP == nil{
		req.IP("127.0.0.1")
	}

	r := bankidm.AuthRequest{
		EndUserIP:   req.Who.IP.String(),
	}

	if req.Who.SSN != ""{
		r.SSN = req.Who.SSN
	}
	resp, err := e.parent.AuthInit(ctx, r)
	if err != nil{
		return
	}
	in = &eid.Inter{
		Req:      req,
		Mode:     eid.AUTH,
		Ref:      resp.OrderRef.OrderRef,
		Inferred: resp.AutoStartToken,
		QRData:   fmt.Sprintf("bankid:///?autostarttoken=%s", resp.AutoStartToken),
	}

	return in, nil
}


func (e eeid) SignInit(ctx context.Context, req *eid.Req) (in *eid.Inter, err error) {
	if req.Who == nil{
		req.Inferred()
	}
	if req.Who.IP == nil{
		req.IP("127.0.0.1")
	}


	r := bankidm.SignRequest{
		EndUserIP:  req.Who.IP.String(),
		SSN: req.Who.SSN,
	}


	if req.Payload != nil{
		r.UserVisibleData = req.Payload.Text
		r.UserNonVisibleData = req.Payload.Data
	}

	resp, err := e.parent.SignInit(ctx, r)
	if err != nil{
		return
	}

	in = &eid.Inter{
		Req:      req,
		Mode:     eid.SIGN,
		Ref:      resp.OrderRef.OrderRef,
		Inferred: resp.AutoStartToken,
		QRData:   fmt.Sprintf("bankid:///?autostarttoken=%s", resp.AutoStartToken),
	}

	return

}

func bidResToEidRes(res *bankidm.CollectResponse) (resp *eid.Resp, err error){

	resp = &eid.Resp{}

	switch res.Status {
	case bankidm.STATUS_PENDING:
		resp.Status = eid.STATUS_PENDING
	case bankidm.STATUS_FAILED:
		switch res.HintCode {
		case bankidm.HINT_EXPIRED:
			resp.Status = eid.STATUS_EXPIRED
		case bankidm.HINT_USER_CANCEL:
			resp.Status = eid.STATUS_CANCELED
		case bankidm.HINT_CANCELED:
			resp.Status = eid.STATUS_RP_CANCELED
		default:
			resp.Status = eid.STATUS_FAILED
		}
	case bankidm.STATUS_COMPLETE:
		resp.Status = eid.STATUS_APPROVED

		resp.Info.SSN = res.CompletionData.User.PersonalNumber
		resp.Info.SSNCountry = "SE"
		resp.Info.Name = res.CompletionData.User.GivenName
		resp.Info.Surname = res.CompletionData.User.Surname
		resp.Info.IP = net.ParseIP(res.CompletionData.Device.IPAddress)
		resp.Info.DateOfBirth, _ = time.Parse("20060102", res.CompletionData.User.PersonalNumber[:8])

		resp.Extra = map[string]interface{}{
			"cert": res.CompletionData.Cert,
			"fullName": res.CompletionData.User.Name,
		}
		resp.Signature, err = json.Marshal(struct {
			Signature string `json:"signature"`
			OCSPResponse string `json:"ocspResponse"`
		}{
			Signature:res.CompletionData.Signature,
			OCSPResponse:res.CompletionData.OCSPResponse,
		})
	default:
		resp.Status = eid.STATUS_UNKNOWN
	}
	return
}


func (e eeid) Peek(ctx context.Context, in *eid.Inter) (resp *eid.Resp, err error) {
	res, err :=  e.parent.api.Collect(in.Ref)
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

func (e eeid) Cancel(intermediate *eid.Inter) error {
	return e.parent.api.Cancel(intermediate.Ref)
}


