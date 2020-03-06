package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"twofer/example/webauthn/dao"
	"twofer/grpc/gw6n"
)

var (
	serverAddr = "twoferd:43210"
	_client    gw6n.WebAuthnClient
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
	_client = gw6n.NewWebAuthnClient(conn)

	m := mux.NewRouter()
	m.HandleFunc("/register", registerBegin).Methods("GET")
	m.HandleFunc("/register", registerFinish).Methods("POST")
	m.HandleFunc("/login", loginBegin).Methods("GET")
	m.HandleFunc("/login", loginFinish).Methods("POST")
	m.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	http.Handle("/", m)
	_ = http.ListenAndServe(":8080", nil)
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
	resp, err := _client.AuthInit(context.Background(), &gw6n.AuthInitReq{
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

	_, err = _client.AuthFinal(context.Background(), &gw6n.FinalReq{
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

	resp, err := _client.EnrollInit(context.Background(), &gw6n.EnrollInitReq{
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
	resp, err := _client.EnrollFinal(context.Background(), &gw6n.FinalReq{
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