package file

import (
	"testing"
)

// TestFormat 测试 format 函数是否正确返回文件名的扩展名
func TestFormat(t *testing.T) {
	// 定义一个测试用例结构体，包含输入文件名和期望的扩展名
	tests := []struct {
		input  string
		expect string
	}{
		// 测试用例 1：输入为空字符串，期望返回空字符串
		{
			input:  "",
			expect: "",
		},
		// 测试用例 2：输入为空格，期望返回空字符串
		{
			input:  " ",
			expect: "",
		},
		// 测试用例 3：输入为点号，期望返回空字符串
		{
			input:  ".",
			expect: "",
		},
		// 测试用例 4：输入为字母加点号，期望返回空字符串
		{
			input:  "a.",
			expect: "",
		},
		// 测试用例 5：输入为点号加字母，期望返回字母
		{
			input:  ".b",
			expect: "b",
		},
		// 测试用例 6：输入为字母加点号加字母，期望返回最后一个字母
		{
			input:  "a.b",
			expect: "b",
		},
	}
	// 遍历测试用例
	for _, v := range tests {
		// 调用 format 函数，获取实际返回的扩展名
		content := format(v.input)
		// 比较实际返回的扩展名和期望的扩展名是否一致
		if got, want := content, v.expect; got != want {
			// 如果不一致，记录错误信息
			t.Errorf("expect %v,got %v", want, got)
		}
	}
}
