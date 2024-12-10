package file

import "strings"

// format 函数接受一个文件名作为参数，并返回该文件的扩展名
func format(name string) string {
	// 使用 strings.Split 函数将文件名分割成多个部分，使用 "." 作为分隔符
	if p := strings.Split(name, "."); len(p) > 1 {
		// 如果分割后的部分数量大于 1，则返回最后一个部分，即文件的扩展名
		return p[len(p)-1]
	}
	// 如果文件名中没有 "."，则返回空字符串
	return ""
}
