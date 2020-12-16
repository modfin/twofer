package freja

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"twofer/eid/freja/frejam"
)

type API struct {
	parent *Client
}

//// Authentication

func (a *API) AuthInitRequest(ctx context.Context, authReq frejam.AuthRequest) (authRef string, err error) {

	strreq, err := authReq.Marshal()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", a.parent.baseURL+initAuthURL, bytes.NewBuffer([]byte(strreq)))
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "text")
	resp, err := a.parent.client.Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		e := frejam.FrejaError{}
		err = json.Unmarshal(b, &e)
		if err == nil {
			err = e
		}
		return "", err
	}

	var ref frejam.AuthRef
	err = json.Unmarshal(b, &ref)
	if err != nil {
		return "", err
	}

	return ref.AuthRef, nil
}

// includePrevious: string, mandatory. Must be equal to ALL. Indicates that the complete list of authentications
// successfully initiated by the relying party within the last 10 minutes will be returned, including results for
// previously completed authentication results that have been reported through an earlier call to one of the methods
// for getting authentication results.
func (a *API) AuthGetResults(ctx context.Context) ([]frejam.AuthResponse, error) {
	//content := []byte(`{"includePrevious":"ALL"}`)
	//payload := fmt.Sprintf( "getOneAuthResultRequest=%s", base64.StdEncoding.EncodeToString(content))
	payload := `getAuthResultsRequest=eyJpbmNsdWRlUHJldmlvdXMiOiJBTEwifQ==`

	req, err := http.NewRequest("POST", a.parent.baseURL+getAuthResultsURL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "text")
	resp, err := a.parent.client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {

		e := frejam.FrejaError{}
		err = json.Unmarshal(b, &e)
		if err == nil {
			err = e
		}
		return nil, err
	}

	res := struct {
		AuthenticationResults []frejam.AuthResponse `json:"authenticationResults"`
	}{}

	err = json.Unmarshal(b, &res)

	if err != nil {
		return nil, err
	}

	return res.AuthenticationResults, nil
}

// authRef: string, mandatory . The value must be equal to an authentication reference previously returned from a
// call to the Initiate authentication method. As mentioned above, authentications are short-lived and, once initiated
// by a relying party, must be confirmed by an end user within two minutes. Consequently, fetching the result of an
// authentication for a given authentication reference is only possible within 10 minutes of the call to Initiate
// authentication method that returned the said reference.
func (a *API) AuthGetOneResult(ctx context.Context, authRef string) (*frejam.AuthResponse, error) {
	content := []byte(fmt.Sprintf(`{"authRef":"%s"}`, authRef))
	payload := fmt.Sprintf("getOneAuthResultRequest=%s", base64.StdEncoding.EncodeToString(content))

	req, err := http.NewRequest("POST", a.parent.baseURL+getOneAuthResultURL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "text")
	resp, err := a.parent.client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {

		e := frejam.FrejaError{}
		err = json.Unmarshal(b, &e)
		if err == nil {
			err = e
		}
		return nil, err
	}

	res := frejam.AuthResponse{}

	err = json.Unmarshal(b, &res)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (a *API) AuthCancelRequest(ctx context.Context, authRef string) error {
	content := []byte(fmt.Sprintf(`{"authRef":"%s"}`, authRef))
	payload := fmt.Sprintf("cancelAuthRequest=%s", base64.StdEncoding.EncodeToString(content))

	req, err := http.NewRequest("POST", a.parent.baseURL+cancelAuthURL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "text")
	resp, err := a.parent.client.Do(req)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		e := frejam.FrejaError{}
		err = json.Unmarshal(b, &e)
		if err == nil {
			err = e
		}
		return err
	}
	return nil
}

////// Signing

func (a *API) SignInitRequest(ctx context.Context, signReq frejam.SignRequest) (signRef string, err error) {

	strreq, err := signReq.Marshal()
	if err != nil {
		return "", err
	}

	// implement context and add a deadline

	req, err := http.NewRequest("POST", a.parent.baseURL+initSignURL, bytes.NewBuffer([]byte(strreq)))
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "text")
	resp, err := a.parent.client.Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		e := frejam.FrejaError{}
		err = json.Unmarshal(b, &e)
		if err == nil {
			err = e
		}
		return "", err
	}

	var ref frejam.SignRef
	err = json.Unmarshal(b, &ref)
	if err != nil {
		return "", err
	}

	return ref.SignRef, nil
}

func (a *API) SignGetOneResult(ctx context.Context, signRef string) (*frejam.SignResponse, error) {
	content := []byte(fmt.Sprintf(`{"signRef":"%s"}`, signRef))
	payload := fmt.Sprintf("getOneSignResultRequest=%s", base64.StdEncoding.EncodeToString(content))

	req, err := http.NewRequest("POST", a.parent.baseURL+getOneSignResultURL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "text")
	resp, err := a.parent.client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {

		e := frejam.FrejaError{}
		err = json.Unmarshal(b, &e)
		if err == nil {
			err = e
		}
		return nil, err
	}

	res := frejam.SignResponse{}

	err = json.Unmarshal(b, &res)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (a *API) SignGetResults(ctx context.Context) ([]frejam.SignResponse, error) {
	//content := []byte(`{"includePrevious":"ALL"}`)
	//payload := fmt.Sprintf( "getOneAuthResultRequest=%s", base64.StdEncoding.EncodeToString(content))
	payload := `getSignResultsRequest=eyJpbmNsdWRlUHJldmlvdXMiOiJBTEwifQ==`

	req, err := http.NewRequest("POST", a.parent.baseURL+getSignResultsURL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "text")
	resp, err := a.parent.client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {

		e := frejam.FrejaError{}
		err = json.Unmarshal(b, &e)
		if err == nil {
			err = e
		}
		return nil, err
	}

	res := struct {
		SignatureResults []frejam.SignResponse `json:"signatureResults"`
	}{}

	err = json.Unmarshal(b, &res)

	if err != nil {
		return nil, err
	}

	return res.SignatureResults, nil
}

func (a *API) SignCancelRequest(ctx context.Context, signRef string) error {
	content := []byte(fmt.Sprintf(`{"signRef":"%s"}`, signRef))
	payload := fmt.Sprintf("cancelSignRequest=%s", base64.StdEncoding.EncodeToString(content))

	req, err := http.NewRequest("POST", a.parent.baseURL+cancelSignURL, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Add("content-type", "text")
	resp, err := a.parent.client.Do(req)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		e := frejam.FrejaError{}
		err = json.Unmarshal(b, &e)
		if err == nil {
			err = e
		}
		return err
	}
	return nil
}
