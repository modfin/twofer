package test

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/modfin/twofer/internal/bankid"
	"github.com/modfin/twofer/internal/sse"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (s *IntegrationTestSuite) TestAuthCallOnce() {
	authRequest := &bankid.AuthSignRequest{
		EndUserIp: "127.0.0.1",
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(authRequest)
	if err != nil {
		s.NoError(err, "error reading auth request into buffer")
	}

	resp, err := http.Post(s.twoferURL+"/bankid/v6/auth?type=once", "application/json", &buf)
	if err != nil {
		s.NoError(err, "error sending auth request")
	}

	if resp.StatusCode != http.StatusOK {
		s.FailNow("Received invalid status code from auth endpoint", strconv.Itoa(resp.StatusCode), s.twoferURL+"/bankid/v6/auth")
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		s.NoError(err, "error reading auth request response")
	}

	var res bankid.AuthSignAPIResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		s.NoError(err, "error unmarshaling auth response")
	}

	s.True(res.OrderRef != "")

	truth, ok := s.bankidv6.Orders[res.OrderRef]
	s.True(ok, "no matching order in bankid fake")

	mac := hmac.New(sha256.New, []byte(truth.QrStartSecret))
	mac.Write([]byte(strconv.Itoa(0)))
	qrAuthCode := mac.Sum(nil)

	authCode := hex.EncodeToString(qrAuthCode)
	qr := "bankid." + truth.QrStartToken + ".0." + authCode
	s.Equal(qr, res.QR)

	s.Equal("bankid:///?autostarttoken="+truth.AutoStartToken+"&redirect=null", res.URI)
}

func (s *IntegrationTestSuite) TestAuthCall() {
	authRequest := &bankid.AuthSignRequest{
		EndUserIp: "127.0.0.1",
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(authRequest)
	if err != nil {
		s.NoError(err, "error reading auth request into buffer")
	}

	resp, err := http.Post(s.twoferURL+"/bankid/v6/auth", "application/json", &buf)
	if err != nil {
		s.NoError(err, "error sending auth request")
	}

	// Default to splitting on each line
	scanner := bufio.NewScanner(resp.Body)
	defer resp.Body.Close()
	for i := 0; i < 2; i++ { // Server should run for 30 seconds, but let's not wait that long for tests. If pattern holds for 2 messages it should be fine
		msg := sse.Event{}

		scanRes := scanner.Scan()
		if !scanRes {
			s.FailNow("Server-Sent Events message has no content")
		}

		idLine := scanner.Text()
		idSplit := strings.Split(idLine, ": ")
		s.Equal("id", idSplit[0])
		s.Equal(strconv.Itoa(i), idSplit[1])

		msg.Id = idSplit[1]

		scanRes = scanner.Scan()
		if !scanRes {
			s.FailNow("Server-Sent Events message stopped after id")
		}

		eventLine := scanner.Text()
		eventSplit := strings.Split(eventLine, ": ")
		s.Equal("event", eventSplit[0])
		s.Equal("message", eventSplit[1])

		msg.Event = eventSplit[1]

		scanRes = scanner.Scan()
		if !scanRes {
			s.FailNow("Server-Sent Events message stopped after event")
		}

		dataLine := scanner.Text()
		dataSplit := strings.Split(dataLine, ": ")
		s.Equal("data", dataSplit[0])

		msg.Data = dataSplit[1]

		// SSE messages ends in extra empty line, so scan once more
		scanRes = scanner.Scan()
		if !scanRes {
			s.FailNow("Server-Sent Events message stopped after data")
		}

		var res bankid.AuthSignAPIResponse
		err = json.Unmarshal([]byte(msg.Data), &res)
		if err != nil {
			s.NoError(err, "error unmarshaling SSE data for msg")
		}

		s.True(res.OrderRef != "")

		truth, ok := s.bankidv6.Orders[res.OrderRef]
		s.True(ok, "no matching order in bankid fake")

		mac := hmac.New(sha256.New, []byte(truth.QrStartSecret))
		mac.Write([]byte(strconv.Itoa(i)))
		qrAuthCode := mac.Sum(nil)

		authCode := hex.EncodeToString(qrAuthCode)
		qr := "bankid." + truth.QrStartToken + "." + strconv.Itoa(i) + "." + authCode
		s.Equal(qr, res.QR)

		s.Equal("bankid:///?autostarttoken="+truth.AutoStartToken+"&redirect=null", res.URI)
	}
}

func (s *IntegrationTestSuite) TestCollect() {
	authRequest := &bankid.AuthSignRequest{
		EndUserIp: "127.0.0.1",
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(authRequest)
	if err != nil {
		s.NoError(err, "error reading auth request into buffer")
	}

	resp, err := http.Post(s.twoferURL+"/bankid/v6/auth?type=once", "application/json", &buf)
	if err != nil {
		s.NoError(err, "error sending auth request")
	}

	if resp.StatusCode != http.StatusOK {
		s.FailNow("Received invalid status code from auth endpoint", strconv.Itoa(resp.StatusCode), s.twoferURL+"/bankid/v6/auth")
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		s.NoError(err, "error reading auth request response")
	}

	var res bankid.AuthSignAPIResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		s.NoError(err, "error unmarshaling auth response")
	}

	if _, ok := s.bankidv6.Orders[res.OrderRef]; !ok {
		s.FailNow("Order ref could not be found in fake bankid current orders map")
	}

	// Send collect request for the auth request
	collectRequest := &bankid.CollectRequest{
		OrderRef: res.OrderRef,
	}

	var collectBuf bytes.Buffer
	err = json.NewEncoder(&collectBuf).Encode(collectRequest)
	if err != nil {
		s.NoError(err, "error reading collect request into buffer")
	}

	collectResp, err := http.Post(s.twoferURL+"/bankid/v6/collect", "application/json", &collectBuf)
	if err != nil {
		s.NoError(err, "error sending collect request")
	}

	if collectResp.StatusCode != http.StatusOK {
		s.FailNow("Received invalid status code from auth endpoint", strconv.Itoa(collectResp.StatusCode), s.twoferURL+"/bankid/v6/cancel")
	}

	collectBody, err := io.ReadAll(collectResp.Body)
	defer collectResp.Body.Close()
	if err != nil {
		s.NoError(err, "error reading collect request response")
	}

	var collectRes bankid.CollectResponse
	err = json.Unmarshal(collectBody, &collectRes)
	if err != nil {
		s.NoError(err, "error unmarshaling auth response")
	}

	o, ok := s.bankidv6.Orders[collectRes.OrderRef]

	if !ok {
		s.FailNow("Order no longer in fake bankid orders map")
	}

	s.Equal("pending", string(o.Status))
	s.Equal("outstandingTransaction", o.HintCode)
}

func (s *IntegrationTestSuite) TestCancel() {
	authRequest := &bankid.AuthSignRequest{
		EndUserIp: "127.0.0.1",
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(authRequest)
	if err != nil {
		s.NoError(err, "error reading auth request into buffer")
	}

	resp, err := http.Post(s.twoferURL+"/bankid/v6/auth?type=once", "application/json", &buf)
	if err != nil {
		s.NoError(err, "error sending auth request")
	}

	if resp.StatusCode != http.StatusOK {
		s.FailNow("Received invalid status code from auth endpoint", strconv.Itoa(resp.StatusCode), s.twoferURL+"/bankid/v6/auth")
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		s.NoError(err, "error reading auth request response")
	}

	var res bankid.AuthSignAPIResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		s.NoError(err, "error unmarshaling auth response")
	}

	if _, ok := s.bankidv6.Orders[res.OrderRef]; !ok {
		s.FailNow("Order ref could not be found in fake bankid current orders map")
	}

	// Send cancel request for the auth request
	cancelRequest := &bankid.CancelRequest{
		OrderRef: res.OrderRef,
	}

	var cancelBuf bytes.Buffer
	err = json.NewEncoder(&cancelBuf).Encode(cancelRequest)
	if err != nil {
		s.NoError(err, "error reading cancel request into buffer")
	}

	resp, err = http.Post(s.twoferURL+"/bankid/v6/cancel", "application/json", &cancelBuf)
	if err != nil {
		s.NoError(err, "error sending cancel request")
	}

	if resp.StatusCode != http.StatusNoContent {
		s.FailNow("Received invalid status code from auth endpoint", strconv.Itoa(resp.StatusCode), s.twoferURL+"/bankid/v6/cancel")
	}

	o, ok := s.bankidv6.Orders[res.OrderRef]

	if !ok {
		s.FailNow("Order no longer in fake bankid orders map")
	}

	s.Equal("failed", string(o.Status))
	s.Equal("userCancel", o.HintCode)
}

func (s *IntegrationTestSuite) TestSignCallOnce() {
	authRequest := &bankid.AuthSignRequest{
		EndUserIp:       "127.0.0.1",
		UserVisibleData: "Text shown to user",
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(authRequest)
	if err != nil {
		s.NoError(err, "error reading sign request into buffer")
	}

	resp, err := http.Post(s.twoferURL+"/bankid/v6/sign?type=once", "application/json", &buf)
	if err != nil {
		s.NoError(err, "error sending sign request")
	}

	if resp.StatusCode != http.StatusOK {
		s.FailNow("Received invalid status code from sign endpoint", strconv.Itoa(resp.StatusCode), s.twoferURL+"/bankid/v6/sign")
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		s.NoError(err, "error reading sign request response")
	}

	var res bankid.AuthSignAPIResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		s.NoError(err, "error unmarshaling sign response")
	}

	s.True(res.OrderRef != "")

	truth, ok := s.bankidv6.Orders[res.OrderRef]
	s.True(ok, "no matching order in bankid fake")

	mac := hmac.New(sha256.New, []byte(truth.QrStartSecret))
	mac.Write([]byte(strconv.Itoa(0)))
	qrAuthCode := mac.Sum(nil)

	signCode := hex.EncodeToString(qrAuthCode)
	qr := "bankid." + truth.QrStartToken + ".0." + signCode
	s.Equal(qr, res.QR)

	s.Equal("bankid:///?autostarttoken="+truth.AutoStartToken+"&redirect=null", res.URI)
}

func (s *IntegrationTestSuite) TestSignCall() {
	signRequest := &bankid.AuthSignRequest{
		EndUserIp:       "127.0.0.1",
		UserVisibleData: "Text shown to user",
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(signRequest)
	if err != nil {
		s.NoError(err, "error reading sign request into buffer")
	}

	resp, err := http.Post(s.twoferURL+"/bankid/v6/sign", "application/json", &buf)
	if err != nil {
		s.NoError(err, "error sending sign request")
	}

	// Default to splitting on each line
	scanner := bufio.NewScanner(resp.Body)
	defer resp.Body.Close()
	for i := 0; i < 2; i++ { // Server should run for 30 seconds, but let's not wait that long for tests. If pattern holds for 2 messages it should be fine
		msg := sse.Event{}

		scanRes := scanner.Scan()
		if !scanRes {
			s.FailNow("Server-Sent Events message has no content")
		}

		idLine := scanner.Text()
		idSplit := strings.Split(idLine, ": ")
		s.Equal("id", idSplit[0])
		s.Equal(strconv.Itoa(i), idSplit[1])

		msg.Id = idSplit[1]

		scanRes = scanner.Scan()
		if !scanRes {
			s.FailNow("Server-Sent Events message stopped after id")
		}

		eventLine := scanner.Text()
		eventSplit := strings.Split(eventLine, ": ")
		s.Equal("event", eventSplit[0])
		s.Equal("message", eventSplit[1])

		msg.Event = eventSplit[1]

		scanRes = scanner.Scan()
		if !scanRes {
			s.FailNow("Server-Sent Events message stopped after event")
		}

		dataLine := scanner.Text()
		dataSplit := strings.Split(dataLine, ": ")
		s.Equal("data", dataSplit[0])

		msg.Data = dataSplit[1]

		// SSE messages ends in extra empty line, so scan once more
		scanRes = scanner.Scan()
		if !scanRes {
			s.FailNow("Server-Sent Events message stopped after data")
		}

		var res bankid.AuthSignAPIResponse
		err = json.Unmarshal([]byte(msg.Data), &res)
		if err != nil {
			s.NoError(err, "error unmarshaling SSE data for msg")
		}

		s.True(res.OrderRef != "")

		truth, ok := s.bankidv6.Orders[res.OrderRef]
		s.True(ok, "no matching order in bankid fake")

		mac := hmac.New(sha256.New, []byte(truth.QrStartSecret))
		mac.Write([]byte(strconv.Itoa(i)))
		qrAuthCode := mac.Sum(nil)

		signCode := hex.EncodeToString(qrAuthCode)
		qr := "bankid." + truth.QrStartToken + "." + strconv.Itoa(i) + "." + signCode
		s.Equal(qr, res.QR)

		s.Equal("bankid:///?autostarttoken="+truth.AutoStartToken+"&redirect=null", res.URI)
	}
}

func (s *IntegrationTestSuite) TestChange() {
	authRequest := &bankid.AuthSignRequest{
		EndUserIp: "127.0.0.1",
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(authRequest)
	if err != nil {
		s.NoError(err, "error reading auth request into buffer")
	}

	resp, err := http.Post(s.twoferURL+"/bankid/v6/auth?type=once", "application/json", &buf)
	if err != nil {
		s.NoError(err, "error sending auth request")
	}

	if resp.StatusCode != http.StatusOK {
		s.FailNow("Received invalid status code from auth endpoint", strconv.Itoa(resp.StatusCode), s.twoferURL+"/bankid/v6/auth")
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		s.NoError(err, "error reading auth request response")
	}

	var res bankid.AuthSignAPIResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		s.NoError(err, "error unmarshaling auth response")
	}

	if _, ok := s.bankidv6.Orders[res.OrderRef]; !ok {
		s.FailNow("Order ref could not be found in fake bankid current orders map")
	}

	// Start change request while we wait for cancel
	go func() {
		changeRequest := &bankid.ChangeRequest{
			OrderRef: res.OrderRef,
		}

		var changeBuf bytes.Buffer
		err = json.NewEncoder(&changeBuf).Encode(changeRequest)
		if err != nil {
			s.NoError(err, "error reading cancel request into buffer")
		}

		resp, err = http.Post(s.twoferURL+"/bankid/v6/change", "application/json", &changeBuf)
		if err != nil {
			s.NoError(err, "error sending change request")
		}

		if resp.StatusCode != http.StatusOK {
			s.FailNow("Received invalid status code from change endpoint", strconv.Itoa(resp.StatusCode), s.twoferURL+"/bankid/v6/cancel")
		}

		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			s.NoError(err, "error reading change request response")
		}

		var changeRes bankid.CollectResponse
		err = json.Unmarshal(body, &changeRes)
		if err != nil {
			s.NoError(err, "error unmarshaling change response")
		}

		s.Equal(changeRes.OrderRef, res.OrderRef)
		s.Equal("failed", string(changeRes.Status))       // Because of cancel
		s.Equal("userCancel", string(changeRes.HintCode)) // Because of cancel
	}()

	// Send cancel request for the auth request
	cancelRequest := &bankid.CancelRequest{
		OrderRef: res.OrderRef,
	}

	var cancelBuf bytes.Buffer
	err = json.NewEncoder(&cancelBuf).Encode(cancelRequest)
	if err != nil {
		s.NoError(err, "error reading cancel request into buffer")
	}

	resp, err = http.Post(s.twoferURL+"/bankid/v6/cancel", "application/json", &cancelBuf)
	if err != nil {
		s.NoError(err, "error sending cancel request")
	}
	if resp.StatusCode != http.StatusNoContent {
		s.FailNow("Received invalid status code from auth endpoint", strconv.Itoa(resp.StatusCode), s.twoferURL+"/bankid/v6/cancel")
	}

	o, ok := s.bankidv6.Orders[res.OrderRef]

	if !ok {
		s.FailNow("Order no longer in fake bankid orders map")
	}

	s.Equal("failed", string(o.Status))
	s.Equal("userCancel", o.HintCode)
}
