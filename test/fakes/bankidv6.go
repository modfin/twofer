package fakes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/modfin/twofer/internal/bankid"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type BankIDV6Fake struct {
	mut    sync.Mutex
	server *http.Server
	URL    string
	Orders map[string]*order
}

type orderStatus string

const (
	pending  orderStatus = "pending"
	failed   orderStatus = "failed"
	complete orderStatus = "complete"
)

type order struct {
	Ref             string      `json:"orderRef"`
	Status          orderStatus `json:"status"`
	HintCode        string      `json:"hintCode"`
	IP              string
	UserVisibleData string
	AutoStartToken  string
	QrStartToken    string
	QrStartSecret   string
}

type errorMessage struct {
	ErrorCode string `json:"errorcode"`
	Reason    string `json:"reason"`
}

func CreateBankIDV6Fake() *BankIDV6Fake {
	mux := http.ServeMux{}

	fake := BankIDV6Fake{
		Orders: make(map[string]*order, 0),
		server: &http.Server{
			Addr:    ":8998",
			Handler: &mux,
		},
		URL: "http://127.0.0.1:8998",
	}

	mux.HandleFunc(bankid.AuthUrl, fake.handleAuth)
	mux.HandleFunc(bankid.SignUrl, fake.handleSign)
	mux.HandleFunc(bankid.CollectUrl, fake.handleCollect)
	mux.HandleFunc(bankid.CancelUrl, fake.handleCancel)

	return &fake
}

func (fake *BankIDV6Fake) Start() error {
	return fake.server.ListenAndServe()
}

func (fake *BankIDV6Fake) Stop(deadline time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()

	return fake.server.Shutdown(ctx)
}

type authResponse struct {
	OrderRef       string `json:"orderRef"`
	AutoStartToken string `json:"autoStartToken"`
	QrStartToken   string `json:"qrStartToken"`
	QrStartSecret  string `json:"qrStartSecret"`
}

type authReq struct {
	EndUserIp string `json:"endUserIp"`
}

func (fake *BankIDV6Fake) handleAuth(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		e := errorMessage{
			ErrorCode: "fake failed to read bytes",
			Reason:    err.Error(),
		}
		respond(w, e, 400)
	}

	var req authReq
	err = json.Unmarshal(bytes, &req)
	if err != nil {
		e := errorMessage{
			ErrorCode: "fake failed to unmarshal",
			Reason:    err.Error(),
		}
		respond(w, e, 400)
	}

	if req.EndUserIp == "" {
		e := errorMessage{
			ErrorCode: "invalidParameters",
			Reason:    "no endUserIp",
		}
		respond(w, e, 400)
	}

	o := order{
		Ref:            uuid.NewString(),
		Status:         pending,
		HintCode:       "outstandingTransaction",
		IP:             req.EndUserIp,
		QrStartToken:   uuid.NewString(),
		QrStartSecret:  uuid.NewString(),
		AutoStartToken: uuid.NewString(),
	}

	fake.mut.Lock()
	fake.Orders[o.Ref] = &o
	fake.mut.Unlock()

	response := &authResponse{
		OrderRef:       o.Ref,
		AutoStartToken: o.AutoStartToken,
		QrStartToken:   o.QrStartToken,
		QrStartSecret:  o.QrStartSecret,
	}

	respond(w, response, http.StatusOK)
}

type signReq struct {
	EndUserIp       string `json:"endUserIp"`
	UserVisibleData string `json:"userVisibleData"`
}

func (fake *BankIDV6Fake) handleSign(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		e := errorMessage{
			ErrorCode: "fake failed to read bytes",
			Reason:    err.Error(),
		}
		respond(w, e, 400)
	}

	var req signReq
	err = json.Unmarshal(bytes, &req)
	if err != nil {
		e := errorMessage{
			ErrorCode: "fake failed to unmarshal",
			Reason:    err.Error(),
		}
		respond(w, e, 400)
	}

	if req.EndUserIp == "" {
		e := errorMessage{
			ErrorCode: "invalidParameters",
			Reason:    "empty endUserIp",
		}
		respond(w, e, 400)
	}

	if req.UserVisibleData == "" {
		e := errorMessage{
			ErrorCode: "invalidParameters",
			Reason:    "empty user visible data",
		}
		respond(w, e, 400)
	}

	o := order{
		Ref:             uuid.NewString(),
		Status:          pending,
		HintCode:        "outstandingTransaction",
		IP:              req.EndUserIp,
		UserVisibleData: req.UserVisibleData,
		QrStartToken:    uuid.NewString(),
		QrStartSecret:   uuid.NewString(),
		AutoStartToken:  uuid.NewString(),
	}

	fake.mut.Lock()
	fake.Orders[o.Ref] = &o
	fake.mut.Unlock()

	response := &authResponse{
		OrderRef:       o.Ref,
		AutoStartToken: o.AutoStartToken,
		QrStartToken:   o.QrStartToken,
		QrStartSecret:  o.QrStartSecret,
	}

	respond(w, response, http.StatusOK)
}

type collectReq struct {
	OrderRef string `json:"orderRef"`
}

func (fake *BankIDV6Fake) handleCollect(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		e := errorMessage{
			ErrorCode: "fake failed to read bytes",
			Reason:    err.Error(),
		}
		respond(w, e, 400)
	}

	var req collectReq
	err = json.Unmarshal(bytes, &req)
	if err != nil {
		e := errorMessage{
			ErrorCode: "fake failed to unmarshal",
			Reason:    err.Error(),
		}
		respond(w, e, 400)
	}

	o, ok := fake.Orders[req.OrderRef]
	if !ok {
		e := errorMessage{
			ErrorCode: "invalidParameters",
			Reason:    err.Error(),
		}
		respond(w, e, 404)
	}

	respond(w, o, http.StatusOK)
}

type empty struct{}

func (fake *BankIDV6Fake) handleCancel(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		e := errorMessage{
			ErrorCode: "fake failed to read bytes",
			Reason:    err.Error(),
		}
		respond(w, e, 400)
	}

	var req collectReq
	err = json.Unmarshal(bytes, &req)
	if err != nil {
		e := errorMessage{
			ErrorCode: "fake failed to unmarshal",
			Reason:    err.Error(),
		}
		respond(w, e, 400)
	}

	fake.mut.Lock()
	_, ok := fake.Orders[req.OrderRef]
	if !ok {
		respond(w, empty{}, http.StatusNotFound)
	}

	fake.Orders[req.OrderRef].Status = failed
	fake.Orders[req.OrderRef].HintCode = "userCancel"

	fake.mut.Unlock()

	respond(w, empty{}, http.StatusOK)
}

func respond(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
	}
}
