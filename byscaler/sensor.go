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

type Sensor struct{
	Addr  int32 //传感器地址
	Value int32 //传感器的值
	State  State //状态.
	Timeout int32 //超时计数器.
	TimeStamp int64 //采集的时间戳.
	Online bool
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