package twofer

import (
	"twofer/httpclients"
)

type Client struct {
	EID *httpclients.EidClient
	Pwd *httpclients.PwdClient
	Otp *httpclients.OtpClient
	Qr  *httpclients.QrClient
}

func NewClient(baseurl string) Client {
	return Client{
		EID: httpclients.NewEidClient(baseurl),
		Pwd: httpclients.NewPwdClient(baseurl),
		Otp: httpclients.NewOtpClient(baseurl),
		Qr:  httpclients.NewQrClient(baseurl),
	}
}
