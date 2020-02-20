package eid

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
	"twofer/twoferrpc"
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
	Inferred   bool   `json:"inferred"`
	SSN        string `json:"ssn"`
	SSNCountry string `json:"ssn_country"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	IP         net.IP `json:"ip"`
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
	Name        string    `json:"name"`
	Surname     string    `json:"surname"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

type Resp struct {
	Inter *Inter `json:"inter"`

	Status    Status                 `json:"status"`
	Info      Info                   `json:"info"`
	Signature []byte                 `json:"signature"`
	Extra     map[string]interface{} `json:"extra"`
}

func FromGrpcReq(req *twoferrpc.Req) (r Req, err error) {
	if req == nil {
		err = errors.New("could not convert nil to request")
		return
	}

	if req.Who == nil {
		err = errors.New("who must be defined in request")
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
	if req.Payload != nil {
		r.Payload = &Payload{
			Text: req.Payload.Text,
			Data: req.Payload.Data,
		}
	}
	return
}

func FromGrpcInter(inter *twoferrpc.Inter) (i Inter, err error) {
	if inter == nil {
		err = errors.New("ALL IS NOT WELL")
		return
	}
	req, err := FromGrpcReq(inter.Req)
	if err != nil {
		return
	}
	i.Req = &req
	i.Inferred = inter.Inferred
	i.QRData = inter.QrData
	i.Ref = inter.Ref
	switch inter.Mode {
	case twoferrpc.Inter_AUTH:
		i.Mode = AUTH
	case twoferrpc.Inter_SIGN:
		i.Mode = SIGN
	default:
		fmt.Println("WAAAT")
	}
	return
}

func ToGrpcInter(inter *Inter, provider string) (i twoferrpc.Inter, err error) {
	if inter.Req == nil {
		err = errors.New("SOMETHING'S NOT WHAT WE EXPECT")
		return
	}

	user := toGrpcUser(*inter.Req.Who)
	payload := toGrpcPayload(*inter.Req.Payload)

	i.Req = &twoferrpc.Req{
		Provider: &twoferrpc.Provider{
			Name: provider,
		},
		Who:     &user,
		Payload: &payload,
	}
	switch inter.Mode {
	case AUTH:
		i.Mode = twoferrpc.Inter_AUTH
	case SIGN:
		i.Mode = twoferrpc.Inter_SIGN
	default:
		fmt.Println("WAAAT")
	}
	i.Ref = inter.Ref
	i.Inferred = inter.Inferred
	i.QrData = inter.QRData
	return
}

func ToGrpcResp(res *Resp) (r twoferrpc.Resp, e error) {
	/*	TODO
		Error
		Extra
	*/
	if res == nil {
		e = errors.New("SOMETHING'S BROKEN")
		return
	}
	inter, err := ToGrpcInter(res.Inter, "todo")
	if err != nil {
		return
	}
	user := toGrpcUser(res.Info.User)

	r.Inter = &inter
	r.Signature = res.Signature
	r.Info = &user

	// DISGUSTINGTOWN
	switch res.Status {
	case STATUS_UNKNOWN:
		r.Status = twoferrpc.Resp_STATUS_UNKNOWN
	case STATUS_PENDING:
		r.Status = twoferrpc.Resp_STATUS_PENDING
	case STATUS_CANCELED:
		r.Status = twoferrpc.Resp_STATUS_CANCELED
	case STATUS_RP_CANCELED:
		r.Status = twoferrpc.Resp_STATUS_RP_CANCELED
	case STATUS_EXPIRED:
		r.Status = twoferrpc.Resp_STATUS_EXPIRED
	case STATUS_APPROVED:
		r.Status = twoferrpc.Resp_STATUS_APPROVED
	case STATUS_REJECTED:
		r.Status = twoferrpc.Resp_STATUS_REJECTED
	case STATUS_FAILED:
		r.Status = twoferrpc.Resp_STATUS_FAILED
	default:
		err = errors.New("THIS SHOULDN'T HAPPEN")
		return
	}
	return
}

func toGrpcUser(who User) (u twoferrpc.User) {
	u.Inferred = who.Inferred
	u.Ssn = who.SSN
	u.SsnCountry = who.SSNCountry
	u.Email = who.Email
	u.Phone = who.Phone
	u.Ip = who.IP.String()
	return
}

func toGrpcPayload(l Payload) (p twoferrpc.Req_Payload) {
	p.Text = l.Text
	p.Data = l.Data
	return
}
