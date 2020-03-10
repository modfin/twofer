package freja

import (
	"context"
	"fmt"
	"net/url"
	"time"
	"twofer/eid"
	"twofer/eid/freja/frejam"
)

type eeid struct {
	parent *Client
}

func (e eeid) Name() string {
	return Name()
}

func (e eeid) AuthInit(ctx context.Context, req *eid.Req) (in *eid.Inter, err error) {

	if req.Who == nil {
		req.Inferred()
	}

	r := frejam.AuthRequest{}

	if req.Who.Inferred {
		r.UseInferred()
	} else if req.Who.SSNCountry != "" && req.Who.SSN != "" {
		err = r.UseSSN(req.Who.SSNCountry, req.Who.SSN)
		if err != nil {
			return nil, err
		}
	} else if req.Who.Phone != "" {
		r.UsePhone(req.Who.Phone)
	} else if req.Who.Email != "" {
		r.UseEmail(req.Who.Email)
	} else {
		r.UseInferred()
	}

	r.AttributesToReturn = []frejam.Attribute{
		{frejam.ATTR_BASIC_USER_INFO},
		{frejam.ATTR_SSN},
		{frejam.ATTR_DATE_OF_BIRTH},
		{frejam.ATTR_EMAIL_ADDRESS},
	}

	authRef, err := e.parent.AuthInit(ctx, r)
	if err != nil {
		return
	}
	in = &eid.Inter{
		Req:      req,
		Mode:     eid.AUTH,
		Ref:      authRef,
		Inferred: authRef,
		URI:      fmt.Sprintf("frejaeid://bindUserToTransaction?transactionReference=%s", url.QueryEscape(authRef)),
	}
	return in, nil

}

func (e eeid) SignInit(ctx context.Context, req *eid.Req) (in *eid.Inter, err error) {
	if req.Who == nil {
		req.Inferred()
	}

	r := frejam.SignRequest{}

	if req.Who.SSNCountry != "" && req.Who.SSN != "" {
		err = r.UseSSN(req.Who.SSNCountry, req.Who.SSN)
		if err != nil {
			return nil, err
		}
	} else if req.Who.Phone != "" {
		r.UsePhone(req.Who.Phone)
	} else if req.Who.Email != "" {
		r.UseEmail(req.Who.Email)
	} else {
		return nil, fmt.Errorf("ssn, phone or email must be supplied for signing")
	}

	r.AttributesToReturn = []frejam.Attribute{
		{frejam.ATTR_BASIC_USER_INFO},
		{frejam.ATTR_SSN},
		{frejam.ATTR_DATE_OF_BIRTH},
		{frejam.ATTR_EMAIL_ADDRESS},
	}

	if req.Payload != nil {
		r.DataToSign.Text = req.Payload.Text
		r.DataToSign.BinaryData = req.Payload.Data
	}

	signRef, err := e.parent.SignInit(ctx, r)
	if err != nil {
		return
	}
	in = &eid.Inter{
		Req:      req,
		Mode:     eid.SIGN,
		Ref:      signRef,
		Inferred: signRef,
		URI:      fmt.Sprintf("frejaeid://bindUserToTransaction?transactionReference=%s", url.QueryEscape(signRef)),
	}
	return in, nil

}

func (e eeid) Cancel(intermediate *eid.Inter) error {
	switch intermediate.Mode {
	case eid.SIGN:
		return e.parent.api.SignCancelRequest(intermediate.Ref)
	case eid.AUTH:
		return e.parent.api.AuthCancelRequest(intermediate.Ref)
	}
	return fmt.Errorf("mode %s is not supported to cancel", intermediate.Mode)
}

func (e eeid) Peek(ctx context.Context, inter *eid.Inter) (resp *eid.Resp, err error) {
	resp = &eid.Resp{
		Inter: inter,
	}

	var attr frejam.RequestedAttributes
	switch inter.Mode {
	case eid.AUTH:
		res, err := e.parent.api.AuthGetOneResult(inter.Ref)
		if err != nil {
			return nil, err
		}
		resp.Status = mapStatus(res.Status)
		resp.Signature = []byte(res.Details)
		attr = res.RequestedAttributes

	case eid.SIGN:
		res, err := e.parent.api.SignGetOneResult(inter.Ref)
		if err != nil {
			return nil, err
		}
		resp.Status = mapStatus(res.Status)
		resp.Signature = []byte(res.Details)
		attr = res.RequestedAttributes
	default:
		return nil, fmt.Errorf("mode %s is not valid for collect", inter.Mode)
	}

	resp.Info.SSN = attr.SSN.SSN
	resp.Info.SSNCountry = attr.SSN.Country
	resp.Info.Email = attr.EmailAddress
	resp.Info.Name = attr.BasicUserInfo.Name
	resp.Info.Surname = attr.BasicUserInfo.Surname
	resp.Info.DateOfBirth, _ = time.Parse("2006-01-02", attr.DateOfBirth)

	return
}

func (e eeid) Collect(ctx context.Context, inter *eid.Inter, cancelOnErr bool) (resp *eid.Resp, err error) {

	resp = &eid.Resp{
		Inter: inter,
	}

	var attr frejam.RequestedAttributes
	switch inter.Mode {
	case eid.AUTH:
		res, err := e.parent.AuthCollect(ctx, inter.Ref, cancelOnErr)
		if err != nil {
			return nil, err
		}
		resp.Status = mapStatus(res.Status)
		resp.Signature = []byte(res.Details)
		attr = res.RequestedAttributes

	case eid.SIGN:
		res, err := e.parent.SignCollect(ctx, inter.Ref, cancelOnErr)
		if err != nil {
			return nil, err
		}
		resp.Status = mapStatus(res.Status)
		resp.Signature = []byte(res.Details)
		attr = res.RequestedAttributes
	default:
		return nil, fmt.Errorf("mode %s is not valid for collect", inter.Mode)
	}

	resp.Info.SSN = attr.SSN.SSN
	resp.Info.SSNCountry = attr.SSN.Country
	resp.Info.Email = attr.EmailAddress
	resp.Info.Name = attr.BasicUserInfo.Name
	resp.Info.Surname = attr.BasicUserInfo.Surname
	resp.Info.DateOfBirth, _ = time.Parse("2006-01-02", attr.DateOfBirth)

	return
}

func mapStatus(status frejam.Status) eid.Status {
	switch status {
	case frejam.STATUS_STARTED, frejam.STATUS_DELIVERED_TO_MOBILE:
		return eid.STATUS_PENDING
	case frejam.STATUS_CANCELED:
		return eid.STATUS_CANCELED
	case frejam.STATUS_RP_CANCELED:
		return eid.STATUS_RP_CANCELED
	case frejam.STATUS_EXPIRED:
		return eid.STATUS_EXPIRED
	case frejam.STATUS_APPROVED:
		return eid.STATUS_APPROVED
	case frejam.STATUS_REJECTED:
		return eid.STATUS_REJECTED
	default:
		return eid.STATUS_UNKNOWN
	}

}
