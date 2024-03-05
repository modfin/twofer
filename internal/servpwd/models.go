package models

type EnrollReq struct {
	Password string `json:"password,omitempty"`
}

type AuthReq struct {
	Password string `json:"password,omitempty"`
	UserBlob string `json:"userBlob,omitempty"`
}

type Res struct {
	Valid   bool   `json:"valid,omitempty"`
	Message string `json:"message,omitempty"`
}

type Blob struct {
	UserBlob string `json:"userBlob,omitempty"`
}
