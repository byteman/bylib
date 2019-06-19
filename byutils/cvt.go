package byutil

import (
	"encoding/binary"
	"math"
)

//各种字节序的转换函数.


func BytesToFloat32(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)

	return math.Float32frombits(bits)
}
func Abs(a,b int)int{
	if a > b {
		return a-b
	}
	return b-a
}
func GetLittleUInt32(result []byte)uint32{
	return uint32(uint32(result[3])<<24)+
		uint32(uint32(result[2])<<16)+
		uint32(uint32(result[1])<<8)+
		uint32(result[0])
}
func GetLittleInt16(result []byte)int16{
	return int16(int16(result[1])<<8)+int16(result[0])
}
func GetBigEndianInt16(result []byte)int16{
	return int16(int16(result[0])<<8)+int16(result[1])
}

func Uint16Big2LittleEndian(v uint16 )uint16{
	hi := byte(v>>8)
	lo := byte(v)
	return uint16(lo)<<8 | uint16(hi)
}