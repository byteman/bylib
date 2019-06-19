package byopenwrt

import (
	"bylib/byutils"
	"github.com/go-cmd/cmd"
	"path/filepath"
	"time"
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
func ResetAfter(delayS int)  {

	time.AfterFunc(time.Duration(delayS)*time.Second,func(){
		Reset()
	})

}

func SDExist()bool{

	files, _ := filepath.Glob("/dev/mmcblk0*")
	if len(files) == 0{
		return false
	}
	if exist,_:=byutil.PathExists("/mnt/mmc/weizhi_sdcard.txt");!exist{
		return false
	}
	return true
}
func USBExist()bool{

	files, _ := filepath.Glob("/dev/sd*")
	if len(files) == 0{
		return false
	}
	if exist,_:=byutil.PathExists("/mnt/usb/weizhi_usb.txt");!exist{
		return false
	}
	return true
}