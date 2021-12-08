package httpclients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"twofer/qr"
)

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
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return qr.QRData{}, err
	}
	err = json.Unmarshal(b, &qrData)
	if err != nil {
		return qr.QRData{}, err
	}
	return qrData, nil
}