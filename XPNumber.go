package XPSuperKit

import (
	"time"
	"math/rand"
)

type XPNumberImpl struct {

}

func (n *XPNumberImpl) RandomNumber(max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max)
}
