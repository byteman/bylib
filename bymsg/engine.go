package bymsg

//消息引擎,一个消息引擎可以注册1个通道、1个协议分析器、1个

type Engine struct{

}

func (eng *Engine)RegisterHandler(cmd int, handler Handler){

}
//发送消息
func (eng *Engine)SendMessage(message Message)error{
	return nil
}