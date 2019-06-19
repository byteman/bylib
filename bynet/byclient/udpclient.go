package byclient

import (
	"bylib/bylog"
	"bylib/bynet/byudp"
	"bylib/byproto"
	"errors"
	"fmt"
	"net"
)

type UDPClientConfig struct{
	IpAddr string
	Port   int
}
/**
UDP的发送和接收管理框架
1.发送数据，可以支持缓存发送，多线程发送等
2.接收，可以支持安装消息回调函数.
 */
type UDPClient struct{
	Config UDPClientConfig
	Conn 	*byudp.ClientConn
	Handlers map[int]byproto.Handler
	Codec  byudp.Codec
}


func (c *UDPClient) SendBytes(data []byte) error {
	panic("implement me")
}
func (c *UDPClient) SendMessage(msg byudp.Message) error {
	return c.Conn.Write(msg)
}

func (c *UDPClient)SetConfig(config UDPClientConfig)error{
	c.Config.IpAddr = config.IpAddr
	c.Config.Port   = config.Port
	return nil
}
func (u *UDPClient)Start()error {

	localip := net.ParseIP(":")
	remoteip := net.ParseIP(u.Config.IpAddr)
	lAddr := &net.UDPAddr{IP: localip, Port: 8888}
	rAddr := &net.UDPAddr{IP: remoteip, Port: u.Config.Port}

	c, err := net.DialUDP("udp", lAddr,rAddr)

	if err != nil {
		return err
	}

	onConnect := byudp.OnConnectOption(func(c byudp.WriteCloser) bool {
		bylog.Debug("on connect")
		return true
	})

	onError := byudp.OnErrorOption(func(c byudp.WriteCloser) {
		bylog.Error("on error")
	})

	onClose := byudp.OnCloseOption(func(c byudp.WriteCloser) {
		bylog.Info("on close")
	})

	onMessage := byudp.OnMessageOption(func(msg byudp.Message, c byudp.WriteCloser) {
		bylog.Debug("------------------------ recv")
		if handler,ok:=u.Handlers[int(msg.MessageNumber())];ok{

			g:=msg.(*byproto.WzMsg)
			handler(g)
		}
	})
	codec:=byudp.CustomCodecOption(byproto.WzProtoCodec{})
	options := []byudp.ServerOption{
		codec,
		onConnect,
		onError,
		onClose,
		onMessage,
		byudp.ReconnectOption(),
	}

	conn := byudp.NewClientConn(0, c, options...)
	if conn == nil {

		bylog.Error("New client err")
		return errors.New("New client error")
	}
	u.Conn = conn
	//byudp.Register(MSG_RT_WGT,)
	conn.Start()

	return nil
}

func (s *UDPClient)ReStart()error{
	return s.Stop()

}
func (c *UDPClient)Stop()error{

	if c.Conn!=nil{
		c.Conn.Close()
	}

	return nil

}
//安装接收句柄
func (c *UDPClient)InstallHandler(cmd int,handler byproto.Handler)error{
	if _,ok:=c.Handlers[cmd];ok{
		return fmt.Errorf("cmd[%d] has exist",cmd)
	}
	c.Handlers[cmd] = handler
	return nil
}

func NewUdpClient(config UDPClientConfig,codec byudp.Codec)*UDPClient {
	return &UDPClient{
		Config:config,
		Handlers:make(map[int]byproto.Handler),
		Codec:codec,
	}
}