package bydefine

//伟志通用modbus寄存器.
const(
	REG_WGT=0  //重量寄存器.
	REG_STATE=2 //重量状态.
	REG_DOT=3 //小数点点位
	REG_DIV_HIGH=8
	REG_CALIB=20
	REG_FULL_SPAN=26

	REG_ADDR=30
	REG_4B_CORN_K=36 //40037/40038 传感器1号角差系数(1000代表1.000)
	REG_2B_AUTO_CORN=44 //自动角差控制：0：启动标定；1：标定传感器1；2：标定传感器2；3：标定传感器3；4：标定传感器4；5：结束标定
	REG_2B_SENSOR_NUM=46//传感器个数
	REG_4B_CHANNEL_AD=49 //第一路通道的AD值 整数值
	REG_2B_ALALOG_LO=59 //第一路通道的AD值 整数值
	REG_2B_ALALOG_HI=60 //第一路通道的AD值 整数值
	REG_2B_ALALOG_EXIT=61 //退出模拟模式.
	REG_4B_SUM_WGT=62 // 总输出重量.
	REG_2B_INPUT=64 // 外部IO输入 0 无输入 1 上升沿 2 下降沿.
	REG_2B_SAVE_FACTORY=66 //保存校准值
	REG_2B_LOAD_FACTORY=67 //恢复校准值

	REG_TEDS_SAVE =79 //保存teds参数
	REG_TEDS_LOAD=80 //加载TEDS参数
	REG_4B_TEDS_SERIAL_NO =87
	REG_4B_TEDS_DATE=89 //时间日期.
	REG_4B_TEDS_CAP_F32=91 //满量程
	REG_4B_TEDS_SEN_F32=93 //灵敏度
	REG_2B_CHANGE_CHAN=101 //切换通道.
	REG_2B_ALALOG_FIX_VALUE=95 //模拟量4ma-20ma标定值.
)