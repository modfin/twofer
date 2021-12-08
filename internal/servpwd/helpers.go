package servpwd

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"hash"
	"twofer/grpc/gpwd"
)


func GetHmacDigest(password string, salt string, hashFunc hash.Hash, hashCount int) string {
	digest := []byte(password)
	saltBytes := []byte(salt)
	for i := 0; i < hashCount; i++ {
		hashFunc.Reset()
		hashFunc.Write(saltBytes)
		hashFunc.Write([]byte(password))
		digest = hashFunc.Sum(nil)
	}
	return Base64Encode(digest)
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
	case gpwd.Alg_SHA_256:
		return sha256.New()
	case gpwd.Alg_SHA_512:
		return sha512.New()
	}
	panic("unreached")
}
