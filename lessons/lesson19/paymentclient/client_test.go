package paymentclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestClientPaySuccess(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pay" {
			t.Fatalf("expected path /pay, got %s", r.URL.Path)
		}
		var input PayInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PayOutput{TradeNo: "T-" + input.OrderNo, Status: "paid"})
	})

	client := New(server.URL, time.Second)
	out, err := client.Pay(context.Background(), PayInput{OrderNo: "ORD1", AmountCent: 100})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.TradeNo != "T-ORD1" || out.Status != "paid" {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestClientRefundSuccess(t *testing.T) {
	var receivedInput RefundInput
	server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/refund" {
			t.Fatalf("expected path /refund, got %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&receivedInput); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(RefundOutput{TradeNo: receivedInput.TradeNo, Status: "refunded"})
	})

	client := New(server.URL, time.Second)
	out, err := client.Refund(context.Background(), RefundInput{
		TradeNo:    "T-ORD1",
		OrderNo:    "ORD1",
		AmountCent: 100,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out.TradeNo != "T-ORD1" || out.Status != "refunded" {
		t.Fatalf("unexpected output: %+v", out)
	}
	if receivedInput.OrderNo != "ORD1" || receivedInput.AmountCent != 100 {
		t.Fatalf("unexpected received input: %+v", receivedInput)
	}
}

func TestClientRefundBadStatus(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	})

	client := New(server.URL, time.Second)
	_, err := client.Refund(context.Background(), RefundInput{
		TradeNo:    "T-FAIL",
		OrderNo:    "FAIL",
		AmountCent: 100,
	})
	if err == nil {
		t.Fatal("expected error for 502 status, got nil")
	}
}

func TestClientRefundInvalidJSON(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not-json"))
	})

	client := New(server.URL, time.Second)
	_, err := client.Refund(context.Background(), RefundInput{
		TradeNo:    "T",
		OrderNo:    "O",
		AmountCent: 1,
	})
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

func TestClientRefundMissingFields(t *testing.T) {
	server := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tradeNo":"","status":""}`))
	})

	client := New(server.URL, time.Second)
	_, err := client.Refund(context.Background(), RefundInput{
		TradeNo:    "T",
		OrderNo:    "O",
		AmountCent: 1,
	})
	if err == nil {
		t.Fatal("expected error for missing fields, got nil")
	}
}

func TestClientRefundNetworkFailure(t *testing.T) {
	client := New("http://127.0.0.1:1", 200*time.Millisecond)
	_, err := client.Refund(context.Background(), RefundInput{
		TradeNo:    "T",
		OrderNo:    "O",
		AmountCent: 1,
	})
	if err == nil {
		t.Fatal("expected network error, got nil")
	}
}
