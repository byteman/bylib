package byutil

import (
	"fmt"
	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)
func ParseMac(strMac string,split string)(mac []byte ,err error){
	macArr:=strings.Split(strMac,split)

	for _,v:=range macArr{
		x,err:=strconv.ParseInt(v,16,32)
		if err!=nil{
			return nil,err
		}

		mac=append(mac,byte(x))
	}
	return mac,nil
}
func MT7688_GetMAC()(mac []byte ,err error){

	aps:=cmd.NewCmd("cat","/sys/class/net/eth0/address")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return nil,status.Error
	}
	if len(status.Stdout) <= 0{
		return nil,errors.New("no output")
	}
	return ParseMac(status.Stdout[0],":")

}
func MT7688_FormataMAC()(string ,error){
	mac,err:=MT7688_GetMAC()
	if err!=nil{
		return "",err
	}
	if len(mac)!=6{
		return "",errors.New("mac len must = 6")
	}
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x",mac[0],mac[1],mac[2],mac[3],mac[4],mac[5]),nil

}