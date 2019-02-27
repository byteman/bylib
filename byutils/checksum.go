package byutil

func Xor(data []byte)byte{
	var sum=data[0]
	tmp:=data[1:]
	for _,v:=range tmp {
		sum = sum ^ v
	}
	return sum
}