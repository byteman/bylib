package byscaler

//传感器协议接口.
type SensorProtocol interface {
	marshal()[]byte
}
type Addr int32

type State struct{
	Error  bool		//设备故障
	Overflow bool 	//重量溢出
	Still bool 		//是否稳定
	Zero bool 		//是否零位.
}
//单个传感器
type Sensor struct{
	Addr  int32 //传感器地址
	Value int32 //传感器的值
	State  State //状态.
	Timeout int32 //超时计数器.
	TimeStamp int64 //采集的时间戳.
	Online bool
}
//传感器集合/称，一般对应一个集散器连接n个传感器
type SensorSet struct{
	Addr int32 //集散器ID
	//Num int //传感器个数.
	Online bool
	TimeStamp int64 //上次的时间戳
	Timeout int32 //超时计数器.

	Sensors map[int]*Sensor //传感器地址和传感器信息的map
}
func (ss *SensorSet)SensorNum()int{
	return len(ss.Sensors)
}

func (ss *SensorSet)Clear(){

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
//拷贝数据
func (ss *SensorSet)Copy(s *SensorSet){
	sn:=s.SensorNum()
	for i:=1; i <= sn; i++{
		if _,ok:=ss.Sensors[i];!ok{
			//没有数据
			ss.Sensors[i] = NewSensor(int32(i))
		}
		ss.Sensors[i].Value = s.Sensors[i].Value
		ss.Sensors[i].Addr = s.Sensors[i].Addr
		ss.Sensors[i].State = s.Sensors[i].State
		ss.Sensors[i].TimeStamp = s.Sensors[i].TimeStamp
		ss.Sensors[i].Timeout = s.Sensors[i].Timeout
	}
}
//比较两组传感器值，如果超过范围
func (ss *SensorSet)Compare(old *SensorSet,diff int32)(sg []*Sensor){

	//if len(old.Sensors) != len(ss.Sensors){
	//	//数据都不一样长，先拷贝
	//	old.Copy(ss)
	//	return
	//}
	sn:=ss.SensorNum()
	//直接遍历当前数据
	for i:= 1; i <= sn; i++{
		if _,ok:=old.Sensors[i];!ok{
			//没有数据
			old.Sensors[i] = NewSensor(int32(i))
			old.Sensors[i].Value = ss.Sensors[i].Value
			old.Sensors[i].Addr = ss.Sensors[i].Addr
			old.Sensors[i].State = ss.Sensors[i].State
			old.Sensors[i].TimeStamp = ss.Sensors[i].TimeStamp
			old.Sensors[i].Timeout = ss.Sensors[i].Timeout
			continue
		}
		df:=ss.Sensors[i].Value-old.Sensors[i].Value
		if df < 0{
			df=-df
		}

		//bylog.Debug("new=%d old=%d diff=%d %d",ss.Sensors[i].Value,old.Sensors[i].Value,df,diff)


		if df >= diff{

			sg=append(sg,&Sensor{
				Addr: int32(i),
				Value:df,
				State:ss.Sensors[i].State,
				TimeStamp:ss.Sensors[i].TimeStamp,
				Timeout:ss.Sensors[i].Timeout,
			})
		}
		//更新旧的值
		old.Sensors[i].Value = ss.Sensors[i].Value
		old.Sensors[i].Addr = ss.Sensors[i].Addr
		old.Sensors[i].State = ss.Sensors[i].State
		old.Sensors[i].TimeStamp = ss.Sensors[i].TimeStamp
		old.Sensors[i].Timeout = ss.Sensors[i].Timeout
	}
	return
}
func NewSensorSet(addr int32,num int )*SensorSet  {
	ss:=&SensorSet{
		Addr:addr,
		Sensors:make(map[int]*Sensor),
	}

	return ss
}
func NewSensor(addr int32)*Sensor  {
	return &Sensor{
		Value:0,
		Addr:addr,
		State:State{

		},
		Online:false,
		Timeout:0,
	}
}
//传感器管理接口集合
type SensorIO interface {
	//修改传感器地址.
	ModifyAddr(oldAddr, newAddr int32)error
	//更新测量信息
	Measure(ss *Sensor)(error)
	//标定系数
	CalZero(addr int32)error
	//标定重量
	CalKg(addr int32, kg int32)error
}