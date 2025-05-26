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
	authSignTestFn    func(*testing.T) authSignFn
	watchTestFn       func(*testing.T) watchFn
	newContextFn      func(r *http.Request, w http.ResponseWriter) echo.Context
	responseCheckFn   func(*testing.T, context.Context, io.ReadCloser, authSignTest)
	responseCheckFnV3 func(*testing.T, context.Context, io.ReadCloser, authSignTestV3)
	authSignTest      struct {
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
	authSignTestV3 struct {
		name           string
		encoder        NewStreamEncoder
		testDecoder    responseCheckFnV3
		request        api.BankIdv6AuthSignRequestV3
		params         *url.Values
		authSign       authSignTestFn
		watch          watchTestFn
		wantHTTPStatus int
		wantEvents     []sse.Event
		wantResponses  []api.BankIdV6AuthSignResponseV3
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
		if cnt < len(tt.wantResponses) && !reflect.DeepEqual(res, tt.wantResponses[cnt]) {
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

func Test_authSignV3Stream(t *testing.T) {
	e := echo.New()
	for _, tt := range []authSignTestV3{
		{
			name:        "happy_auth_flow_sse_stream",
			encoder:     sse.NewEncoder,
			testDecoder: sseCheckV3,
			request:     api.BankIdv6AuthSignRequestV3{EndUserIp: testIP},
			authSign:    authSignTestMock(authResponseOK, nil),
			watch: watchMock([]bankid.CollectResponse{
				pendingOutstandingTransaction,
				pendingUserSign,
				completeOK,
			}, nil),
			wantHTTPStatus: http.StatusOK,
			wantEvents: []sse.Event{
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.0.dc69358e712458a66a7525beef148ae8526b1c71610eff2c16cdffb4cdac9bf8"}`, ID: "0"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.1.949d559bf23403952a94d103e67743126381eda00f0b3cbddbf7c96b1adcbce2"}`, ID: "1"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.2.a9e5ec59cb4eee4ef4117150abc58fad7a85439a6a96ccbecc3668b41795b3f3"}`, ID: "2"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.3.96077d77699971790b46ee1f04ff1e44fe96b0602c9c51e4ca9c6d031c7c3bb7"}`, ID: "3"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.4.1d9a7e5dd98d08cb393f73c63ce032df0c9433512153ab9fb040b96cd45b1b11"}`, ID: "4"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.5.56a7bb043d51f8c7aa6828689767b412179a727a6d4e9b7e1c15ded30061bd2f"}`, ID: "5"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.6.51e9a2ea531b5ca7334fd8dd050bd592b8d235d6584ea6b251f0eec4d434267b"}`, ID: "6"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.7.e6a7d5c37920aeb22ea554716fde4dcd42665d5d641a41f459cc9cda03472d31"}`, ID: "7"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.8.d4bbdfa6fe217349d4128429c17bb30cb5b1bc79b54128ef0022478951a273d4"}`, ID: "8"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.9.85fc3076b6a1d9ec5ff73014b006035375c8e9dd34bbbaea60c08e931a6eceeb"}`, ID: "9"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.10.2822ca616ce1e64a1c171df69154ebc5adef4011244c867d6ad88a02db178962"}`, ID: "10"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.11.6c7b122b2bca1ce8ff0c6f24ffe52e69c649b2790fe8485d75f38ae05edd256e"}`, ID: "11"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.12.7b2410a2fdbae51a1f3c5c1e223752d3840ea4664444ceb81319f0707d219a3c"}`, ID: "12"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.13.e6a3ed34b582ea3cc4b905fb2f4d3e9e657999cc4d1898ff3f457964a25b899a"}`, ID: "13"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.14.6e97534d44bb9f38d239f707aad95412cec222dd53e57259320e41e990700a0e"}`, ID: "14"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.15.3ebd4094cce9a3f993e86a5f940da8a94b5be8db42e5d5141aa6b60f51c1c2eb"}`, ID: "15"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.16.3de9e9322183ebe9c102b3349565c32ebb5439da9307f0593962974796892599"}`, ID: "16"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.17.03642ea1c1249141a2f5706acf8cabdb474259afab830e4d94fb95a5229bbf6f"}`, ID: "17"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.18.4556427effb48bcb95c0bcb4208df6984a0dd4c360d9ece47e9efc823419333d"}`, ID: "18"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.19.8755fe51cbd0feb281f85f4a4c18a40f1a30d98480c1b3d18ba1fa54a30ce705"}`, ID: "19"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.20.c5f95da618d92fca54d233363f6d89cecfdca389b8dc331d4ea39d1a24ab78b1"}`, ID: "20"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.21.634cc7d85fa5cb3190c58ff3c821a582c71fd4dcbce18e2560e9f83a92d9f2d1"}`, ID: "21"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.22.fdb8c5e87e2680adde2c5903b892b5e979bc4e7ec1aa7274113efd6048614b87"}`, ID: "22"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.23.d5e52872f000c57a7b03cb7306f18641d3498a0943aedce2d9d471973ef50ce1"}`, ID: "23"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.24.4d0ced42119b047b4e2cc5d01cf3f39f1126969c3e0c9b15c09259fd3584446d"}`, ID: "24"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.25.517b8eda4a7a0de9ba6b603a93d1fc1e3b75d943e308fd28cc04d6e4ab5a5487"}`, ID: "25"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.26.2359ce7475a218e1c346e24caf8dc748636fca4dbaba2c2a40813457ee88b949"}`, ID: "26"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.27.438a3e6e4ab4456ba5251d539c449b03d20022a8446729d549e9d0174d497194"}`, ID: "27"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.28.325b0142d18e1b77bb1a7eb694bbc1c7a30391bda91bf25f3e32dd47284f6978"}`, ID: "28"},
				{Event: "message", Data: `{"orderRef":"131daac9-16c6-4618-beb0-365768f37288","uri":"bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null","qr":"bankid.67df3917-fa0d-44e5-b327-edcc928297f8.29.26049f2bc12d5b43ebe8bf701e7725abf015d2d8347aa04c924b5793ee4a196c"}`, ID: "29"},
			},
		},
		{
			name:        "happy_auth_flow_ndjson_stream",
			encoder:     ndjson.NewEncoder,
			testDecoder: ndjsonCheckV3,
			request:     api.BankIdv6AuthSignRequestV3{EndUserIp: testIP},
			authSign:    authSignTestMock(authResponseOK, nil),
			watch: watchMock([]bankid.CollectResponse{
				pendingOutstandingTransaction,
				pendingUserSign,
				completeOK,
			}, nil),
			wantHTTPStatus: http.StatusOK,
			wantResponses: []api.BankIdV6AuthSignResponseV3{
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.0.dc69358e712458a66a7525beef148ae8526b1c71610eff2c16cdffb4cdac9bf8"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.1.949d559bf23403952a94d103e67743126381eda00f0b3cbddbf7c96b1adcbce2"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.2.a9e5ec59cb4eee4ef4117150abc58fad7a85439a6a96ccbecc3668b41795b3f3"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.3.96077d77699971790b46ee1f04ff1e44fe96b0602c9c51e4ca9c6d031c7c3bb7"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.4.1d9a7e5dd98d08cb393f73c63ce032df0c9433512153ab9fb040b96cd45b1b11"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.5.56a7bb043d51f8c7aa6828689767b412179a727a6d4e9b7e1c15ded30061bd2f"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.6.51e9a2ea531b5ca7334fd8dd050bd592b8d235d6584ea6b251f0eec4d434267b"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.7.e6a7d5c37920aeb22ea554716fde4dcd42665d5d641a41f459cc9cda03472d31"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.8.d4bbdfa6fe217349d4128429c17bb30cb5b1bc79b54128ef0022478951a273d4"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.9.85fc3076b6a1d9ec5ff73014b006035375c8e9dd34bbbaea60c08e931a6eceeb"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.10.2822ca616ce1e64a1c171df69154ebc5adef4011244c867d6ad88a02db178962"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.11.6c7b122b2bca1ce8ff0c6f24ffe52e69c649b2790fe8485d75f38ae05edd256e"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.12.7b2410a2fdbae51a1f3c5c1e223752d3840ea4664444ceb81319f0707d219a3c"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.13.e6a3ed34b582ea3cc4b905fb2f4d3e9e657999cc4d1898ff3f457964a25b899a"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.14.6e97534d44bb9f38d239f707aad95412cec222dd53e57259320e41e990700a0e"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.15.3ebd4094cce9a3f993e86a5f940da8a94b5be8db42e5d5141aa6b60f51c1c2eb"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.16.3de9e9322183ebe9c102b3349565c32ebb5439da9307f0593962974796892599"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.17.03642ea1c1249141a2f5706acf8cabdb474259afab830e4d94fb95a5229bbf6f"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.18.4556427effb48bcb95c0bcb4208df6984a0dd4c360d9ece47e9efc823419333d"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.19.8755fe51cbd0feb281f85f4a4c18a40f1a30d98480c1b3d18ba1fa54a30ce705"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.20.c5f95da618d92fca54d233363f6d89cecfdca389b8dc331d4ea39d1a24ab78b1"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.21.634cc7d85fa5cb3190c58ff3c821a582c71fd4dcbce18e2560e9f83a92d9f2d1"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.22.fdb8c5e87e2680adde2c5903b892b5e979bc4e7ec1aa7274113efd6048614b87"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.23.d5e52872f000c57a7b03cb7306f18641d3498a0943aedce2d9d471973ef50ce1"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.24.4d0ced42119b047b4e2cc5d01cf3f39f1126969c3e0c9b15c09259fd3584446d"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.25.517b8eda4a7a0de9ba6b603a93d1fc1e3b75d943e308fd28cc04d6e4ab5a5487"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.26.2359ce7475a218e1c346e24caf8dc748636fca4dbaba2c2a40813457ee88b949"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.27.438a3e6e4ab4456ba5251d539c449b03d20022a8446729d549e9d0174d497194"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.28.325b0142d18e1b77bb1a7eb694bbc1c7a30391bda91bf25f3e32dd47284f6978"},
				{OrderRef: "131daac9-16c6-4618-beb0-365768f37288", URI: "bankid:///?autostarttoken=7c40b5c9-fa74-49cf-b98c-bfe651f9a7c6\u0026redirect=null", QR: "bankid.67df3917-fa0d-44e5-b327-edcc928297f8.29.26049f2bc12d5b43ebe8bf701e7725abf015d2d8347aa04c924b5793ee4a196c"},
			},
		},
	} {
		t.Run(tt.name, testAuthSignV3(tt, e.NewContext))
	}
}

func testAuthSignV3(tt authSignTestV3, newContext newContextFn) func(t *testing.T) {
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

		echoHandler := authSignV3(tt.authSign(t), time.Millisecond, tt.encoder)
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

func ndjsonCheckV3(t *testing.T, ctx context.Context, body io.ReadCloser, tt authSignTestV3) {
	responseChan := ndjson.NewReader[api.BankIdV6AuthSignResponseV3](ctx, body)
	var cnt int
	for res := range responseChan {
		//t.Logf("got response: %v", res)
		if cnt < len(tt.wantResponses) && !reflect.DeepEqual(res, tt.wantResponses[cnt]) {
			t.Errorf(" got event: %v\nwant event: %v", res, tt.wantResponses[cnt])
		}
		cnt++
	}
	if cnt != len(tt.wantResponses) {
		t.Errorf("got %d events, want %d", cnt, len(tt.wantEvents))
	}
}

func sseCheckV3(t *testing.T, ctx context.Context, body io.ReadCloser, tt authSignTestV3) {
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
