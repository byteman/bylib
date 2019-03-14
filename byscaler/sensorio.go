package byscaler

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
