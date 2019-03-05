package byudp

import (
	"bufio"
	"fmt"
	"github.com/leesper/holmes"
	"github.com/leesper/tao/examples/chat"
	"net"
	"os"
	"testing"
)

func TestUdp(t *testing.T) {
	defer holmes.Start().Stop()
	holmes.Infof("start")
	c, err := net.Dial("udp", "127.0.0.1:12345")
	if err != nil {
		holmes.Fatalln(err)
	}
	fmt.Println("c=",c)
	onConnect := OnConnectOption(func(c WriteCloser) bool {
		holmes.Infoln("on connect")
		return true
	})

	onError := OnErrorOption(func(c WriteCloser) {
		holmes.Infoln("on error")
	})

	onClose := OnCloseOption(func(c WriteCloser) {
		holmes.Infoln("on close")
	})

	onMessage := OnMessageOption(func(msg Message, c WriteCloser) {
		fmt.Print(msg.(chat.Message).Content)
	})

	options := []ServerOption{
		onConnect,
		onError,
		onClose,
		onMessage,
		ReconnectOption(),
	}

	conn := NewClientConn(0, c, options...)
	if conn==nil{
		fmt.Println("new client err")
		return
	}
	defer conn.Close()

	conn.Start()
	for {
		reader := bufio.NewReader(os.Stdin)
		talk, _ := reader.ReadString('\n')
		if talk == "bye\n" {
			break
		} else {
			msg := chat.Message{
				Content: talk,
			}
			if err := conn.Write(msg); err != nil {
				holmes.Infoln("error", err)
			}
		}
	}
	fmt.Println("goodbye")
}
