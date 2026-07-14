package example

import "errors"

// CalculateTotal returns the total amount in cents.
func CalculateTotal(priceCent int64, quantity int64) (int64, error) {
	if priceCent <= 0 {
		return 0, errors.New("商品价格必须大于0")
	}
	if quantity <= 0 {
		return 0, errors.New("购买数量必须大于0")
	}
	return priceCent * quantity, nil
}
