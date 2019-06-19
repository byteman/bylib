package bybus



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

type ModbusServer interface {
	RegisterHandler(addr, quality uint16 , readFunc  ModbusReadHandleFunc, writeFunc ModbusWriteHandleFunc )
	ReadInputRegsToBuffer(addr int, reg int,nr int) []byte
	WriteHoldingRegisters(address uint16, values []uint16) error
}

