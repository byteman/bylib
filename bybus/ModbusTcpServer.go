package bybus

import (
	"bylib/bylog"
	"bylib/byutils"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/tbrandon/mbserver"
	"sync"
)


type MBTcpServer struct{
	mbserv *mbserver.Server //modbus 服务器
	holdMux sync.Mutex
	readWriteHandler map[uint16]*ModbusReadWriteHandler
}

func (s *MBTcpServer) RegisterHandler(addr, quality uint16, readFunc ModbusReadHandleFunc, writeFunc ModbusWriteHandleFunc) {
	s.readWriteHandler[addr] = &ModbusReadWriteHandler{
		Addr:addr,
		Quality:quality,
		ReadFunc:readFunc,
		WriteFunc:writeFunc,
	}
}

//更新服务器上寄存器的值.
func (s *MBTcpServer)WriteHoldingRegisters(address uint16 ,values []uint16)error  {
	s.holdMux.Lock()
	defer func() {
		s.holdMux.Unlock()
	}()
	reg := int(address)
	for i,r:=range values{
		s.mbserv.HoldingRegisters[reg+i] = r
	}
	return nil

}
func (s *MBTcpServer)ReadInputRegsToBuffer(addr int, reg int,nr int) []byte {
	register :=  addr+reg
	endRegister := register + nr
	return mbserver.Uint16ToBytes(s.mbserv.HoldingRegisters[register:endRegister])
}
func (s *MBTcpServer)Close(){
	if s.mbserv!=nil{
		s.mbserv.Close()
	}
}
func registerAddressAndValues(frame mbserver.Framer) (int, []uint16) {
	data := frame.GetData()
	register := int(binary.BigEndian.Uint16(data[0:2]))
	return register, byutil.MbBytesToUint16(data[5:])
}
func registerAddressAndValue(frame mbserver.Framer) (int, uint16) {
	data := frame.GetData()
	register := int(binary.BigEndian.Uint16(data[0:2]))

	value := binary.BigEndian.Uint16(data[2:4])
	return register, value
}
func registerAddressAndNumber(frame mbserver.Framer) (addr,register int, numRegs int, endRegister int) {
	data := frame.GetData()
	register = int(binary.BigEndian.Uint16(data[0:2]))
	numRegs = int(binary.BigEndian.Uint16(data[2:4]))
	if tcp, ok := frame.(*mbserver.TCPFrame); ok {
		addr = int(tcp.Device)
	}
	endRegister = register + numRegs
	return addr,register, numRegs, endRegister
}

func (mb *MBTcpServer)readHoldingRegisters(address, quantity uint16) ([]uint16, error) {
	//bylog.Info("ReadInputRegisters from %v, quantity %v\n", address, quantity)

	for reg:=address; reg < (address+quantity);{
		if v,ok:=mb.readWriteHandler[reg];ok{
			if v.ReadFunc!=nil{
				values,err:=v.ReadFunc()
				if err!=nil {
					return nil,err
				}
				copy(mb.mbserv.HoldingRegisters[reg:],values)
				reg+=v.Quality
				continue
			}
		}
		reg+=1

	}

	return mb.mbserv.HoldingRegisters[address : address+quantity], nil
}


func (mb *MBTcpServer)writeHoldingRegisters(address uint16, values []uint16) error {
	//bylog.Info("WriteInputRegisters from %v, quantity %v value=%v\n", address, len(values),values)

	for reg:=address; reg < (address+uint16(len(values)));{
		if v,ok:=mb.readWriteHandler[reg];ok{
			if v.WriteFunc!=nil{
				off:=reg-address
				if off > uint16(len(values)){
					return errors.New("writeHoldingRegisters address invalid")
				}
				//寄存器的长度由各自函数去处理.
				if err:=v.WriteFunc(values[off:off+v.Quality]);err!=nil{
					return err
				}
				reg+=v.Quality
				continue
			}
		}
		reg+=1
	}

	return nil
}
//写单个寄存器.
func (mb *MBTcpServer)handleWriteHolding(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	mb.holdMux.Lock()
	defer func() {
		mb.holdMux.Unlock()
	}()
	register,values:=registerAddressAndValue(frame)
	//bylog.Debug("handleWriteHolding register=%d values=% x",register,values)

	if err:=mb.writeHoldingRegisters(uint16(register),[]uint16{values});err!=nil{
		bylog.Error("writeHoldingRegisters err=%v",err)
	}
	return mbserver.WriteHoldingRegister(s ,frame)

}
//0x10 写多个寄存器
func (mb *MBTcpServer)handleWriteMultiHolding(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
	mb.holdMux.Lock()
	defer func() {
		mb.holdMux.Unlock()
	}()
	register,values:=registerAddressAndValues(frame)
	//bylog.Debug("handleWriteMultiHolding register=%d values=% x",register,values)
	if err:=mb.writeHoldingRegisters(uint16(register),values);err!=nil{
		bylog.Error("handleWriteMultiHolding err=%v",err)
	}
	return mbserver.WriteHoldingRegisters(s ,frame)

}

//0x3 读多个寄存器 只需要把holdingRegister中的数据返回出去就可以了
func (mb *MBTcpServer)handleReadHolding(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {

	mb.holdMux.Lock()
	defer func() {
		mb.holdMux.Unlock()
	}()
	_, register, numRegs, endRegister := registerAddressAndNumber(frame)

	//bylog.Debug("addr=%d ,reg=%d num=%d,end=%d",addr,register,numRegs,endRegister)

	endRegister = register + numRegs
	if endRegister > 65536 {
		return []byte{}, &mbserver.IllegalDataAddress
	}
	//检测对应的寄存器是否有过滤函数，有的话执行过滤函数，拷贝过滤函数的结果到对应的holding寄存器，最后统一的返回出去.
	_,err:=mb.readHoldingRegisters(uint16(register), uint16(numRegs))
	if err!=nil{
		return nil,&mbserver.IllegalDataValue
	}
	return append([]byte{byte(numRegs * 2)},
		mbserver.Uint16ToBytes(s.HoldingRegisters[register:endRegister])...),
		&mbserver.Success
}
func NewModbusTcpServer(port int) *MBTcpServer{
	serv:=MBTcpServer{
		mbserv :mbserver.NewServer(),
		readWriteHandler : make(map[uint16]*ModbusReadWriteHandler),
	}
	err := serv.mbserv.ListenTCP(fmt.Sprintf(":%d",port))
	if err != nil {
		bylog.Error("ListenTCP %d error %v",port,err)
		return nil
	}

	serv.mbserv.RegisterFunctionHandler(3,serv.handleReadHolding)
	serv.mbserv.RegisterFunctionHandler(6,serv.handleWriteHolding)
	serv.mbserv.RegisterFunctionHandler(0x10,serv.handleWriteMultiHolding)
	return &serv
}
