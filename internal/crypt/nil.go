package crypt

type NilStore struct {
}

func (n NilStore) Encrypt(plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

func (n NilStore) Decrypt(ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}
