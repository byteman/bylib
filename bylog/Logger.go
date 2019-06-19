package bylog

import (
	"fmt"
	"sync"
)
var logger ByLogger
var debug = false
var _level = DEBUG_LEVEL
var lock sync.Mutex
const (
	FATAL_LEVEL=0
	ERR_LEVEL=iota
	WARN_LEVEL=iota
	INFO_LEVEL=iota
	DEBUG_LEVEL=iota
)

type ByLogger interface {
	WriteLevel(int, []byte) error

	// Write is used to write a message at the default level
	Write([]byte) (int, error)

	// Close is used to close the connection to the logger
	Close() error
}
func Fatal(format string, a ...interface{}) {
	lock.Lock()
	defer lock.Unlock()
	if  FATAL_LEVEL > _level{
		return
	}
	if logger!=nil{
		if debug{
			fmt.Println(fmt.Sprintf(format,a...))
		}
		logger.WriteLevel(FATAL_LEVEL,[]byte(fmt.Sprintf(format,a...)))
	}else{
		fmt.Println(fmt.Sprintf(format,a...))
	}
}
func Debug(format string, a ...interface{})  {
	lock.Lock()
	defer lock.Unlock()
	if  DEBUG_LEVEL > _level{
		return
	}
	if logger!=nil{
		if debug{
			fmt.Println(fmt.Sprintf(format,a...))
		}
		logger.WriteLevel(DEBUG_LEVEL,[]byte(fmt.Sprintf(format,a...)))
	}else{
		fmt.Println(fmt.Sprintf(format,a...))
	}

}
func Warn(format string, a ...interface{})  {
	lock.Lock()
	defer lock.Unlock()
	if  WARN_LEVEL > _level{
		return
	}
	if logger!=nil{
		if debug{
			fmt.Println(fmt.Sprintf(format,a...))
		}
		logger.WriteLevel(WARN_LEVEL,[]byte(fmt.Sprintf(format,a...)))
	}else{
		fmt.Println(fmt.Sprintf(format,a...))
	}
}
func Info(format string, a ...interface{})  {
	lock.Lock()
	defer lock.Unlock()
	if  INFO_LEVEL > _level{
		return
	}
	if logger!=nil{
		if debug{
			fmt.Println(fmt.Sprintf(format,a...))
		}
		logger.WriteLevel(INFO_LEVEL,[]byte(fmt.Sprintf(format,a...)))
	}else{
		fmt.Println(fmt.Sprintf(format,a...))
	}
}
func Error(format string, a ...interface{})  {
	lock.Lock()
	defer lock.Unlock()
	if  ERR_LEVEL > _level{
		return
	}
	if logger!=nil{
		if debug{
			fmt.Println(fmt.Sprintf(format,a...))
		}
		logger.WriteLevel(ERR_LEVEL,[]byte(fmt.Sprintf(format,a...)))
	}else{
		fmt.Println(fmt.Sprintf(format,a...))
	}
}

func InitLogger(enable bool,level int)error{
	var err error
	debug = enable
	_level = level

	logger,err=NewSysLogger("user","logger")
	if err!=nil{
		fmt.Println("Create Logger failed ",err)
	}
	return err
}
func SetLogger(log ByLogger)  {
	lock.Lock()
	defer lock.Unlock()
	logger = log
}