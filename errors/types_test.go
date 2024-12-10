package errors

import (
	"testing"
)

// TestTypes 测试 errors 包中的错误类型是否能够被正确识别
func TestTypes(t *testing.T) {
	// 定义一个错误列表，包含各种类型的错误
	var (
		input = []error{
			BadRequest("reason_400", "message_400"),
			Unauthorized("reason_401", "message_401"),
			Forbidden("reason_403", "message_403"),
			NotFound("reason_404", "message_404"),
			Conflict("reason_409", "message_409"),
			InternalServer("reason_500", "message_500"),
			ServiceUnavailable("reason_503", "message_503"),
			GatewayTimeout("reason_504", "message_504"),
			ClientClosed("reason_499", "message_499"),
		}
		// 定义一个函数列表，包含各种类型错误的判断函数
		output = []func(error) bool{
			IsBadRequest,
			IsUnauthorized,
			IsForbidden,
			IsNotFound,
			IsConflict,
			IsInternalServer,
			IsServiceUnavailable,
			IsGatewayTimeout,
			IsClientClosed,
		}
	)

	// 遍历错误列表和函数列表，检查每个错误是否能够被正确识别
	for i, in := range input {
		// 如果函数列表中的函数不能识别错误列表中的错误，则输出错误信息
		if !output[i](in) {
			t.Errorf("not expect: %v", in)
		}
	}
}
