package freja

import (
	"context"
	"fmt"
	"net/url"
	"twofer/eid"
	"twofer/eid/freja/frejam"
)

type eeid struct {
	parent *Client
}

func (e eeid) Name() string {
	return Name()
}

func (e eeid) AuthInit(ctx context.Context, req eid.Request) (in eid.Intermediate, err error) {

	r := frejam.AuthRequest{}
	r.UserInfoType = frejam.UIT_INFERRED

	if req.User != nil && req.User.SSN != nil{
		r.UserInfoType = frejam.UIT_SSN
		r.UserInfo = req.User.SSN.SSN
	}
	authRef, err := e.parent.AuthInit(ctx, r)
	if err != nil{
		return
	}
	in.Mode = eid.AUTH
	in.Ref = authRef
	in.Inferred = authRef
	in.QRData = fmt.Sprintf("frejaeid://bindUserToTransaction?transactionReference=%s", url.QueryEscape(in.Inferred))

	return in, nil
}

func (e eeid) AuthCollect(ctx context.Context, req eid.Intermediate) (resp eid.Response, err error) {
	panic("not implmented")
}

func (e eeid) AuthCancel(intermediate eid.Intermediate) error {
	return e.parent.api.AuthCancelRequest(intermediate.Ref)

}

func (e eeid) SignInit(ctx context.Context, req eid.Request) (in eid.Intermediate, err error) {
	panic("not implmented")

}

func (e eeid) SignCollect(ctx context.Context, req eid.Intermediate) (resp eid.Response, err error) {
	panic("not implmented")
}

func (e eeid) SignCancel(intermediate eid.Intermediate) error {
	return e.parent.api.SignCancelRequest(intermediate.Ref)
}
