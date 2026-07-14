package main

import (
	"context"
	"os"
	"sync"
	"time"

	"goframe-study/lessons/lesson19/paymentclient"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type order struct {
	ID         string `json:"id"`
	OrderNo    string `json:"orderNo"`
	AmountCent int64  `json:"amountCent"`
	Status     string `json:"status"`
	TradeNo    string `json:"tradeNo,omitempty"`
}

type apiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

var (
	mu        sync.Mutex
	orders    = map[string]*order{}
	callbacks = map[string]bool{}
)

func main() {
	orders["1"] = &order{ID: "1", OrderNo: "O202607140001", AmountCent: 9900, Status: "pending"}
	orders["2"] = &order{ID: "2", OrderNo: "O202607140002", AmountCent: 19900, Status: "pending"}

	server := g.Server()
	server.SetPort(8009)

	server.Group("/api", func(group *ghttp.RouterGroup) {
		group.GET("/orders/:id", handleGetOrder)
		group.POST("/orders/:id/pay", handlePayOrder)
		group.POST("/orders/:id/refund", handleRefundOrder)
		group.POST("/payments/callback", handlePaymentCallback)
	})

	server.Run()
}

func handleGetOrder(r *ghttp.Request) {
	id := r.GetRouter("id").String()

	mu.Lock()
	defer mu.Unlock()

	item, ok := orders[id]
	if !ok {
		writeJSON(r, 51, "order not found", nil)
		return
	}
	writeJSON(r, 0, "ok", item)
}

func handlePayOrder(r *ghttp.Request) {
	id := r.GetRouter("id").String()
	baseURL := getPaymentBaseURL(r.Context())
	client := paymentclient.New(baseURL, 1500*time.Millisecond)

	mu.Lock()
	item, ok := orders[id]
	if !ok {
		mu.Unlock()
		writeJSON(r, 51, "order not found", nil)
		return
	}
	if item.Status == "paid" {
		out := *item
		mu.Unlock()
		writeJSON(r, 0, "order already paid", out)
		return
	}

	orderNo := item.OrderNo
	if mockOrderNo := r.GetQuery("mockOrderNo").String(); mockOrderNo != "" {
		orderNo = mockOrderNo
	}
	amountCent := item.AmountCent
	mu.Unlock()

	payResult, err := client.Pay(r.Context(), paymentclient.PayInput{
		OrderNo:    orderNo,
		AmountCent: amountCent,
	})
	if err != nil {
		g.Log().Error(r.Context(), err)
		writeJSON(r, 52, err.Error(), nil)
		return
	}

	mu.Lock()
	item.Status = payResult.Status
	item.TradeNo = payResult.TradeNo
	out := *item
	mu.Unlock()

	writeJSON(r, 0, "ok", g.Map{
		"order":   out,
		"payment": payResult,
	})
}

func handleRefundOrder(r *ghttp.Request) {
	id := r.GetRouter("id").String()
	baseURL := getPaymentBaseURL(r.Context())
	client := paymentclient.New(baseURL, 1500*time.Millisecond)

	mu.Lock()
	item, ok := orders[id]
	if !ok {
		mu.Unlock()
		writeJSON(r, 51, "order not found", nil)
		return
	}
	if item.Status == "refunded" {
		out := *item
		mu.Unlock()
		writeJSON(r, 0, "已经退款", out)
		return
	}
	if item.Status != "paid" {
		status := item.Status
		mu.Unlock()
		writeJSON(r, 54, "订单状态不允许退款: "+status, nil)
		return
	}

	orderNo := item.OrderNo
	if mockOrderNo := r.GetQuery("mockOrderNo").String(); mockOrderNo != "" {
		orderNo = mockOrderNo
	}
	tradeNo := item.TradeNo
	amountCent := item.AmountCent
	mu.Unlock()

	refundResult, err := client.Refund(r.Context(), paymentclient.RefundInput{
		TradeNo:    tradeNo,
		OrderNo:    orderNo,
		AmountCent: amountCent,
	})
	if err != nil {
		g.Log().Error(r.Context(), err)
		writeJSON(r, 55, err.Error(), nil)
		return
	}

	mu.Lock()
	item.Status = refundResult.Status
	out := *item
	mu.Unlock()

	writeJSON(r, 0, "ok", g.Map{
		"order":  out,
		"refund": refundResult,
	})
}

func getPaymentBaseURL(ctx context.Context) string {
	const defaultBaseURL = "http://127.0.0.1:9001"

	if env := os.Getenv("PAYMENT_BASEURL"); env != "" {
		return env
	}
	value, err := g.Cfg().Get(ctx, "payment.baseURL")
	if err != nil || value.String() == "" {
		return defaultBaseURL
	}
	return value.String()
}

func handlePaymentCallback(r *ghttp.Request) {
	tradeNo := r.Get("tradeNo").String()
	orderNo := r.Get("orderNo").String()
	if tradeNo == "" || orderNo == "" {
		writeJSON(r, 53, "tradeNo and orderNo are required", nil)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if callbacks[tradeNo] {
		writeJSON(r, 0, "duplicate callback ignored", g.Map{
			"tradeNo":   tradeNo,
			"duplicate": true,
		})
		return
	}
	callbacks[tradeNo] = true

	for _, item := range orders {
		if item.OrderNo == orderNo {
			item.Status = "paid"
			item.TradeNo = tradeNo
			writeJSON(r, 0, "callback handled", g.Map{
				"tradeNo":   tradeNo,
				"duplicate": false,
				"order":     item,
			})
			return
		}
	}

	writeJSON(r, 51, "order not found", nil)
}

func writeJSON(r *ghttp.Request, code int, message string, data any) {
	r.Response.WriteJson(apiResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
