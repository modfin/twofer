package bankid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"twofer/eid/bankid/bankidm"
)

type API struct {
	parent *Client
}



func (a *API) Auth(request bankidm.AuthRequest) (r *bankidm.AuthResponse, err error){

	data, err := request.Marshal()

	res, err := a.parent.client.Post(a.parent.baseURL + authURL, "application/json", bytes.NewBuffer(data))
	if err != nil{
		return nil, err
	}

	resdata, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return nil, err
	}

	if res.StatusCode != 200{
		var err1 bankidm.BankIdError
		err = json.Unmarshal(resdata, &err1)
		if err == nil{
			err = err1
		}
		return nil, err
	}

	var resp bankidm.AuthResponse
	err = json.Unmarshal(resdata, &resp)
	return &resp, err


}

func (a *API) Sign(request bankidm.SignRequest) (r *bankidm.SignResponse, err error){

	data, err := request.Marshal()

	res, err := a.parent.client.Post(a.parent.baseURL + signURL, "application/json", bytes.NewBuffer(data))
	if err != nil{
		return nil, err
	}

	resdata, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return nil, err
	}

	if res.StatusCode != 200{
		var err1 bankidm.BankIdError
		err = json.Unmarshal(resdata, &err1)
		if err == nil{
			err = err1
		}
		return nil, err
	}

	var resp bankidm.SignResponse
	err = json.Unmarshal(resdata, &resp)
	return &resp, err
}


func (a *API) Collect(orderRef string) (r *bankidm.CollectResponse, err error){

	data := []byte(fmt.Sprintf(`{"orderRef":"%s"}`, orderRef))
	res, err := a.parent.client.Post(a.parent.baseURL + collectURL, "application/json", bytes.NewBuffer(data))
	if err != nil{
		return nil, err
	}

	resdata, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return nil, err
	}

	if res.StatusCode != 200{
		var err1 bankidm.BankIdError
		err = json.Unmarshal(resdata, &err1)
		if err == nil{
			err = err1
		}
		return nil, err
	}

	var resp bankidm.CollectResponse
	err = json.Unmarshal(resdata, &resp)
	return &resp, err
}

func (a *API) Cancel(orderRef string) (err error){
	data := []byte(fmt.Sprintf(`{"orderRef":"%s"}`, orderRef))

	res, err := a.parent.client.Post(a.parent.baseURL + cancelURL, "application/json", bytes.NewBuffer(data))
	if err != nil{
		return err
	}

	resdata, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return err
	}

	if res.StatusCode != 200{
		var err1 bankidm.BankIdError
		err = json.Unmarshal(resdata, &err1)
		if err == nil{
			err = err1
		}
		return err
	}

	return nil
}