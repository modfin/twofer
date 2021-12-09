package httpclients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"twofer/grpc/gpwd"
)

type PwdClient struct {
	c       *http.Client
	baseUrl string
}

func NewPwdClient(baseurl string) *PwdClient {
	return &PwdClient{
		c:       http.DefaultClient,
		baseUrl: baseurl,
	}
}

func (c *PwdClient) Enroll(ctx context.Context, req *gpwd.EnrollReq) (gpwd.Blob, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return gpwd.Blob{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/pwd/enroll")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return gpwd.Blob{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return gpwd.Blob{}, err
	}
	var userBlob gpwd.Blob
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return gpwd.Blob{}, err
	}
	err = json.Unmarshal(b, &userBlob)
	if err != nil {
		return gpwd.Blob{}, err
	}
	return userBlob, nil
}

func (c *PwdClient) Auth(ctx context.Context, req *gpwd.AuthReq) (gpwd.Res, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return gpwd.Res{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/pwd/auth")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return gpwd.Res{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return gpwd.Res{}, err
	}
	var userRes gpwd.Res
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return gpwd.Res{}, err
	}
	err = json.Unmarshal(b, &userRes)
	if err != nil {
		return gpwd.Res{}, err
	}
	return userRes, nil
}
