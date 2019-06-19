package byutil

import (
	"fmt"
	"math/big"
	"net"
)

func InetNtoA(ip int64) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func InetAtoN(ip string) int32 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return int32(ret.Int64())
}

func FormatMacFromUint16(mac []uint16)string{
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",byte(mac[0]>>8),byte(mac[0]),
		byte(mac[1]>>8),byte(mac[1]),
		byte(mac[2]>>8),byte(mac[2]))
}