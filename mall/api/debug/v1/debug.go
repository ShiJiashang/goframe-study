package v1

import "github.com/gogf/gf/v2/frame/g"

type ErrorReq struct {
	g.Meta `path:"/debug/error" method:"get" tags:"Debug" summary:"Debug error and log"`

	Type string `json:"type" in:"query" d:"ok" dc:"错误类型：ok/notfound/wrap/panic"`
}

type ErrorRes struct {
	TraceID string `json:"traceId" dc:"链路ID"`
	Message string `json:"message" dc:"调试消息"`
}

type PaymentParseReq struct {
	g.Meta `path:"/debug/payment/parse" method:"post" tags:"Debug" summary:"解析第三方支付数据"`

	Payload g.Map `json:"payload" v:"required#支付数据不能为空" dc:"第三方支付原始数据"`
}

type PaymentParseRes struct {
	TradeNo         string   `json:"tradeNo" dc:"第三方支付单号"`
	AmountCent      int64    `json:"amountCent" dc:"支付金额，单位为分"`
	PaidAt          string   `json:"paidAt" dc:"格式化后的支付时间"`
	PaidAtTimestamp int64    `json:"paidAtTimestamp" dc:"支付时间秒级时间戳"`
	Success         bool     `json:"success" dc:"是否支付成功"`
	Tags            []string `json:"tags" dc:"去重后的支付标签"`
	Channel         string   `json:"channel" dc:"支付渠道"`
	FeeCent         int64    `json:"feeCent" dc:"手续费，单位为分"`
	NetAmountCent   int64    `json:"netAmountCent" dc:"净支付金额，单位为分"`
}
