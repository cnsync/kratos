package status

import (
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
)

// TestToGRPCCode 测试 ToGRPCCode 函数是否能够正确转换 HTTP 状态码到 gRPC 状态码
func TestToGRPCCode(t *testing.T) {
	// 定义测试用例结构体
	tests := []struct {
		// 测试用例名称
		name string
		// HTTP 状态码
		code int
		// 期望的 gRPC 状态码
		want codes.Code
	}{
		// 测试用例 1：HTTP 状态码 200，期望的 gRPC 状态码 codes.OK
		{"http.StatusOK", http.StatusOK, codes.OK},
		// 测试用例 2：HTTP 状态码 400，期望的 gRPC 状态码 codes.InvalidArgument
		{"http.StatusBadRequest", http.StatusBadRequest, codes.InvalidArgument},
		// 测试用例 3：HTTP 状态码 401，期望的 gRPC 状态码 codes.Unauthenticated
		{"http.StatusUnauthorized", http.StatusUnauthorized, codes.Unauthenticated},
		// 测试用例 4：HTTP 状态码 403，期望的 gRPC 状态码 codes.PermissionDenied
		{"http.StatusForbidden", http.StatusForbidden, codes.PermissionDenied},
		// 测试用例 5：HTTP 状态码 404，期望的 gRPC 状态码 codes.NotFound
		{"http.StatusNotFound", http.StatusNotFound, codes.NotFound},
		// 测试用例 6：HTTP 状态码 409，期望的 gRPC 状态码 codes.Aborted
		{"http.StatusConflict", http.StatusConflict, codes.Aborted},
		// 测试用例 7：HTTP 状态码 429，期望的 gRPC 状态码 codes.ResourceExhausted
		{"http.StatusTooManyRequests", http.StatusTooManyRequests, codes.ResourceExhausted},
		// 测试用例 8：HTTP 状态码 500，期望的 gRPC 状态码 codes.Internal
		{"http.StatusInternalServerError", http.StatusInternalServerError, codes.Internal},
		// 测试用例 9：HTTP 状态码 501，期望的 gRPC 状态码 codes.Unimplemented
		{"http.StatusNotImplemented", http.StatusNotImplemented, codes.Unimplemented},
		// 测试用例 10：HTTP 状态码 503，期望的 gRPC 状态码 codes.Unavailable
		{"http.StatusServiceUnavailable", http.StatusServiceUnavailable, codes.Unavailable},
		// 测试用例 11：HTTP 状态码 504，期望的 gRPC 状态码 codes.DeadlineExceeded
		{"http.StatusGatewayTimeout", http.StatusGatewayTimeout, codes.DeadlineExceeded},
		// 测试用例 12：自定义状态码 499，期望的 gRPC 状态码 codes.Canceled
		{"StatusClientClosed", ClientClosed, codes.Canceled},
		// 测试用例 13：自定义状态码 100000，期望的 gRPC 状态码 codes.Unknown
		{"else", 100000, codes.Unknown},
	}
	// 遍历测试用例
	for _, tt := range tests {
		// 运行测试用例
		t.Run(tt.name, func(t *testing.T) {
			// 调用 ToGRPCCode 函数
			got := ToGRPCCode(tt.code)
			// 比较实际结果和预期结果
			if got != tt.want {
				// 如果不相等，记录错误信息
				t.Errorf("GRPCCodeFromStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromGRPCCode(t *testing.T) {
	// 定义测试用例结构体
	tests := []struct {
		// 测试用例名称
		name string
		// gRPC 状态码
		code codes.Code
		// 期望的 HTTP 状态码
		want int
	}{
		// 测试用例 1：gRPC 状态码 codes.OK，期望的 HTTP 状态码 http.StatusOK
		{"codes.OK", codes.OK, http.StatusOK},
		// 测试用例 2：gRPC 状态码 codes.Canceled，期望的 HTTP 状态码 ClientClosed
		{"codes.Canceled", codes.Canceled, ClientClosed},
		// 测试用例 3：gRPC 状态码 codes.Unknown，期望的 HTTP 状态码 http.StatusInternalServerError
		{"codes.Unknown", codes.Unknown, http.StatusInternalServerError},
		// 测试用例 4：gRPC 状态码 codes.InvalidArgument，期望的 HTTP 状态码 http.StatusBadRequest
		{"codes.InvalidArgument", codes.InvalidArgument, http.StatusBadRequest},
		// 测试用例 5：gRPC 状态码 codes.DeadlineExceeded，期望的 HTTP 状态码 http.StatusGatewayTimeout
		{"codes.DeadlineExceeded", codes.DeadlineExceeded, http.StatusGatewayTimeout},
		// 测试用例 6：gRPC 状态码 codes.NotFound，期望的 HTTP 状态码 http.StatusNotFound
		{"codes.NotFound", codes.NotFound, http.StatusNotFound},
		// 测试用例 7：gRPC 状态码 codes.AlreadyExists，期望的 HTTP 状态码 http.StatusConflict
		{"codes.AlreadyExists", codes.AlreadyExists, http.StatusConflict},
		// 测试用例 8：gRPC 状态码 codes.PermissionDenied，期望的 HTTP 状态码 http.StatusForbidden
		{"codes.PermissionDenied", codes.PermissionDenied, http.StatusForbidden},
		// 测试用例 9：gRPC 状态码 codes.Unauthenticated，期望的 HTTP 状态码 http.StatusUnauthorized
		{"codes.Unauthenticated", codes.Unauthenticated, http.StatusUnauthorized},
		// 测试用例 10：gRPC 状态码 codes.ResourceExhausted，期望的 HTTP 状态码 http.StatusTooManyRequests
		{"codes.ResourceExhausted", codes.ResourceExhausted, http.StatusTooManyRequests},
		// 测试用例 11：gRPC 状态码 codes.FailedPrecondition，期望的 HTTP 状态码 http.StatusBadRequest
		{"codes.FailedPrecondition", codes.FailedPrecondition, http.StatusBadRequest},
		// 测试用例 12：gRPC 状态码 codes.Aborted，期望的 HTTP 状态码 http.StatusConflict
		{"codes.Aborted", codes.Aborted, http.StatusConflict},
		// 测试用例 13：gRPC 状态码 codes.OutOfRange，期望的 HTTP 状态码 http.StatusBadRequest
		{"codes.OutOfRange", codes.OutOfRange, http.StatusBadRequest},
		// 测试用例 14：gRPC 状态码 codes.Unimplemented，期望的 HTTP 状态码 http.StatusNotImplemented
		{"codes.Unimplemented", codes.Unimplemented, http.StatusNotImplemented},
		// 测试用例 15：gRPC 状态码 codes.Internal，期望的 HTTP 状态码 http.StatusInternalServerError
		{"codes.Internal", codes.Internal, http.StatusInternalServerError},
		// 测试用例 16：gRPC 状态码 codes.Unavailable，期望的 HTTP 状态码 http.StatusServiceUnavailable
		{"codes.Unavailable", codes.Unavailable, http.StatusServiceUnavailable},
		// 测试用例 17：gRPC 状态码 codes.DataLoss，期望的 HTTP 状态码 http.StatusInternalServerError
		{"codes.DataLoss", codes.DataLoss, http.StatusInternalServerError},
		// 测试用例 18：自定义状态码 10000，期望的 HTTP 状态码 http.StatusInternalServerError
		{"else", codes.Code(10000), http.StatusInternalServerError},
	}
	// 遍历测试用例
	for _, tt := range tests {
		// 运行测试用例
		t.Run(tt.name, func(t *testing.T) {
			// 调用 FromGRPCCode 函数
			if got := FromGRPCCode(tt.code); got != tt.want {
				// 如果不相等，记录错误信息
				t.Errorf("StatusFromGRPCCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
