package serveid

import (
	"fmt"
	"golang.org/x/net/context"
	"twofer/eid"
	"twofer/grpc/geid"
)

func New() *Server {
	return &Server{
		EID: eid.New(),
	}
}

type Server struct {
	*eid.EID
}

func (s Server) GetProviders(context.Context, *geid.Empty) (*geid.Providers, error) {

	fmt.Println("Request GetProviders")

	prov := &geid.Providers{}

	for _, v := range s.List() {
		p := &geid.Provider{Name: v}
		prov.Providers = append(prov.Providers, p)
	}
	return prov, nil
}

func (s Server) AuthInit(ctx context.Context, req *geid.Req) (in *geid.Inter, err error) {
	cli, err := s.Get(req.Provider.Name)
	if err != nil {
		return
	}
	eidReq, err := eid.FromGrpcReq(req, cli)
	if err != nil {
		return
	}
	authInit, err := cli.AuthInit(ctx, &eidReq)
	if err != nil {
		return
	}
	grpcInter, err := eid.ToGrpcInter(authInit)
	if err != nil {
		return
	}
	return &grpcInter, nil
}

func (s Server) SignInit(ctx context.Context, req *geid.Req) (in *geid.Inter, err error) {
	fmt.Println("Someone's asking about things")
	cli, err := s.Get(req.Provider.Name)
	if err != nil {
		return
	}
	fmt.Printf("Using provider %s to do sign request\n", cli.Name())
	eidReq, err := eid.FromGrpcReq(req, cli)
	if err != nil {
		return
	}
	signInit, err := cli.SignInit(ctx, &eidReq)
	if err != nil {
		return
	}
	grpcInter, err := eid.ToGrpcInter(signInit)
	if err != nil {
		return
	}
	return &grpcInter, nil
}

func (s Server) Collect(ctx context.Context, inter *geid.Inter) (r *geid.Resp, err error) {
	cli, err := s.Get(inter.Req.Provider.Name)
	if err != nil {
		return
	}
	eidInter, err := eid.FromGrpcInter(inter, cli)
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

func (s Server) Peek(ctx context.Context, inter *geid.Inter) (res *geid.Resp, err error) {
	cli, err := s.Get(inter.Req.Provider.Name)
	if err != nil {
		return
	}
	eidInter, err := eid.FromGrpcInter(inter, cli)
	if err != nil {
		return
	}
	peek, err := cli.Peek(ctx, &eidInter)
	if err != nil {
		return
	}
	grpcRes, err := eid.ToGrpcResp(peek)
	if err != nil {
		return
	}
	return &grpcRes, nil
}

func (s Server) Cancel(_ context.Context, inter *geid.Inter) (emp *geid.Empty, err error) {
	cli, err := s.Get(inter.Req.Provider.Name)
	if err != nil {
		return
	}
	eidCancel, err := eid.FromGrpcInter(inter, cli)
	if err != nil {
		return
	}
	err = cli.Cancel(&eidCancel)
	if err != nil {
		return
	}
	return &geid.Empty{}, nil
}
