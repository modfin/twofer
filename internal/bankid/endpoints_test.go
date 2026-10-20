package bankid

import (
	"context"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/modfin/twofer/bankidtest"
)

// The v1 auth and sign handlers stop reading the watch channel after their last
// QR frame, so the watcher has to end without a receiver.
func TestWatchForChange_ends_without_a_receiver(t *testing.T) {
	fake := bankidtest.New()
	defer fake.Close()
	api := NewAPI(http.DefaultClient, fake.URL(), 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	res, err := api.Auth(ctx, &AuthSignRequest{EndUserIp: "127.0.0.1", Requirement: Requirement{PersonalNumber: "199001010099"}})
	if err != nil {
		t.Fatalf("auth: %v", err)
	}

	_ = api.WatchForChange(ctx, res.OrderRef)
	cancel()

	deadline := time.Now().Add(2 * time.Second)
	for watcherRunning() {
		if time.Now().After(deadline) {
			t.Fatal("WatchForChange goroutine still running after ctx was cancelled")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func watcherRunning() bool {
	buf := make([]byte, 1<<20)
	return strings.Contains(string(buf[:runtime.Stack(buf, true)]), "WatchForChange.func1")
}
