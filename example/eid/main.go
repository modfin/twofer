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
	"twofer/example/eid/dao"
	"twofer/grpc/geid"
	"twofer/grpc/gqr"
)

var (
	serverAddr = "twoferd:43210"
	_eidClient geid.EIDClient
	_qrQlient  gqr.QRClient
)

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

	m := mux.NewRouter()
	m.HandleFunc("/providers", getProviders)

	m.HandleFunc("/inferred", initInferred)
	m.HandleFunc("/start-auth", ssnAuthInit)
	m.HandleFunc("/start-sign", ssnSignInit)

	m.HandleFunc("/qrimage/{ref}", qrImage)
	m.HandleFunc("/peek/{ref}", peek)
	m.HandleFunc("/collect/{ref}", collect)
	m.HandleFunc("/cancel", cancel)

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
	res := AuthResult{Status: collect.Status.String()}
	ar, err := json.Marshal(res)
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

type AuthResult struct {
	Status string `json:"status"`
}
