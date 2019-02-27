package bylog

import (
	"fmt"
	"github.com/hashicorp/go-syslog"
)
var logger gsyslog.Syslogger
var debug = false
var _level = DEBUG_LEVEL
const (
	ERR_LEVEL=0
	INFO_LEVEL=iota
	DEBUG_LEVEL=iota
)
func Debug(format string, a ...interface{})  {

	if  DEBUG_LEVEL > _level{
		return
	}
	if logger!=nil{
		if debug{
			fmt.Println(fmt.Sprintf(format,a...))
		}
		logger.WriteLevel(gsyslog.LOG_DEBUG,[]byte(fmt.Sprintf(format,a...)))
	}else{
		fmt.Println(fmt.Sprintf(format,a...))
	}

}
func Info(format string, a ...interface{})  {
	if  INFO_LEVEL > _level{
		return
	}
	if logger!=nil{
		if debug{
			fmt.Println(fmt.Sprintf(format,a...))
		}
		logger.WriteLevel(gsyslog.LOG_INFO,[]byte(fmt.Sprintf(format,a...)))
	}else{
		fmt.Println(fmt.Sprintf(format,a...))
	}
}
func Error(format string, a ...interface{})  {

	if  ERR_LEVEL > _level{
		return
	}
	if logger!=nil{
		if debug{
			fmt.Println(fmt.Sprintf(format,a...))
		}
		logger.WriteLevel(gsyslog.LOG_ERR,[]byte(fmt.Sprintf(format,a...)))
	}else{
		fmt.Println(fmt.Sprintf(format,a...))
	}
}

func InitLogger(enable bool,level int)error{
	var err error
	debug = enable
	_level = level
	logger,err=gsyslog.NewLogger(gsyslog.LOG_DEBUG,"USER","mbserver")
	if err!=nil{
		fmt.Println("Create Logger failed ",err)
	}
	return err

}