package bankid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync/atomic"
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

// collectStub answers BankID collect with a pending order, except call number
// completeOn, which takes delay and completes it.
func collectStub(t *testing.T, completeOn int32, delay time.Duration) *httptest.Server {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if calls.Add(1) != completeOn {
			_, _ = w.Write([]byte(`{"orderRef":"ref-1","status":"pending","hintCode":"outstandingTransaction"}`))
			return
		}
		time.Sleep(delay)
		_, _ = w.Write([]byte(`{"orderRef":"ref-1","status":"complete","completionData":{"user":{"personalNumber":"199001010099"}}}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestChangeV3_max_wait_returns_the_current_state(t *testing.T) {
	srv := collectStub(t, -1, 0)
	api := NewAPI(http.DefaultClient, srv.URL, 10*time.Millisecond)

	start := time.Now()
	res, err := api.ChangeV3(context.Background(), &ChangeRequest{OrderRef: "ref-1", WaitUntilFinished: true}, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if res.Status != Pending {
		t.Errorf("got status %s, want pending", res.Status)
	}
	if took := time.Since(start); took > time.Second {
		t.Errorf("returned after %v, want about the 50ms max wait", took)
	}
}

// A poll still running at the deadline may carry the order's only finished
// state, so it must not be cut off.
func TestChangeV3_max_wait_keeps_the_poll_in_flight(t *testing.T) {
	srv := collectStub(t, 2, 200*time.Millisecond)
	api := NewAPI(http.DefaultClient, srv.URL, 10*time.Millisecond)

	res, err := api.ChangeV3(context.Background(), &ChangeRequest{OrderRef: "ref-1", WaitUntilFinished: true}, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if res.Status != Complete {
		t.Errorf("got status %s, want complete", res.Status)
	}
}
