package byopenwrt

import (
	"fmt"
	"testing"
)

func TestGetLocalNowTime(t *testing.T) {

	fmt.Println(GetLocalNowTime().Unix())
}
