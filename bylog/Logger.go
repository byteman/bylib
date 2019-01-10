package bylog

import (
	"fmt"
	"github.com/hashicorp/go-syslog"
)
var logger gsyslog.Syslogger
var debug = false
var level = 0
func Debug(format string, a ...interface{})  {


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
	logger,err=gsyslog.NewLogger(gsyslog.LOG_DEBUG,"USER","mbserver")
	if err!=nil{
		fmt.Println("Create Logger failed ",err)
	}
	return err

}