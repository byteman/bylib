package byutil

//拷贝字符串中的内容到一个byte数组中.
func CopyString2Array(src string, dst []byte)int{
	return copy(dst, src)
}