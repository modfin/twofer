package eid

import (
	"context"
	"net"
	"time"
)

type ToEID interface {
	EID() Client
}

type Client interface {
	Name() string

	AuthInit(ctx context.Context, req *Req) (*Inter, error)
	SignInit(ctx context.Context, req *Req) (*Inter, error)

	Peek(ctx context.Context, req *Inter) (*Resp, error)
	Collect(ctx context.Context, req *Inter, cancelOnErr bool) (*Resp, error)
	Cancel(intermediate *Inter) error
}

func Request() *Req {
	return &Req{}
}
type Req struct {
	Who     *User `json:"who"`
	Payload *Payload `json:"payload"`
}

func (r *Req) ensureWho() {
	if r.Who == nil {
		r.Who = &User{}
	}
}
func (r *Req) ensurePayload() {
	if r.Payload == nil {
		r.Payload = &Payload{}
	}
}

func (r *Req) IP(ip string) *Req {
	r.ensureWho()
	r.Who.IP = net.ParseIP(ip)
	return r
}

func (r *Req) Inferred() *Req {
	r.ensureWho()
	r.Who.Inferred = true
	return r
}
func (r *Req) SSN(ssn string, country string) *Req {
	r.ensureWho()
	r.Who.SSN = ssn
	r.Who.SSNCountry = country
	return r
}
func (r *Req) Email(email string) *Req {
	r.ensureWho()
	r.Who.Email = email
	return r
}
func (r *Req) Phone(phone string) *Req {
	r.ensureWho()
	r.Who.Phone = phone
	return r
}

func (r *Req) SignText(text string) *Req {
	r.ensurePayload()
	r.Payload.Text = text
	return r
}

func (r *Req) SignData(data []byte) *Req {
	r.ensurePayload()
	r.Payload.Data = data
	return r
}

type User struct {
	Inferred   bool `json:"inferred"`
	SSN        string `json:"ssn"`
	SSNCountry string`json:"ssn_country"`
	Email      string`json:"email"`
	Phone      string`json:"phone"`
	IP         net.IP`json:"ip"`
}
type Payload struct {
	Text string  `json:"text"`// Display text of what is being signed.
	Data []byte `json:"data"`// Preferable a digest of a document
}

type Mode string

const (
	AUTH Mode = "AUTH"
	SIGN Mode = "SIGN"
)

type Inter struct {
	Req *Req `json:"req"`

	Mode     Mode `json:"mode"`
	Ref      string `json:"ref"`
	Inferred string `json:"inferred"`
	QRData   string `json:"qr_data"`
}

type Status string

const (
	STATUS_UNKNOWN Status = "UNKNOWN"

	STATUS_PENDING     Status = "PENDING"
	STATUS_CANCELED    Status = "CANCELED"
	STATUS_RP_CANCELED Status = "RP_CANCELED"
	STATUS_EXPIRED     Status = "EXPIRED"
	STATUS_APPROVED    Status = "APPROVED"
	STATUS_REJECTED    Status = "REJECTED"
	STATUS_FAILED      Status = "FAILED"
)

type Info struct {
	User
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

type Resp struct {
	Inter *Inter `json:"inter"`

	Status    Status `json:"status"`
	Info      Info `json:"info"`
	Signature []byte `json:"signature"`
	Extra     map[string]interface{} `json:"extra"`
}
