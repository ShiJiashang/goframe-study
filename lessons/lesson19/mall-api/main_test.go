package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"goframe-study/lessons/lesson19/paymentclient"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func resetState() {
	mu.Lock()
	defer mu.Unlock()
	orders = map[string]*order{}
	callbacks = map[string]bool{}
}

var portCounter int32 = 18009

func allocPort() int {
	return int(atomic.AddInt32(&portCounter, 1))
}

func startMallServer(t *testing.T, paymentBaseURL string) string {
	t.Helper()
	resetState()
	mu.Lock()
	orders["1"] = &order{ID: "1", OrderNo: "O202607140001", AmountCent: 9900, Status: "pending"}
	orders["2"] = &order{ID: "2", OrderNo: "O202607140002", AmountCent: 19900, Status: "pending"}
	orders["3"] = &order{ID: "3", OrderNo: "O202607140003", AmountCent: 5000, Status: "pending"}
	mu.Unlock()

	t.Setenv("PAYMENT_BASEURL", paymentBaseURL)

	port := allocPort()
	server := g.Server(fmt.Sprintf("mall-test-%d", port))
	server.SetPort(port)
	server.SetDumpRouterMap(false)
	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.GET("/orders/:id", handleGetOrder)
		group.POST("/orders/:id/pay", handlePayOrder)
		group.POST("/orders/:id/refund", handleRefundOrder)
		group.POST("/payments/callback", handlePaymentCallback)
	})
	go func() {
		_ = server.Start()
	}()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/api/orders/1")
		if err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Cleanup(func() {
		_ = server.Shutdown()
	})
	return baseURL
}

type fakePayment struct {
	server   *httptest.Server
	payCalls atomic.Int32
	refCalls atomic.Int32
}

func newFakePayment(t *testing.T) *fakePayment {
	t.Helper()
	fp := &fakePayment{}
	mux := http.NewServeMux()
	mux.HandleFunc("/pay", func(w http.ResponseWriter, r *http.Request) {
		fp.payCalls.Add(1)
		var req struct {
			OrderNo    string `json:"orderNo"`
			AmountCent int64  `json:"amountCent"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if req.OrderNo == "FAIL" {
			http.Error(w, "fail", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tradeNo": "MOCK-" + req.OrderNo,
			"status":  "paid",
		})
	})
	mux.HandleFunc("/refund", func(w http.ResponseWriter, r *http.Request) {
		fp.refCalls.Add(1)
		var req struct {
			TradeNo    string `json:"tradeNo"`
			OrderNo    string `json:"orderNo"`
			AmountCent int64  `json:"amountCent"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if req.OrderNo == "FAIL" {
			http.Error(w, "fail", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tradeNo": req.TradeNo,
			"status":  "refunded",
		})
	})
	fp.server = httptest.NewServer(mux)
	t.Cleanup(fp.server.Close)
	return fp
}

type response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func httpPost(t *testing.T, url string) response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out response
	if err = json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode response: %v body=%s", err, string(body))
	}
	return out
}

func httpGet(t *testing.T, url string) response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out response
	if err = json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode response: %v body=%s", err, string(body))
	}
	return out
}

func TestRefundFullFlow(t *testing.T) {
	fp := newFakePayment(t)
	mallURL := startMallServer(t, fp.server.URL)

	res := httpPost(t, mallURL+"/api/orders/1/pay")
	if res.Code != 0 {
		t.Fatalf("pay failed: %+v", res)
	}
	if fp.payCalls.Load() != 1 {
		t.Fatalf("expected 1 pay call, got %d", fp.payCalls.Load())
	}

	getRes := httpGet(t, mallURL+"/api/orders/1")
	if !strings.Contains(string(getRes.Data), `"status":"paid"`) {
		t.Fatalf("expected paid status, data=%s", string(getRes.Data))
	}

	refundRes := httpPost(t, mallURL+"/api/orders/1/refund")
	if refundRes.Code != 0 {
		t.Fatalf("refund failed: %+v", refundRes)
	}
	if fp.refCalls.Load() != 1 {
		t.Fatalf("expected 1 refund call, got %d", fp.refCalls.Load())
	}

	getRes = httpGet(t, mallURL+"/api/orders/1")
	if !strings.Contains(string(getRes.Data), `"status":"refunded"`) {
		t.Fatalf("expected refunded status, data=%s", string(getRes.Data))
	}

	dupRes := httpPost(t, mallURL+"/api/orders/1/refund")
	if dupRes.Code != 0 {
		t.Fatalf("repeat refund unexpected code: %+v", dupRes)
	}
	if !strings.Contains(dupRes.Message, "已经退款") {
		t.Fatalf("expected '已经退款' message, got %q", dupRes.Message)
	}
	if fp.refCalls.Load() != 1 {
		t.Fatalf("expected still 1 refund call after duplicate, got %d", fp.refCalls.Load())
	}
}

func TestRefundPendingRejected(t *testing.T) {
	fp := newFakePayment(t)
	mallURL := startMallServer(t, fp.server.URL)

	res := httpPost(t, mallURL+"/api/orders/2/refund")
	if res.Code == 0 {
		t.Fatalf("expected non-zero code for pending refund, got %+v", res)
	}
	if fp.refCalls.Load() != 0 {
		t.Fatalf("expected 0 refund calls for pending order, got %d", fp.refCalls.Load())
	}
}

func TestRefundFAILProducesControlledError(t *testing.T) {
	fp := newFakePayment(t)
	mallURL := startMallServer(t, fp.server.URL)

	payRes := httpPost(t, mallURL+"/api/orders/3/pay")
	if payRes.Code != 0 {
		t.Fatalf("pay failed: %+v", payRes)
	}

	res := httpPost(t, mallURL+"/api/orders/3/refund?mockOrderNo=FAIL")
	if res.Code == 0 {
		t.Fatalf("expected non-zero code for FAIL refund, got %+v", res)
	}
	if res.Message == "" {
		t.Fatal("expected non-empty error message")
	}
	getRes := httpGet(t, mallURL+"/api/orders/3")
	if !strings.Contains(string(getRes.Data), `"status":"paid"`) {
		t.Fatalf("expected order to remain paid after FAIL refund, data=%s", string(getRes.Data))
	}
}

func TestPaymentClientRefundIntegration(t *testing.T) {
	fp := newFakePayment(t)
	client := paymentclient.New(fp.server.URL, time.Second)

	out, err := client.Refund(context.Background(), paymentclient.RefundInput{
		TradeNo:    "MOCK-ORD",
		OrderNo:    "ORD",
		AmountCent: 100,
	})
	if err != nil {
		t.Fatalf("refund err: %v", err)
	}
	if out.Status != "refunded" {
		t.Fatalf("expected refunded status, got %s", out.Status)
	}
}
