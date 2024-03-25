package twofer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/modfin/twofer/internal/serveid"
	"github.com/modfin/twofer/internal/servotp"
	"github.com/modfin/twofer/internal/servpwd"
	"github.com/modfin/twofer/internal/servqr"
	"io"
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

func (c *EidClient) Providers(ctx context.Context) (serveid.Providers, error) {
	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/providers")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return serveid.Providers{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return serveid.Providers{}, err
	}
	var geidProviders serveid.Providers
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return serveid.Providers{}, err
	}
	err = json.Unmarshal(b, &geidProviders)
	if err != nil {
		return serveid.Providers{}, err
	}
	return geidProviders, nil
}

func (c *EidClient) AuthInit(ctx context.Context, req *serveid.Req) (serveid.Inter, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return serveid.Inter{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/auth")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return serveid.Inter{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return serveid.Inter{}, err
	}
	var inter serveid.Inter
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return serveid.Inter{}, err
	}
	err = json.Unmarshal(b, &inter)
	if err != nil {
		return serveid.Inter{}, err
	}
	return inter, nil
}

func (c *EidClient) SignInit(ctx context.Context, req *serveid.Req) (serveid.Inter, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return serveid.Inter{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/sign")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return serveid.Inter{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return serveid.Inter{}, err
	}
	var inter serveid.Inter
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return serveid.Inter{}, err
	}
	err = json.Unmarshal(b, &inter)
	if err != nil {
		return serveid.Inter{}, err
	}
	return inter, nil
}

func (c *EidClient) Collect(ctx context.Context, req *serveid.Inter) (serveid.Resp, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return serveid.Resp{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/collect")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return serveid.Resp{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return serveid.Resp{}, err
	}
	var geidResp serveid.Resp
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return serveid.Resp{}, err
	}
	err = json.Unmarshal(b, &geidResp)
	if err != nil {
		return serveid.Resp{}, err
	}
	return geidResp, nil
}

func (c *EidClient) Change(ctx context.Context, req *serveid.Inter) (serveid.Resp, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return serveid.Resp{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/change")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return serveid.Resp{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return serveid.Resp{}, err
	}
	var geidResp serveid.Resp
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return serveid.Resp{}, err
	}
	err = json.Unmarshal(b, &geidResp)
	if err != nil {
		return serveid.Resp{}, err
	}
	return geidResp, nil
}

func (c *EidClient) Peek(ctx context.Context, req *serveid.Inter) (serveid.Resp, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return serveid.Resp{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/eid/peek")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return serveid.Resp{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return serveid.Resp{}, err
	}
	var geidResp serveid.Resp
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return serveid.Resp{}, err
	}
	err = json.Unmarshal(b, &geidResp)
	if err != nil {
		return serveid.Resp{}, err
	}
	return geidResp, nil
}

func (c *EidClient) Cancel(ctx context.Context, req *serveid.Inter) error {
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

func (c *OtpClient) Enroll(ctx context.Context, req *servotp.Enrollment) (servotp.EnrollmentResponse, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return servotp.EnrollmentResponse{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/otp/enroll")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return servotp.EnrollmentResponse{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return servotp.EnrollmentResponse{}, err
	}
	var userEnrollmentResponse servotp.EnrollmentResponse
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return servotp.EnrollmentResponse{}, err
	}
	err = json.Unmarshal(b, &userEnrollmentResponse)
	if err != nil {
		return servotp.EnrollmentResponse{}, err
	}
	return userEnrollmentResponse, nil
}

func (c *OtpClient) Auth(ctx context.Context, req *servotp.Credentials) (servotp.AuthResponse, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return servotp.AuthResponse{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/otp/auth")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return servotp.AuthResponse{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return servotp.AuthResponse{}, err
	}
	var userAuthResponse servotp.AuthResponse
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return servotp.AuthResponse{}, err
	}
	err = json.Unmarshal(b, &userAuthResponse)
	if err != nil {
		return servotp.AuthResponse{}, err
	}
	return userAuthResponse, nil
}

func (c *OtpClient) GetQRImage(ctx context.Context, req *servotp.Credentials) (servqr.Image, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return servqr.Image{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/otp/qr")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return servqr.Image{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return servqr.Image{}, err
	}
	var qrImage servqr.Image
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return servqr.Image{}, err
	}
	err = json.Unmarshal(b, &qrImage)
	if err != nil {
		return servqr.Image{}, err
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

func (c *PwdClient) Enroll(ctx context.Context, req *servpwd.EnrollReq) (servpwd.Blob, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return servpwd.Blob{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/pwd/enroll")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return servpwd.Blob{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return servpwd.Blob{}, err
	}
	var userBlob servpwd.Blob
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return servpwd.Blob{}, err
	}
	err = json.Unmarshal(b, &userBlob)
	if err != nil {
		return servpwd.Blob{}, err
	}
	return userBlob, nil
}

func (c *PwdClient) Auth(ctx context.Context, req *servpwd.AuthReq) (servpwd.Res, error) {
	bs, err := json.Marshal(req)
	if err != nil {
		return servpwd.Res{}, err
	}
	buf := bytes.NewBuffer(bs)

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/pwd/auth")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return servpwd.Res{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return servpwd.Res{}, err
	}
	var userRes servpwd.Res
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return servpwd.Res{}, err
	}
	err = json.Unmarshal(b, &userRes)
	if err != nil {
		return servpwd.Res{}, err
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

func (c *QrClient) GetQrData(ctx context.Context, uriBody string) (servqr.QRData, error) {
	buf := bytes.NewBuffer([]byte(uriBody))

	u := fmt.Sprintf("%s/%s", c.baseUrl, "v1/qr")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, buf)
	if err != nil {
		return servqr.QRData{}, err
	}
	resp, err := c.c.Do(hreq)
	if err != nil {
		return servqr.QRData{}, err
	}
	var qrData servqr.QRData
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return servqr.QRData{}, err
	}
	err = json.Unmarshal(b, &qrData)
	if err != nil {
		return servqr.QRData{}, err
	}
	return qrData, nil
}
