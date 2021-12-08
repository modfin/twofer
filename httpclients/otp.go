package httpclients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"twofer/grpc/gotp"
)

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
	b, err := ioutil.ReadAll(resp.Body)
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
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return gotp.AuthResponse{}, err
	}
	err = json.Unmarshal(b, &userAuthResponse)
	if err != nil {
		return gotp.AuthResponse{}, err
	}
	return userAuthResponse, nil
}