package byproto

type Handler func (msg CommMsg)error

type CommMsg interface {
	//把通用消息中的数据进行绑定到具体的数据
	BindBinary(v interface{})error
	//消息编号.
	MessageNumber() int32
}

type MsgProtocal interface {
	//注册需要接收的消息包.
	RegisterHandler(cmd int, handler Handler)
	//序列话和打包数据包.
	SendPacket(slaveAddr ,cmd  int, data []byte)error
}