package bybus

import (
	"bylib/bylog"
	"bylib/byutils"
	"github.com/pkg/errors"
	"github.com/tarm/serial"
	"github.com/xiegeo/modbusone"
	"sync"
)
const size = 0x10000

//读寄存器处理函数
type ModbusReadHandleFunc func ()([]uint16,error )
//写寄存器处理函数.
type ModbusWriteHandleFunc func ([]uint16)(error)

type ModbusReadWriteHandler struct{
	Addr 	uint16 //寄存器地址
	Quality uint16 //寄存器个数
	ReadFunc ModbusReadHandleFunc //读处理函数.
	WriteFunc ModbusWriteHandleFunc //写处理函数.
}
type MBServer struct{
	//mbserv *mbserver.Server //modbus 服务器j
	Port int
	Baud int
	//inputRegisters [size]uint16
	holdingRegisters [size]uint16
	holdMux sync.Mutex
	readWriteHandler map[uint16]*ModbusReadWriteHandler
}

func (mb *MBServer)RegisterHandler(addr, quality uint16 ,
	readFunc  ModbusReadHandleFunc,
	writeFunc ModbusWriteHandleFunc ){
	mb.readWriteHandler[addr] = &ModbusReadWriteHandler{
		Addr:addr,
		Quality:quality,
		ReadFunc:readFunc,
		WriteFunc:writeFunc,
	}
}

func (mb *MBServer)ReadInputRegsToBuffer(addr int, reg int,nr int) []byte {
	register :=  addr+reg
	endRegister := register + nr
	mb.holdMux.Lock()
	defer func() {
		mb.holdMux.Unlock()
	}()
	return byutil.MbUint16ToBytes(mb.holdingRegisters[register:endRegister])
}
func (mb *MBServer)Close(){

}

func (mb *MBServer)Open(port,baud int)error{
	go mb.open(port,baud)
	return nil
}

func (mb *MBServer)WriteHoldingRegisters(address uint16, values []uint16) error {
	//bylog.Info("WriteInputRegisters from %v, quantity %v value=%v\n", address, len(values),values)
	mb.holdMux.Lock()
	defer func() {
		mb.holdMux.Unlock()
	}()
	for reg:=address; reg < (address+uint16(len(values)));{
		if v,ok:=mb.readWriteHandler[reg];ok{
			if v.WriteFunc!=nil{
				off:=reg-address
				if off > uint16(len(values)){
					return errors.New("address invalid")
				}
				//寄存器的长度由各自函数去处理.
				if err:=v.WriteFunc(values[off:]);err!=nil{
					return err
				}
				reg+=v.Quality
				continue
			}
		}
		reg+=1
	}
	for i, v := range values {
		mb.holdingRegisters[address+uint16(i)] = v
	}

	return nil
}

//lcd显示屏读取寄存器.
func (mb *MBServer)readHoldingRegisters(address, quantity uint16) ([]uint16, error) {
	//bylog.Info("ReadInputRegisters from %v, quantity %v\n", address, quantity)
	mb.holdMux.Lock()
	defer func() {
		mb.holdMux.Unlock()
	}()
	for reg:=address; reg < (address+quantity);{
			if v,ok:=mb.readWriteHandler[reg];ok{
				if v.ReadFunc!=nil{
					values,err:=v.ReadFunc()
					if err!=nil {
						return nil,err
					}
					copy(mb.holdingRegisters[reg:],values)
					reg+=v.Quality
					continue
				}
			}
			reg+=1

	}

	return mb.holdingRegisters[address : address+quantity], nil
}
func (mb *MBServer)open(port,baud int)error{
	config := serial.Config{
		Name:     byutil.GetUartName(port),
		Baud:     38400,
		StopBits: serial.StopBits(1),
		Parity:'N',
	}

	s, err := serial.OpenPort(&config)
	if err != nil {
		bylog.Error( "open serial error: %v\n", err)
		return err
	}
	com := modbusone.NewSerialContext(s, int64(baud))
	defer func() {
		bylog.Debug("%+v\n", com.Stats())
		com.Close()
	}()


	id, err := modbusone.Uint64ToSlaveID(1)
	if err != nil {
		bylog.Error("set slaveID error: %v\n", err)
		return err
	}


	device := modbusone.NewRTUServer(com, id)

	h := modbusone.SimpleHandler{


		ReadHoldingRegisters: mb.readHoldingRegisters,
		WriteHoldingRegisters:mb.WriteHoldingRegisters,

		OnErrorImp: func(req modbusone.PDU, errRep modbusone.PDU) {
			bylog.Error("error received: %v from req: %v\n", errRep, req)
		},
	}
	err = device.Serve(&h)
	if err != nil {
		bylog.Error("serve error: %v\n", err)
		return err
	}
	return nil
}
func NewModbusServer() *MBServer{

	return &MBServer{
		readWriteHandler:make(map[uint16]*ModbusReadWriteHandler),
	}
}
