package wifi

//{
//"ssid":"111",
//"signal",100， //信号强度
//"auth":"aes", //加密方式
//"mac":"xx:xx:xx", //mac地址
//}
//Ap信号
type ApSignal struct{
	SSID string `json:"ssid"` //ssid
	BSSID string `json:"bssid"` //就是ap的mac地址
	Signal int `json:"signal"` //信号强度
	Channle int `json:"channle"` //信道
	Security string `json:"security"` //加密方式
}


type ApSignalList []ApSignal

func (apl ApSignalList)AddApSignal(aps *ApSignal)  {

}
//在信号列表中查找
func (apl ApSignalList)Find(ssid string)(ap ApSignal,find bool)  {
	for _,ap:=range apl{
		if ap.SSID == ssid{
			return ap,true
		}
	}
	return ap,false

}