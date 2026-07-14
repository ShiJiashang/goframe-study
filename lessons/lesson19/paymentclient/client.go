package paymentclient

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/gclient"
)

type PayInput struct {
	OrderNo    string `json:"orderNo"`
	AmountCent int64  `json:"amountCent"`
}

type PayOutput struct {
	TradeNo string `json:"tradeNo"`
	Status  string `json:"status"`
}

type RefundInput struct {
	TradeNo    string `json:"tradeNo"`
	OrderNo    string `json:"orderNo"`
	AmountCent int64  `json:"amountCent"`
}

type RefundOutput struct {
	TradeNo string `json:"tradeNo"`
	Status  string `json:"status"`
}

type Client struct {
	http *gclient.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 1500 * time.Millisecond
	}
	return &Client{
		http: gclient.New().
			ContentJson().
			Timeout(timeout).
			Prefix(strings.TrimRight(baseURL, "/")),
	}
}

func (c *Client) Pay(ctx context.Context, input PayInput) (*PayOutput, error) {
	response, err := c.http.Post(ctx, "/pay", input)
	if err != nil {
		return nil, gerror.Wrap(err, "request payment service failed")
	}
	defer response.Close()

	body := response.ReadAll()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, gerror.Newf(
			"payment service returned status=%d body=%s",
			response.StatusCode,
			string(body),
		)
	}

	var output PayOutput
	if err = gjson.DecodeTo(body, &output); err != nil {
		return nil, gerror.Wrap(err, "decode payment response failed")
	}
	if output.TradeNo == "" || output.Status == "" {
		return nil, gerror.New("payment response misses tradeNo or status")
	}
	return &output, nil
}

func (c *Client) Refund(ctx context.Context, input RefundInput) (*RefundOutput, error) {
	response, err := c.http.Post(ctx, "/refund", input)
	if err != nil {
		return nil, gerror.Wrap(err, "request refund service failed")
	}
	defer response.Close()

	body := response.ReadAll()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, gerror.Newf(
			"refund service returned status=%d body=%s",
			response.StatusCode,
			string(body),
		)
	}

	var output RefundOutput
	if err = gjson.DecodeTo(body, &output); err != nil {
		return nil, gerror.Wrap(err, "decode refund response failed")
	}
	if output.TradeNo == "" || output.Status == "" {
		return nil, gerror.New("refund response misses tradeNo or status")
	}
	return &output, nil
}
