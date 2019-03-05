package byutil

import (
	"bylib/bylog"
	"time"
)
//打印二进制数据
func HexDump(msg string, data []byte){
	bylog.Debug("%s % X",msg,data)
}
type TickInfo struct{
	start int64
	count int
}
var tickMap map[string]*TickInfo

func init(){
	tickMap=make(map[string]*TickInfo)
}
func FreqPrintf(name string){

	if _,ok:=tickMap[name];!ok{
		tickMap[name] = &TickInfo{
			start:time.Now().Unix(),
			count:0,
		}
	}
	item:=tickMap[name]
	item.count++

	diff:=time.Now().Unix() - item.start
	if diff==0{
		return
	}
	freq := int64(item.count) /  diff
	bylog.Debug("[%s]---> freq=%d", name,freq )
}