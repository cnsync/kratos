package validate

import (
	"context"

	"github.com/cnsync/kratos/errors"
	"github.com/cnsync/kratos/middleware"
)

// validator 是一个接口，定义了 Validate 方法，用于验证对象
type validator interface {
	Validate() error
}

// Validator 是一个验证中间件，用于在处理请求之前验证请求对象
func Validator() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 检查请求对象是否实现了 validator 接口
			if v, ok := req.(validator); ok {
				// 调用 Validate 方法进行验证
				if err := v.Validate(); err != nil {
					// 如果验证失败，返回错误
					return nil, errors.BadRequest("VALIDATOR", err.Error()).WithCause(err)
				}
			}
			// 如果验证通过或者请求对象未实现 validator 接口，调用下一个中间件或处理函数
			return handler(ctx, req)
		}
	}
}
