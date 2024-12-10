package binding

import (
	"net/http"
	"net/url"

	"github.com/cnsync/kratos/encoding"
	"github.com/cnsync/kratos/encoding/form"
	"github.com/cnsync/kratos/errors"
)

// BindQuery 函数用于将 URL 中的查询参数绑定到指定的目标对象上。
// 参数：
//   - vars：包含查询参数的 url.Values 对象。
//   - target：目标对象的指针，查询参数将被绑定到该对象上。
//
// 返回值：
//   - error：如果在绑定过程中发生错误，返回相应的错误信息；否则返回 nil。
func BindQuery(vars url.Values, target interface{}) error {
	// 使用 form 编码解码器将查询参数编码为字节数组
	if err := encoding.GetCodec(form.Name).Unmarshal([]byte(vars.Encode()), target); err != nil {
		// 如果发生错误，返回一个包含错误代码和错误信息的 BadRequest 错误
		return errors.BadRequest("CODEC", err.Error())
	}
	// 如果没有错误，返回 nil
	return nil
}

// BindForm 函数用于将 HTTP 请求中的表单数据绑定到指定的目标对象上。
// 参数：
//   - req：包含表单数据的 http.Request 对象。
//   - target：目标对象的指针，表单数据将被绑定到该对象上。
//
// 返回值：
//   - error：如果在绑定过程中发生错误，返回相应的错误信息；否则返回 nil。
func BindForm(req *http.Request, target interface{}) error {
	// 解析请求中的表单数据
	if err := req.ParseForm(); err != nil {
		// 如果解析表单数据时发生错误，返回该错误
		return err
	}
	// 使用 form 编码解码器将表单数据编码为字节数组
	if err := encoding.GetCodec(form.Name).Unmarshal([]byte(req.Form.Encode()), target); err != nil {
		// 如果发生错误，返回一个包含错误代码和错误信息的 BadRequest 错误
		return errors.BadRequest("CODEC", err.Error())
	}
	// 如果没有错误，返回 nil
	return nil
}
