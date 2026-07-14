// Package orderapi 是 lesson21 用来演示"用接口替换外部依赖"的教学包。
// 它把 lesson19 的支付客户端抽象成 PaymentClient 接口，让 handler 测试
// 可以传入 fake 实现，不依赖真实 HTTP 或 mock-payment 进程。
package orderapi

import (
	"context"
	"errors"

	"github.com/gogf/gf/v2/net/ghttp"

	"goframe-study/lessons/lesson19/paymentclient"
)

// PaymentClient 是 handler 依赖的最小接口。
// lesson19 的 *paymentclient.Client 天然满足这个接口（隐式实现）。
type PaymentClient interface {
	Pay(ctx context.Context, input paymentclient.PayInput) (*paymentclient.PayOutput, error)
}

type Order struct {
	ID         string `json:"id"`
	OrderNo    string `json:"orderNo"`
	AmountCent int64  `json:"amountCent"`
	Status     string `json:"status"`
	TradeNo    string `json:"tradeNo,omitempty"`
}

type OrderStore interface {
	Get(id string) (*Order, bool)
	Save(o *Order)
}

type apiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// HandlePay 返回一个绑定了 store/client 的 handler。
// 通过参数注入依赖，方便在测试里替换。
func HandlePay(store OrderStore, client PaymentClient) func(r *ghttp.Request) {
	return func(r *ghttp.Request) {
		id := r.GetRouter("id").String()
		order, ok := store.Get(id)
		if !ok {
			writeJSON(r, 51, "order not found", nil)
			return
		}
		if order.Status == "paid" {
			writeJSON(r, 0, "ok", order)
			return
		}
		if order.Status != "pending" {
			writeJSON(r, 52, "order cannot be paid", nil)
			return
		}

		out, err := client.Pay(r.Context(), paymentclient.PayInput{
			OrderNo:    order.OrderNo,
			AmountCent: order.AmountCent,
		})
		if err != nil {
			writeJSON(r, 55, err.Error(), nil)
			return
		}

		order.Status = "paid"
		order.TradeNo = out.TradeNo
		store.Save(order)
		writeJSON(r, 0, "ok", order)
	}
}

// ErrPaymentFailed 是测试里 fake 客户端最常复用的错误。
var ErrPaymentFailed = errors.New("payment failed")

func writeJSON(r *ghttp.Request, code int, message string, data any) {
	r.Response.WriteJson(apiResponse{Code: code, Message: message, Data: data})
}
