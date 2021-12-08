package servpwd

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
	"twofer/grpc/gpwd"
)


func GetHmacDigest(password string, salt string, hashFunc hash.Hash, hashCount uint) string {
	digest := password
	for i := uint(0); i < hashCount; i++ {
		hashFunc.Reset()
		hashFunc.Write(Base64Decode(salt))
		hashFunc.Write([]byte(password))
		bMac := hashFunc.Sum(nil)
		digest = Base64Encode(bMac)
	}
	return digest
}

func Base64Encode(b64 []byte) string {
	return base64.StdEncoding.EncodeToString(b64)
}

func Base64Decode(b64 string) []byte {
	b, _ := base64.StdEncoding.DecodeString(b64)
	return b
}

func GenerateRandomBase64Bytes(bytes int) string {
	b := make([]byte, bytes)
	rand.Read(b)
	return Base64Encode(b)
}

func Hash(a gpwd.Alg) hash.Hash {
	switch a {
	case gpwd.Alg_SHA_1:
		return sha1.New()
	case gpwd.Alg_SHA_256:
		return sha256.New()
	case gpwd.Alg_SHA_512:
		return sha512.New()
	}
	panic("unreached")
}
