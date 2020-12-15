package dao

import (
	"sync"
)

var mu sync.Mutex
var qrStore = map[string]QrData{}
var interStore = map[string][]byte{}

type QrData struct {
	Reference string
	Image     []byte
}

func GetQr(ref string) (qr QrData, ok bool) {
	mu.Lock()
	defer mu.Unlock()
	qr, ok = qrStore[ref]
	return
}

func SetQr(qr QrData) {
	mu.Lock()
	defer mu.Unlock()
	qrStore[qr.Reference] = qr
}

func GetInter(ref string) (i []byte, ok bool) {
	mu.Lock()
	defer mu.Unlock()
	i, ok = interStore[ref]
	return
}

func SetInter(ref string, i []byte) {
	mu.Lock()
	defer mu.Unlock()
	interStore[ref] = i
}
