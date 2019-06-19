package byscaler

import (
	"bylib/bylog"
	"bylib/byopenwrt"
	"bylib/byutils"
	"fmt"
	"github.com/goburrow/serial"
	"github.com/pkg/errors"
	"io"
	"sync"
	"time"
)

//adt03的管理类.
//adt03是串口接口的，所以需要依赖一个串口类
type ADT03 struct {
	Port serial.Port
	PortMutex sync.Mutex
	TotalRx int
	ValidRx int
	Timeout int
}
func (f *ADT03)WriteBuffer(addr int32,cmd []byte)error{

	//bylog.Debug("write buffer %d -> % x",addr, cmd)
	if _,err:=f.Port.Write(cmd);err!=nil{
		bylog.Error("Write sensor[%d] err %v",addr,err)

		return err
	}
	return nil
}

func (f *ADT03)calibrateZero(addr int32)error{

	return f.writeInt(addr,0x4,0x00010101)

}
func (f *ADT03)calibrateK(addr int32,weight int32)error{

	return f.writeInt(addr,0x7,weight)

}
func (f *ADT03)sendReadWeight(addr int32)error{

	return f.writeCmd(addr,0x82)
}

func (f *ADT03)getReadWeight(ss *Sensor)(error){
	var result [8]byte
	if n,err:=f.readBuffer(ss.Addr,result[:]);err!=nil || n!= 8{

		if err!=nil{
			//bylog.Error("read sensor[%d] err %v",ss.Addr,err)
			return err
		}
		if n!=8{
			return fmt.Errorf("read sensor[%d] length = %d",ss.Addr,n)
		}

	}
	if result[0] != 0x55 || result[1] != 0xAA{
		//错误的数据格式.
		return fmt.Errorf("error data %v",result)
	}
	if int32(result[2]) != ss.Addr{
		return fmt.Errorf("mismatch %v addr want=%d,get=%d",result,ss.Addr,result[2])
	}


	//零位指示
	if result[6]&0x1 != 0{
		ss.State.Zero = true
	}else{
		ss.State.Zero = false
	}
	//bit1 稳定
	if result[6]&0x2 != 0{
		ss.State.Still = true
	}else{
		ss.State.Still = false
	}
	//bit2 超载故障
	if result[6]&0x4 != 0{
		ss.State.Overflow = true
	}else{
		ss.State.Overflow = false
	}
	//bit4 传感器故障
	if result[6]&0x10 != 0{
		ss.State.SensorErr = true
	}else{
		ss.State.SensorErr = false
	}

	//var uv = uint16(byutil.GetLittleInt16(result[4:6]))
	//传送出来的数是无符号2字节整数，加上符号位指示能表示的范围是 -65535~65535
	if result[6]&0x20 != 0{
		ss.State.Positive = false
		//var  v int32 = 0xFFFF0000
		var  ux  = int32(uint16(byutil.GetLittleInt16(result[4:6])))
		if ux == 0{
			ss.Value = 0
		}else{
			ss.Value = -ux
		}

	}else{
		ss.State.Positive = true
		ss.Value = int32(uint16(byutil.GetLittleInt16(result[4:6])))
	}
	//if ss.Addr == 8{
	//	bylog.Debug("value=%d result=% x",ss.Value,result)
	//}

	//如果有传感器错误，直接输出一个最大值.
	if ss.State.SensorErr{
		ss.Value = 65535
	}
	return nil
}
func (f *ADT03)writeCmd(addr int32,cmd byte)(err error) {
	buff:=[]byte{0xFE,0x7F,byte(addr),cmd,0}
	//求异或.
	buff[4] = byutil.Xor(buff[0:4])

	return f.WriteBuffer(addr,buff)
}
func (f *ADT03)writeInt(addr int32,cmd byte,value int32)(err error) {

	//求异或.
	f.PortMutex.Lock()
	defer func() {
		f.PortMutex.Unlock()
	}()
	buff:=[8]byte{0xFE,0x7F,byte(addr),cmd,0,0,0,0}
	buff[4] = byte(value&0xff)
	buff[5] = byte((value>>8)&0xff)
	buff[6] = byte((value>>16)&0xff)

	buff[7]=byutil.Xor(buff[0:7])
	bylog.Debug("writeInt % x",buff)

	f.WriteBuffer(addr,buff[:])
	var result [8]byte
	if _,err:=f.readBuffer(addr,result[:]);err!=nil{
		bylog.Error("writeInt failed %v",err)
		return err
	}

	return nil
}
func (f *ADT03)readBuffer(addr int32,result []byte)(n int, err error){

	if n,err=io.ReadFull(f.Port, result[:]);err!=nil{
		return 0,err
	}
	if n!=len(result){
		return n,errors.New("len not match")
	}
	if result[0] != 0x55 || result[1] != 0xAA{
		//错误的数据格式.
		return 0,fmt.Errorf("error data %v",result)
	}

	return n,nil
}
func (f *ADT03)getInt24(buf []byte)(value int32) {

	value = int32(buf[2])<<16
	value = value + (int32(buf[1])<<8)
	value = value + int32(buf[0])
	return
}
func (f *ADT03)readInt(addr int32,cmd byte)(value int32,err error){
	f.PortMutex.Lock()
	defer func() {
		f.PortMutex.Unlock()
	}()
	f.writeCmd(addr,cmd)
	var result [8]byte

	if _,err:=f.readBuffer(addr,result[:]);err!=nil{
		return 0,errors.New("read failed")
	}
	//bylog.Debug("readInt=% x",result)
	value = f.getInt24(result[4:7])
	return value,nil
}
func (f *ADT03)sim(ss *Sensor)error{
	ss.Value = 0
	ss.State.Still = true
	ss.Timeout = 0
	ss.TimeStamp = byopenwrt.GetLocalNowTime().Unix()
	ss.Online = true
	return nil
}
func (f *ADT03)readSensor(ss *Sensor)(err error){
	//FE 7F 01 83 03

	f.TotalRx++
	f.PortMutex.Lock()
	defer func() {
		f.PortMutex.Unlock()
	}()

	//return f.sim(ss)
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

func (adt *ADT03) CalibFact(addr int32,action int32) (err error) {
	if action == 0{
		return adt.writeInt(addr,0x70,0x010101)
	}else{
		return adt.writeInt(addr,0x71,0x010101)
	}

}

//修改地址
func (ADT03) ModifyAddr(oldAddr, newAddr int32) error {
	panic("implement me")
}

func (adt *ADT03) Measure(ss *Sensor) (err error) {
	return adt.readSensor(ss)
}

func (adt *ADT03) GetAD(addr int32) (value int32,err error) {
	return adt.readInt(addr,0x83)
}
//mv * 100000
func (adt *ADT03) SetMv(addr int32,mv int32) (err error) {
	return adt.writeInt(addr,0xA,mv)
}
func (adt *ADT03) GetMv(addr int32) (mv int32,err error) {
	return adt.readInt(addr,0x86)
}
func (adt *ADT03) SetFull(addr int32,value int32) (err error) {
	return adt.writeInt(addr,0x9,value)
}
func (adt *ADT03) GetFull(addr int32) (value int32,err error) {
	return adt.readInt(addr,0x85)
}
func (adt *ADT03) CalZero(addr int32) error {
	return adt.calibrateZero(addr)
}

func (adt *ADT03) CalKg(addr int32, kg int32) error {
	return adt.calibrateK(addr,kg)
}
func (adt *ADT03)Open(port,baud int)error{
	config := serial.Config{
		Address:  byutil.GetUartName(port),
		BaudRate: baud,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
		Timeout:  100 * time.Millisecond,
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
func NewADT03()*ADT03{
	return &ADT03{}
}