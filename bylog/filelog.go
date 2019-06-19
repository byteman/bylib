package bylog

import (
	"log"
	"os"
)

type ByFileLogger struct {
	logFile *os.File
	l *log.Logger
}
func NewFileLoggr(file string)(ByLogger,error)  {
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("file open error : %v", err)
	}
	l := log.New(f,"[DEBUG] ", log.Ldate|log.Ltime|log.Llongfile)

	return &ByFileLogger{
		l:l,
	},nil

}
func (l *ByFileLogger) WriteLevel(prio int, data []byte) error {
	if l.l!=nil{
		l.l.Println(string(data))
	}
	return nil
}

func (l *ByFileLogger) Write(data []byte) (int, error) {
	if l.l!=nil{
		l.l.Println(string(data))
	}
	return len(data),nil
}

func (l *ByFileLogger) Close() error {
	if l.logFile!=nil{
		l.logFile.Close()
	}
	return nil
}

