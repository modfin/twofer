package eid

import "context"

type ToEID interface {
	EID() Client
}

type Client interface {
	Name() string

	AuthInit(ctx context.Context, req Request) (Intermediate, error)
	AuthCollect(ctx context.Context, req Intermediate) (Response, error)
	AuthCancel(intermediate Intermediate) error

	SignInit(ctx context.Context, req Request) (Intermediate, error)
	SignCollect(ctx context.Context, req Intermediate) (Response, error)
	SignCancel(intermediate Intermediate) error
}

type Request struct {
	User   *User
	UserIP string // User ip if available

	SignText string // Display text of what is being signed.
	SignData []byte // Preferable a digest of a document
}

type User struct {
	SSN   *SSN
	Email string
	Phone string
}

type SSN struct {
	SSN     string
	Country string
}

type Mode string

const (
	AUTH Mode = "auth"
	SIGN Mode = "sign"
)

type Intermediate struct {
	Mode     Mode
	Ref      string
	Inferred string
	QRData   string

	Internal map[string]interface{}
}

type Status string

const (
	STATUS_UNKNOWN     Status = "UNKNOWN"

	STATUS_PENDING     Status = "PENDING"
	STATUS_CANCELED    Status = "CANCELED"
	STATUS_RP_CANCELED Status = "RP_CANCELED"
	STATUS_EXPIRED     Status = "EXPIRED"
	STATUS_APPROVED    Status = "APPROVED"
	STATUS_REJECTED    Status = "REJECTED"
	STATUS_FAILED      Status = "FAILED"
)


type UserInfo struct{
	User
	Name string
	GivenName string
	Surname string
}

type Response struct {
	Intermediate Intermediate
	Status Status
	UserInfo UserInfo
	Signature []byte
	Extra map[string]interface{}
}
