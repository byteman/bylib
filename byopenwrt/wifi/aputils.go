package wifi

import (
	"bufio"
	"bylib/bylog"
	"bylib/byutils"
	"fmt"
	"io"
	"strings"
)

type ApConnState struct {
	Connected bool `json:"connected"` //是否连接成功
	Ap ApSignal `json:"ap"` //连接成功的Ap信息
}


func ReadLine(msg string)(lines []string, err error) {


	buf := bufio.NewReader(strings.NewReader(msg))
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)

		if err != nil {

			if err == io.EOF {
				lines=append(lines,line)
				return lines,nil
			}
			return lines,err
		}
		lines=append(lines,line)
	}
	return lines,nil
}
func parseApState(lines []string,state *ApConnState)error{
	if len(lines)!=12{
		state.Connected=false
		return fmt.Errorf("line must 12")
	}

	line1:=strings.Fields(lines[0])
	if len(line1) != 3{
		return fmt.Errorf("line1 must 3")
	}
	state.Ap.SSID=strings.Trim(line1[2],"\"")
	line2:=strings.Fields(lines[1])
	if len(line2) != 3{
		return fmt.Errorf("line2 must 3 %d",len(line2))
	}
	state.Ap.BSSID = strings.Trim(line2[2]," ")

	fmt.Println(line1)
	return nil
}
//获取指定ap的连接状态
func GetApConnState(name string)ApConnState{
	apsta:=ApConnState{
		Connected:false,
	}
	stdout,_,err:=byutil.Run("iwinfo",name,"info")
	if err!=nil{
		return  apsta
	}
	if err:=parseApState(stdout,&apsta);err!=nil{
		bylog.Error("pasreAp err=%s",err)
		return apsta
	}
	if apsta.Ap.BSSID=="00:00:00:00:00:00"{
		return apsta
	}
	apsta.Connected = true
	return apsta
}
func WifiReload()error{
	_,_,err:=byutil.Run("/sbin/wifi","reload")
	return err
}