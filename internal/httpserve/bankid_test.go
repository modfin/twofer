package httpserve

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/modfin/twofer/api"
	"github.com/modfin/twofer/internal/bankid"
	"github.com/modfin/twofer/stream/ndjson"
	"github.com/modfin/twofer/stream/sse"
)

const (
	qrTestPeriod = time.Millisecond * 100
	testIP       = "127.0.0.1"

	// Test ID's taken from example in /auth API documentation
	testAuthOrderRef       = "131daac9-16c6-4618-beb0-365768f37288"
	testAuthAutoStartToken = "7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6"
	testAuthQrStartToken   = "67df3917-fa0d-44e5-b327-edcc928297f8"
	testAuthQrStartSecret  = "d28db9a7-4cde-429e-a983-359be676944c"
)

type (
	authSignTestFn  func(*testing.T) authSignFn
	watchTestFn     func(*testing.T) watchFn
	newContextFn    func(r *http.Request, w http.ResponseWriter) echo.Context
	responseCheckFn func(*testing.T, context.Context, io.ReadCloser, authSignTest)
	authSignTest    struct {
		name           string
		encoder        NewStreamEncoder
		testDecoder    responseCheckFn
		request        bankid.AuthSignRequest
		params         *url.Values
		authSign       authSignTestFn
		watch          watchTestFn
		wantHTTPStatus int
		wantEvents     []sse.Event
		wantResponses  []api.BankIdV6Response
	}
)

var (
	authResponseOK = bankid.AuthSignResponse{
		OrderRef:       testAuthOrderRef,
		AutoStartToken: testAuthAutoStartToken,
		QrStartToken:   testAuthQrStartToken,
		QrStartSecret:  testAuthQrStartSecret,
	}

	pendingOutstandingTransaction = bankid.CollectResponse{OrderRef: testAuthOrderRef, Status: bankid.Pending, HintCode: bankid.OutstandingTransaction}
	pendingUserSign               = bankid.CollectResponse{OrderRef: testAuthOrderRef, Status: bankid.Pending, HintCode: bankid.UserSign}
	completeOK                    = bankid.CollectResponse{
		OrderRef: testAuthOrderRef,
		Status:   bankid.Complete,
		// Data below taken from /collect API example
		CompletionData: bankid.CompletionData{
			User: bankid.User{
				PersonalNumber: "190000000000",
				Name:           "Karl Karlsson",
				GivenName:      "Karl",
				SurName:        "Karlsson",
			},
			Device: bankid.Device{
				IpAddress: testIP,
			},
			BankIdIssueDate: "2020-02-01",
			Signature:       "<base64-encoded data>", // OK since we don't try to decode it
			OcspResponse:    "<base64-encoded data>", // OK since we don't try to decode it
		},
	}
)

func Test_authSign(t *testing.T) {
	e := echo.New()
	for _, tt := range []authSignTest{
		{
			name:        "happy_auth_flow_sse_stream",
			encoder:     sse.NewEncoder,
			testDecoder: sseCheck,
			request:     bankid.AuthSignRequest{EndUserIp: testIP},
			authSign:    authSignTestMock(authResponseOK, nil),
			watch: watchMock([]bankid.CollectResponse{
				pendingOutstandingTransaction,
				pendingUserSign,
				completeOK,
			}, nil),
			wantHTTPStatus: http.StatusOK,
			wantEvents: []sse.Event{
				{Event: "qrcode", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.0.dc69358e712458a66a7525beef148ae8526b1c71610eff2c16cdffb4cdac9bf8"}`},
				{Event: "qrcode", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.1.949d559bf23403952a94d103e67743126381eda00f0b3cbddbf7c96b1adcbce2"}`},
				{Event: "qrcode", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.2.a9e5ec59cb4eee4ef4117150abc58fad7a85439a6a96ccbecc3668b41795b3f3"}`},
				{Event: "status", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","status":"pending","hintCode":"outstandingTransaction"}`},
				{Event: "qrcode", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.3.96077d77699971790b46ee1f04ff1e44fe96b0602c9c51e4ca9c6d031c7c3bb7"}`},
				{Event: "qrcode", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.4.1d9a7e5dd98d08cb393f73c63ce032df0c9433512153ab9fb040b96cd45b1b11"}`},
				{Event: "status", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","status":"pending","hintCode":"userSign"}`},
				{Event: "status", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","status":"complete","completionData":{"user":{"personalNumber":"190000000000","name":"Karl Karlsson","givenName":"Karl","surName":"Karlsson"},"device":{"ipAddress":"127.0.0.1"},"bankIdIssueDate":"2020-02-01","stepUp":{},"signature":"\u003cbase64-encoded data\u003e","ocspResponse":"\u003cbase64-encoded data\u003e"}}`},
			},
		},
		{
			name:        "happy_auth_flow_ndjson_stream",
			encoder:     ndjson.NewEncoder,
			testDecoder: ndjsonCheck,
			request:     bankid.AuthSignRequest{EndUserIp: testIP},
			authSign:    authSignTestMock(authResponseOK, nil),
			watch: watchMock([]bankid.CollectResponse{
				pendingOutstandingTransaction,
				pendingUserSign,
				completeOK,
			}, nil),
			wantHTTPStatus: http.StatusOK,
			wantResponses: []api.BankIdV6Response{
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.0.dc69358e712458a66a7525beef148ae8526b1c71610eff2c16cdffb4cdac9bf8"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.1.949d559bf23403952a94d103e67743126381eda00f0b3cbddbf7c96b1adcbce2"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.2.a9e5ec59cb4eee4ef4117150abc58fad7a85439a6a96ccbecc3668b41795b3f3"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", Status: "pending", HintCode: "outstandingTransaction"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.3.96077d77699971790b46ee1f04ff1e44fe96b0602c9c51e4ca9c6d031c7c3bb7"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.4.1d9a7e5dd98d08cb393f73c63ce032df0c9433512153ab9fb040b96cd45b1b11"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", Status: "pending", HintCode: "userSign"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", Status: "complete", CompletionData: &api.BankIdV6CompletionData{
					User:            api.BankIdV6User{PersonalNumber: "190000000000", Name: "Karl Karlsson", GivenName: "Karl", SurName: "Karlsson"},
					Device:          api.BankIdV6Device{IpAddress: "127.0.0.1"},
					BankIdIssueDate: "2020-02-01",
					Signature:       "\u003cbase64-encoded data\u003e",
					OcspResponse:    "\u003cbase64-encoded data\u003e"}},
			},
		},
	} {
		t.Run(tt.name, testAuthSign(tt, e.NewContext))
	}
}

func testAuthSign(tt authSignTest, newContext newContextFn) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		bodyData, err := json.Marshal(tt.request)
		if err != nil {
			t.Fatalf("ERROR: Failed to marshal request: %v", err)
		}

		u, err := url.Parse("http://test.local/api/someurl")
		if err != nil {
			t.Fatalf("ERROR: Failed to parse URL: %v", err)
		}

		if tt.params != nil {
			u.RawQuery = tt.params.Encode()
		}

		req := httptest.NewRequest("", u.String(), bytes.NewReader(bodyData))
		res := httptest.NewRecorder()
		ctx := newContext(req, res)

		echoHandler := authSign(tt.authSign(t), tt.watch(t), qrTestPeriod, tt.encoder)
		err = echoHandler(ctx)
		if err != nil {
			t.Fatalf("got error: %v", err)
		}

		response := res.Result()
		defer func() { _ = response.Body.Close() }()
		if response.StatusCode != tt.wantHTTPStatus {
			t.Errorf("ERROR: got HTTP status code: %d, want: %d\n", response.StatusCode, tt.wantHTTPStatus)
		}

		tt.testDecoder(t, req.Context(), response.Body, tt)
	}
}

func authSignTestMock(response bankid.AuthSignResponse, testErr error) authSignTestFn {
	return func(t *testing.T) authSignFn {
		return func(ctx context.Context, asr *bankid.AuthSignRequest) (*bankid.AuthSignResponse, error) {
			if asr.EndUserIp != testIP {
				t.Errorf("authFn got EndUserIp: '%s', want: '%s'", asr.EndUserIp, testIP)
			}

			err := asr.ValidateAuthRequest()
			if err != nil {
				return nil, err
			}

			return &response, testErr
		}
	}
}

func watchMock(responses []bankid.CollectResponse, testErr error) watchTestFn {
	return func(t *testing.T) watchFn {
		return func(ctx context.Context, orderRef string) (<-chan bankid.Change, error) {
			if orderRef != testAuthOrderRef {
				t.Errorf("authwatch got orderRef: %s, want: %s", orderRef, testAuthOrderRef)
			}

			collectRequest := &bankid.CollectRequest{OrderRef: orderRef}
			err := collectRequest.Validate()
			if err != nil {
				return nil, err
			}

			if testErr != nil {
				return nil, testErr
			}

			watch := make(chan bankid.Change, 1) // Make it a buffered channel so that we can post initial state before we return

			sendChange := func(change bankid.CollectResponse) {
				select {
				case watch <- bankid.Change{CollectResponse: change}:
				case <-time.After(time.Second):
				}
			}

			go func() {
				defer close(watch)
				time.Sleep(time.Millisecond * 25)
				for _, change := range responses {
					time.Sleep(time.Millisecond * 200)
					sendChange(change)
					if change.Status != bankid.Pending {
						return
					}
				}
			}()

			return watch, nil
		}
	}
}

func ndjsonCheck(t *testing.T, ctx context.Context, body io.ReadCloser, tt authSignTest) {
	responseChan := ndjson.NewReader[api.BankIdV6Response](ctx, body)
	var cnt int
	for res := range responseChan {
		//t.Logf("got response: %v", res)
		if cnt < len(tt.wantEvents) && !reflect.DeepEqual(res, tt.wantResponses[cnt]) {
			t.Errorf(" got event: %v\nwant event: %v", res, tt.wantResponses[cnt])
		}
		cnt++
	}
}

func sseCheck(t *testing.T, ctx context.Context, body io.ReadCloser, tt authSignTest) {
	events := sse.NewReader(ctx, body)
	var cnt int
	for e := range events {
		// t.Logf("%v: got event: %v", time.Since(start), e)
		if cnt < len(tt.wantEvents) && !reflect.DeepEqual(e, tt.wantEvents[cnt]) {
			t.Errorf(" got event: %v\nwant event: %v", e, tt.wantEvents[cnt])
		}
		cnt++
	}
	if cnt != len(tt.wantEvents) {
		t.Errorf("got %d events, want %d", cnt, len(tt.wantEvents))
	}

}
