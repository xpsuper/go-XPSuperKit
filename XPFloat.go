package XPSuperKit

import (
	"time"
	"math/rand"
)

type XPFloatImpl struct {

}

func (f *XPFloatImpl) EqualsFloat32(a, b float32) bool {
	const ePSINON float32 = 0.00001;
	c := a - b;
	return (c >= - ePSINON) && (c <= ePSINON);
}

func (f *XPFloatImpl) EqualsFloat64(a, b float64) bool {
	const ePSINON float64 = 0.00000001;
	c := a - b;
	return (c >= - ePSINON) && (c <= ePSINON);
}

// 获取范围为[0.0, 1.0)，类型为float32的随机小数

func RandFloat32() float32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Float32()
}

// 获取范围为[0.0, 1.0)，类型为float64的随机小数

func RandFloat64() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Float64()
}
