package byopenwrt

import (
	"bylib/bylog"
	"bylib/byutils"
	"fmt"
	"github.com/go-cmd/cmd"
	"strconv"
	"strings"
	"time"
)


//把硬件时钟设置为系统时钟
func SyncHwClockFromSysClock()error{
	aps:=cmd.NewCmd("hwclock","-w")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return status.Error
	}
	return nil
}
//设置系统时钟
func SetSysDateTime(odt byutil.OpDateTime)error{
	date:=fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d" ,int(2000+int(odt.Year)),odt.Month,odt.Day,odt.Hour,odt.Minute,odt.Second)
	//time:=fmt.Sprintf("%02d:%02d:%02d",odt.Hour,odt.Minute,odt.Second)
	//date := time.Date(int(odt.Year),
	//	time.Month((odt.Month)),
	//	int(odt.Day),
	//	int(odt.Hour),
	//	int(odt.Minute),
	//	int(odt.Second),0,time.Local).Format("01/02/2006 15:04:05")
	bylog.Debug("date=%s",date)
	aps:=cmd.NewCmd("/bin/date",date)

	//等待aps完成
	status := <-aps.Start()
	//这个命令有点奇怪，无论是执行成功失败都会返回成功.
	//bylog.Debug("%v",status.Error)
	//bylog.Debug("%v %v" ,status.Stdout,status.Stderr)
	if status.Error!=nil{
		return status.Error
	}
	if len(status.Stderr) > 0{
		return fmt.Errorf("%v",strings.Join(status.Stderr,"\r\n"))
	}

	return nil
}
func SetHardDateTime(odt byutil.OpDateTime)error{
	return nil
}
func GetHardDateTime(odt *byutil.OpDateTime)(err error){
	return nil
}
func MbDateTimeBytes()[]uint16{
	buf:=[6]byte{}
	odt:=byutil.OpDateTime{}
	err:=GetSysDateTime(&odt)
	if err!=nil{
		bylog.Error("GetSysDateTime err=%v",err)
		return byutil.MbNowBytes()
	}
	buf[0] = byte(odt.Year)
	buf[1] = byte(odt.Month)
	buf[2] = byte(odt.Day)
	buf[3] = byte(odt.Hour)
	buf[4] = byte(odt.Minute)
	buf[5] = byte(odt.Second)

	return byutil.MbBytesToUint16(buf[:])

}
func GetSysDateTime(odt *byutil.OpDateTime)error{
	//x:=time.Now() //这种方式获取到的时间是utc时间
	aps:=cmd.NewCmd("date","+%Y-%m-%d-%H-%M-%S")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return status.Error
	}
	if len(status.Stdout) <= 0{
		return fmt.Errorf("output empty")
	}


	datetime:=[6]int{}
	dates:=strings.Split(status.Stdout[0],"-")
	//bylog.Debug("time=%s",status.Stdout[0])
	for i,v:=range dates{
		x,err:=strconv.ParseInt(v,10,32)
		if err!=nil{
			return err
		}
		datetime[i] = int(x)
	}
	//bylog.Debug("datetime=% x",datetime)
	odt.Year = uint8(datetime[0]-2000)
	odt.Month= uint8(datetime[1])
	odt.Day =uint8(datetime[2])
	odt.Hour=uint8(datetime[3])
	odt.Minute=uint8(datetime[4])
	odt.Second=uint8(datetime[5])
	return nil
}
//在无涯的openwrt板子上，无法获取到本地的 时区，所以手工指定一个时区，但是必须把时区文件
//拷贝到板子里面. 设置GOROOT的目录$GOROOT/lib/time/zoneinfo.zip
//还有一个更好的办法，直接设置一个固定的时区.
//无论哪种时区，影响的只是显示的时间，但是不会影响unix输出，各个时区的unix输出都是一样的.
var  shanghaiLoc *time.Location = nil
func GetLocalNowTime()time.Time{
	if shanghaiLoc==nil{
		shanghaiLoc = time.FixedZone("CST", 8*3600)       // 东八
		//shanghaiLoc,err=time.LoadLocation("Asia/Shanghai")

	}
	return time.Now().In(shanghaiLoc)

}