package crypt

import (
	"encoding/binary"
	"testing"
)

func TestStore(t *testing.T) {

	store, err := New([]string{
		"1:aes:uTdWcGl+cOnIgHoGnuBF3w==",
	})

	if err != nil {
		t.Fatalf("crypt.New(...) returned error: %v", err)
	}

	encrypted, err := store.Encrypt([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	plain, err := store.Decrypt(encrypted)

	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != "hello" {
		t.Log(`expected "hello", got `, string(plain))
		t.FailNow()
	}

	store, err = New([]string{
		"1:aes:uTdWcGl+cOnIgHoGnuBF3w==",
		"2:aes:vqs/8Sk7H2cpzHvd3lLPn8lOK/j3g/8s",
	})
	if err != nil {
		t.Fatal(err)
	}

	plain, err = store.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != "hello" {
		t.Log(`expected "hello", got `, string(plain))
		t.FailNow()
	}

	encrypted, err = store.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}

	if binary.BigEndian.Uint32(encrypted[:4]) != 2 {
		t.Log(`expected "2", got `, string(plain))
		t.FailNow()
	}

	store, err = New([]string{
		"1:aes:uTdWcGl+cOnIgHoGnuBF3w==",
		"2:aes:vqs/8Sk7H2cpzHvd3lLPn8lOK/j3g/8s",
		"3:aes:atFSKGC+DOD7+WOF/OLordrPbNVIHQvNnMkcRC2qEvI=",
	})
	if err != nil {
		t.Fatal(err)
	}

	plain, err = store.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != "hello" {
		t.Log(`expected "hello", got `, string(plain))
		t.FailNow()
	}

	encrypted, err = store.Encrypt(plain)
	if err != nil {
		t.Fatal(err)
	}

	if binary.BigEndian.Uint32(encrypted[:4]) != 3 {
		t.Log(`expected "3", got `, string(plain))
		t.FailNow()
	}

	plain, err = store.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if string(plain) != "hello" {
		t.Log(`expected "hello", got `, string(plain))
		t.FailNow()
	}
}
