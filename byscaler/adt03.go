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

//adt03的管理类.
//adt03是串口接口的，所以需要依赖一个串口类
type ADT03 struct {
	Port serial.Port
	PortMutex sync.Mutex
	TotalRx int
	ValidRx int
	Timeout int
}
func (f *ADT03)WriteCmd(addr int32,cmd []byte)error{
	f.PortMutex.Lock()
	defer func() {
		f.PortMutex.Unlock()
	}()
	if _,err:=f.Port.Write(cmd);err!=nil{
		bylog.Error("Write sensor[%d] err %v",addr,err)

		return err
	}
	return nil
}
func (f *ADT03)ReadCmd(addr int32,result []byte)(n int, err error){
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
func (f *ADT03)calibrateZero(addr int32)error{

	cmd:=[8]byte{0xFE,0x7F,byte(addr),0x4,1,1,1,0}
	cmd[7]=byutil.Xor(cmd[0:7])
	bylog.Debug("CalibrateZero %v",cmd)
	if err:=f.WriteCmd(addr,cmd[:]);err!=nil{
		return err
	}

	return nil
}
func (f *ADT03)calibrateK(addr int32,weight int32)error{

	cmd:=[8]byte{0xFE,0x7F,byte(addr),0x7,0,0,0,0}
	cmd[4] = byte(weight&0xff)
	cmd[5] = byte((weight>>8)&0xff)
	cmd[6] = byte((weight>>16)&0xff)

	cmd[7]=byutil.Xor(cmd[0:7])
	bylog.Debug("CalibrateK %v",cmd)
	if err:=f.WriteCmd(addr,cmd[:]);err!=nil{
		return err
	}

	return nil

}
func (f *ADT03)sendReadWeight(addr int32)error{
	cmd:=[]byte{0xFE,0x7F,byte(addr),0x82,0}
	//求异或.
	cmd[4] = byutil.Xor(cmd[0:4])

	return f.WriteCmd(addr,cmd)
}

func (f *ADT03)getReadWeight(ss *Sensor)(error){
	var result [8]byte
	if n,err:=f.readCmd(ss.Addr,result[:]);err!=nil || n!= 8{

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

	ss.Value = int32(byutil.GetLittleInt16(result[4:6]))
	//fmt.Println("weight=",ss.Weight)
	if result[6]&0x2 != 0{
		ss.State.Still = true
	}else{
		ss.State.Still = false
	}
	//0x111000 芯片故障|传感器故障|开机零点故障
	if result[6]&0x38 != 0{
		ss.State.Error = true
	}else{
		ss.State.Error = false
	}
	return nil
}
func (f *ADT03)writeCmd(addr int,cmd []byte)error{
	f.PortMutex.Lock()
	defer func() {
		f.PortMutex.Unlock()
	}()
	if _,err:=f.Port.Write(cmd);err!=nil{
		bylog.Error("Write sensor[%d] err %v",addr,err)

		return err
	}
	return nil
}
func (f *ADT03)readCmd(addr int32,result []byte)(n int, err error){
	f.PortMutex.Lock()
	defer func() {
		f.PortMutex.Unlock()
	}()
	if n,err=io.ReadFull(f.Port, result[:]);err!=nil{

		if err!=nil{
			return 0,err
		}
	}

	return n,nil
}
func (f *ADT03)readSensor(ss *Sensor)(err error){
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
func (ADT03) ModifyAddr(oldAddr, newAddr int32) error {
	panic("implement me")
}

func (adt *ADT03) Measure(ss *Sensor) (err error) {
	return adt.readSensor(ss)
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