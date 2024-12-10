package httputil

import (
	"strings"
)

const (
	// baseContentType 定义了基础的内容类型前缀
	baseContentType = "application"
)

// ContentType 函数用于返回带有基础前缀的内容类型。
// 参数：
//   - subtype：内容类型的子类型。
//
// 返回值：
//   - string：带有基础前缀的内容类型字符串。
func ContentType(subtype string) string {
	return baseContentType + "/" + subtype
}

// ContentSubtype 函数用于从给定的内容类型中提取内容子类型。
// 参数：
//   - contentType：有效的内容类型字符串，必须以 application/ 开头。
//
// 返回值：
//   - string：提取出的内容子类型，如果内容类型无效，则返回空字符串。
func ContentSubtype(contentType string) string {
	// 在 contentType 中查找第一个 "/" 的位置
	left := strings.Index(contentType, "/")
	// 如果没有找到 "/"，则返回空字符串
	if left == -1 {
		return ""
	}
	// 在 contentType 中查找第一个 ";" 的位置
	right := strings.Index(contentType, ";")
	// 如果没有找到 ";"，则将 right 设置为 contentType 的长度
	if right == -1 {
		right = len(contentType)
	}
	// 如果 right 小于 left，则返回空字符串
	if right < left {
		return ""
	}
	// 返回 contentType 中从 left+1 到 right 的子串，即内容子类型
	return contentType[left+1 : right]
}
