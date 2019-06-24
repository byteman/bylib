
默认的组播地址：239.255.255.250 端口 9999


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