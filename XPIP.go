package XPSuperKit

import (
	"net"
)

type XPIPImpl struct {

}

// 将ip地址转换为长整型
func (ip *XPIPImpl)IP2Long(ipstr string) int64 {
	if ip := net.ParseIP(ipstr); ip != nil {
		var n uint32
		ipBytes := ip.To4()
		for i := uint8(0); i <= 3; i++ {
			n |= uint32(ipBytes[i]) << ((3 - i) * 8)
		}
		return int64(n)
	}
	return 0
}

// 将长整型转换为ip地址
func (ip *XPIPImpl)Long2IP(long int64) string {
	ipBytes := net.IP{}
	for i := uint(0); i <= 3; i++ {
		ipBytes = append(ipBytes, byte(long>>((3-i)*8)))
	}
	return ipBytes.String()
}
