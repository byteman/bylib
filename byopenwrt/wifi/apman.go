package wifi

import (
	"bylib/bylog"
	"bylib/byutils"
	"encoding/json"
	"fmt"
	"github.com/go-cmd/cmd"
	"runtime"
	"strconv"
	"strings"
	"time"
)
type ConnResult struct{
	Result string `json:"result"`
	Message string `json:"message"`
	IP string `json:"ip"`
	Connect bool `json:"connect"`
}
//ap管理器
type  ApManager struct{
	Config ApConfig //配置文件.
}
func (self *ApManager)UciGetString(key string)(str string,err error){
	aps:=cmd.NewCmd("/sbin/uci","get",key)

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		bylog.Error("UciGetString %s err=%v",key,status.Error)
		return "",status.Error
		//return ctx.Json(400,status.Error)
	}
	if len(status.Stdout) <= 0{
		return "",fmt.Errorf("UciGetString %s empty",key)
	}
	return status.Stdout[0],nil
}
//列出目前配置了的多连接Ap列表
func (self *ApManager)ListAp()[]*ApInfo{
	//apcli=`ifconfig | grep apcli0`

	//ip=`ifconfig apcli0 | grep 'inet addr:'| cut -d: -f2 | awk '{ print $1}'`
	//LogDebug("ListAp")
	self.ClearApState()
	ap,err:=self.GetCurrAp()
	if err!=nil{
		bylog.Error("GetCurAp err=%v",err)
	}else{
		//找到一个当前选择项
		if c:=self.FindAp(ap.SSID);c!=nil{
			c.Connected = ap.Connected
			c.Selected = true
			c.IP = ap.IP
		}
	}
	return self.Config.ApList
}
func (self *ApManager)GetCurrApStatus()(*ConnResult){
	aps:=cmd.NewCmd("/usr/sbin/checkwifi","1")
	result:=&ConnResult{}
	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		bylog.Error("checkwifi err=%v",status.Error)
		return nil
		//return ctx.Json(400,status.Error)
	}

	if len(status.Stdout) > 0{
		//for _, line := range status.Stdout {
		//	fmt.Println(line)
		//}
		if err:=json.Unmarshal([]byte(status.Stdout[0]),&result);err!=nil{
			bylog.Error("checkwifi err=%v",err)
			return nil
		}else{
			//LogDebug("result=%v",result)
			if result.Result=="success"{
				result.Connect = true
			}else{
				result.Connect = false
			}
			return result
		}
	}
	return nil

}
//func (self *ApManager)UciGetInt(key string)(int,err error){
//	return 0,nil
//}
//获取当前配置了的ap
func (self *ApManager)GetCurrAp()(ap *ApInfo,err error){
	//uci get wireless.@wifi-iface[0].ApCliSsid
	ap=&ApInfo{}

	ap.SSID,err = self.UciGetString("wireless.@wifi-iface[0].ApCliSsid")
	if err!=nil{
		return
	}
	//LogDebug("ssid=%s",ap.SSID)
	ap.PassWord,err = self.UciGetString("wireless.@wifi-iface[0].ApCliPassWord")
	if err!=nil{
		return
	}
	//LogDebug("password=%s",ap.PassWord)
	//var result *ConnResult
	result:=self.GetCurrApStatus()
	if result !=nil{
		ap.IP = result.IP
		ap.Connected = result.Connect
		ap.Selected = true
		//LogDebug("ip=%s ,conn=%v",ap.IP,ap.Connected)
	}
	return
}
func (self *ApManager)FindAp(ssid string)*ApInfo{
	for _,ap:=range self.Config.ApList{
		if ap.SSID == ssid{
			return ap
		}
	}
	return nil
}
func (self *ApManager)ClearApState(){
	for _,ap:=range self.Config.ApList{
		ap.Selected = false
		ap.Connected = false
		ap.IP=""

	}
}

//扫描获取ap信号列表
func (self *ApManager)ScanApList()(aplist []ApSignal,err error){


	if runtime.GOOS == "windows" {
		for i:=0 ; i < 5; i++{
			aplist = append(aplist,ApSignal{
				SSID:fmt.Sprintf("test%d",i+1),

			})
		}

		return aplist,nil
	}
	//iwpriv ra0 set SiteSurvey=1
	aps:=cmd.NewCmd("iwpriv","ra0","set","SiteSurvey=1")

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{

		return nil,byutil.FormatError("iwpriv ra0 set SiteSurvey=1",status.Error)
	}
	aps=cmd.NewCmd("sleep","5")

	//等待aps完成
	status = <-aps.Start()
	if status.Error!=nil{
		return nil,byutil.FormatError("sleep err",status.Error)
	}
	aps=cmd.NewCmd("iwpriv","ra0","get_site_survey")

	//等待aps完成
	status = <-aps.Start()
	if status.Error!=nil{
		return nil,byutil.FormatError("iwpriv ra0 get_site_survey",status.Error)
	}

	//iwpriv ra0 get_site_survey
	for i, line := range status.Stdout {

		if i < 2{
			continue
		}
		//line = strings.TrimSpace(line)
		apstr := strings.Fields(line)
		//apstr:=strings.Split(line," ") //这个分隔有问题
		if len(apstr) != 8{
			bylog.Error("%s len not 8",apstr)
			continue
		}
		//fmt.Printf("index=%d line=%s cont=%d\n",i,line,len(apstr))
		//fmt.Printf("1=%s\n",apstr[0])
		//fmt.Printf("2=%s\n",apstr[1])
		//fmt.Printf("3=%s\n",apstr[2])
		//fmt.Printf("4=%s\n",apstr[3])
		//fmt.Printf("5=%s\n",apstr[4])
		//fmt.Println(apstr)
		ap:=ApSignal{}
		var err error
		ap.Channle,err =strconv.Atoi(apstr[0])
		if err!=nil{
			bylog.Error("channel err=%s",err)
			continue
		}


		ap.SSID = apstr[1]
		ap.BSSID = apstr[2]
		ap.Security = apstr[3]
		ap.Signal,err =strconv.Atoi(apstr[4])
		if err!=nil{
			bylog.Error("Signal err=%s",err)
			continue
		}
		aplist=append(aplist,ap)

	}
	//bylog.Debug("aplist=%+v",aplist)
	return aplist,nil

}
//列举连接ap列表

func (self *ApManager)ConnectAp(ssid,passwd string)error{
	aps:=cmd.NewCmd("/usr/sbin/setwifi",ssid,passwd)

	//等待aps完成
	status := <-aps.Start()
	if status.Error!=nil{
		bylog.Error("set wifi err=%v",status.Error)
	}
	return status.Error
}


//添加一个新的ap
func (self *ApManager)AddAp(ap ApInfo)error{

	self.Config.AddAp(&ap)
	return self.Config.SaveApList()
}


//删除某个ap
func (self *ApManager)RemoveAp(ap ApInfo)error {

	self.Config.RemoveAp(ap.SSID)
	return self.Config.SaveApList()
}
//测试外网地址.
func (self *ApManager)ping(url string)bool{
	return true
}
func (s *ApManager)findConnectAps(api []*ApInfo, aps ApSignalList)(apis []*ApInfo){
	for _,ai:=range api{
		if _,find:=aps.Find(ai.SSID);find{
			//信号找找到了该ID，检查该信号上次连接时间
			apis=append(apis,ai)
		}
	}
	return
}
func (s *ApManager)findMinTimeAp(apis []*ApInfo)( ai *ApInfo){

	for i,ap:=range apis{
		if i == 0{
			ai = ap
			continue
		}
		if ap.ConnStamp < ai.ConnStamp{
			ai = ap
		}
	}
	return
}
//从aps列表中查找一个可连接的信号
func (s *ApManager)findNextAp(api []*ApInfo, aps ApSignalList )*ApInfo{

	apis:=s.findConnectAps(api,aps)
	//找不到信号，或者找到的信号长度为0
	if apis==nil || len(apis) ==0 {
		return nil
	}
	return s.findMinTimeAp(apis)
	//找到了可以连接的ap列表，查找其中时间最小的
	//for _,ai:=range api{
	//	if _,find:=aps.Find(ai.SSID);find{
	//		//信号找找到了该ID，检查该信号上次连接时间
	//		if ai.Beyond(s.Config.RetryTime){
	//			return ai
	//		}
	//	}
	//}
	//return nil
}
/**
自动多连接线程
1.每隔1分钟读取当前ap连接情况，检查wifi是否已经启动成功，不成功就等待，因为wifi重启需要一定的时间。
2.ifconfig apcli0 查看是否有连接成功的ap，如果成功就跳过
3.没有连接成功的话，调用aps获取在线ap列表，比较本地连接列表中，找到待连接列表中信号最强的一个ap，并且连接次数最小的，如果没有搜索到一个，跳转到第一步，否则下一步
4.把当前ap切换为最强信号ap，并且把该ap的连接次数加1，重启wifi ，连接，跳转到第一步


 */
func (self *ApManager)runMultiConn(){
	for{
		//每隔一段时间检测一次ap连接状态.
		bylog.Debug("retry=%d",self.Config.RetryTime)
		time.Sleep(time.Duration(self.Config.RetryTime) * time.Second)
		ap:=GetApConnState("apcli0")
		bylog.Debug("%+v",ap)
		if ap.Connected && self.ping(""){
			//已经连接成功了，可以判断外网情况，但是有时候不需要测试外网情况
			bylog.Debug("wifi has connected to %s",ap.Ap.SSID)
			continue
		}
		//网络不通，搜索可连接的信号，连接下一个.
		var aps ApSignalList
		var err error
		if aps,err=self.ScanApList();err!=nil{
			//一个都搜索不到,跳到下次
			bylog.Error("can not find ap list %s",err)
			continue
		}
		//搜索到了一个，从本地连接列表中查找一个满足要求的ap

		api:=self.findNextAp(self.Config.ApList,aps)
		if api==nil{
			//找不到一个可以连接的对象
			bylog.Error("can not find nextAp")
			continue
		}
		bylog.Debug("find next ap=%s ready connect",api.SSID)
		//找到了可连接的ap,连接他.有可能这个ap是目前正在连接的.那么直接重启网络就可以了
		curAp,err:=self.GetCurrAp()

		if err!=nil{
			bylog.Error("GetCurrAp err=%s ",err)
			//没有配置过AP，切换ap
			if err:=self.ConnectAp(api.SSID,api.PassWord);err!=nil{
				bylog.Error("ConnectAp err=%s",err)
			}
		}else{
			//配置过AP，并且跟目前不一致，那么切换ap
			bylog.Debug("curAp=%s connect ap=%s",curAp.SSID,api.SSID)
			if curAp.SSID != api.SSID{
				//不是同一个，切换用户名和密码
				if err:=self.ConnectAp(api.SSID,api.PassWord);err!=nil{
					bylog.Error("ConnectAp err=%s",err)
				}
			}
		}
		//更新时间戳.
		api.ConnStamp = time.Now().Unix()
		//重启网络
		bylog.Debug("WifiReload")
		WifiReload()

	}
}


//启动AP管理器
func (self *ApManager)Start(enHttp bool)error{
	bylog.Debug("ApManager Start------")

	if err:=self.Config.LoadConfig();err!=nil{
		bylog.Error("LoadConfig failed %v",err)
	}
	//获取当前配置好了的AP
	ap,err:=self.GetCurrAp()
	bylog.Debug("Find current Ap = %v",ap)
	if err == nil && ap!=nil{
		if self.FindAp(ap.SSID) == nil{
			bylog.Debug("Save Current Ap ")
			//在已保存待连接WIFI列表中找不到当前AP，则加入并保存
			self.Config.AddAp(ap)
			self.Config.SaveApList()
		}
	}
	if enHttp{
		ApHttpInit()
	}

	go self.runMultiConn()
	return nil
}
var apm ApManager
func DefaultApAdmin()*ApManager{

	return &apm
}