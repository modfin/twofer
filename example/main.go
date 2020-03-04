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
	"twofer/example/dao"
	"twofer/twoferrpc/gw6n"
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
	m.HandleFunc("/register/{userId}", registerBegin).Methods("GET")
	m.HandleFunc("/register/finish/{userId}", registerFinish).Methods("POST")
	m.HandleFunc("/login/begin/{userId}", loginBegin).Methods("GET")
	m.HandleFunc("/login/finish/{userId}", loginFinish).Methods("POST")
	m.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static"))))

	http.Handle("/", m)
	_ = http.ListenAndServe(":8080", nil)
}

func loginFinish(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	u, ok := dao.Get(userId)

	if !ok {
		httpError(w, errors.New("no user "+userId))
		return
	}

	session := []byte(r.Header.Get("WebAuthn-Session"))
	userBlob := u.Blob
	signature, err := ioutil.ReadAll(r.Body)
	if err != nil {
		httpError(w, err)
		return
	}

	_, err = _client.FinishLogin(context.Background(), &gw6n.FinishLoginRequest{
		Session:   session,
		UserBlob:  userBlob,
		Signature: signature,
	})
	if err != nil {
		httpError(w, err)
		return
	}

}

func loginBegin(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	userId := vars["userId"]
	u, ok := dao.Get(userId)
	if !ok {
		httpError(w, errors.New("no user "+userId))
		return
	}
	resp, err := _client.BeginLogin(context.Background(), &gw6n.BeginLoginRequest{
		UserBlob: u.Blob,
	})
	if err != nil {
		httpError(w, err)
		return
	}

	w.Header().Set("WebAuthn-Session", string(resp.Session))
	w.Header().Set("Content-Type:", "application/json")
	_, _ = w.Write(resp.Json)
}

func registerBegin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	u, _ := dao.Get(userId)

	resp, err := _client.BeginRegister(context.Background(), &gw6n.BeginRegisterRequest{
		User: &gw6n.User{
			Id: userId,
		},
		UserBlob: u.Blob,
	})
	if err != nil {
		httpError(w, err)
		return
	}

	w.Header().Set("WebAuthn-Session", string(resp.Session))
	w.Header().Set("Content-Type:", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write(resp.Json)
}

func registerFinish(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	u, _ := dao.Get(userId)

	session := []byte(r.Header.Get("WebAuthn-Session"))
	signature, err := ioutil.ReadAll(r.Body)
	userBlob := u.Blob

	if err != nil {
		httpError(w, err)
		return
	}
	resp, err := _client.FinishRegister(context.Background(), &gw6n.FinishRegisterRequest{
		UserBlob:  userBlob,
		Session:   session,
		Signature: signature,
	})

	if err != nil {
		httpError(w, err)
		return
	}

	dao.Set(dao.User{
		Id:   userId,
		Blob: resp.UserBlob,
	})

	w.WriteHeader(200)
}
