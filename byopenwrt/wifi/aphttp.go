package wifi

import (
	"bylib/byhttp"
	"bylib/bylog"
)

//扫描ap列表
func scanApList(ctx *byhttp.MuxerContext)error{

	var aplist []ApSignal
	var err error
	if aplist,err=DefaultApAdmin().ScanApList();err!=nil{
		bylog.Error("scan err=%s",err)
		ctx.Json(400,err)
	}
	bylog.Debug("aplist=%v",aplist)
	return ctx.Json(200,aplist)
}
func listAp(ctx *byhttp.MuxerContext)error{

	return ctx.Json(200,DefaultApAdmin().ListAp())
}
func connectAp(ctx *byhttp.MuxerContext)error {

	ap:=ApInfo{}
	if err:=ctx.BindJson(&ap);err!=nil {
		return ctx.Json(400, err)
	}

	if err:=DefaultApAdmin().ConnectAp(ap.SSID,ap.PassWord);err!=nil {
		return ctx.Json(400, err)
	}
	return ctx.Json(200,"OK")
}
//添加一个AP
func addAp(ctx *byhttp.MuxerContext)error{

	ap:=ApInfo{}
	if err:=ctx.BindJson(&ap);err!=nil{
		return ctx.Json(400,err)
	}
	if err:=DefaultApAdmin().AddAp(ap);err!=nil{
		return ctx.Json(400,err)
	}
	return ctx.Json(200,"OK")
}
func removeAp(ctx *byhttp.MuxerContext)error {
	ap:=ApInfo{}
	if err:=ctx.BindJson(&ap);err!=nil{
		ctx.Json(400,err)
	}
	if err:=DefaultApAdmin().RemoveAp(ap);err!=nil{
		return ctx.Json(400,err)
	}
	return ctx.Json(200,"OK")
}

func ApHttpInit(){
	byhttp.GetMuxServer().Get("/ap/scan",scanApList)
	byhttp.GetMuxServer().Get("/ap/list",listAp)
	byhttp.GetMuxServer().Post("/ap/connect",connectAp)
	byhttp.GetMuxServer().Post("/ap/add",addAp)
	byhttp.GetMuxServer().Post("/ap/remove",removeAp)
}