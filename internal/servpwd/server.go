package servpwd

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
	"github.com/modfin/twofer/grpc/gpwd"
	"github.com/modfin/twofer/internal/crypt"
)

type PWDConfig struct {
	DefaultAlg          gpwd.Alg
	DefaultHashCount    int
	DefaultBCryptCost   int
	DefaultSCryptN      int
	DefaultSCryptR      int
	DefaultSCryptP      int
	DefaultSCryptKeyLen int
}

type Server struct {
	store crypt.Store
	conf  PWDConfig
}

func New(conf PWDConfig, keys []string) (*Server, error) {
	s := &Server{conf: conf}
	var err error
	if len(keys) > 0 {
		s.store, err = crypt.New(keys)
		if err != nil {
			return nil, err
		}
	}

	if s.store == nil {
		s.store = &crypt.NilStore{}
	}
	fmt.Printf("	- Using PWD default alg: %s", conf.DefaultAlg)
	return s, nil
}

type wrapper struct {
	Digest      string   `json:"digest"`
	Salt        string   `json:"salt"`
	Alg         gpwd.Alg `json:"alg"`
	AlgMetadata []byte   `json:"alg_metadata"`
}

type shaMetadata struct {
	HashCount int `json:"hash_count"`
}

type bcryptMetadata struct {
	Cost int `json:"cost"`
}

type scryptMetadata struct {
	N      int `json:"default_s_crypt_n"`
	R      int `json:"default_s_crypt_r"`
	P      int `json:"default_s_crypt_p"`
	KeyLen int `json:"default_s_crypt_key_len"`
}

func (s *Server) Enroll(ctx context.Context, enReq *gpwd.EnrollReq) (*gpwd.Blob, error) {
	var o wrapper
	o.Alg = s.conf.DefaultAlg

	switch s.conf.DefaultAlg {
	case gpwd.Alg_SHA_256:
		fallthrough
	case gpwd.Alg_SHA_512:
		o.Salt = GenerateRandomBase64Bytes(32)
		o.Digest = GetHmacDigest(enReq.Password, o.Salt, Hash(s.conf.DefaultAlg), s.conf.DefaultHashCount)
		metadataBytes, err := json.Marshal(shaMetadata{HashCount: s.conf.DefaultHashCount})
		if err != nil {
			return nil, err
		}
		o.AlgMetadata = metadataBytes

	case gpwd.Alg_SCrypt:
		o.Salt = GenerateRandomBase64Bytes(32)
		dk, err := scrypt.Key([]byte(enReq.Password), []byte(o.Salt), s.conf.DefaultSCryptN, s.conf.DefaultSCryptR, s.conf.DefaultSCryptP, s.conf.DefaultSCryptKeyLen)
		if err != nil {
			return nil, err
		}
		o.Digest = Base64Encode(dk)
		metadataBytes, err := json.Marshal(scryptMetadata{
			N:      s.conf.DefaultSCryptN,
			R:      s.conf.DefaultSCryptR,
			P:      s.conf.DefaultSCryptP,
			KeyLen: s.conf.DefaultSCryptKeyLen,
		})
		if err != nil {
			return nil, err
		}
		o.AlgMetadata = metadataBytes
	case gpwd.Alg_BCrypt:
		dk, err := bcrypt.GenerateFromPassword([]byte(enReq.Password), s.conf.DefaultBCryptCost)
		if err != nil {
			return nil, err
		}
		o.Digest = Base64Encode(dk)
	}

	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	b, err = s.store.Encrypt(b)
	if err != nil {
		return nil, err
	}

	return &gpwd.Blob{UserBlob: Base64Encode(b)}, nil
}

func (s *Server) Auth(ctx context.Context, authReq *gpwd.AuthReq) (*gpwd.Res, error) {
	sec := Base64Decode(authReq.UserBlob)
	sec, err := s.store.Decrypt(sec)
	if err != nil {
		fmt.Println("decrypt error")
		return nil, err
	}

	var v wrapper
	err = json.Unmarshal(sec, &v)
	if err != nil {
		fmt.Println("Unmarshal error")
		return nil, err
	}

	var valid bool

	switch v.Alg {
	case gpwd.Alg_SHA_256:
		fallthrough
	case gpwd.Alg_SHA_512:
		var _shaMetadata shaMetadata
		err = json.Unmarshal(v.AlgMetadata, &_shaMetadata)
		if err != nil {
			return nil, err
		}
		newDigest := GetHmacDigest(authReq.Password, v.Salt, Hash(v.Alg), _shaMetadata.HashCount)
		valid = hmac.Equal(Base64Decode(newDigest), Base64Decode(v.Digest))
	case gpwd.Alg_SCrypt:
		var _scryptMetadata scryptMetadata
		err = json.Unmarshal(v.AlgMetadata, &_scryptMetadata)
		if err != nil {
			return nil, err
		}
		newDigest, err := scrypt.Key([]byte(authReq.Password), []byte(v.Salt), _scryptMetadata.N, _scryptMetadata.R, _scryptMetadata.P, _scryptMetadata.KeyLen)
		if err != nil {
			return nil, err
		}
		valid = hmac.Equal(newDigest, Base64Decode(v.Digest))

	case gpwd.Alg_BCrypt:
		err = bcrypt.CompareHashAndPassword(Base64Decode(v.Digest), []byte(authReq.Password))
		if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
			return nil, err
		}
		valid = err != bcrypt.ErrMismatchedHashAndPassword
	}

	return &gpwd.Res{Valid: valid, Message: "password processed"}, nil

}

func (s *Server) Upgrade(_ context.Context, req *gpwd.Blob) (*gpwd.Blob, error) {
	sec := Base64Decode(req.UserBlob)
	sec, err := s.store.Decrypt(sec)
	if err != nil {
		fmt.Println("decrypt error")
		return nil, err
	}
	b, err := s.store.Encrypt(sec)
	if err != nil {
		return nil, err
	}
	return &gpwd.Blob{UserBlob: Base64Encode(b)}, nil
}
