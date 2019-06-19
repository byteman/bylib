package byscaler

//传感器协议接口.
type SensorProtocol interface {
	marshal()[]byte
}
type Addr int32

type State struct{
	Positive  bool //正负号 0 正 1 负数
	SensorErr bool 	//传感器故障、脱落 bit5  0 正常 1 故障
	Error  bool		//设备故障. bit3
	Overflow bool 	//重量溢出. bit2
	Still bool 		//是否稳定.	bit1
	Zero bool 		//是否零位. bit0
}
//单个传感器
type Sensor struct{
	Addr  int32 `json:"addr"`//传感器地址
	Value int32 `json:"value"`//传感器的原始值
	CalcValue int32 `json:"calc_value"`//计算值 = 原始值-零点
	State  State //状态.
	StateValue uint8 `json:"state_value"`
	Timeout int32 `json:"timeout"`//超时计数器.
	TimeStamp int64 `json:"time_stamp"`//采集的时间戳.
	Online bool `json:"online"`
	StillCount uint8 //稳定计数器
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