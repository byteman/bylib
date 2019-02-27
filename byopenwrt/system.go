package byopenwrt

import (
	"github.com/go-cmd/cmd"
)

func Reset()error  {
	aps:=cmd.NewCmd("reboot")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return status.Error
	}
	return nil

}