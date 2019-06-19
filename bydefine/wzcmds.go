package bydefine

//伟志通用命令码
const (
	PROTO_CMD_EMPTY = 0x00+iota
	//参数读取和写入命令
	M2S_ACK_READ_PAR		//主机回复从机的读参数命令，返回主机的参数信息
	M2S_ACK_WRITE_PAR	//主机回复从机的写参数命令，返回写入的结果
	M2S_ASK_READ_PAR		//主机请求从机发送参数命令
	M2S_ASK_WRITE_PAR	//主机请求写参数到从机


	S2M_ACK_READ_PAR		//从机回复主机的读参数命令，返回从机的参数信息
	S2M_ACK_WRITE_PAR	//从机回应回复主机的写参数命令
	S2M_ASK_READ_PAR		//从机请求读主机的参数
	S2M_ASK_WRITE_PAR	//从机请求写参数到主机

	//重量和状态更新

	M2S_ACK_UPDATE_WET	//主机回复更新重量的结果，是否更新成功
	M2S_ASK_UPDATE_WET	//主机请求更新重量到从机内部。

	S2M_ACK_UPDATE_WET	//从机回复主机更新重量的请求，返回从机的重量
	S2M_ASK_UPDATE_WET	//从机请求更新重量到主机中.携带从机的重量

	//数据和状态更新

	M2S_ACK_DATA	//主机回复更新数据的结果，是否更新成功
	M2S_ASK_DATA	//主机请求更新数据到从机内部。

	S2M_ACK_DATA	//从机回复主机更新数据信息的请求，返回从机的数据
	S2M_ASK_DATA	//从机请求更新数据到主机中.携带从机的数据
	//通用控制命令

	M2S_ACK_DO_CMD	//主机响应从机控制主机后的结果
	M2S_ASK_DO_CMD		//主机请求控制从机执行命令
	S2M_ACK_DO_CMD	//从机回复主机自己执行命令的结果
	S2M_ASK_DO_CMD	//从机请求主机去执行命令

	//标定
	M2S_ASK_UPDATE_CAL	//请求更新标定
	S2M_ASK_UPDATE_CAL  //请求更新标定

	//用户自定义命令区	 从0x40开始定义

)
