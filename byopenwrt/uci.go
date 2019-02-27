package byopenwrt

import (
	"fmt"
	"github.com/go-cmd/cmd"
)

func UciGetString(key string)(str string,err error){
	aps:=cmd.NewCmd("/sbin/uci","get",key)

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return "",status.Error
	}
	if len(status.Stdout) <= 0{
		return "",fmt.Errorf("UciGetString %s empty",key)
	}
	return status.Stdout[0],nil
}
func UciCommit()error{
	aps:=cmd.NewCmd("/sbin/uci","commit")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return status.Error
	}
	return nil
}
func UciSetString(key string,value string)(err error){
	aps:=cmd.NewCmd("/sbin/uci","set",fmt.Sprintf("%s=%s",key,value))

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return status.Error
	}
	return nil
}
