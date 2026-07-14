package example

import (
	"testing"

	"github.com/gogf/gf/v2/test/gtest"
)

func TestCalculateTotalWithGTest(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		total, err := CalculateTotal(2500, 3)
		t.AssertNil(err)
		t.AssertEQ(total, int64(7500))
	})
}
