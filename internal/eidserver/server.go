package eidserver

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"twofer/eid"
	"twofer/twoferrpc"
)

func New() *Server{
	return &Server{
		EID: eid.New(),
	}
}

type Server struct {
	*eid.EID
}

func (s Server) GetProviders(context.Context, *twoferrpc.Empty) (*twoferrpc.Providers, error) {

	fmt.Println("Request GetProviders")

	prov := &twoferrpc.Providers{}

	for _, v := range s.List(){
		p := &twoferrpc.Provider{Name: v}
		prov.Providers = append(prov.Providers, p)
	}
	return prov, nil
}

func (s Server) AuthInit(ctx context.Context, req *twoferrpc.Req) (in *twoferrpc.Inter, err error) {
	cli, err := s.Get(req.Provider.Name)
	if err != nil{
		return
	}

	d, err := json.Marshal(req)
	if err != nil{
		return
	}

	r := eid.Req{}

	err = json.Unmarshal(d, &r)
	if err != nil{
		return
	}

	i, err := cli.AuthInit(ctx, &r)
	if err != nil{
		return
	}
	d, err = json.Marshal(i)
	if err != nil{
		return
	}
	_ = json.Unmarshal(d, &in)
	//if err != nil{
	//	return
	//}
	in.Mode = twoferrpc.Inter_AUTH

	return
}

func (s Server) SignInit(ctx context.Context, req *twoferrpc.Req) (in *twoferrpc.Inter, err error) {
	cli, err := s.Get(req.Provider.Name)
	if err != nil{
		return
	}

	d, err := json.Marshal(req)
	if err != nil{
		return
	}

	r := eid.Req{}

	err = json.Unmarshal(d, &r)
	if err != nil{
		return
	}

	i, err := cli.AuthInit(ctx, &r)
	if err != nil{
		return
	}
	d, err = json.Marshal(i)
	if err != nil{
		return
	}
	err = json.Unmarshal(d, &in)
	if err != nil{
		return
	}
	in.Mode = twoferrpc.Inter_SIGN
	return
}

func (s Server) Collect(context.Context, *twoferrpc.Inter) (*twoferrpc.Resp, error) {
	panic("implement me")
}

func (s Server) Peek(context.Context, *twoferrpc.Inter) (*twoferrpc.Resp, error) {
	panic("implement me")
}

func (s Server) Cancel(context.Context, *twoferrpc.Inter) (*twoferrpc.Error, error) {
	panic("implement me")
}
