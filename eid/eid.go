package eid

import (
	"fmt"
)

func New() *EID {
	e := EID{
		providers: map[string]Client{},
	}
	return &e
}

type EID struct {
	providers map[string]Client
}

func (e *EID) Add(provider ToEID){
	c := provider.EID()
	e.providers[c.Name()] = c
}


func (e *EID) Get(name string) (Client, error){
	c := e.providers[name]
	if c == nil{
		return nil, fmt.Errorf("could not find eid provider %s", name)
	}
	return c, nil
}












