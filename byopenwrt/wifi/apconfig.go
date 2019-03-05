package wifi

import (
	"bylib/bylog"
	"bylib/byutils"
	"encoding/json"
	"io/ioutil"
	"sync"
)

type ApConfig struct{
	ApList []*ApInfo `json:"aplist"` //ap连接列表
	RetryTime int `json:"retry_time"` //AP重连和检测时间秒
	Lock sync.Mutex
}
//列举保存所有的ap列表
func (c *ApConfig)GetApList()[]*ApInfo{
	return c.ApList
}
//加载ap配置.
func (c *ApConfig)LoadConfig()error{
	c.RetryTime=30 //默认60s
	path,err:=byutil.GetCurrentPath()
	if err!=nil{
		bylog.Error("getCurrentPath %v",err)
	}
	data, err := ioutil.ReadFile(path+"/ap.json")
	if err != nil {
		return err
	}
	var tmp []*ApInfo

	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	c.Lock.Lock()
	c.ApList = tmp
	c.Lock.Unlock()
	return nil
}
//保存ap列表
func (c *ApConfig)SaveApList()error{

	data, err := json.Marshal(&c.ApList)
	if err != nil {
		return err
	}

	path,err:=byutil.GetCurrentPath()
	if err!=nil{

	}

	return ioutil.WriteFile(path+"/ap.json",data,666)
}

//添加一个ap
func (c *ApConfig)AddAp(ap *ApInfo)error  {
	c.Lock.Lock()
	c.ApList = append(c.ApList,ap)
	c.Lock.Unlock()
	return nil
}
func remove(slice []*ApInfo, ssid string) []*ApInfo{
	if len(slice) == 0 {
		return slice
	}
	tmp:=make([]*ApInfo,0)
	for _, v := range slice {
		if v.SSID != ssid {
			tmp = append(tmp,v)
		}
	}
	return tmp
}

//删除一个ap
func (c *ApConfig)RemoveAp(ssid string)error  {
	c.Lock.Lock()
	c.ApList = remove(c.ApList,ssid)
	c.Lock.Unlock()
	return nil
}
