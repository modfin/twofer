package dao

import "sync"

var mu sync.Mutex
var store = map[string]User{}

type User struct {
	Id   string
	Blob []byte
}

func Get(userId string) (u User, ok bool) {
	mu.Lock()
	defer mu.Unlock()
	u, ok = store[userId]
	return
}

func Set(user User) {
	mu.Lock()
	defer mu.Unlock()
	store[user.Id] = user
}
