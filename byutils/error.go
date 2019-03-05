package byutil

import (
	"bylib/bylog"
	"fmt"
)

func FormatError(msg string,err error)error{
	err2:=fmt.Errorf( "%s %s",msg,err.Error)
	bylog.Error("%s",err2.Error())
	return err2
}