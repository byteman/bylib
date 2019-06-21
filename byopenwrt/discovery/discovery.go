package discovery

import (
	"bylib/bydefine"
	"bylib/bylog"
	"bylib/byopenwrt"
	"encoding/json"
	"github.com/tidwall/gjson"
)

type Request struct{
	Cmd string `json:"cmd"`
}
type Response struct{
	Cmd 	string `json:"cmd"`
	Error 	int `json:"error"`
	Message string `json:"message"`
}

type RequestNetCfg struct{
	Request
	IfName string `json:"if_name"`
}
type ResponseDiscovery struct{
	Response
	NetCfg byopenwrt.NetConfig `json:"net_cfg"`
	Product bydefine.ProductModel `json:"product"`
}
func ErrorResponse(cmd string,errCode int, msg string)error{
	resp:=Response{
		Cmd:cmd,
		Error:errCode,
		Message:msg,
	}
	data,err:=json.Marshal(&resp)
	if err!=nil{
		return err
	}
	return discovery.Send([]byte(data))
}
func SuccResponse(resp interface{})error{
	data,err:=json.Marshal(resp)
	if err!=nil{
		return err
	}
	return discovery.Send(data)
}
var discovery MultiCaster
var product = bydefine.ProductModel{
	Name:"unknown",
	SerialNo:"123456",
}
/**
请求网络发现.
{
	"cmd":"request_discovery",
	"if_name":"wan"
}
响应读取网络配置
{
	"cmd":"response_discovery",
	"error":0,
	"message":"ok",
	"if_name":"wan"
	"net_cfg":{
		"ip":"10.10.10.2",
		"mac":"00:11:22:33:44:55",
		"netmask":"255.0.0.0",
	}
}
 */
func discoveryNotify(discovered Discovered) {
	bylog.Debug("MultiCastNotify addr=%s %s",discovered.Address,string(discovered.Payload))

	result:=gjson.Parse(string(discovered.Payload))

	cmd:=result.Get("cmd").String()

	switch cmd {
	case "request_netcfg":
		discoveryNetCfg(cmd,result)
		break
	case "modify_netcfg":
		modifyNetCfg(cmd,result)
		break
	}

}
/**
请求修改ip地址
{
	"cmd":"modify_netcfg",
	"ifname":"wan",
	"local_ip":"10.10.10.2",
	"mac":"00:11:22:33:44:55",
	"netmask":"255.0.0.0"
}
响应修改命令成功
{
	"cmd":"modify_netcfg",
	"error":0,
	"message":"ok",
}
 */
//修改网络配置.
func modifyNetCfg(cmd string,msg gjson.Result)error  {
	ifname:=msg.Get("ifname").String()
	netcfg:=byopenwrt.NetConfig{}
	if err:=byopenwrt.GetNetWorkConfig(ifname,&netcfg);err!=nil{
		return ErrorResponse(cmd,1,err.Error())
	}
	if msg.Get("local_ip").Exists(){
		netcfg.Ip = msg.Get("local_ip").String()
	}
	if msg.Get("mac").Exists(){
		netcfg.Mac = msg.Get("mac").String()
	}
	if msg.Get("netmask").Exists(){
		netcfg.Mask = msg.Get("netmask").String()
	}
	if err:=byopenwrt.SetNetWork(ifname,&netcfg);err!=nil{
		return ErrorResponse(cmd,2, err.Error())
	}
	return nil
}


func discoveryNetCfg(cmd string,msg gjson.Result)error  {

	ifname:=msg.Get("ifname").String()
	netcfg:=byopenwrt.NetConfig{}
	if err:=byopenwrt.GetNetWorkConfig(ifname,&netcfg);err!=nil{
		return ErrorResponse(cmd,1,err.Error())
	}

	return SuccResponse(ResponseDiscovery{
		Response:Response{
			Cmd:cmd,
			Error:0,
			Message:"ok",
		},
		NetCfg:netcfg,
		Product:product,
	})
}
func Discovery(model bydefine.ProductModel){
	product = model
	if err:=discovery.Listen(Settings{
		Notify: discoveryNotify,
	});err!=nil{
		bylog.Error("Discovery listen err=%v",err)
	}
}
func DiscoveryStop(){

}