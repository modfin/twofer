package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"twofer/example/eid/dao"
	"twofer/grpc/geid"
	"twofer/grpc/gqr"
	"twofer/grpc/gw6n"
)

var (
	serverAddr   = "twoferd:43210"
	_eidClient   geid.EIDClient
	_qrQlient    gqr.QRClient
	_authnClient gw6n.WebAuthnClient
)

func httpError(writer http.ResponseWriter, err error) {
	writer.WriteHeader(500)
	writer.Write([]byte(err.Error()))
}
func main() {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	fmt.Println("Dialing grpc..")
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	fmt.Println("connected")
	defer conn.Close()
	_eidClient = geid.NewEIDClient(conn)
	_qrQlient = gqr.NewQRClient(conn)
	_authnClient = gw6n.NewWebAuthnClient(conn)
	m := mux.NewRouter()

	// Eid
	m.HandleFunc("/eid/providers", getProviders)
	m.HandleFunc("/eid/inferred", initInferred)
	m.HandleFunc("/eid/start-auth", ssnAuthInit)
	m.HandleFunc("/eid/start-sign", ssnSignInit)
	m.HandleFunc("/eid/qrimage/{ref}", qrImage)
	m.HandleFunc("/eid/peek/{ref}", peek)
	m.HandleFunc("/eid/collect/{ref}", collect)
	m.HandleFunc("/eid/cancel", cancel)

	// Twofer
	m.HandleFunc("/authn/register", registerBegin).Methods("GET")
	m.HandleFunc("/authn/register", registerFinish).Methods("POST")
	m.HandleFunc("/authn/login", loginBegin).Methods("GET")
	m.HandleFunc("/authn/login", loginFinish).Methods("POST")

	m.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	http.Handle("/", m)
	_ = http.ListenAndServe(":8080", nil)
}

func internalError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	_, _ = w.Write([]byte(err.Error()))
}

func getProviders(w http.ResponseWriter, r *http.Request) {
	p, err := _eidClient.GetProviders(context.Background(), &geid.Empty{})
	if err != nil {
		internalError(w, err)
		return
	}
	pr, err := json.Marshal(p)
	_, _ = w.Write(pr)
}

func initInferred(w http.ResponseWriter, r *http.Request) {
	pr := r.URL.Query().Get("provider")
	fmt.Printf("Provider=%s\n", pr)
	i, err := _eidClient.AuthInit(context.Background(), &geid.Req{
		Provider: &geid.Provider{
			Name: pr,
		},
		Who: &geid.User{
			Inferred: true,
		},
	})
	if err != nil {
		internalError(w, err)
		return
	}
	m, err := json.Marshal(i)
	if err != nil {
		internalError(w, err)
		return
	}
	dao.SetInter(i.Ref, m)
	qrr, err := _qrQlient.Generate(context.Background(), &gqr.Data{
		RecoveryLevel: 0,
		Size:          0,
		Data:          i.URI,
	})
	if err != nil {
		internalError(w, err)
		return
	}
	dao.SetQr(dao.QrData{
		Reference: i.Ref,
		Image:     qrr.Data,
	})
	ir, err := json.Marshal(i)
	_, _ = w.Write(ir)
}

func ssnAuthInit(w http.ResponseWriter, r *http.Request) {
	pr := r.URL.Query().Get("provider")
	u := &geid.User{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		internalError(w, err)
		return
	}
	err = json.Unmarshal(b, u)
	if err != nil {
		internalError(w, err)
		return
	}

	init, err := _eidClient.AuthInit(context.Background(), &geid.Req{
		Provider: &geid.Provider{
			Name: pr,
		},
		Who:     u,
		Payload: nil,
	})
	fmt.Printf("Provider=%s\n", pr)
	if err != nil {
		internalError(w, err)
		return
	}
	m, err := json.Marshal(init)
	if err != nil {
		internalError(w, err)
		return
	}
	dao.SetInter(init.Ref, m)
	ir, err := json.Marshal(init)
	_, _ = w.Write(ir)
}

func peek(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ref := vars["ref"]
	inter, ok := dao.GetInter(ref)
	if !ok {
		internalError(w, errors.New("no ref stored for ref={ref}"))
	}
	i := geid.Inter{}
	err := json.Unmarshal(inter, &i)
	if err != nil {
		internalError(w, err)
		return
	}
	peek, err := _eidClient.Peek(context.Background(), &i)
	if err != nil {
		internalError(w, err)
		return
	}
	pr, err := json.Marshal(peek)
	if err != nil {
		internalError(w, err)
		return
	}
	_, _ = w.Write(pr)
}

func collect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ref := vars["ref"]
	inter, ok := dao.GetInter(ref)
	if !ok {
		internalError(w, errors.New("no ref stored for ref={ref}"))
	}
	i := geid.Inter{}
	err := json.Unmarshal(inter, &i)
	if err != nil {
		internalError(w, err)
		return
	}
	collect, err := _eidClient.Collect(context.Background(), &i)
	if err != nil {
		internalError(w, err)
		return
	}
	ar, err := json.Marshal(collect)
	if err != nil {
		internalError(w, err)
		return
	}
	_, _ = w.Write(ar)
}

func cancel(w http.ResponseWriter, r *http.Request) {
	ref := r.URL.Query().Get("ref")

	inter, ok := dao.GetInter(ref)
	fmt.Printf("waaaat")
	if !ok {
		internalError(w, errors.New("no ref stored for ref={ref}"))
	}
	i := geid.Inter{}
	err := json.Unmarshal(inter, &i)
	if err != nil {
		internalError(w, err)
		return
	}
	_, err = _eidClient.Cancel(context.Background(), &i)
	if err != nil {
		internalError(w, err)
		return
	}
}

func ssnSignInit(w http.ResponseWriter, r *http.Request) {
	pr := r.URL.Query().Get("provider")
	u := &geid.Req{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		internalError(w, err)
		return
	}
	err = json.Unmarshal(b, u)
	if err != nil {
		internalError(w, err)
		return
	}

	init, err := _eidClient.SignInit(context.Background(), &geid.Req{
		Provider: &geid.Provider{
			Name: pr,
		},
		Who:     u.Who,
		Payload: u.Payload,
	})
	fmt.Printf("Provider=%s\n", pr)
	if err != nil {
		internalError(w, err)
		return
	}
	m, err := json.Marshal(init)
	if err != nil {
		internalError(w, err)
		return
	}
	dao.SetInter(init.Ref, m)
	ir, err := json.Marshal(init)
	_, _ = w.Write(ir)
}

func qrImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ref := vars["ref"]
	fmt.Printf(ref)
	qr, ok := dao.GetQr(ref)
	if !ok {
		internalError(w, errors.New("BONKERS, AS USUAL"))
	}

	w.Header().Set("Content-Type:", "image/png")
	_, _ = w.Write(qr.Image)
}

func createConfig(r *http.Request) *gw6n.Config {
	if r.Host != "localhost:8080" {
		name := strings.Split(r.Host, ":")[0]
		return &gw6n.Config{
			RPID:          name,
			RPDisplayName: name,
			RPOrigin:      "http://" + r.Host,
		}
	}
	return nil
}

func loginBegin(w http.ResponseWriter, r *http.Request) {
	var cfg = createConfig(r)

	userId := r.Header.Get("user")
	u, ok := dao.Get(userId)
	if !ok {
		httpError(w, errors.New("no user "+userId))
		return
	}
	resp, err := _authnClient.AuthInit(context.Background(), &gw6n.AuthInitReq{
		// The credentials for the user provided in the latest performed enrollment finish step
		UserBlob: u.Blob,

		// Optional, will use server default if not set
		Cfg: cfg,
	})
	if err != nil {
		httpError(w, err)
		return
	}

	// The session will be used to create the credentials and must be provided in the finish step.
	// While setting it as a header or a cookie works, the session can become quite big
	// and we suggest you persist it in a session store where it can later be retrived, eg. Redis, memcached or such.
	w.Header().Set("WebAuthn-Session", string(resp.Session))

	w.Header().Set("Content-Type:", "application/json")
	fmt.Printf("%s", string(resp.Json))

	_, _ = w.Write(resp.Json)
}

func loginFinish(w http.ResponseWriter, r *http.Request) {
	var cfg = createConfig(r)

	session := []byte(r.Header.Get("WebAuthn-Session"))
	signature, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(w, err)
		return
	}

	_, err = _authnClient.AuthFinal(context.Background(), &gw6n.FinalReq{
		// The session that was created in the registerBegin begin step
		Session: session,

		// The signature provided from the frontend.
		Signature: signature,

		// Optional, will use server default if not set
		Cfg: cfg,
	})
	if err != nil {
		httpError(w, err)
		return
	}

}

func registerBegin(w http.ResponseWriter, r *http.Request) {

	var cfg = createConfig(r)

	userId := r.Header.Get("user")
	u, _ := dao.Get(userId)

	resp, err := _authnClient.EnrollInit(context.Background(), &gw6n.EnrollInitReq{
		// This is primarily used for first enrollment.
		// if the user does exist, only adding the blob works fine
		User: &gw6n.User{
			Id: userId,
		},

		// if the user already have credentials and want to enroll new once, attach the current one
		UserBlob: u.Blob,

		// Optional, will use server default if not set
		Cfg: cfg,
	})
	if err != nil {
		httpError(w, err)
		return
	}

	// The session will be used to create the credentials and must be provided in the finish step.
	// While setting it as a header or a cookie works, the session can become quite big
	// and we suggest you persist it in a session store where it can later be retrived, eg. Redis, memcached or such.
	w.Header().Set("WebAuthn-Session", string(resp.Session))
	w.Header().Set("Content-Type:", "application/json")
	w.WriteHeader(200)
	fmt.Printf("%s", string(resp.Json))
	_, _ = w.Write(resp.Json)
}

func registerFinish(w http.ResponseWriter, r *http.Request) {

	var cfg = createConfig(r)

	userId := r.Header.Get("user")

	session := []byte(r.Header.Get("WebAuthn-Session"))
	signature, err := ioutil.ReadAll(r.Body)

	if err != nil {
		httpError(w, err)
		return
	}
	resp, err := _authnClient.EnrollFinal(context.Background(), &gw6n.FinalReq{
		// The session that was created in the registerBegin begin step
		Session: session,

		// The signature provided from the frontend.
		Signature: signature,

		// Optional, will use server default if not set
		Cfg: cfg,
	})

	if err != nil {
		httpError(w, err)
		return
	}

	dao.Set(dao.User{
		Id: userId,

		//Make sure to persist the UserBlob along with the user. This Blob is needed to later authenticate the user.
		Blob: resp.UserBlob,
	})

	w.WriteHeader(200)
}
