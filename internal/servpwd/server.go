package servpwd

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"twofer/grpc/gpwd"
	"twofer/internal/crypt"
)

type PWDConfig struct {
	DefaultAlg       gpwd.Alg
	DefaultHashCount uint
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
	return s, nil
}

type wrapper struct {
	Digest    string   `json:"digest"`
	Salt      string   `json:"salt"`
	Alg       gpwd.Alg `json:"alg"`
	HashCount uint     `json:"hash_count"`
}

func (s *Server) Enroll(ctx context.Context, enReq *gpwd.EnrollReq) (*gpwd.Blob, error) {
	salt := GenerateRandomBase64Bytes(32)
	digest := GetHmacDigest(enReq.Password, salt, Hash(s.conf.DefaultAlg), s.conf.DefaultHashCount)

	o := wrapper{
		Digest:    digest,
		Salt:      salt,
		Alg:       s.conf.DefaultAlg,
		HashCount: s.conf.DefaultHashCount,
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

	newDigest := GetHmacDigest(authReq.Password, v.Salt, Hash(v.Alg), v.HashCount)
	valid := hmac.Equal(Base64Decode(newDigest), Base64Decode(v.Digest))

	return &gpwd.Res{Valid: valid, Message: "password processed"}, nil

}

func (s *Server) Upgrade(context.Context, *gpwd.Blob) (*gpwd.Blob, error) {
	panic("implement me")
}
