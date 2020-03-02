package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net/http"
	"twofer/example/samplem"
	"twofer/twoferrpc"
)

var (
	serverAddr                = "127.0.0.1:43210"
	_client                   twoferrpc.WebauthnClient
	_userCache                = make(map[string]*twoferrpc.UserInfo)
	_registrationSessionStore = make(map[string]*twoferrpc.SessionData)
	_loginSessionStore        = make(map[string]*twoferrpc.SessionData)
)

func main() {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	fmt.Println("Dialing grpc..")
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	_client = twoferrpc.NewWebauthnClient(conn)

	m := mux.NewRouter()
	m.HandleFunc("/register/{userId}", registerBegin).Methods("GET")
	m.HandleFunc("/register/finish/{userId}", registerFinish).Methods("POST")
	m.HandleFunc("/login/begin/{userId}", loginBegin).Methods("GET")
	m.HandleFunc("/login/finish/{userId}", loginFinish).Methods("POST")
	m.HandleFunc("/users", getUsers).Methods("GET")
	m.HandleFunc("/sessions", getSessions).Methods("GET")

	http.Handle("/", m)
	_ = http.ListenAndServe(":8080", nil)
}

func loginFinish(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	userId := vars["userId"]
	enableCors(&writer)
	tu := _userCache[userId]
	session := _loginSessionStore[userId]
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(400)
		return
	}
	loginRequest := twoferrpc.FinishLoginRequest{
		User:    tu,
		Session: session,
		Blob:    string(body),
	}
	_, err = _client.FinishLogin(context.Background(), &loginRequest)
	if err != nil {
		writer.WriteHeader(500)
		return
	}
	writer.WriteHeader(200)
	return
}

func loginBegin(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	userId := vars["userId"]
	enableCors(&writer)
	tu := _userCache[userId]
	login, err := _client.BeginLogin(context.Background(), &twoferrpc.BeginLoginRequest{
		User: tu,
	})
	if err != nil {
		return
	}
	_loginSessionStore[userId] = login.SessionData
	_userCache[userId] = login.User

	raw := json.RawMessage(login.PublicKey)
	response := samplem.LoginResponse{PublicKey: raw}
	fullResponse, _ := json.Marshal(response)

	_, _ = writer.Write(fullResponse)
}

func getUsers(writer http.ResponseWriter, _ *http.Request) {
	users, _ := json.Marshal(_userCache)
	_, _ = writer.Write(users)
}

func getSessions(writer http.ResponseWriter, _ *http.Request) {
	users, _ := json.Marshal(_registrationSessionStore)
	_, _ = writer.Write(users)
}

func registerBegin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	enableCors(&w)
	var tu = &twoferrpc.UserInfo{}
	tu = _userCache[userId]
	if tu == nil {
		tu = &twoferrpc.UserInfo{
			Id:                 userId,
			Name:               "Name",
			DisplayName:        "Display Me",
			AllowedCredentials: nil,
		}
	}
	register, err := _client.BeginRegister(context.Background(), &twoferrpc.BeginRegisterRequest{User: tu})
	if err != nil {
		w.WriteHeader(500)
		return
	}
	_registrationSessionStore[userId] = register.SessionData
	_userCache[userId] = register.User

	message := json.RawMessage(register.PublicKey)
	cr := samplem.RegisterResponse{PublicKey: message}

	response, err := json.Marshal(cr)
	if err != nil {
		return
	}
	w.WriteHeader(200)
	_, _ = w.Write(response)
}

func registerFinish(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	enableCors(&w)
	u := _userCache[userId]
	sessionData := _registrationSessionStore[userId]
	delete(_registrationSessionStore, userId)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	request := &twoferrpc.FinishRegisterRequest{
		SessionData: sessionData,
		User:        u,
		Blob:        string(body),
	}
	register, err := _client.FinishRegister(context.Background(), request)
	if err != nil {
		return
	}
	_userCache[userId] = register.User
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
