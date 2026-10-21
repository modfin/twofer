package bankid

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modfin/twofer/bankidtest"
)

const (
	pendingReply  = `{"orderRef":"ref-1","status":"pending","hintCode":"outstandingTransaction"}`
	completeReply = `{"orderRef":"ref-1","status":"complete","completionData":{"user":{"personalNumber":"199001010099"}}}`
	// a poll interval no test waits out
	noPoll = time.Hour
)

// bankidStub answers BankID collect with reply(n) for the n:th call.
func bankidStub(t *testing.T, reply func(n int32) string) *httptest.Server {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(reply(calls.Add(1))))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func fakeOrder(t *testing.T, poll time.Duration) (*bankidtest.BankID, *API, string) {
	t.Helper()
	fake := bankidtest.New()
	t.Cleanup(fake.Close)
	api := NewAPI(http.DefaultClient, fake.URL(), poll)
	res, err := api.Auth(context.Background(), &AuthSignRequest{EndUserIp: "127.0.0.1"})
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	return fake, api, res.OrderRef
}

// The v1 auth and sign handlers stop reading the watch channel after their last
// QR frame, so the watcher has to complete its send without a receiver.
func TestWatchForChange_ends_without_a_receiver(t *testing.T) {
	_, api, ref := fakeOrder(t, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	watch := api.WatchForChange(ctx, ref)
	cancel()

	deadline := time.Now().Add(2 * time.Second)
	for len(watch) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("watcher never completed its send without a receiver")
		}
		time.Sleep(10 * time.Millisecond)
	}
	<-watch
	if _, ok := <-watch; ok {
		t.Error("watch channel not closed after its only send")
	}
}

func TestChangeV3_returns_on_a_hint_change(t *testing.T) {
	fake, api, ref := fakeOrder(t, 10*time.Millisecond)

	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = fake.HintOrder(ref, "userSign")
	}()
	res, err := api.ChangeV3(context.Background(), &ChangeRequest{OrderRef: ref}, 0)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if res.Status != Pending || res.HintCode != "userSign" {
		t.Errorf("got %s/%s, want pending/userSign", res.Status, res.HintCode)
	}
}

func TestChangeV3_until_finished_waits_out_hint_changes(t *testing.T) {
	fake, api, ref := fakeOrder(t, 10*time.Millisecond)

	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = fake.HintOrder(ref, "userSign")
		time.Sleep(50 * time.Millisecond)
		_ = fake.CompleteOrder(ref, bankidtest.By("199001010099"))
	}()
	res, err := api.ChangeV3(context.Background(), &ChangeRequest{OrderRef: ref, WaitUntilFinished: true}, 0)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if res.Status != Complete {
		t.Errorf("got %s/%s, want complete", res.Status, res.HintCode)
	}
}

// v1 Change shares the loop: with onFinished it must not return on a hint change.
func TestChange_until_finished_waits_out_hint_changes(t *testing.T) {
	fake, api, ref := fakeOrder(t, 10*time.Millisecond)

	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = fake.HintOrder(ref, "userSign")
		time.Sleep(50 * time.Millisecond)
		_ = fake.CompleteOrder(ref, bankidtest.By("199001010099"))
	}()
	res, err := api.Change(context.Background(), &ChangeRequest{OrderRef: ref, WaitUntilFinished: true})
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if res.Status != Complete {
		t.Errorf("got %s/%s, want complete", res.Status, res.HintCode)
	}
}

// The reply at the bound is the current state, not the last polled one.
func TestChangeV3_max_wait_returns_the_current_state(t *testing.T) {
	fake, api, ref := fakeOrder(t, noPoll)

	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = fake.HintOrder(ref, "userSign")
	}()
	start := time.Now()
	res, err := api.ChangeV3(context.Background(), &ChangeRequest{OrderRef: ref, WaitUntilFinished: true}, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if res.Status != Pending || res.HintCode != "userSign" {
		t.Errorf("got %s/%s, want pending/userSign", res.Status, res.HintCode)
	}
	if took := time.Since(start); took > time.Second {
		t.Errorf("returned after %v, want about the 200ms max wait", took)
	}
}

func TestChangeV3_negative_max_wait_returns_at_once(t *testing.T) {
	_, api, ref := fakeOrder(t, noPoll)

	start := time.Now()
	res, err := api.ChangeV3(context.Background(), &ChangeRequest{OrderRef: ref, WaitUntilFinished: true}, -time.Second)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if res.Status != Pending {
		t.Errorf("got status %s, want pending", res.Status)
	}
	if took := time.Since(start); took > time.Second {
		t.Errorf("returned after %v, want at once", took)
	}
}

// The bound counts from the request, so a slow first collect is part of it.
func TestChangeV3_max_wait_includes_the_first_collect(t *testing.T) {
	srv := bankidStub(t, func(n int32) string {
		if n == 1 {
			time.Sleep(300 * time.Millisecond)
		}
		return pendingReply
	})
	api := NewAPI(http.DefaultClient, srv.URL, noPoll)

	start := time.Now()
	_, err := api.ChangeV3(context.Background(), &ChangeRequest{OrderRef: "ref-1", WaitUntilFinished: true}, 300*time.Millisecond)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if took := time.Since(start); took > 500*time.Millisecond {
		t.Errorf("returned after %v, want about the 300ms max wait", took)
	}
}

// A poll still running at the deadline may carry the order's only finished
// state, so it must not be cut off.
func TestChangeV3_max_wait_keeps_the_poll_in_flight(t *testing.T) {
	srv := bankidStub(t, func(n int32) string {
		if n == 2 {
			time.Sleep(200 * time.Millisecond)
			return completeReply
		}
		return pendingReply
	})
	api := NewAPI(http.DefaultClient, srv.URL, 10*time.Millisecond)

	res, err := api.ChangeV3(context.Background(), &ChangeRequest{OrderRef: "ref-1", WaitUntilFinished: true}, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if res.Status != Complete {
		t.Errorf("got status %s, want complete", res.Status)
	}
}

// BankID hands out a finished order once; a caller that went away before
// receiving it can still collect it.
func TestCollect_keeps_a_finished_order(t *testing.T) {
	fake, api, ref := fakeOrder(t, noPoll)
	if err := fake.CompleteOrder(ref, bankidtest.By("199001010099")); err != nil {
		t.Fatal(err)
	}

	for i := range 2 {
		res, err := api.Collect(context.Background(), &CollectRequest{OrderRef: ref})
		if err != nil {
			t.Fatalf("collect %d: %v", i, err)
		}
		if res.Status != Complete {
			t.Errorf("collect %d: got status %s, want complete", i, res.Status)
		}
	}
}

// A caller cut while the completing poll is in flight must not lose the order.
func TestCollect_survives_a_cut_caller(t *testing.T) {
	srv := bankidStub(t, func(n int32) string {
		if n == 1 {
			time.Sleep(200 * time.Millisecond)
			return completeReply
		}
		return `{"errorCode":"invalidParameters","details":"No such order"}`
	})
	api := NewAPI(http.DefaultClient, srv.URL, noPoll)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, _ = api.Collect(ctx, &CollectRequest{OrderRef: "ref-1"})

	res, err := api.Collect(context.Background(), &CollectRequest{OrderRef: "ref-1"})
	if err != nil {
		t.Fatalf("collect after the cut: %v", err)
	}
	if res.Status != Complete {
		t.Errorf("got status %s, want complete", res.Status)
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) == false {
		t.Error("the first collect was expected to outlive its caller")
	}
}
