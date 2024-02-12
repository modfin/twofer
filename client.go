package twofer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/modfin/twofer/grpc/geid"
	"github.com/modfin/twofer/grpc/gotp"
	"github.com/modfin/twofer/grpc/gpwd"
	"github.com/modfin/twofer/grpc/gqr"
	"github.com/modfin/twofer/qr"
	"net/http"
)

type Client struct {
	EID *EidClient
	Pwd *PwdClient
	Otp *OtpClient
	Qr  *QrClient
}

func NewClient(baseurl string) Client {
	return Client{
		EID: NewEidClient(baseurl),
		Pwd: NewPwdClient(baseurl),
		Otp: NewOtpClient(baseurl),
		Qr:  NewQrClient(baseurl),
	}
}

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
	b, err := io.ReadAll(resp.Body)
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
	b, err := io.ReadAll(resp.Body)
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
	b, err := io.ReadAll(resp.Body)
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
	b, err := io.ReadAll(resp.Body)
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
	b, err := io.ReadAll(resp.Body)
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
	b, err := io.ReadAll(resp.Body)
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

type OtpClient struct {
	c       *http.Client
	baseUrl string
}

func NewOtpClient(baseurl string) *OtpClient {
	return &OtpClient{
		c:       http.DefaultClient,
		baseUrl: baseurl,
	}
}

func (c *OtpClient) Enroll(ctx context.Context, req *gotp.Enrollment) (gotp.EnrollmentResponse, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return gotp.EnrollmentResponse{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/otp/enroll")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return gotp.EnrollmentResponse{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return gotp.EnrollmentResponse{}, err
	}
	var userEnrollmentResponse gotp.EnrollmentResponse
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return gotp.EnrollmentResponse{}, err
	}
	err = json.Unmarshal(b, &userEnrollmentResponse)
	if err != nil {
		return gotp.EnrollmentResponse{}, err
	}
	return userEnrollmentResponse, nil
}

func (c *OtpClient) Auth(ctx context.Context, req *gotp.Credentials) (gotp.AuthResponse, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return gotp.AuthResponse{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/otp/auth")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return gotp.AuthResponse{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return gotp.AuthResponse{}, err
	}
	var userAuthResponse gotp.AuthResponse
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return gotp.AuthResponse{}, err
	}
	err = json.Unmarshal(b, &userAuthResponse)
	if err != nil {
		return gotp.AuthResponse{}, err
	}
	return userAuthResponse, nil
}

func (c *OtpClient) GetQRImage(ctx context.Context, req *gotp.Credentials) (gqr.Image, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return gqr.Image{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/otp/qr")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return gqr.Image{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return gqr.Image{}, err
	}
	var qrImage gqr.Image
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return gqr.Image{}, err
	}
	err = json.Unmarshal(b, &qrImage)
	if err != nil {
		return gqr.Image{}, err
	}
	return qrImage, nil
}

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
	b, err := io.ReadAll(resp.Body)
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
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return gpwd.Res{}, err
	}
	err = json.Unmarshal(b, &userRes)
	if err != nil {
		return gpwd.Res{}, err
	}
	return userRes, nil
}

type QrClient struct {
	c       *http.Client
	baseUrl string
}

func NewQrClient(baseurl string) *QrClient {
	return &QrClient{
		c:       http.DefaultClient,
		baseUrl: baseurl,
	}
}

func (c *QrClient) GetQrData(ctx context.Context, uriBody string) (qr.QRData, error) {
	buf := bytes.NewBuffer([]byte(uriBody))

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/qr")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return qr.QRData{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return qr.QRData{}, err
	}
	var qrData qr.QRData
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return qr.QRData{}, err
	}
	err = json.Unmarshal(b, &qrData)
	if err != nil {
		return qr.QRData{}, err
	}
	return qrData, nil
}
