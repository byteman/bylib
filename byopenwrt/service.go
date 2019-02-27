package byopenwrt

import "github.com/go-cmd/cmd"

func ServiceRestart(name string)error{
	aps:=cmd.NewCmd("/etc/init.d/"+name,"restart")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return status.Error
	}
	return nil
}
func ServiceStop(name string)error{
	aps:=cmd.NewCmd("/etc/init.d/"+name,"stop")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return status.Error
	}
	return nil
}
func ServiceStart(name string)error{
	aps:=cmd.NewCmd("/etc/init.d/"+name,"start")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return status.Error
	}
	return nil
}