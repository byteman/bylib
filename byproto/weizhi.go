package byproto

import (
	"bylib/bylog"
	"bylib/byutils"
	"bytes"
	"context"
	"encoding/binary"
	"github.com/goburrow/serial"
	"sync"
	"time"
)

//伟志自定义串口协议分析器.
const MIN_MSG_SIZE=8
const WGT_MSG_SIZE=18
type WzMsg struct{
	SlaveAddr int	//从机地址
	Cmd int32	//命令类型.
	Data *bytes.Buffer
}

func (w *WzMsg) BindBinary(v interface{}) error {
	return binary.Read(w.Data,binary.LittleEndian,v)
}

func (w *WzMsg) MessageNumber() int32 {
	return w.Cmd
}

func (m *WzMsg) Serialize() ([]byte, error) {
	panic("implement me")
}

//伟志协议
type WzProtocal struct {
	Port serial.Port
	Device string
	Baud int
	RxCnt int
	RxLen int
	Ctx context.Context
	Cancel context.CancelFunc
	Handlers map[int]Handler
	MsgChan chan WzMsg //消息通道，可以收到多个消息后，一起处理.
	Buffer bytes.Buffer
	Lock sync.Mutex
}


func (wp *WzProtocal) RegisterHandler(cmd int, handler Handler) {
	wp.Handlers[cmd] = handler

}
/*
名称			帧头起始标志	从机地址	数据帧长度	命令号	消息体	校验(CRC16)
长度（字节）	2(AA,55)	1			2			1	N		2
*/
func (wp *WzProtocal) SendPacket(slaveAddr ,cmd int, data []byte,flag ...bool)error {
	var result [65535]byte
	var n uint16 = 0
	result[0] = 0xAA
	result[1] = 0x55
	result[2] = uint8(slaveAddr)
	result[5] = byte(cmd)
	if data != nil{
		copy(result[6:] , data)
		n = uint16(len(data) + 8)
	}else{
		n = 8
	}

	//这种方法更优美,写入长度
	binary.LittleEndian.PutUint16(result[3:5],n)
	crc:=byutil.CRC16BigEndian(result[:n-2])
	//写入校验和
	binary.LittleEndian.PutUint16(result[n-2:],crc)
	if len(flag) > 0{
		bylog.Debug("write=% x",result[:n])
	}

	wp.Lock.Lock()
	_,err:=wp.Port.Write(result[:n])
	wp.Lock.Unlock()

	return  err
}
//实现了一个Read接口
func (wp *WzProtocal)Read(data []byte)(int ,error){
	panic("implement me")
}
func (wp* WzProtocal) parse(rx byte) {

	RxCnt:= wp.Buffer.Len()
	if RxCnt == 0 &&  rx == 0xAA {
		wp.Buffer.WriteByte(rx)
	} else if RxCnt == 1 {
		if  rx == 0x55 {
			wp.Buffer.WriteByte(rx)
		} else {
			//协议不对，清空数据
			wp.Buffer.Truncate(0)
		}
	} else if RxCnt == 2{
		wp.Buffer.WriteByte(rx)
	} else if RxCnt == 3{
		wp.Buffer.WriteByte(rx)
		wp.RxLen = int(rx)
	} else if RxCnt  == 4{
		//计算数据区长度.
		wp.Buffer.WriteByte(rx)
		wp.RxLen = wp.RxLen  + (int(rx)<< 8)
		if wp.RxLen < MIN_MSG_SIZE{
			//长度指示数据不正确，清空数据
			wp.Buffer.Truncate(0)
		}
	} else if RxCnt >= 5 {

		wp.Buffer.WriteByte(rx)
		//收到的数据长度>=指示长度,收满数据了
		if wp.Buffer.Len() >= wp.RxLen {
			var data = make([]byte, wp.Buffer.Len())
			copy(data,wp.Buffer.Bytes())
			//bylog.Debug("len=%d",len(data))
			crc := byutil.CRC16BigEndian(data[:len(data)-2])

			rcv_crc :=  uint16(data[wp.RxLen-1] )
			rcv_crc <<= 8
			rcv_crc += uint16(data[wp.RxLen-2] )
			//判断校验和是否正确.
			if crc == rcv_crc {
				wz:=WzMsg{
					Data:bytes.NewBuffer(data[6:]),
					Cmd:int32(data[5]),
					SlaveAddr:int(data[2]),
				}
				//投递消息入管道
				wp.MsgChan<-wz
			}else{
				bylog.Error("weizhi protocal parse crc err % x != % x",rcv_crc, crc)
			}
			//数据满了，一定要清空数据.
			wp.Buffer.Truncate(0)
		}
	}


}

//读取线程不断读取100个字节，读取到超时时间100毫秒，然后把数据丢到一个fifo中
//分析线程不断进行协议分析，找到一个完整的协议包后，回调对应的已经注册的处理函数
//接收线程，
func (wp *WzProtocal) receive() {
	for{
		var data [WGT_MSG_SIZE]byte
		var n = 0
		var err error
		if n,err = wp.Port.Read(data[:]);err!=nil{
			//bylog.Error("read %v",err)
			continue
		}
		//bylog.Debug("Receive % x",data[:n])
		for _,ch:=range data[:n]{
			wp.parse(ch)
		}
	}
}
func (wp *WzProtocal)doWork(msg WzMsg){
	for k,v:=range wp.Handlers{
		//bylog.Debug("k=%d cmd=%d",k,msg.Cmd)
		if k == int(msg.Cmd){
			v(&msg)
			return
		}
	}
	bylog.Warn("msg[%d] can not find handler",msg.Cmd)
}
//分析线程，如果没有数据就阻塞住
func (wp *WzProtocal) dispatch() {

	for {
		select {
		case msg := <-wp.MsgChan:  //如果有数据，下面打印。但是有可能ch一直没数据
			//bylog.Debug("msg[%d] data=% x", msg.Cmd,msg.Data.Bytes())
			wp.doWork(msg)

		case <-wp.Ctx.Done(): //上面的ch如果一直没数据会阻塞，那么select也会检测其他case条件，检测到后3秒超时
			bylog.Debug("WzProtocal quit")
			return
		}
	}



}
func (wp *WzProtocal) Start()error {

	config := serial.Config{
		Address:  wp.Device,
		BaudRate: wp.Baud,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N", //这里配置很重要，默认是E校验
		Timeout:  100 * time.Millisecond,
		RS485:serial.RS485Config{
			Enabled:false,
		},
	}
	port,err:= serial.Open(&config)
	if err!=nil{
		bylog.Debug("open %s baud=%d err=%v",config.Address,config.BaudRate,err)
		return err
	}
	bylog.Debug("open %s baud=%d Success",config.Address,config.BaudRate)
	wp.Port = port

	go wp.receive()
	go wp.dispatch()
	return nil
}
func (wp *WzProtocal) Stop() {
	panic("implement me")
}
//创建一个伟志自定义协议分析器,可以给分析器注册消息
func NewWeiZhiProtocal(device string ,baud int)*WzProtocal  {
	wz:= &WzProtocal{
		Device:device,
		Baud:baud,
		Handlers:make(map[int]Handler),
		MsgChan:make(chan WzMsg, 100),

	}

	wz.Ctx,wz.Cancel = context.WithCancel(context.Background())
	return wz
}