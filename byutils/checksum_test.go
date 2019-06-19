package byutil

import "testing"

func Test_Xor2(t *testing.T) {
	//cmd:=[8]byte{0xFE,0x7F,1,0x4,0,0,0,0}
	//cmd[7]=Xor(cmd[0:6])
	//t.Logf("xor=%d",cmd[7])
}
func TestHello(t *testing.T)  {

}
type MyPerson struct {
	Age int32
	Tall int8
}
func TestEncodeStruct(t *testing.T)  {
	p:=MyPerson{
		Age:10,
		Tall:1,
	}
	data:=encode(p)

	t.Logf("data= % x ",data)
}