package byopenwrt

import (
	"bylib/bylog"
	"fmt"
	"github.com/go-cmd/cmd"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"net"
	"strings"
)
const(
	NetTypeStatic=0
	NetTypeDhcp=iota
	NetTypeUnknown=iota
)
type NetConfig struct{
	Type int
	Ip string
	Mask string
	Gateway string
	Mac string
}
//设置网络参数
func SetNetWork(ifname string,cfg *NetConfig)error{
	bylog.Debug("SetNetWork2=%+v",cfg)
	var netype="static"
	//暂时不支持DHCP方式.
	//switch cfg.Type {
	//case NetTypeDhcp:
	//	netype="dhcp"
	//case NetTypeStatic:
	//	netype="static"
	//default:
	//
	//}
	if err:=UciSetString(fmt.Sprintf("network.%s.proto",ifname),netype);err!=nil{
		bylog.Error("network.lan.proto=%s %v",netype,err)
		return err
	}
	if err:=UciSetString(fmt.Sprintf("network.%s.netmask",ifname),cfg.Mask);err!=nil{
		bylog.Error("network.lan.netmask=%s %v",cfg.Mask,err)
		return err
	}
	if err:=UciSetString(fmt.Sprintf("network.%s.ipaddr",ifname),cfg.Ip);err!=nil{
		bylog.Error("network.lan.ipaddr=%s %v",cfg.Ip,err)
		return err
	}
	if err:=UciSetString(fmt.Sprintf("network.%s.macaddr",ifname),cfg.Mac);err!=nil{
		bylog.Error("network.lan.mac=%s %v",cfg.Mac,err)
		return err
	}

	if err:=UciCommit();err!=nil{
		bylog.Error("UciCommit failed %s %v",cfg.Ip,err)
		return err
	}
	//return nil
	////重新启动服务.
	return ServiceRestart("network")

}
//ubus call network.interface.wan status | grep nexthop | grep -oE '([0-9]{1,3}.){3}.[0-9]{1,3}'

func GetNetWorkStatus(ifname string,cfg *NetConfig)error{
	aps:=cmd.NewCmd("ubus","call",fmt.Sprintf("network.interface.%s",ifname),"status")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		return status.Error
	}
	if len(status.Stdout) <= 0{
		return fmt.Errorf("output empty")
	}
	output:=strings.Join(status.Stdout,"")
	cfg.Ip = gjson.Get(output,"ipv4-address.0.address").String()
	mask := gjson.Get(output,"ipv4-address.0.mask").Uint()
	cmask:=net.CIDRMask(int(mask),32)
	cfg.Mask = fmt.Sprintf("%d.%d.%d.%d",cmask[0],cmask[1],cmask[2],cmask[3])

	return nil
}
func GetNetWorkStatus2(ifname string,cfg *NetConfig)error{
	itf,er:=net.InterfaceByName(ifname)
	if er!=nil{
		return er
	}
	addrs,er:=itf.Addrs()
	if er!=nil{
		return er
	}
	if len(addrs) < 1{
		return errors.New("no addr")
	}
	return nil

}

func CheckNetWorkConfig(cfg *NetConfig)error{
	return nil
	address := net.ParseIP(cfg.Ip)
	if address == nil {
		return errors.New("ip format error")
	}
	mask := net.ParseIP(cfg.Mask)
	if mask == nil {
		return errors.New("mask format error")
	}
	return nil
}
//获取网络参数
func GetNetWorkConfig(ifname string,cfg *NetConfig)error{

	if err:=CheckNetWorkConfig(cfg);err!=nil{
		return err
	}

	ip,err:=UciGetString(fmt.Sprintf("network.%s.ipaddr",ifname))
	if err!=nil{
		return err
	}
	cfg.Ip = ip
	bylog.Debug("ip=%s",ip)
	netype,err:=UciGetString(fmt.Sprintf("network.%s.proto",ifname))
	if err!=nil{
		return err
	}
	switch netype{
	case "static":
		cfg.Type = NetTypeStatic
	case "dhcp":
		cfg.Type=NetTypeDhcp
	default:
		cfg.Type=NetTypeUnknown
	}
	mask,err:=UciGetString(fmt.Sprintf("network.%s.netmask",ifname))
	if err!=nil{
		return err
	}
	cfg.Mask = mask

	mac,err:=UciGetString(fmt.Sprintf("network.%s.macaddr",ifname))
	if err!=nil{
		return err
	}
	cfg.Mac = mac
	return nil
}

//修改wan或者lan接口对应的设备
//ifname 接口名称
//mode wan lan
func ChangeIfDevice(ifname string,device string)error{
	UciSetString(fmt.Sprintf("network.%s.ifname",ifname),device)
	return nil
}
//修改eth01的mode为lan还是wan
//eth0.2是一个vlan 对应的是以太网的端口0,也就是我们板子上目前使用的口子
//eth0.1是另外一个vlan，对应的是以太网的端口1-4
func ChangeEth02Mode(mode string)error{
	switch mode {
	case "lan": //把网口设置为lan口，也就是可以分配ip给连上的电脑,路由器模式用这个.
		ChangeIfDevice("lan","eth0.2")
		ChangeIfDevice("wan","eth0.1")
	case "wan": //wan口，从上级获取ip，或者设置静态ip.iot模式时用这个
		ChangeIfDevice("lan","eth0.1")
		ChangeIfDevice("wan","eth0.2")
	}
	UciCommit()
	return ServiceRestart("network")

}