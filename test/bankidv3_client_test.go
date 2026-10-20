package test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/modfin/twofer/api"
	"github.com/modfin/twofer/bankidclient"
	"github.com/modfin/twofer/bankidtest"
)

const clientTestIP = "127.0.0.1"

func (s *IntegrationTestSuite) client() (*bankidclient.BankIdV3Client, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	s.T().Cleanup(cancel)
	return bankidclient.NewBankIdV3Client(s.twoferURL), ctx
}

func clientAuthRequest(pnr string) api.BankIdv6AuthSignRequestV3 {
	return api.BankIdv6AuthSignRequestV3{
		EndUserIp:        clientTestIP,
		PersonalNumber:   pnr,
		OrderTokenExpire: time.Minute,
	}
}

func expectedQR(o bankidtest.Order, i int) string {
	mac := hmac.New(sha256.New, []byte(o.QrStartSecret))
	mac.Write([]byte(strconv.Itoa(i)))
	return "bankid." + o.QrStartToken + "." + strconv.Itoa(i) + "." + hex.EncodeToString(mac.Sum(nil))
}

func (s *IntegrationTestSuite) assertQRFrames(stream *bankidclient.AuthV3Stream, frames int) {
	var orderRef string
	for i := 0; i < frames; i++ {
		f, err := stream.Next()
		s.Require().NoError(err, "frame #%d", i)
		if i == 0 {
			orderRef = f.OrderRef
		}
		s.Equal(orderRef, f.OrderRef)
		s.NotEmpty(f.OrderToken)

		o, ok := s.bankidv6.Order(f.OrderRef)
		s.Require().True(ok, "no matching order in bankid fake")
		s.Equal(expectedQR(o, i), f.QR)
		s.Equal("bankid:///?autostarttoken="+o.AutoStartToken+"&redirect=null", f.URI)
	}
}

func (s *IntegrationTestSuite) TestClientAuthV3Stream() {
	const pnr = "199001010101"
	c, ctx := s.client()

	stream, err := c.AuthV3(ctx, clientAuthRequest(pnr))
	s.Require().NoError(err)
	defer stream.Close()

	s.assertQRFrames(stream, 2)

	auths := s.bankidv6.Auths(pnr)
	s.Require().Len(auths, 1)
	s.Equal(clientTestIP, auths[0].EndUserIP)
}

func (s *IntegrationTestSuite) TestClientAuthV3StreamNDJSON() {
	const pnr = "199001010102"
	_, ctx := s.client()

	stream, err := bankidclient.NewBankIdV3Client(s.twoferNDJSON.URL).AuthV3(ctx, clientAuthRequest(pnr))
	s.Require().NoError(err)
	defer stream.Close()

	s.assertQRFrames(stream, 2)
}

func (s *IntegrationTestSuite) TestClientAuthV3Once() {
	const pnr = "199001010103"
	c, ctx := s.client()

	ar, err := c.AuthV3Once(ctx, clientAuthRequest(pnr))
	s.Require().NoError(err)
	s.NotEmpty(ar.OrderToken)

	o, ok := s.bankidv6.Order(ar.OrderRef)
	s.Require().True(ok, "no matching order in bankid fake")
	s.Equal(pnr, o.PersonalNumber)
	s.Equal(expectedQR(o, 0), ar.QR)
	s.Len(s.bankidv6.Auths(pnr), 1)
}

func (s *IntegrationTestSuite) TestClientCollectV3RoundTrip() {
	const pnr = "199001010104"
	c, ctx := s.client()

	ar, err := c.AuthV3Once(ctx, clientAuthRequest(pnr))
	s.Require().NoError(err)

	cr, err := c.CollectV3(ctx, api.BankIdv6CollectRequestV3{OrderToken: ar.OrderToken, EndUserIp: clientTestIP})
	s.Require().NoError(err)
	s.Equal(ar.OrderRef, cr.OrderRef)
	s.Equal(api.StatusPending, cr.Status)
	s.Equal("outstandingTransaction", cr.HintCode)
	s.Nil(cr.CompletionData)

	s.Require().NoError(s.bankidv6.Complete(pnr))
	cr, err = c.CollectV3(ctx, api.BankIdv6CollectRequestV3{OrderToken: ar.OrderToken, EndUserIp: clientTestIP, WaitForChange: true})
	s.Require().NoError(err)
	s.Equal(ar.OrderRef, cr.OrderRef)
	s.Equal(api.StatusComplete, cr.Status)
	s.Require().NotNil(cr.CompletionData)
	s.Equal(pnr, cr.CompletionData.User.PersonalNumber)
	s.Equal(clientTestIP, cr.CompletionData.Device.IpAddress)
}

func (s *IntegrationTestSuite) TestClientCancelV3() {
	const pnr = "199001010105"
	c, ctx := s.client()

	ar, err := c.AuthV3Once(ctx, clientAuthRequest(pnr))
	s.Require().NoError(err)

	err = c.CancelV3(ctx, api.BankIdv6CancelRequestV3{OrderToken: ar.OrderToken, EndUserIp: clientTestIP})
	s.Require().NoError(err)
	o, ok := s.bankidv6.Order(ar.OrderRef)
	s.Require().True(ok)
	s.True(o.Cancelled)

	_, err = c.CollectV3(ctx, api.BankIdv6CollectRequestV3{OrderToken: ar.OrderToken, EndUserIp: clientTestIP})
	var be *bankidclient.BankIdError
	s.Require().ErrorAs(err, &be)
	s.Equal("collectV3", be.Endpoint)
	s.Equal(http.StatusBadRequest, be.HTTPStatus)
	s.Equal(api.ErrorOriginBankIDv6, be.Origin)
	s.Equal(http.StatusBadRequest, be.StatusCode)
	s.Equal("invalidParameters", be.Code)
	s.Equal("No such order", be.Detail)
}

func (s *IntegrationTestSuite) TestClientAuthV3AlreadyInProgress() {
	const pnr = "199001010106"
	c, ctx := s.client()

	_, err := c.AuthV3Once(ctx, clientAuthRequest(pnr))
	s.Require().NoError(err)

	_, err = c.AuthV3Once(ctx, clientAuthRequest(pnr))
	var be *bankidclient.BankIdError
	s.Require().ErrorAs(err, &be)
	s.Equal(api.ErrorOriginBankIDv6, be.Origin)
	s.Equal("alreadyInProgress", be.Code)

	// BankID aborted the first order too, so a retry goes through
	_, err = c.AuthV3Once(ctx, clientAuthRequest(pnr))
	s.NoError(err)
	s.Len(s.bankidv6.Auths(pnr), 3)
}

func (s *IntegrationTestSuite) TestClientOrderTokenBoundToIP() {
	const pnr = "199001010107"
	const otherIP = "198.51.100.7"
	c, ctx := s.client()

	ar, err := c.AuthV3Once(ctx, clientAuthRequest(pnr))
	s.Require().NoError(err)

	_, err = c.CollectV3(ctx, api.BankIdv6CollectRequestV3{OrderToken: ar.OrderToken, EndUserIp: otherIP})
	var be *bankidclient.BankIdError
	s.Require().ErrorAs(err, &be)
	s.Equal(http.StatusBadRequest, be.HTTPStatus)
	s.Equal(api.ErrorOriginTwofer, be.Origin)
	s.Equal("order ip mismatch", be.Code)
	s.Equal("order token ip mismatch with request ip", be.Detail)

	err = c.CancelV3(ctx, api.BankIdv6CancelRequestV3{OrderToken: ar.OrderToken, EndUserIp: otherIP})
	s.Require().ErrorAs(err, &be)
	s.Equal("cancelV3", be.Endpoint)
	s.Equal("order token ip mismatch with request ip", be.Detail)
}

func (s *IntegrationTestSuite) TestClientSameDeviceChecksDeviceIP() {
	tests := []struct {
		name     string
		pnr      string
		deviceIP string
		other    bool
	}{
		{name: "signed on this device", pnr: "199001010108", deviceIP: clientTestIP},
		{name: "signed from another ip", pnr: "199001010109", deviceIP: "198.51.100.9", other: true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			c, ctx := s.client()
			req := clientAuthRequest(tt.pnr)
			req.SameDevice = true
			ar, err := c.AuthV3Once(ctx, req)
			s.Require().NoError(err)

			s.Require().NoError(s.bankidv6.Complete(tt.pnr, bankidtest.FromIP(tt.deviceIP)))
			cr, err := c.CollectV3(ctx, api.BankIdv6CollectRequestV3{OrderToken: ar.OrderToken, EndUserIp: clientTestIP})
			if !tt.other {
				s.Require().NoError(err)
				s.Equal(api.StatusComplete, cr.Status)
				return
			}
			var be *bankidclient.BankIdError
			s.Require().ErrorAs(err, &be)
			s.Equal(api.ErrorOriginTwofer, be.Origin)
			s.Equal("order token ip mismatch with device ip", be.Detail)
		})
	}
}

func (s *IntegrationTestSuite) TestClientErrorBodyIsCapped() {
	body := strings.Repeat("x", 10000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := bankidclient.NewBankIdV3Client(srv.URL).AuthV3Once(ctx, clientAuthRequest("199001010110"))
	var be *bankidclient.BankIdError
	s.Require().True(errors.As(err, &be), "%v", err)
	s.Equal(http.StatusBadGateway, be.HTTPStatus)
	s.Equal(body[:4096], be.Body)
	s.Empty(be.Origin)
}

// twofer ends the stream after ~30s, too long to wait for against the real handlers.
func (s *IntegrationTestSuite) TestClientAuthV3StreamEnds() {
	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{name: "sse", contentType: "text/event-stream", body: "id: 0\ndata: {\"orderRef\":\"a\"}\n\nid: 1\ndata: {\"orderRef\":\"b\"}\n\n"},
		{name: "ndjson", contentType: "application/x-json-stream", body: "{\"orderRef\":\"a\"}\n{\"orderRef\":\"b\"}\n"},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			stream, err := bankidclient.NewBankIdV3Client(srv.URL).AuthV3(ctx, clientAuthRequest("199001010111"))
			s.Require().NoError(err)
			defer stream.Close()

			for _, want := range []string{"a", "b"} {
				f, err := stream.Next()
				s.Require().NoError(err)
				s.Equal(want, f.OrderRef)
			}
			_, err = stream.Next()
			s.ErrorIs(err, io.EOF)
		})
	}
}
