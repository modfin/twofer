package eid

import (
	"fmt"
	"sort"
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

type Empty struct{}

func (e *EID) Add(provider Client) {
	e.providers[provider.Name()] = provider
}

func (e *EID) List() []string {
	var names []string
	for key := range e.providers {
		names = append(names, key)
	}
	sort.Strings(names)
	return names
}

func (e *EID) Get(name string) (Client, error) {
	c, found := e.providers[name]
	if !found {
		return nil, fmt.Errorf("could not find eid provider %s", name)
	}
	return c, nil
}
