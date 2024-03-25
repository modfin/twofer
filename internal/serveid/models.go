package serveid

import (
	"errors"
	"github.com/modfin/twofer/internal/eid"
	"net"
)

type Inter_Mode int32

const (
	INTER_AUTH Inter_Mode = 0
	INTER_SIGN Inter_Mode = 1
)

type RESP_STATUS int32

const (
	RESP_STATUS_UNKNOWN      RESP_STATUS = 0
	RESP_STATUS_PENDING      RESP_STATUS = 1
	RESP_STATUS_ONGOING      RESP_STATUS = 2
	RESP_STATUS_APPROVED     RESP_STATUS = 3
	RESP_STATUS_CANCELED     RESP_STATUS = 4
	RESP_STATUS_RP_CANCELED  RESP_STATUS = 5
	RESP_STATUS_EXPIRED      RESP_STATUS = 6
	RESP_STATUS_REJECTED     RESP_STATUS = 7
	RESP_STATUS_FAILED       RESP_STATUS = 8
	RESP_STATUS_START_FAILED RESP_STATUS = 9
)

var RESP_STATUS_NAME = map[int32]string{
	0: "STATUS_UNKNOWN",
	1: "STATUS_PENDING",
	2: "STATUS_ONGOING",
	3: "STATUS_APPROVED",
	4: "STATUS_CANCELED",
	5: "STATUS_RP_CANCELED",
	6: "STATUS_EXPIRED",
	7: "STATUS_REJECTED",
	8: "STATUS_FAILED",
	9: "STATUS_START_FAILED",
}
var RESP_STATUS_VALUE = map[string]int32{
	"STATUS_UNKNOWN":      0,
	"STATUS_PENDING":      1,
	"STATUS_ONGOING":      2,
	"STATUS_APPROVED":     3,
	"STATUS_CANCELED":     4,
	"STATUS_RP_CANCELED":  5,
	"STATUS_EXPIRED":      6,
	"STATUS_REJECTED":     7,
	"STATUS_FAILED":       8,
	"STATUS_START_FAILED": 9,
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

type Providers struct {
	Providers []*Provider `json:"providers,omitempty"`
}

func (m *Providers) GetProviders() []*Provider {
	if m != nil {
		return m.Providers
	}
	return nil
}

type Provider struct {
	Name string `json:"name,omitempty"`
}

type User struct {
	Inferred    bool   `json:"inferred,omitempty"`
	Ssn         string `json:"ssn,omitempty"`
	SsnCountry  string `json:"ssn_country,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Ip          string `json:"ip,omitempty"`
	Name        string `json:"name,omitempty"`
	Surname     string `json:"surname,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
}

type Req struct {
	Provider *Provider    `json:"provider,omitempty"`
	Who      *User        `json:"who,omitempty"`
	Payload  *Req_Payload `json:"payload,omitempty"`
}

type Req_Payload struct {
	Text string `json:"text,omitempty"`
	Data []byte `json:"data,omitempty"`
}

type Inter struct {
	Req      *Req       `json:"req,omitempty"`
	Mode     Inter_Mode `json:"mode,omitempty"`
	Ref      string     `json:"ref,omitempty"`
	Inferred string     `json:"inferred,omitempty"`
	URI      string     `json:"URI,omitempty"`
	Internal []byte     `json:"internal,omitempty"`
}

type Resp struct {
	Inter     *Inter      `json:"inter,omitempty"`
	Status    RESP_STATUS `json:"status,omitempty"`
	Info      *User       `json:"info,omitempty"`
	Signature []byte      `json:"signature,omitempty"`
	Extra     []byte      `json:"extra,omitempty"`
}

func FromReq(req *Req) (r eid.Req, err error) {
	if req == nil {
		err = errors.New("could not convert nil to request")
		return
	}

	if req.Provider == nil {
		err = errors.New("provider must be defined")
		return
	}

	if req.Who == nil {
		err = errors.New("who must be defined")
		return
	}

	r.Who = &eid.User{
		Inferred:   req.Who.Inferred,
		SSN:        req.Who.Ssn,
		SSNCountry: req.Who.SsnCountry,
		Email:      req.Who.Email,
		Phone:      req.Who.Phone,
		IP:         net.ParseIP(req.Who.Ip),
	}

	if req.Payload != nil {
		r.Payload = &eid.Payload{
			Text: req.Payload.Text,
			Data: req.Payload.Data,
		}
	}

	r.Provider = &eid.Provider{Name: req.Provider.Name}

	return
}

func FromInter(inter *Inter) (i eid.Inter, err error) {
	if inter == nil {
		err = errors.New("there must be an intermediate to collect")
		return
	}
	req, err := FromReq(inter.Req)
	if err != nil {
		return
	}
	i.Req = &req
	i.Inferred = inter.Inferred
	i.URI = inter.URI
	i.Ref = inter.Ref
	switch inter.Mode {
	case INTER_AUTH:
		i.Mode = eid.AUTH
	case INTER_SIGN:
		i.Mode = eid.SIGN
	default:
		err = errors.New("there should be a mode present")
	}
	return
}

func ToInter(inter *eid.Inter) (i Inter, err error) {
	if inter.Req == nil {
		err = errors.New("req is required")
		return
	}
	if inter.Req.Who == nil {
		err = errors.New("who must be defined")
		return
	}
	inter.Req.EnsurePayload()
	payload := toPayload(*inter.Req.Payload)
	user := toUser(*inter.Req.Who)
	i.Req = &Req{
		Provider: &Provider{
			Name: inter.Req.Provider.Name,
		},
		Who:     &user,
		Payload: &payload,
	}
	switch inter.Mode {
	case eid.AUTH:
		i.Mode = INTER_AUTH
	case eid.SIGN:
		i.Mode = INTER_SIGN
	default:
		err = errors.New("this should never happen")
	}
	i.Ref = inter.Ref
	i.Inferred = inter.Inferred
	i.URI = inter.URI
	return
}

func ToResp(res *eid.Resp) (r Resp, e error) {
	if res == nil {
		e = errors.New("SOMETHING'S BROKEN")
		return
	}
	inter, err := ToInter(res.Inter)
	if err != nil {
		return
	}
	user := toUser(res.Info)

	r.Inter = &inter
	r.Signature = res.Signature
	r.Info = &user

	switch res.Status {
	case eid.STATUS_UNKNOWN:
		r.Status = RESP_STATUS_UNKNOWN
	case eid.STATUS_PENDING:
		r.Status = RESP_STATUS_PENDING
	case eid.STATUS_ONGOING:
		r.Status = RESP_STATUS_ONGOING
	case eid.STATUS_CANCELED:
		r.Status = RESP_STATUS_CANCELED
	case eid.STATUS_RP_CANCELED:
		r.Status = RESP_STATUS_RP_CANCELED
	case eid.STATUS_EXPIRED:
		r.Status = RESP_STATUS_EXPIRED
	case eid.STATUS_APPROVED:
		r.Status = RESP_STATUS_APPROVED
	case eid.STATUS_REJECTED:
		r.Status = RESP_STATUS_REJECTED
	case eid.STATUS_FAILED:
		r.Status = RESP_STATUS_FAILED
	case eid.STATUS_START_FAILED:
		r.Status = RESP_STATUS_START_FAILED
	default:
		err = errors.New("this should never happen")
		return
	}
	return
}

func toUser(who eid.User) (u User) {
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

func toPayload(l eid.Payload) (p Req_Payload) {
	p.Text = l.Text
	p.Data = l.Data
	return
}
