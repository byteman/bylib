package bymsg

type Handler func (msg CommMsg)error

type CommMsg interface {
	//把通用消息中的数据进行绑定到具体的数据
	BindBinary(v interface{})error
	//消息编号.
	MessageNumber() int32
}
//消息接口，每个消息都有自己的序列化函数
type Message interface {
	MessageNumber() int32
	Serialize() ([]byte, error)
}
type MsgProtocal interface {
	//注册需要接收的消息包.
	RegisterHandler(cmd int, handler Handler)
	//序列话和打包数据包.
	SendPacket(slaveAddr ,cmd  int, data []byte)error
}