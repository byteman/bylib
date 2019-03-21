package byopenwrt

import (
	"bylib/byhttp"
	"bylib/bylog"
	"bylib/byutils"
	"fmt"
	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type SysAdmin struct{
	UploadFlag bool
}

//获取系统时间
func (s *SysAdmin)getSysTime(ctx *byhttp.MuxerContext)error {
	dt:=byutil.OpDateTime{

	}
	if err:=GetSysDateTime(&dt);err!=nil{
		bylog.Error("GetSysDateTime err=%v",err)
		return ctx.Json(400,err)
	}
	return ctx.Json(200,dt)
}
//设置系统时间
func (s *SysAdmin)setSysTime(ctx *byhttp.MuxerContext)error {

	dt:=byutil.OpDateTime{

	}
	if err:=ctx.BindJson(&dt);err!=nil{
		bylog.Error("setSysTime BindJson err=%v",err)
		return ctx.Json(400,err)
	}
	if err:=SetSysDateTime(dt);err!=nil{
		bylog.Error("SetSysDateTime error=%v",err)
		return ctx.Json(400,err)
	}
	return ctx.Json(200,"ok")
}
//上传文件
func (s *SysAdmin)doUpload(dst string)error{
	//执行命令
	//sysupgrade -n -v xx.bin
	s.UploadFlag = true
	aps:=cmd.NewCmd("touch","/tmp/failsafe")
	status := <-aps.Start()
	if status.Error!=nil{
		bylog.Error("sysupgrade err=%v",status.Error)

	}
	aps=cmd.NewCmd("/sbin/sysupgrade","-n","-v",dst)

	//等待aps完成
	status = <-aps.Start()
	if status.Error!=nil{
		bylog.Error("sysupgrade err=%v",status.Error)
		return status.Error
	}
	msg:=strings.Join(status.Stdout,"")
	if strings.Contains(msg,"Invalid image type"){
		return errors.New("错误的文件格式")
	}

	return status.Error
}
func (s *SysAdmin)UploadFile(ctx *byhttp.MuxerContext)error{
	file, err := ctx.FormFile("file")
	if err!=nil{
		bylog.Error("FromFile err %v",err)
		return ctx.Json(400,err)
	}
	bylog.Debug("upload file name=%s",file.Filename)

	dst:="/tmp/"+file.Filename
	err=ctx.SaveUploadedFile(file,dst)
	if err!=nil{
		bylog.Error("SaveUploadedFile %s err %v",dst,err)
		return ctx.Json(400,err)
	}
	if err:=s.doUpload(dst);err!=nil{
		return ctx.Json(400,err)
	}

	return ctx.Json(200,"ok")

}
func (s *SysAdmin)HookSignal(){
	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c,  syscall.SIGTERM,  syscall.SIGKILL,syscall.SIGHUP, syscall.SIGINT,  syscall.SIGQUIT)

	go func() {
		for sig := range c {
			switch sig {
			case syscall.SIGTERM, syscall.SIGKILL,syscall.SIGHUP, syscall.SIGINT,  syscall.SIGQUIT:
				bylog.Debug("Receive SIGTERM,SIGKILL,SIGHUP")
				if s.UploadFlag{
					bylog.Debug("Ignore for upload")
				}else{
					//Debug("Kill me")
					//os.Exit(0)
				}

			default:
				fmt.Println("other", s)
			}
		}
	}()
}

func (self *SysAdmin)ResetDevice(ctx *byhttp.MuxerContext)error{
	aps:=cmd.NewCmd("reboot","")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		ctx.Json(400,status.Error)
	}

	return ctx.Json(200,"OK")
}
func (s *SysAdmin)Start()error{
	//s.HookSignal()
	byhttp.GetMuxServer().Post("/upload",s.UploadFile)
	byhttp.GetMuxServer().Get("/device/reset",s.ResetDevice)
	byhttp.GetMuxServer().Get("/time",s.getSysTime)
	byhttp.GetMuxServer().Post("/time",s.setSysTime)

	return nil
}
func SysAdminGet()*SysAdmin{

	return &SysAdmin{
		UploadFlag:false,
	}
}