package bankid

import (
	"context"
	"encoding/json"
	"fmt"
	"twofer/eid"
	"twofer/eid/bankid/bankidm"
)




type eeid struct {
	parent *Client	
}

func (e eeid) Name() string {
	return Name()
}

func (e eeid) AuthInit(ctx context.Context, req eid.Request) (in eid.Intermediate, err error) {

	if req.UserIP == "" {
		req.UserIP = "127.0.0.1"
	}
	r := bankidm.AuthRequest{
		EndUserIP:   req.UserIP,
	}

	if req.User != nil && req.User.SSN != nil{
		r.SSN = req.User.SSN.SSN
	}
	resp, err := e.parent.AuthInit(ctx, r)
	if err != nil{
		return
	}
	in.Mode = eid.AUTH
	in.Ref = resp.OrderRef.OrderRef
	in.Inferred = resp.AutoStartToken
	in.QRData = fmt.Sprintf("bankid:///?autostarttoken=%s", in.Inferred)

	// Todo impalement QR

	return in, nil
}

func (e eeid) AuthCollect(ctx context.Context, req eid.Intermediate) (resp eid.Response, err error) {
	resp.Intermediate = req

	res, err := e.parent.Collect(ctx, req.Ref)
	if err != nil {
		return
	}

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

		resp.UserInfo.SSN = &eid.SSN{
			Country: "SE",
			SSN: res.CompletionData.User.PersonalNumber,
		}
		resp.UserInfo.Name = res.CompletionData.User.Name
		resp.UserInfo.Surname = res.CompletionData.User.Surname
		resp.UserInfo.GivenName = res.CompletionData.User.GivenName
		resp.Extra = map[string]interface{}{
			"ipAddress": res.CompletionData.Device.IPAddress,
			"cert": res.CompletionData.Cert,
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

func (e eeid) AuthCancel(intermediate eid.Intermediate) error {
	return e.parent.api.Cancel(intermediate.Ref)
}

func (e eeid) SignInit(ctx context.Context, req eid.Request) (in eid.Intermediate, err error) {
	if req.UserIP == "" {
		req.UserIP = "127.0.0.1"
	}
	r := bankidm.SignRequest{
		EndUserIP:   req.UserIP,

	}
	if req.User != nil && req.User.SSN != nil{
		r.SSN = req.User.SSN.SSN
	}

	if len(req.SignText) > 0 {
		r.UserVisibleData = req.SignText
	}
	if len(req.SignData) > 0 {
		r.UserNonVisibleData = req.SignData
	}

	resp, err := e.parent.SignInit(ctx, r)
	if err != nil{
		return
	}
	in.Mode = eid.SIGN
	in.Ref = resp.OrderRef.OrderRef
	in.Inferred = resp.AutoStartToken

	// Todo impalement QR
	return

}

func (e eeid) SignCollect(ctx context.Context, req eid.Intermediate) (resp eid.Response, err error) {
	resp.Intermediate = req

	res, err := e.parent.Collect(ctx, req.Ref)
	if err != nil {
		return
	}

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

		resp.UserInfo.SSN = &eid.SSN{
			Country: "SE",
			SSN: res.CompletionData.User.PersonalNumber,
		}
		resp.UserInfo.Name = res.CompletionData.User.Name
		resp.UserInfo.Surname = res.CompletionData.User.Surname
		resp.UserInfo.GivenName = res.CompletionData.User.GivenName
		resp.Extra = map[string]interface{}{
			"ipAddress": res.CompletionData.Device.IPAddress,
			"cert": res.CompletionData.Cert,
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

func (e eeid) SignCancel(intermediate eid.Intermediate) error {
	return e.parent.api.Cancel(intermediate.Ref)
}




