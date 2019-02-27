package byutil

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}
func GetUartName(nr int)string{
	var name string
	switch runtime.GOOS {
	case "windows":
		name = fmt.Sprintf("COM%d",nr)
		break
	case "linux":
		name = fmt.Sprintf("/dev/ttyS%d",nr)
		break
	default:
		name=fmt.Sprintf("/dev/ttyUSB%d",nr)
	}
	return name
}