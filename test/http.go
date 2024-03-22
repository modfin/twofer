package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

func sendHTTPRequest[T any](server, method, endpoint string, cookie *http.Cookie, bodyBytes []byte) (*T, int, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s/%s", server, strings.TrimLeft(endpoint, "/")), bytes.NewBufferString(string(bodyBytes)))
	if err != nil {
		return nil, 0, err
	}
	req.Close = true // Do not reuse connection
	req.Header.Add("Content-Type", "application/json")

	cookies := make([]*http.Cookie, 0, 1)
	cookies = append(cookies, cookie)

	serverUrl, err := url.Parse(server)
	if err != nil {
		return nil, 0, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, 0, err
	}

	client := http.Client{
		Jar: jar,
	}
	client.Jar.SetCookies(serverUrl, cookies)

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request, err: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	var res T
	err = json.Unmarshal(respBody, &res)
	return &res, resp.StatusCode, err
}
