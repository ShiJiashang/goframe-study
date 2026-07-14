package example

import (
	"strings"
	"testing"
)

func TestCalculateTotal(t *testing.T) {
	tests := []struct {
		name      string
		priceCent int64
		quantity  int64
		want      int64
		wantErr   bool
		errSubstr string
	}{
		{name: "正常计算", priceCent: 1999, quantity: 2, want: 3998},
		{name: "价格 1 数量 1", priceCent: 1, quantity: 1, want: 1},
		{name: "大数不溢出", priceCent: 1_000_000, quantity: 1_000, want: 1_000_000_000},
		{name: "价格为零", priceCent: 0, quantity: 2, wantErr: true, errSubstr: "商品价格"},
		{name: "价格为负", priceCent: -100, quantity: 2, wantErr: true, errSubstr: "商品价格"},
		{name: "数量为零", priceCent: 1999, quantity: 0, wantErr: true, errSubstr: "购买数量"},
		{name: "数量为负", priceCent: 1999, quantity: -3, wantErr: true, errSubstr: "购买数量"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := CalculateTotal(test.priceCent, test.quantity)
			if (err != nil) != test.wantErr {
				t.Fatalf("错误状态不符：err=%v wantErr=%v", err, test.wantErr)
			}
			if err != nil && test.errSubstr != "" && !strings.Contains(err.Error(), test.errSubstr) {
				t.Errorf("错误消息不含 %q：err=%v", test.errSubstr, err)
			}
			if got != test.want {
				t.Errorf("金额不符：got=%d want=%d", got, test.want)
			}
		})
	}
}
