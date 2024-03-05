package serveid

import (
	"context"
	"fmt"
	"github.com/modfin/twofer/internal/eid"
)

func New() *Server {
	return &Server{
		EID: eid.New(),
	}
}

type Server struct {
	*eid.EID
}

func (s *Server) Get(provider string) (eid.Client, error) {
	return s.EID.Get(provider)
}

func (s *Server) GetProviders() ([]eid.Provider, error) {
	prov := make([]eid.Provider, 0)

	for _, v := range s.List() {
		p := eid.Provider{Name: v}
		prov = append(prov, p)
	}
	return prov, nil
}

func (s Server) AuthInit(ctx context.Context, req *eid.Req) (*eid.Inter, error) {
	cli, err := s.Get(req.Provider.Name)
	if err != nil {
		return nil, err
	}

	return cli.AuthInit(ctx, req)
}

func (s Server) SignInit(ctx context.Context, req *eid.Req) (*eid.Inter, error) {
	fmt.Println("Someone's asking about things")
	cli, err := s.Get(req.Provider.Name)
	if err != nil {
		return nil, err
	}

	return cli.SignInit(ctx, req)
}

func (s Server) Collect(ctx context.Context, inter *eid.Inter) (*eid.Resp, error) {
	cli, err := s.Get(inter.Req.Provider.Name)
	if err != nil {
		return nil, err
	}

	return cli.Collect(ctx, inter, false)
}

func (s Server) Change(ctx context.Context, inter *eid.Inter) (*eid.Resp, error) {
	cli, err := s.Get(inter.Req.Provider.Name)
	if err != nil {
		return nil, err
	}

	return cli.Change(ctx, inter, false)
}

func (s Server) Peek(ctx context.Context, inter *eid.Inter) (*eid.Resp, error) {
	cli, err := s.Get(inter.Req.Provider.Name)
	if err != nil {
		return nil, err
	}

	return cli.Peek(ctx, inter)
}

func (s Server) Cancel(ctx context.Context, inter *eid.Inter) (*eid.Empty, error) {
	cli, err := s.Get(inter.Req.Provider.Name)
	if err != nil {
		return nil, err
	}

	err = cli.Cancel(ctx, inter)
	if err != nil {
		return nil, err
	}

	return &eid.Empty{}, nil
}
