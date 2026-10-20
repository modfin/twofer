// Package bankidtest fakes the BankID RP API v6 endpoints twofer calls, so
// tests can run twofer's real handlers without BankID.
package bankidtest

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

const (
	statusPending  = "pending"
	statusComplete = "complete"
	statusFailed   = "failed"
)

// BankID fakes BankID. Orders stay pending until a test completes or fails
// them, as when nobody has scanned yet.
type BankID struct {
	srv *httptest.Server

	mu     sync.Mutex
	orders map[string]*order // by orderRef, kept after BankID forgets them
	auths  []AuthRequest
}

// AuthRequest is what twofer sent BankID to start an auth order.
type AuthRequest struct {
	EndUserIP      string
	PersonalNumber string
	// UserVisibleData as BankID received it, base64 encoded by twofer
	UserVisibleData string
}

// Order is the fake's record of an order.
type Order struct {
	OrderRef       string
	PersonalNumber string
	EndUserIP      string
	Status         string
	HintCode       string
	AutoStartToken string
	QrStartToken   string
	QrStartSecret  string
	// Cancelled is set when twofer called cancel. BankID forgets the order then.
	Cancelled bool
}

type order struct {
	Order
	deviceIP string
	// gone orders get "No such order" from BankID
	gone bool
}

// CompleteOption changes how BankID reports a completed order.
type CompleteOption func(*order)

// FromIP has BankID report the order signed from ip rather than the endUserIp
// it was started with, as when the autostart link was opened elsewhere.
func FromIP(ip string) CompleteOption {
	return func(o *order) { o.deviceIP = ip }
}

// New starts the fake on a local port. Close it when done.
func New() *BankID {
	b := &BankID{orders: map[string]*order{}}
	mux := http.NewServeMux()
	// twofer pings the base URL at startup
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {})
	mux.HandleFunc("POST /rp/v6.0/auth", b.auth)
	mux.HandleFunc("POST /rp/v6.0/sign", b.sign)
	mux.HandleFunc("POST /rp/v6.0/collect", b.collect)
	mux.HandleFunc("POST /rp/v6.0/cancel", b.cancel)
	b.srv = httptest.NewServer(mux)
	return b
}

// URL is the base URL to configure twofer with.
func (b *BankID) URL() string {
	return b.srv.URL
}

// Port is the local port the fake listens on.
func (b *BankID) Port() int {
	i := strings.LastIndex(b.srv.URL, ":")
	port, _ := strconv.Atoi(b.srv.URL[i+1:])
	return port
}

func (b *BankID) Close() {
	b.srv.Close()
}

// Auths returns the auth requests BankID received for pnr.
func (b *BankID) Auths(pnr string) []AuthRequest {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []AuthRequest
	for _, a := range b.auths {
		if a.PersonalNumber == pnr {
			out = append(out, a)
		}
	}
	return out
}

// Order returns the fake's record of orderRef, also after BankID forgot it.
func (b *BankID) Order(orderRef string) (Order, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	o, ok := b.orders[orderRef]
	if !ok {
		return Order{}, false
	}
	return o.Order, true
}

// Complete signs pnr's pending order.
func (b *BankID) Complete(pnr string, opts ...CompleteOption) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	o, err := b.pending(pnr)
	if err != nil {
		return err
	}
	o.Status, o.HintCode = statusComplete, ""
	for _, opt := range opts {
		opt(o)
	}
	return nil
}

// Fail ends pnr's pending order with hintCode, e.g. userCancel when the user
// cancels in the BankID app.
func (b *BankID) Fail(pnr, hintCode string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	o, err := b.pending(pnr)
	if err != nil {
		return err
	}
	o.Status, o.HintCode = statusFailed, hintCode
	return nil
}

// pending must be called with mu held.
func (b *BankID) pending(pnr string) (*order, error) {
	if pnr == "" {
		return nil, errors.New("no personnummer, orders without one cannot be told apart")
	}
	o := b.pendingOrder(pnr)
	if o == nil {
		return nil, fmt.Errorf("no pending bankid order for %s", pnr)
	}
	return o, nil
}

// pendingOrder must be called with mu held.
func (b *BankID) pendingOrder(pnr string) *order {
	for _, o := range b.orders {
		if !o.gone && o.PersonalNumber == pnr && o.Status == statusPending {
			return o
		}
	}
	return nil
}

type startRequest struct {
	EndUserIP       string `json:"endUserIp"`
	UserVisibleData string `json:"userVisibleData"`
	Requirement     struct {
		PersonalNumber string `json:"personalNumber"`
	} `json:"requirement"`
}

func (b *BankID) auth(w http.ResponseWriter, r *http.Request) {
	var req startRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponse(w, "invalidParameters", "invalid json")
		return
	}

	b.mu.Lock()
	b.auths = append(b.auths, AuthRequest{
		EndUserIP:       req.EndUserIP,
		PersonalNumber:  req.Requirement.PersonalNumber,
		UserVisibleData: req.UserVisibleData,
	})
	b.mu.Unlock()

	b.start(w, req)
}

func (b *BankID) sign(w http.ResponseWriter, r *http.Request) {
	var req startRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponse(w, "invalidParameters", "invalid json")
		return
	}
	if req.UserVisibleData == "" {
		errorResponse(w, "invalidParameters", "Missing userVisibleData")
		return
	}

	b.start(w, req)
}

func (b *BankID) start(w http.ResponseWriter, req startRequest) {
	if req.EndUserIP == "" {
		errorResponse(w, "invalidParameters", "Missing endUserIp")
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Like BankID: a second order for a person with one pending is refused,
	// and the pending one is aborted too.
	pnr := req.Requirement.PersonalNumber
	if o := b.pendingOrder(pnr); pnr != "" && o != nil {
		o.Status, o.HintCode = statusFailed, "cancelled"
		errorResponse(w, "alreadyInProgress", "Order already in progress for pno")
		return
	}

	o := &order{Order: Order{
		OrderRef:       uuid.NewString(),
		PersonalNumber: pnr,
		EndUserIP:      req.EndUserIP,
		Status:         statusPending,
		HintCode:       "outstandingTransaction",
		AutoStartToken: uuid.NewString(),
		QrStartToken:   uuid.NewString(),
		QrStartSecret:  uuid.NewString(),
	}}
	b.orders[o.OrderRef] = o
	jsonResponse(w, map[string]string{
		"orderRef":       o.OrderRef,
		"autoStartToken": o.AutoStartToken,
		"qrStartToken":   o.QrStartToken,
		"qrStartSecret":  o.QrStartSecret,
	})
}

func (b *BankID) collect(w http.ResponseWriter, r *http.Request) {
	ref, ok := orderRef(w, r)
	if !ok {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	o, ok := b.orders[ref]
	if !ok || o.gone {
		errorResponse(w, "invalidParameters", "No such order")
		return
	}

	res := map[string]any{"orderRef": o.OrderRef, "status": o.Status}
	if o.HintCode != "" {
		res["hintCode"] = o.HintCode
	}
	if o.Status == statusComplete {
		deviceIP := o.EndUserIP
		if o.deviceIP != "" {
			deviceIP = o.deviceIP
		}
		res["completionData"] = map[string]any{
			"user":            map[string]string{"personalNumber": o.PersonalNumber, "name": "Test Person", "givenName": "Test", "surname": "Person"},
			"device":          map[string]string{"ipAddress": deviceIP},
			"bankIdIssueDate": "2026-01-01",
			"signature":       base64.StdEncoding.EncodeToString([]byte("signature")),
			"ocspResponse":    base64.StdEncoding.EncodeToString([]byte("ocsp")),
		}
	}
	// BankID hands out a finished order only once
	if o.Status != statusPending {
		o.gone = true
	}
	jsonResponse(w, res)
}

func (b *BankID) cancel(w http.ResponseWriter, r *http.Request) {
	ref, ok := orderRef(w, r)
	if !ok {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	o, ok := b.orders[ref]
	if !ok || o.gone {
		errorResponse(w, "invalidParameters", "No such order")
		return
	}
	o.Cancelled, o.gone = true, true
	jsonResponse(w, map[string]string{})
}

func orderRef(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req struct {
		OrderRef string `json:"orderRef"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errorResponse(w, "invalidParameters", "invalid json")
		return "", false
	}
	return req.OrderRef, true
}

func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// errorResponse answers like BankID. twofer reads the error code only when the
// content type is exactly application/json.
func errorResponse(w http.ResponseWriter, code, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"errorCode": code, "details": details})
}
