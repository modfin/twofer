package eid

import (
	"fmt"
	"sort"
	"strings"
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

func (e *EID) Add(provider ToEID) {
	c := provider.EID()
	e.providers[c.Name()] = c
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
	c := e.providers[name]
	if c == nil {
		return nil, fmt.Errorf("could not find eid provider %s", name)
	}
	return c, nil
}
func (e *EID) GetFromRef(ref string) (Client, error) {
	parts := strings.SplitN(ref, ":", 2)
	return e.Get(parts[0])
}
