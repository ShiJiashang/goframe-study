package paymentclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"goframe-study/lessons/lesson19/paymentclient"
)

// 测试成功分支：payment 服务返回合法 JSON，客户端能解析出 tradeNo/status。
func TestClient_Pay_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pay" {
			t.Fatalf("path=%s want /pay", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("content-type=%s want application/json", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tradeNo":"T100","status":"paid"}`))
	}))
	defer server.Close()

	client := paymentclient.New(server.URL, time.Second)
	output, err := client.Pay(context.Background(), paymentclient.PayInput{
		OrderNo:    "O100",
		AmountCent: 9900,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if output.TradeNo != "T100" || output.Status != "paid" {
		t.Fatalf("output=%+v", output)
	}
}

// 测试 5xx 分支：payment 服务返回 502，客户端应返回带 status/body 的错误。
func TestClient_Pay_StatusNot2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":"upstream unavailable"}`))
	}))
	defer server.Close()

	client := paymentclient.New(server.URL, time.Second)
	_, err := client.Pay(context.Background(), paymentclient.PayInput{OrderNo: "O", AmountCent: 1})
	if err == nil {
		t.Fatalf("expected error for 502, got nil")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Fatalf("err should mention status=502: %v", err)
	}
}

// 测试非法 JSON 分支：HTTP 200 但 body 不是合法 JSON。
func TestClient_Pay_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not-json-at-all`))
	}))
	defer server.Close()

	client := paymentclient.New(server.URL, time.Second)
	_, err := client.Pay(context.Background(), paymentclient.PayInput{OrderNo: "O", AmountCent: 1})
	if err == nil {
		t.Fatalf("expected decode error, got nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("err should mention decode: %v", err)
	}
}

// 测试字段缺失分支：HTTP 200，JSON 合法，但缺 tradeNo。
func TestClient_Pay_MissingFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"paid"}`))
	}))
	defer server.Close()

	client := paymentclient.New(server.URL, time.Second)
	_, err := client.Pay(context.Background(), paymentclient.PayInput{OrderNo: "O", AmountCent: 1})
	if err == nil {
		t.Fatalf("expected missing field error, got nil")
	}
	if !strings.Contains(err.Error(), "tradeNo") {
		t.Fatalf("err should mention tradeNo: %v", err)
	}
}

// 测试超时分支：payment 服务故意休眠超过客户端 timeout，客户端应返回网络错误。
func TestClient_Pay_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tradeNo":"T","status":"paid"}`))
	}))
	defer server.Close()

	client := paymentclient.New(server.URL, 100*time.Millisecond)
	_, err := client.Pay(context.Background(), paymentclient.PayInput{OrderNo: "O", AmountCent: 1})
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "request payment service failed") {
		t.Fatalf("err should mention request payment: %v", err)
	}
}
