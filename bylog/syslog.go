package bylog

import (
	"github.com/hashicorp/go-syslog"
)

type BySysLogger struct {
	logger gsyslog.Syslogger
}

func NewSysLogger(fac,name string) (ByLogger,error) {
	logger,err:=gsyslog.NewLogger(gsyslog.LOG_DEBUG,fac,name)
	if err!=nil{
		//fmt.Println("Create Logger failed ",err)
		return nil,err
	}
	return &BySysLogger{
		logger:logger,
	},nil
}
func (l *BySysLogger) WriteLevel(p int ,data []byte) error {
	return l.logger.WriteLevel(gsyslog.Priority(p), data)
}

func (l *BySysLogger) Write(data []byte) (int, error) {
	return l.logger.Write(data)
}

func (l *BySysLogger) Close() error {
	return  l.logger.Close()
}
