package wifi

import (
	"time"
)

//表示配置的ap连接信息
type ApInfo struct {
	SSID string `json:"ssid"` //连接的ssid
	PassWord string `json:"password"` //密码
	Connected bool `json:"connected"` //是否被连接
	IP string `json:"ip"` //连接的ip地址
	Selected bool `json:"selected"` //是否是当前连接的ssid.
	ConnStamp int64 `json:"-"` //上次连接时间.
}

//比较当前时间-上次连接时间是否超过阀值
func (ai *ApInfo)Beyond(s int)bool{
	if int(time.Now().Unix()-ai.ConnStamp) > s{
		return true
	}
	return false
}