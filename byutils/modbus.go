package byutil

import (
	"bylib/bylog"
	"bylib/byopenwrt"
	"encoding/binary"
	"time"
)

func MbNowBytes()[]uint16{
	buf:=[6]byte{}
	tm:=time.Now()
	buf[0] = byte(tm.Year()-2000)
	buf[1] = byte(tm.Month())
	buf[2] = byte(tm.Day())
	buf[3] = byte(tm.Hour())
	buf[4] = byte(tm.Minute())
	buf[5] = byte(tm.Second())

	return MbBytesToUint16(buf[:])

}
func MbDateTimeBytes()[]uint16{
	buf:=[6]byte{}
	odt:=byopenwrt.OpDateTime{}
	err:=byopenwrt.GetSysDateTime(&odt)
	if err!=nil{
		bylog.Error("GetSysDateTime err=%v",err)
		return MbNowBytes()
	}
	buf[0] = byte(odt.Year)
	buf[1] = byte(odt.Month)
	buf[2] = byte(odt.Day)
	buf[3] = byte(odt.Hour)
	buf[4] = byte(odt.Minute)
	buf[5] = byte(odt.Second)
	//bylog.Debug("buf=% x",buf)
	return MbBytesToUint16(buf[:])

}
func MbInt32ToUint16(value int32)[]uint16{
	var x [2]uint16
	x[0] = uint16((value >> 16)&0xffff)
	x[1] = uint16((value)&0xffff)
	return x[0:]
}
func MbUint16sToInt32(value []uint16)int32{

	x:=int32(value[0])<<16
	x+=int32(value[1])
	return x

}

// BytesToUint16 converts a big endian array of bytes to an array of unit16s
func MbBytesToUint16(bytes []byte) []uint16 {
	values := make([]uint16, len(bytes)/2)

	for i := range values {
		values[i] = binary.BigEndian.Uint16(bytes[i*2 : (i+1)*2])

	}
	return values
}

// Uint16ToBytes converts an array of uint16s to a big endian array of bytes
func MbUint16ToBytes(values []uint16) []byte {
	bytes := make([]byte, len(values)*2)

	for i, value := range values {
		binary.BigEndian.PutUint16(bytes[i*2:(i+1)*2], value)
	}
	return bytes
}

type MbBuffer struct{
	buf []uint16
}
func NewMbBuffer()*MbBuffer{
	return &MbBuffer{
		buf:make([]uint16,0,0),
	}
}
func (mb *MbBuffer)Uint16()[]uint16{
	return mb.buf
}
//写入通用数据类型
func (mb *MbBuffer)Write(value interface{})error{

	switch value.(type) {
	case int32:
		v:=value.(int32)
		mb.buf=append(mb.buf,MbInt32ToUint16(v)...)
	case uint32:
		v:=value.(uint32)
		mb.buf=append(mb.buf,MbInt32ToUint16(int32(v))...)

	case uint16:
		mb.buf=append(mb.buf,value.(uint16))
	case int16:
		mb.buf=append(mb.buf,uint16(value.(int16)))
	case uint8:
		x:=uint16(value.(uint8))
		mb.buf=append(mb.buf,x)
	case int8:
		x:=uint16(value.(int8))
		mb.buf=append(mb.buf,x)
	case int64:
		x:=value.(int64)
		mb.buf=append(mb.buf,MbInt32ToUint16(int32(x>>32))...)
		mb.buf=append(mb.buf,MbInt32ToUint16(int32(x&0xffffffff))...)
	case uint64:
		x:=value.(uint64)
		mb.buf=append(mb.buf,MbInt32ToUint16(int32(x>>32))...)
		mb.buf=append(mb.buf,MbInt32ToUint16(int32(x&0xffffffff))...)

	default:
		bylog.Error("Error type=%T",value)
		panic("unknown type")
	}
	return nil
}