package byscaler

import (
	"bylib/bylog"
	"bylib/byopenwrt"
	"bylib/byutils"
	"fmt"
	"github.com/goburrow/serial"
	"io"
	"sync"
	"time"
)
const(
	CMD_CLEAR_ZERO = 0 //清零
	CMD_READ_WGT = iota //查询单个或者全部称台的重量和状态
	CMD_QUERY_CHANGE_SENSOR=iota //查询重量有更改的称台
	CMD_QUERY_ERROR_SENSOR=iota //查询有故障的传感器.
	CMD_SET_WGT=iota //设置称台重量
	CMD_CUSTOM_READ_AD=105 //读取AD值
	CMD_CUSTOM_CALIB_ZERO=iota //标定零点
	CMD_CUSTOM_CALIB_WGT=iota  //标定重量
	CMD_CUSTOM_WRITE_PARAM=iota  //参数写 传感器个数/阀值/读取时间
	CMD_CUSTOM_READ_PARAM=iota  //参数读 传感器个数/阀值/读取时间
	CMD_CUSTOM_WRITE_SENSOR_ADDR=iota //传感器地址设置，只写

)
//wzk01的管理类. wzk01是一个集散控制器，下面接了n个adt03.
//wzk01是串口接口的，所以需要依赖一个串口类
type WZK01 struct {
	Port serial.Port
	PortMutex sync.Mutex
	TotalRx int
	ValidRx int
	Timeout int
	Buffer []byte //数据
}

func (f *WZK01)ReadCmd(addr int32,result []byte)(n int, err error){
	f.PortMutex.Lock()
	defer func() {
		f.PortMutex.Unlock()
	}()
	if n,err=io.ReadFull(f.Port, result[:]);err!=nil{

		if err!=nil{
			bylog.Error("read sensor[%d] err %v",addr,err)

			return 0,err
		}
	}

	return n,nil
}
func (f *WZK01)calibrateZero(addr int32)error{

	cmd:=[8]byte{0xFE,0x7F,byte(addr),0x4,1,1,1,0}
	cmd[7]=byutil.Xor(cmd[0:7])
	bylog.Debug("CalibrateZero %v",cmd)
	if err:=f.writeCmd(addr,cmd[:]);err!=nil{
		return err
	}

	return nil
}
func (f *WZK01)calibrateK(addr int32,weight int32)error{

	cmd:=[8]byte{0xFE,0x7F,byte(addr),0x7,0,0,0,0}
	cmd[4] = byte(weight&0xff)
	cmd[5] = byte((weight>>8)&0xff)
	cmd[6] = byte((weight>>16)&0xff)

	cmd[7]=byutil.Xor(cmd[0:7])
	bylog.Debug("CalibrateK %v",cmd)
	if err:=f.writeCmd(addr,cmd[:]);err!=nil{
		return err
	}

	return nil

}
//读取重量的命令.
func (f *WZK01)sendReadWeight(addr int32)error{
	//查询全部称台实时重量.
	cmd:=[]byte{0x55,0xFE,0xAA,0,7,byte(addr),CMD_READ_WGT,0,}
	//求异或.

	crc16:=byutil.CRC16BigEndian(cmd)
	cmd=append(cmd,byte(crc16>>8))
	cmd=append(cmd,byte(crc16))

	//cmd=append(cmd,crc16...)

	return f.writeCmd(addr,cmd)
}

func (f *WZK01)getReadWeight(ss *SensorSet)(error){

	var result []byte//[maxSize]byte //最大长度
	var err error
	if result,err=f.readCmd(ss.Addr);err!=nil || result==nil{
		return fmt.Errorf("read sensor[%d] err=%v",ss.Addr,err)
	}
	resize:=len(result)
	if resize < rtuMinSize{
		byutil.HexDump("error minSize  <---",result)
		return fmt.Errorf("read len error")
	}

	if result[0] != 0xAA || result[1] != 0x7F || result[2] != 0x55 {
		//错误的数据格式.
		return fmt.Errorf("error data %v",result)
	}
	crc1:=byutil.CRC16LittleEndian(result[:resize-2])
	crc2:=uint16(byutil.GetLittleInt16(result[resize-2:]))
	if crc1!=crc2{
		bylog.Error("size=%d data=% X",resize,result)
		return fmt.Errorf("crc err % x != % x",crc1, crc2)
	}
	sn:=(len(result) - rtuMinSize)/4
	//bylog.Debug("sensor addr=%d % X",ss.Addr,result)

	//byutil.FreqPrintf("sn")
	//bylog.Debug("sn=%d",sn)
	for i:=0; i < sn; i++{
		addr:=int(result[9+i*4])
		//bylog.Debug("addr=%d",addr)
		//判断地址是否已经存在了.
		if _,ok:=ss.Sensors[addr];!ok{
			ss.Sensors[addr] = NewSensor(int32(addr))
		}
		ss.Sensors[addr].Addr = int32(addr)

		var state = result[10+i*4]
		ss.Sensors[addr].StateValue = state

		//fmt.Println("weight=",ss.Weight)
		if state&0x1 != 0{
			ss.Sensors[addr].State.Zero = true
		}else{
			ss.Sensors[addr].State.Zero = false
		}
		if state&0x2 != 0{
			ss.Sensors[addr].State.Still = true
		}else{
			ss.Sensors[addr].State.Still = false
		}
		//0x111000 芯片故障|传感器故障|开机零点故障
		if state&0x30 != 0{
			ss.Sensors[addr].State.Error = true
		}else{
			ss.Sensors[addr].State.Error = false
		}
		value:=int32(byutil.GetBigEndianInt16(result[11+i*4:13+i*4]))
		ss.Sensors[addr].Value=value
		//if addr == 1{
		//
		//	bylog.Debug("wg=%d state=%x",value,state)
		//}
	}


	return nil
}
func (f *WZK01)writeCmd(addr int32,cmd []byte)error {
	f.PortMutex.Lock()
	defer func() {
		f.PortMutex.Unlock()
	}()
	//byutil.HexDump("send->", cmd)
	if _, err := f.Port.Write(cmd); err != nil {
		bylog.Error("Write sensor[%d] err %v", addr, err)

		return err
	}
	return nil

}

const(
	rtuMaxSize=256
	rtuMinSize=10
)

func (f *WZK01)readCmd(addr int32)(result []byte,err error){
	f.PortMutex.Lock()
	defer func() {
		f.PortMutex.Unlock()
	}()
	var data [rtuMaxSize]byte
	var n int
	var n1 int
	n, err = io.ReadAtLeast(f.Port, data[:], rtuMinSize)
	if err != nil {
		return
	}

	bytesToRead:=int(byutil.GetBigEndianInt16(data[3:5]))+3
	//bylog.Debug("get len=% x % x",bytesToRead,data[:n])
	if n < bytesToRead {
		if bytesToRead > rtuMinSize && bytesToRead <= rtuMaxSize {
			if bytesToRead > n {
				n1, err = io.ReadFull(f.Port, data[n:bytesToRead])
				n += n1
			}
		}
	}

	if err != nil {
		bylog.Error("err=%s",err)
		return
	}
	result = data[:n]

	return
}
func (f *WZK01)readSensor(ss *SensorSet)(err error){
	//FE 7F 01 83 03

	f.TotalRx++
	//发送读取命令
	//Debug("Write sensor[%d] %v",ss.Addr,cmd)
	if err:=f.sendReadWeight(ss.Addr);err!=nil{
		//bylog.Error("Write sensor[%d] err %v",ss.Addr,err)
		ss.Timeout++
		return err
	}
	if err=f.getReadWeight(ss);err!=nil {
		ss.Timeout++
		//bylog.Error("read sensor[%d] err %v",ss.Addr,err)
		return err
	}
	f.ValidRx++
	ss.Timeout = 0
	ss.TimeStamp = byopenwrt.GetLocalNowTime().Unix()
	ss.Online = true
	return nil
}
//修改地址
func (WZK01) ModifyAddr(oldAddr, newAddr int32) error {
	panic("implement me")
}
//获取测量重量
//id 集散器的编号

func (adt *WZK01) Measure(ss *SensorSet) (err error) {
	return adt.readSensor(ss)
}

func (adt *WZK01) CalZero(addr int32) error {
	return adt.calibrateZero(addr)
}

func (adt *WZK01) CalKg(addr int32, kg int32) error {
	return adt.calibrateK(addr,kg)
}
func (adt *WZK01)Open(port,baud int)error{
	config := serial.Config{
		Address:  byutil.GetUartName(port),
		BaudRate: baud,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  1000 * time.Millisecond,
	}
	var err error
	_port, err := serial.Open(&config)
	if err != nil {
		bylog.Error("open %s failed",config.Address)
		return err
	}
	adt.Port = _port

	return nil
}
func NewWZK01()*WZK01{
	return &WZK01{}
}