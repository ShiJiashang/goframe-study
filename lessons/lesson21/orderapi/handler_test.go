package orderapi_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"goframe-study/lessons/lesson19/paymentclient"
	"goframe-study/lessons/lesson21/orderapi"
)

// fakePaymentClient 是 orderapi.PaymentClient 的测试替身。
// PayFn 允许每个测试自定义 Pay 的行为；calls 用来断言"被调用次数"。
type fakePaymentClient struct {
	PayFn func(ctx context.Context, input paymentclient.PayInput) (*paymentclient.PayOutput, error)
	calls atomic.Int32
}

func (f *fakePaymentClient) Pay(ctx context.Context, input paymentclient.PayInput) (*paymentclient.PayOutput, error) {
	f.calls.Add(1)
	if f.PayFn == nil {
		return nil, orderapi.ErrPaymentFailed
	}
	return f.PayFn(ctx, input)
}

func (f *fakePaymentClient) Calls() int {
	return int(f.calls.Load())
}

// memStore 是内存版 OrderStore，供 handler 测试使用。
type memStore struct {
	mu     sync.Mutex
	orders map[string]*orderapi.Order
}

func newMemStore(seed ...*orderapi.Order) *memStore {
	s := &memStore{orders: map[string]*orderapi.Order{}}
	for _, o := range seed {
		s.orders[o.ID] = o
	}
	return s
}

func (m *memStore) Get(id string) (*orderapi.Order, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	return o, ok
}

func (m *memStore) Save(o *orderapi.Order) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[o.ID] = o
}

// startTestServer 起一个短生命周期 GoFrame 服务器绑定 handler，返回 baseURL。
func startTestServer(t *testing.T, path string, handler func(r *ghttp.Request)) (string, func()) {
	t.Helper()

	s := g.Server(t.Name())
	s.SetPort(0) // 让内核分配空闲端口
	s.SetDumpRouterMap(false)
	s.BindHandler(path, handler)
	if err := s.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}

	baseURL := "http://127.0.0.1:" + intToStr(s.GetListenedPort())
	waitReady(t, baseURL)
	return baseURL, func() { _ = s.Shutdown() }
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func waitReady(t *testing.T, baseURL string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/__ping")
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func mustPOST(t *testing.T, url string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, string(body))
	}
	return resp.StatusCode, out
}

// —— 成功路径：pending -> paid，Pay 被调用 1 次，store 状态更新为 paid。
func TestHandlePay_Success(t *testing.T) {
	store := newMemStore(&orderapi.Order{
		ID: "1", OrderNo: "O1", AmountCent: 9900, Status: "pending",
	})
	fake := &fakePaymentClient{
		PayFn: func(ctx context.Context, in paymentclient.PayInput) (*paymentclient.PayOutput, error) {
			if in.OrderNo != "O1" || in.AmountCent != 9900 {
				t.Errorf("unexpected input=%+v", in)
			}
			return &paymentclient.PayOutput{TradeNo: "T1", Status: "paid"}, nil
		},
	}

	base, stop := startTestServer(t, "POST:/api/orders/{id}/pay", orderapi.HandlePay(store, fake))
	defer stop()

	code, body := mustPOST(t, base+"/api/orders/1/pay")
	if code != http.StatusOK {
		t.Fatalf("http status=%d body=%v", code, body)
	}
	if got := body["code"]; !floatEq(got, 0) {
		t.Fatalf("code=%v want 0", got)
	}
	if fake.Calls() != 1 {
		t.Fatalf("Pay calls=%d want 1", fake.Calls())
	}
	if got, _ := store.Get("1"); got.Status != "paid" || got.TradeNo != "T1" {
		t.Fatalf("store not updated: %+v", got)
	}
}

// —— 支付服务返回错误：handler 应返回业务错误码，且订单状态不变。
func TestHandlePay_UpstreamError(t *testing.T) {
	store := newMemStore(&orderapi.Order{
		ID: "1", OrderNo: "O1", AmountCent: 9900, Status: "pending",
	})
	fake := &fakePaymentClient{
		PayFn: func(ctx context.Context, in paymentclient.PayInput) (*paymentclient.PayOutput, error) {
			return nil, orderapi.ErrPaymentFailed
		},
	}
	base, stop := startTestServer(t, "POST:/api/orders/{id}/pay", orderapi.HandlePay(store, fake))
	defer stop()

	code, body := mustPOST(t, base+"/api/orders/1/pay")
	if code != http.StatusOK {
		t.Fatalf("http status=%d body=%v", code, body)
	}
	if got := body["code"]; !floatEq(got, 55) {
		t.Fatalf("code=%v want 55", got)
	}
	msg, _ := body["message"].(string)
	if !strings.Contains(msg, "payment failed") {
		t.Fatalf("message=%q should contain payment failed", msg)
	}
	if got, _ := store.Get("1"); got.Status != "pending" || got.TradeNo != "" {
		t.Fatalf("store should be untouched: %+v", got)
	}
}

// —— 订单不存在：51 order not found，Pay 不应被调用。
func TestHandlePay_NotFound(t *testing.T) {
	store := newMemStore()
	fake := &fakePaymentClient{}

	base, stop := startTestServer(t, "POST:/api/orders/{id}/pay", orderapi.HandlePay(store, fake))
	defer stop()

	_, body := mustPOST(t, base+"/api/orders/999/pay")
	if got := body["code"]; !floatEq(got, 51) {
		t.Fatalf("code=%v want 51", got)
	}
	if fake.Calls() != 0 {
		t.Fatalf("Pay should not be called, got %d", fake.Calls())
	}
}

// —— 已支付订单：直接返回 ok，不再调用支付服务（幂等）。
func TestHandlePay_AlreadyPaid(t *testing.T) {
	store := newMemStore(&orderapi.Order{
		ID: "1", OrderNo: "O1", AmountCent: 9900, Status: "paid", TradeNo: "T-prev",
	})
	fake := &fakePaymentClient{}

	base, stop := startTestServer(t, "POST:/api/orders/{id}/pay", orderapi.HandlePay(store, fake))
	defer stop()

	_, body := mustPOST(t, base+"/api/orders/1/pay")
	if got := body["code"]; !floatEq(got, 0) {
		t.Fatalf("code=%v want 0", got)
	}
	if fake.Calls() != 0 {
		t.Fatalf("Pay should not be called for paid order, got %d", fake.Calls())
	}
}

// —— cancelled 订单：52 order cannot be paid，Pay 不应被调用。
func TestHandlePay_CancelledOrder(t *testing.T) {
	store := newMemStore(&orderapi.Order{
		ID: "1", OrderNo: "O1", AmountCent: 9900, Status: "cancelled",
	})
	fake := &fakePaymentClient{}

	base, stop := startTestServer(t, "POST:/api/orders/{id}/pay", orderapi.HandlePay(store, fake))
	defer stop()

	_, body := mustPOST(t, base+"/api/orders/1/pay")
	if got := body["code"]; !floatEq(got, 52) {
		t.Fatalf("code=%v want 52", got)
	}
	if fake.Calls() != 0 {
		t.Fatalf("Pay should not be called for cancelled order, got %d", fake.Calls())
	}
}

func floatEq(v any, want float64) bool {
	f, ok := v.(float64)
	if !ok {
		return false
	}
	return f == want
}
