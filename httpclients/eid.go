package httpclients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"twofer/grpc/geid"
)

type EidClient struct {
	c       *http.Client
	baseUrl string
}

func NewEidClient(baseurl string) *EidClient {
	return &EidClient{
		c:       http.DefaultClient,
		baseUrl: baseurl,
	}
}

func (c *EidClient) Providers(ctx context.Context) (geid.Providers, error) {
	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/providers")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return geid.Providers{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return geid.Providers{}, err
	}
	var geidProviders geid.Providers
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return geid.Providers{}, err
	}
	err = json.Unmarshal(b, &geidProviders)
	if err != nil {
		return geid.Providers{}, err
	}
	return geidProviders, nil
}

func (c *EidClient) AuthInit(ctx context.Context, req *geid.Req) (geid.Inter, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return geid.Inter{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/auth")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return geid.Inter{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return geid.Inter{}, err
	}
	var inter geid.Inter
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return geid.Inter{}, err
	}
	err = json.Unmarshal(b, &inter)
	if err != nil {
		return geid.Inter{}, err
	}
	return inter, nil
}

func (c *EidClient) SignInit(ctx context.Context, req *geid.Req) (geid.Inter, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return geid.Inter{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/sign")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return geid.Inter{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return geid.Inter{}, err
	}
	var inter geid.Inter
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return geid.Inter{}, err
	}
	err = json.Unmarshal(b, &inter)
	if err != nil {
		return geid.Inter{}, err
	}
	return inter, nil
}

func (c *EidClient) Collect(ctx context.Context, req *geid.Inter) (geid.Resp, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return geid.Resp{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/collect")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return geid.Resp{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return geid.Resp{}, err
	}
	var geidResp geid.Resp
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return geid.Resp{}, err
	}
	err = json.Unmarshal(b, &geidResp)
	if err != nil {
		return geid.Resp{}, err
	}
	return geidResp, nil
}

func (c *EidClient) Change(ctx context.Context, req *geid.Inter) (geid.Resp, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return geid.Resp{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/change")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return geid.Resp{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return geid.Resp{}, err
	}
	var geidResp geid.Resp
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return geid.Resp{}, err
	}
	err = json.Unmarshal(b, &geidResp)
	if err != nil {
		return geid.Resp{}, err
	}
	return geidResp, nil
}

func (c *EidClient) Peek(ctx context.Context, req *geid.Inter) (geid.Resp, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return geid.Resp{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/peek")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return geid.Resp{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return geid.Resp{}, err
	}
	var geidResp geid.Resp
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return geid.Resp{}, err
	}
	err = json.Unmarshal(b, &geidResp)
	if err != nil {
		return geid.Resp{}, err
	}
	return geidResp, nil
}

func (c *EidClient) Cancel(ctx context.Context, req *geid.Inter) error {
	bs, err := json.Marshal(req)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/cancel")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unsuccessful cancel, status code: %d", resp.StatusCode)
	}
	return nil
}