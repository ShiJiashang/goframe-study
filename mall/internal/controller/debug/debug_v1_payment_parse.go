package debug

import (
	"context"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	v1 "goframe-study/mall/api/debug/v1"
)

func (c *ControllerV1) PaymentParse(ctx context.Context, req *v1.PaymentParseReq) (res *v1.PaymentParseRes, err error) {
	paymentJSON := gjson.New(req.Payload)

	tradeNo := paymentJSON.Get("tradeNo", "").String()
	if tradeNo == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "tradeNo不能为空")
	}

	amountValue := paymentJSON.Get("amountCent", 0)
	amountCent := gconv.Int64(amountValue.Val())
	if amountCent <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "amountCent必须大于0")
	}

	paidAtText := paymentJSON.Get("paidAt", "").String()
	if paidAtText == "" {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "paidAt不能为空")
	}
	paidAt, parseErr := gtime.StrToTime(paidAtText)
	if parseErr != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, parseErr, "paidAt格式错误")
	}

	success := gconv.Bool(paymentJSON.Get("success", false).Val())
	tags := garray.NewTArrayFrom[string](
		gconv.Strings(paymentJSON.Get("tags", []string{}).Val()),
	).Unique().Slice()
	channel := paymentJSON.Get("channel", "").String()

	feeCent := gconv.Int64(paymentJSON.Get("feeCent", 0).Val())
	/*
	   feeCent 不能小于 0
	   feeCent 不能大于 amountCent
	   netAmountCent = amountCent - feeCent
	*/
	if feeCent < 0 || feeCent > amountCent {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "feeCent必须在0到amountCent之间")
	}

	return &v1.PaymentParseRes{
		TradeNo:         tradeNo,
		AmountCent:      amountCent,
		PaidAt:          paidAt.Format("Y-m-d H:i:s"),
		PaidAtTimestamp: paidAt.Timestamp(),
		Success:         success,
		Tags:            tags,
		Channel:         channel,
		FeeCent:         feeCent,
		NetAmountCent:   amountCent - feeCent,
	}, nil
}
