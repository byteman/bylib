package byutil

import "testing"

func TestModbusBuffer(t *testing.T) {
	buf:=NewMbBuffer()
	buf.Write(uint8(1))
	buf.Write(int8(1))
	buf.Write(uint16(1))
	buf.Write(int16(1))
	buf.Write(uint32(1))
	buf.Write(int32(1))

}
