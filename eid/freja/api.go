package freja

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	// implement context and add a deadline
	resp, err := a.parent.client.Post(a.parent.baseURL+initAuthURL, "text", bytes.NewBuffer([]byte(strreq)))
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
func (a *API) AuthGetResults() ([]frejam.AuthResponse, error) {
	//content := []byte(`{"includePrevious":"ALL"}`)
	//payload := fmt.Sprintf( "getOneAuthResultRequest=%s", base64.StdEncoding.EncodeToString(content))
	payload := `getAuthResultsRequest=eyJpbmNsdWRlUHJldmlvdXMiOiJBTEwifQ==`
	resp, err := a.parent.client.Post(a.parent.baseURL+getAuthResultsURL, "text", bytes.NewBuffer([]byte(payload)))

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
func (a *API) AuthGetOneResult(authRef string) (*frejam.AuthResponse, error) {
	content := []byte(fmt.Sprintf(`{"authRef":"%s"}`, authRef))
	payload := fmt.Sprintf("getOneAuthResultRequest=%s", base64.StdEncoding.EncodeToString(content))

	resp, err := a.parent.client.Post(a.parent.baseURL+getOneAuthResultURL, "text", bytes.NewBuffer([]byte(payload)))

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

func (a *API) AuthCancelRequest(authRef string) error {
	content := []byte(fmt.Sprintf(`{"authRef":"%s"}`, authRef))
	payload := fmt.Sprintf("cancelAuthRequest=%s", base64.StdEncoding.EncodeToString(content))

	resp, err := a.parent.client.Post(a.parent.baseURL+cancelAuthURL, "text", bytes.NewBuffer([]byte(payload)))

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
	resp, err := a.parent.client.Post(a.parent.baseURL+initSignURL, "text", bytes.NewBuffer([]byte(strreq)))
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

func (a *API) SignGetOneResult(signRef string) (*frejam.SignResponse, error) {
	content := []byte(fmt.Sprintf(`{"signRef":"%s"}`, signRef))
	payload := fmt.Sprintf("getOneSignResultRequest=%s", base64.StdEncoding.EncodeToString(content))

	resp, err := a.parent.client.Post(a.parent.baseURL+getOneSignResultURL, "text", bytes.NewBuffer([]byte(payload)))

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

func (a *API) SignGetResults() ([]frejam.SignResponse, error) {
	//content := []byte(`{"includePrevious":"ALL"}`)
	//payload := fmt.Sprintf( "getOneAuthResultRequest=%s", base64.StdEncoding.EncodeToString(content))
	payload := `getSignResultsRequest=eyJpbmNsdWRlUHJldmlvdXMiOiJBTEwifQ==`
	resp, err := a.parent.client.Post(a.parent.baseURL+getSignResultsURL, "text", bytes.NewBuffer([]byte(payload)))

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

func (a *API) SignCancelRequest(signRef string) error {
	content := []byte(fmt.Sprintf(`{"signRef":"%s"}`, signRef))
	payload := fmt.Sprintf("cancelSignRequest=%s", base64.StdEncoding.EncodeToString(content))

	resp, err := a.parent.client.Post(a.parent.baseURL+cancelSignURL, "text", bytes.NewBuffer([]byte(payload)))

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
