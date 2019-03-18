package byscaler

import (
	"bylib/bylog"
	"bylib/byutils"
	"sync"
)

const (
	MAX_STILL_COUNT=3
)

//传感器集合/称，一般对应一个称连接n个传感器
type SensorSet struct{
	Addr int32 `json:"addr"`//集散器ID 
	Online bool `json:"-"`
	TimeStamp int64 `json:"-"`//上次的时间戳 
	Timeout int32 `json:"-"`//超时计数器.
	Diffs map[int]int32 `json:"diffs"` //零点集合.
	Zeros map[int]int32 `json:"zeros"`//零点集合.
	Sensors map[int]*Sensor  //传感器地址和传感器信息的map
	Lock sync.Mutex `json:"-"`
}
func (ss *SensorSet)SensorNum()int{
	return len(ss.Sensors)
}
func (ss *SensorSet)Zero(addr int)int32{
	if z,ok:=ss.Zeros[addr];ok{
		return z
	}
	return 0
}
func (ss *SensorSet)SetDiff(addr int,value int32){
	ss.Diffs[addr] = value
}
func (ss *SensorSet)SetZero(addr int,value int32){
	ss.Zeros[addr] = value
}
func (ss *SensorSet)SetAllZero(value int32){
	ss.Lock.Lock()
	defer ss.Lock.Unlock()
	for _,s:=range ss.Sensors{
		s.CalcValue = value
	}

}
//清零,记录传感器的当前值到zero点列表.
func (ss *SensorSet)Clear(){
	ss.Lock.Lock()
	defer ss.Lock.Unlock()
	for addr,s:=range ss.Sensors{
	 	ss.Zeros[addr]=s.Value
	}
}
func (ss *SensorSet)GetErrSensor()(sensors []*Sensor){
	for _,s:=range ss.Sensors{
		if s.State.Error{
			sensors=append(sensors,&Sensor{
				Addr:s.Addr,
				Value:s.Value,
				CalcValue:s.CalcValue,
				StateValue:s.StateValue,
				State:s.State,

			})
		}

	}
	return
}
func (ss *SensorSet)Update(){
	ss.Lock.Lock()
	defer ss.Lock.Unlock()
	for _,s:=range ss.Sensors{
		s.CalcValue = s.Value
		if z,ok:=ss.Zeros[int(s.Addr)];ok{
			s.CalcValue = s.Value-z
		}

	}
}
func (ss *SensorSet)GetValues()[]uint16{
	sn:=ss.SensorNum()
	var values []uint16
	//传感器的个数，从1开始
	for i:=1; i <= sn; i++{
		values=append(values,uint16(ss.Sensors[i].Value))
	}
	return values
}
func (sset *SensorSet)CopySensorSet(sset2 *SensorSet){
	sset.Lock.Lock()
	defer sset.Lock.Unlock()
	for addr,s:=range sset2.Sensors{
		if s==nil{
			bylog.Debug("s==nil")
			continue
		}
		if _,ok:=sset.Sensors[addr];ok{
			bylog.Debug("copy id=%d addr=%d value=%d",sset.Addr,addr,s.CalcValue)
			sset.Sensors[addr].CalcValue = s.CalcValue
		}

	}
}
//拷贝数据
func (sset *SensorSet)CopySensors()(sr []*Sensor){
	sset.Lock.Lock()
	defer sset.Lock.Unlock()
	for _,s:=range sset.Sensors{

		sr=append(sr,&Sensor{
			Addr:s.Addr,
			Value:s.Value,
			CalcValue:s.CalcValue,
			StateValue:s.StateValue,
			State:s.State,

		})
		//if addr == 1{
		//	bylog.Debug("calc=%d",s.CalcValue)
		//}

	}
	return
}

//比较两组传感器值，如果超过范围
func (sset *SensorSet)Compare(old *SensorSet,diff int32,still uint8)(sg []*Sensor){

	sset.Lock.Lock()
	defer sset.Lock.Unlock()
	//直接遍历当前数据
	for addr,s:=range sset.Sensors{
		//当前重量不稳定，或者错误 不进行重量判断.
		//if sset.Addr==1 && addr==2{
		//	bylog.Debug("w=%d state=%x still=%d",s.Value,s.StateValue,s.StillCount)
		//}
		//if !s.State.Still || s.State.Error {
		if  s.State.Error {
			s.StillCount=0
			continue
		}

		//稳定次数大于某个数
		if s.StillCount < still{
			s.StillCount++
			continue
		}
		s.StillCount=0
		//判断上次记录中是否有该地址.
		if _,ok:=old.Sensors[addr];!ok{
			//没有数据,并且重量稳定,仅仅更新值.

			old.Sensors[addr] = NewSensor(int32(addr))
			old.Sensors[addr].Value = s.Value
			old.Sensors[addr].CalcValue = s.CalcValue
			old.Sensors[addr].Addr = s.Addr
			old.Sensors[addr].State = s.State
			old.Sensors[addr].TimeStamp = s.TimeStamp
			old.Sensors[addr].Timeout = s.Timeout
			old.Sensors[addr].StateValue = s.StateValue

			continue
		}
		df:=byutil.Abs(int(s.CalcValue),int(old.Sensors[addr].CalcValue))

		//bylog.Debug("addr=%d new=%d old=%d diff=%d %d",addr,s.CalcValue,old.Sensors[addr].CalcValue,df,diff)

		if sdiff,ok:=sset.Diffs[addr];ok{
			//如果有传感器自己的sdiff，就用传感器自己的sdiff.

			diff = sdiff
		}else{
			//否则设置成默认的.
			sset.Diffs[addr]= int32(diff)
		}

		if df > int(diff){

			sg=append(sg,&Sensor{
				Addr: int32(addr),
				CalcValue:s.CalcValue - old.Sensors[addr].CalcValue,
				State:s.State,
				StateValue:s.StateValue,
				TimeStamp:s.TimeStamp,
				Timeout:s.Timeout,
			})
			//重量阀值超过旧的值，才更新旧的值
			old.Sensors[addr].Value = s.Value
			old.Sensors[addr].CalcValue = s.CalcValue
			old.Sensors[addr].Addr = s.Addr
			old.Sensors[addr].State = s.State
			old.Sensors[addr].StateValue = s.StateValue

			old.Sensors[addr].TimeStamp = s.TimeStamp
			old.Sensors[addr].Timeout = s.Timeout
		}else{
			//否则如果小于阀值，但是
		}

	}
	return
}
func NewSensorSet(addr int32 )*SensorSet  {
	ss:=&SensorSet{
		Addr:addr,
		Sensors:make(map[int]*Sensor),
		Zeros:make(map[int]int32),
		Diffs:make(map[int]int32),
	}

	return ss
}