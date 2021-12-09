package eid

import (
	"context"
	"errors"
	"net"
	"time"
	"twofer/grpc/geid"
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
	Cancel(ctx context.Context, intermediate *Inter) error
	Change(ctx context.Context, req *Inter, cancelOnErr bool) (*Resp, error)

	Ping() error
}

func Request() *Req {
	return &Req{}
}

type Req struct {
	Provider Client
	Who      *User    `json:"who"`
	Payload  *Payload `json:"payload"`
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

func FromGrpcReq(req *geid.Req, cli Client) (r Req, err error) {
	if req == nil {
		err = errors.New("could not convert nil to request")
		return
	}

	if req.Who == nil {
		err = errors.New("who must be defined")
		return
	}

	r.Who = &User{
		Inferred:   req.Who.Inferred,
		SSN:        req.Who.Ssn,
		SSNCountry: req.Who.SsnCountry,
		Email:      req.Who.Email,
		Phone:      req.Who.Phone,
		IP:         net.ParseIP(req.Who.Ip),
	}
	r.Provider = cli
	if req.Payload != nil {
		r.Payload = &Payload{
			Text: req.Payload.Text,
			Data: req.Payload.Data,
		}
	}
	return
}

func FromGrpcInter(inter *geid.Inter, cli Client) (i Inter, err error) {
	if inter == nil {
		err = errors.New("there must be an intermediate to collect")
		return
	}
	req, err := FromGrpcReq(inter.Req, cli)
	if err != nil {
		return
	}
	i.Req = &req
	i.Inferred = inter.Inferred
	i.URI = inter.URI
	i.Ref = inter.Ref
	switch inter.Mode {
	case geid.Inter_AUTH:
		i.Mode = AUTH
	case geid.Inter_SIGN:
		i.Mode = SIGN
	default:
		err = errors.New("there should be a mode present")
	}
	return
}

func ToGrpcInter(inter *Inter) (i geid.Inter, err error) {
	if inter.Req == nil {
		err = errors.New("req is required")
		return
	}
	if inter.Req.Who == nil {
		err = errors.New("who must be defined")
		return
	}
	inter.Req.ensurePayload()
	payload := toGrpcPayload(*inter.Req.Payload)
	user := toGrpcUser(*inter.Req.Who)
	i.Req = &geid.Req{
		Provider: &geid.Provider{
			Name: inter.Req.Provider.Name(),
		},
		Who:     &user,
		Payload: &payload,
	}
	switch inter.Mode {
	case AUTH:
		i.Mode = geid.Inter_AUTH
	case SIGN:
		i.Mode = geid.Inter_SIGN
	default:
		err = errors.New("this should never happen")
	}
	i.Ref = inter.Ref
	i.Inferred = inter.Inferred
	i.URI = inter.URI
	return
}

func ToGrpcResp(res *Resp) (r geid.Resp, e error) {
	if res == nil {
		e = errors.New("SOMETHING'S BROKEN")
		return
	}
	inter, err := ToGrpcInter(res.Inter)
	if err != nil {
		return
	}
	user := toGrpcUser(res.Info)

	r.Inter = &inter
	r.Signature = res.Signature
	r.Info = &user

	// DISGUSTINGTOWN
	switch res.Status {
	case STATUS_UNKNOWN:
		r.Status = geid.Resp_STATUS_UNKNOWN
	case STATUS_PENDING:
		r.Status = geid.Resp_STATUS_PENDING
	case STATUS_ONGOING:
		r.Status = geid.Resp_STATUS_ONGOING
	case STATUS_CANCELED:
		r.Status = geid.Resp_STATUS_CANCELED
	case STATUS_RP_CANCELED:
		r.Status = geid.Resp_STATUS_RP_CANCELED
	case STATUS_EXPIRED:
		r.Status = geid.Resp_STATUS_EXPIRED
	case STATUS_APPROVED:
		r.Status = geid.Resp_STATUS_APPROVED
	case STATUS_REJECTED:
		r.Status = geid.Resp_STATUS_REJECTED
	case STATUS_FAILED:
		r.Status = geid.Resp_STATUS_FAILED
	case STATUS_START_FAILED:
		r.Status = geid.Resp_STATUS_START_FAILED
	default:
		err = errors.New("this should never happen")
		return
	}
	return
}

func toGrpcUser(who User) (u geid.User) {
	u.Inferred = who.Inferred
	u.Ssn = who.SSN
	u.SsnCountry = who.SSNCountry
	u.Email = who.Email
	u.Phone = who.Phone
	u.Ip = who.IP.String()
	u.Name = who.Name
	u.Surname = who.Surname
	return
}

func toGrpcPayload(l Payload) (p geid.Req_Payload) {
	p.Text = l.Text
	p.Data = l.Data
	return
}
