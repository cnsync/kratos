package binding

import (
	"reflect"
	"regexp"

	"github.com/cnsync/kratos/encoding/form"
	"google.golang.org/protobuf/proto"
)

var reg = regexp.MustCompile(`{[\\.\w]+}`) // 定义一个正则表达式，用于匹配路径模板中的占位符

// EncodeURL 将 proto 消息编码为 URL 路径。
// pathTemplate 是路径模板，msg 是要编码的 proto 消息，needQuery 表示是否需要将剩余的查询参数附加到 URL 中。
func EncodeURL(pathTemplate string, msg interface{}, needQuery bool) string {
	// 如果消息为空或者是空指针，直接返回原始的路径模板
	if msg == nil || (reflect.ValueOf(msg).Kind() == reflect.Ptr && reflect.ValueOf(msg).IsNil()) {
		return pathTemplate
	}

	// 将 proto 消息编码为查询参数，queryParams 是包含所有查询参数的集合
	queryParams, _ := form.EncodeValues(msg)

	// 创建一个映射来存储路径中的占位符参数
	pathParams := make(map[string]struct{})

	// 使用正则表达式替换路径模板中的占位符，将占位符替换为对应的查询参数值
	path := reg.ReplaceAllStringFunc(pathTemplate, func(in string) string {
		// 从占位符中提取出键名
		key := in[1 : len(in)-1]
		// 将键添加到路径参数映射中
		pathParams[key] = struct{}{}
		// 从查询参数中获取对应键的值，并替换占位符
		return queryParams.Get(key)
	})

	// 如果不需要查询参数
	if !needQuery {
		// 如果消息是 proto.Message 类型，则对字段掩码进行编码作为查询参数
		if v, ok := msg.(proto.Message); ok {
			if query := form.EncodeFieldMask(v.ProtoReflect()); query != "" {
				// 如果有字段掩码，返回路径并附加字段掩码
				return path + "?" + query
			}
		}
		// 否则直接返回路径
		return path
	}

	// 如果需要查询参数且查询参数不为空
	if len(queryParams) > 0 {
		// 从查询参数中移除路径中已经使用的占位符参数
		for key := range pathParams {
			delete(queryParams, key)
		}
		// 将剩余的查询参数附加到路径后面
		if query := queryParams.Encode(); query != "" {
			// 返回路径 + 查询字符串
			path += "?" + query
		}
	}

	// 返回最终的路径
	return path
}
