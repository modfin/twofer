package eid

import (
	"context"
	"net"
	"time"
)

type Client interface {
	Name() string

	AuthInit(ctx context.Context, req *Req) (*Inter, error)
	SignInit(ctx context.Context, req *Req) (*Inter, error)

	Peek(ctx context.Context, req *Inter) (*Resp, error)
	Collect(ctx context.Context, req *Inter, cancelOnErr bool) (*Resp, error)
	Cancel(ctx context.Context, intermediate *Inter) error
	Change(ctx context.Context, req *Inter, cancelOnErr bool) (*Resp, error)

	Ping() error
}

type Provider struct {
	Name string
}

type Req struct {
	Provider *Provider `json:"provider"`
	Who      *User     `json:"who"`
	Payload  *Payload  `json:"payload"`
}

func (r *Req) EnsureWho() {
	if r.Who == nil {
		r.Who = &User{}
	}
}

func (r *Req) EnsurePayload() {
	if r.Payload == nil {
		r.Payload = &Payload{}
	}
}

func (r *Req) IP(ip string) *Req {
	r.EnsureWho()
	r.Who.IP = net.ParseIP(ip)
	return r
}

func (r *Req) Inferred() *Req {
	r.EnsureWho()
	r.Who.Inferred = true
	return r
}

func (r *Req) SSN(ssn string, country string) *Req {
	r.EnsureWho()
	r.Who.SSN = ssn
	r.Who.SSNCountry = country
	return r
}

func (r *Req) Email(email string) *Req {
	r.EnsureWho()
	r.Who.Email = email
	return r
}

func (r *Req) Phone(phone string) *Req {
	r.EnsureWho()
	r.Who.Phone = phone
	return r
}

func (r *Req) SignText(text string) *Req {
	r.EnsurePayload()
	r.Payload.Text = text
	return r
}

func (r *Req) SignData(data []byte) *Req {
	r.EnsurePayload()
	r.Payload.Data = data
	return r
}

type User struct {
	Inferred    bool      `json:"inferred"`
	SSN         string    `json:"ssn"`
	SSNCountry  string    `json:"ssn_country"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	IP          net.IP    `json:"ip"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

type Payload struct {
	Text string `json:"text"` // Display text of what is being signed.
	Data []byte `json:"data"` // Preferable a digest of a document
}

type Mode string

const (
	AUTH Mode = "AUTH"
	SIGN Mode = "SIGN"
)

type Inter struct {
	Req *Req `json:"req"`

	Mode     Mode   `json:"mode"`
	Ref      string `json:"ref"`
	Inferred string `json:"inferred"`
	URI      string `json:"URI"`
}

type Status string

const (
	STATUS_UNKNOWN      Status = "UNKNOWN"
	STATUS_PENDING      Status = "PENDING"
	STATUS_ONGOING      Status = "ONGOING"
	STATUS_CANCELED     Status = "CANCELED"
	STATUS_RP_CANCELED  Status = "RP_CANCELED"
	STATUS_EXPIRED      Status = "EXPIRED"
	STATUS_APPROVED     Status = "APPROVED"
	STATUS_REJECTED     Status = "REJECTED"
	STATUS_FAILED       Status = "FAILED"
	STATUS_START_FAILED Status = "START_FAILED"
)

type Resp struct {
	Inter *Inter `json:"inter"`

	Status    Status                 `json:"status"`
	Info      User                   `json:"info"`
	Signature []byte                 `json:"signature"`
	Extra     map[string]interface{} `json:"extra"`
}
