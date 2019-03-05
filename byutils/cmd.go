package byutil

import "github.com/go-cmd/cmd"

func Run(name string,args ...string)(output []string,error []string, err error){
	aps:=cmd.NewCmd(name,args...)

	status := <-aps.Start()

	return status.Stdout,status.Stderr,nil
}