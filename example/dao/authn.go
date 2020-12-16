package dao

import "sync"

var mu2 sync.Mutex
var store = map[string]User{}

type User struct {
	Id   string
	Blob []byte
}

func Get(userId string) (u User, ok bool) {
	mu2.Lock()
	defer mu2.Unlock()
	u, ok = store[userId]
	return
}

func Set(user User) {
	mu2.Lock()
	defer mu2.Unlock()
	store[user.Id] = user
}
