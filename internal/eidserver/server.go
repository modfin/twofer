package eidserver

import (
	"fmt"
	"golang.org/x/net/context"
	"twofer/eid"
	"twofer/twoferrpc"
)

func New() *Server {
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

	for _, v := range s.List() {
		p := &twoferrpc.Provider{Name: v}
		prov.Providers = append(prov.Providers, p)
	}
	return prov, nil
}

func (s Server) AuthInit(ctx context.Context, req *twoferrpc.Req) (in *twoferrpc.Inter, err error) {
	cli, err := s.Get(req.Provider.Name)
	if err != nil {
		return
	}
	eidReq, err := eid.FromGrpcReq(req)
	if err != nil {
		return
	}
	authInit, err := cli.AuthInit(ctx, &eidReq)
	if err != nil {
		return
	}
	grpcInter, err := eid.ToGrpcInter(authInit, req.Provider.Name)
	if err != nil {
		return
	}
	return &grpcInter, nil
}

func (s Server) SignInit(ctx context.Context, req *twoferrpc.Req) (in *twoferrpc.Inter, err error) {
	fmt.Println("Someone's asking about things")
	cli, err := s.Get(req.Provider.Name)
	if err != nil {
		return
	}
	fmt.Printf("Using provider %s to do sign request\n", cli.Name())
	eidReq, err := eid.FromGrpcReq(req)
	if err != nil {
		return
	}
	fmt.Printf("EIDRequest%+v\n", eidReq)
	signInit, err := cli.SignInit(ctx, &eidReq)
	if err != nil {
		return
	}
	fmt.Printf("Sign Init: %+v\n", signInit)
	grpcInter, err := eid.ToGrpcInter(signInit, req.Provider.Name)
	if err != nil {
		return
	}
	return &grpcInter, nil
}

func (s Server) Collect(ctx context.Context, inter *twoferrpc.Inter) (r *twoferrpc.Resp, err error) {
	cli, err := s.Get(inter.Req.Provider.Name)
	if err != nil {
		return
	}
	eidInter, err := eid.FromGrpcInter(inter)
	if err != nil {
		return
	}
	collect, err := cli.Collect(ctx, &eidInter, false)
	if err != nil {
		return
	}
	grpcRes, err := eid.ToGrpcResp(collect)
	if err != nil {
		return
	}
	return &grpcRes, nil
}

func (s Server) Peek(ctx context.Context, inter *twoferrpc.Inter) (res *twoferrpc.Resp, err error) {
	cli, err := s.Get(inter.Req.Provider.Name)
	if err != nil {
		return
	}
	eidInter, err := eid.FromGrpcInter(inter)
	if err != nil {
		return
	}
	collect, err := cli.Peek(ctx, &eidInter)
	if err != nil {
		return
	}
	grpcRes, err := eid.ToGrpcResp(collect)
	if err != nil {
		return
	}
	return &grpcRes, nil
}

func (s Server) Cancel(context.Context, *twoferrpc.Inter) (*twoferrpc.Error, error) {
	panic("implement me")
}
