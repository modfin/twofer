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
	"twofer/twoferrpc"
)

var (
	serverAddr = "twoferd:43210"
	_client    twoferrpc.WebauthnClient
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
	_client = twoferrpc.NewWebauthnClient(conn)

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

	_, err = _client.FinishLogin(context.Background(), &twoferrpc.FinishLoginRequest{
		Session:       session,
		UserBlob:      userBlob,
		UserSignature: signature,
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
	resp, err := _client.BeginLogin(context.Background(), &twoferrpc.BeginLoginRequest{
		UserBlob: u.Blob,
	})
	if err != nil {
		httpError(w, err)
		return
	}

	w.Header().Set("WebAuthn-Session", string(resp.Session))
	w.Header().Set("Content-Type:", "application/json")
	_, _ = w.Write(resp.Response2User)
}

func registerBegin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	u, _ := dao.Get(userId)

	resp, err := _client.BeginRegister(context.Background(), &twoferrpc.BeginRegisterRequest{
		UserId:   userId,
		UserBlob: u.Blob,
	})
	if err != nil {
		httpError(w, err)
		return
	}

	w.Header().Set("WebAuthn-Session", string(resp.Session))
	w.Header().Set("Content-Type:", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write(resp.Response2User)
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
	resp, err := _client.FinishRegister(context.Background(), &twoferrpc.FinishRegisterRequest{
		UserBlob:      userBlob,
		Session:       session,
		UserSignature: signature,
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
