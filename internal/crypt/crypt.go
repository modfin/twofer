package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Store interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type store struct {
	max  uint32
	keys map[uint32]*key
}

type key struct {
	version uint32
	block   cipher.Block
}

func New(keys []string) (Store, error) {
	s := store{
		max:  0,
		keys: map[uint32]*key{},
	}

	for _, k := range keys {
		parts := strings.Split(k, ":")
		if len(parts) != 3 {
			return nil, errors.New("a key shall consist of three portions. version:alg:key")
		}

		vv, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("could not parse version of key to an uint: %v", err)
		}
		rawKey, err := base64.StdEncoding.DecodeString(parts[2])
		if err != nil {
			return nil, fmt.Errorf("key is expected to be in base64 std encoding: %v", err)
		}

		version := uint32(vv)
		var block cipher.Block

		switch strings.ToLower(parts[1]) {
		case "aes":
			block, err = aes.NewCipher(rawKey)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("expected key type to be aes")
		}

		if _, ok := s.keys[version]; ok {
			return nil, fmt.Errorf("there seems to be duplicate version of key %d", version)
		}
		s.keys[version] = &key{
			version: version,
			block:   block,
		}
		if version > s.max {
			s.max = version
		}
	}

	return &s, nil
}

func (s *store) Encrypt(plaintext []byte) ([]byte, error) {
	if len(s.keys) == 0 {
		return nil, errors.New("no keys available")
	}

	block := s.keys[s.max].block
	version := s.max

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	ver := make([]byte, 4)
	binary.BigEndian.PutUint32(ver, version)

	return append(append(ver, nonce...), ciphertext...), nil
}

func (s *store) Decrypt(ciphertext []byte) ([]byte, error) {

	version := binary.BigEndian.Uint32(ciphertext[:4])
	key, ok := s.keys[version]
	if !ok {
		return nil, fmt.Errorf("could not find key %d for ciphertext", version)
	}

	aead, err := cipher.NewGCM(key.block)
	if err != nil {
		return nil, err
	}

	return aead.Open(nil, ciphertext[4:4+aead.NonceSize()], ciphertext[4+aead.NonceSize():], nil)
}
