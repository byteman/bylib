package byproto

import (
	"bylib/bylog"
	"bylib/bynet/byudp"
	"bylib/byutils"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

//货架称通讯协议包编码解码器

type Wk01Codec struct{

}
//解码协议包
/**

头标识	1	固定为0xFE

包长度 	2 	包含从本身开始， 直到报文内容结束的数据长度
命令 	1 	定义见表1-2
唯一识别码	6	网络集采集器唯一识别码
数据信息 	n 	根据命令而定,见表1-3
校验和 	CRC16 	校验和为数据头开始到数据体结束所有字节取
反加1
 */
func (Wk01Codec) Decode(conn net.Conn) (byudp.Message, error) {
	//读取出数据包
	var buf [1024]byte
	n,err:=conn.Read(buf[:])
	if err!=nil{
		bylog.Error("decode err=%v",err)
	}
	byutil.HexDump("recv<---",buf[:n])
	if buf[0] != 0xFE{
		return nil,fmt.Errorf("not find sync header 0xFE")
	}
	len := byutil.GetBigEndianInt16(buf[1:3])
	if int(len+1) != n || n < 12{
		return nil,fmt.Errorf("length not match %d != %d",len, n)
	}
	//msgType:=int32(buf[3])
	crc1:=byutil.CRC16BigEndian(buf[:n-2])
	crc2:=uint16(byutil.GetBigEndianInt16(buf[n-2:]))
	if crc1 != crc2 {
		return nil,fmt.Errorf("crc not match % x != % x",crc1, crc2)
	}
	return DeserializeMyMessage(buf[3:])


}
//编码协议包
/**
头标识 		1 	固定为0x7F
包长度 		2 	包含从本身开始， 直到报文内容结束的数据长度
命令 		1 	定义见表1-2
唯一识别码	6	网络集采集器唯一识别码
秤台数据信息 		定义见表2-2
校验和 		2	CRC16校验和为数据头开始到数据体结束所有字节
 */
func (Wk01Codec) Encode(msg byudp.Message) ([]byte, error) {

	data, err := msg.Serialize()
	if err != nil {
		return nil, err
	}
	var sz uint16
	sz = uint16(len(data) + 5)
	buf := new(bytes.Buffer)

	buf.WriteByte(0x7F)
	binary.Write(buf, binary.BigEndian, sz) //size
	buf.WriteByte(byte(msg.MessageNumber()))
	buf.Write(data)
	crc16:=byutil.CRC16BigEndian(buf.Bytes())
	binary.Write(buf, binary.BigEndian, crc16)

	packet := buf.Bytes()
	return packet, nil
}