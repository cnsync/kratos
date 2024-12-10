package status

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

const (
	// ClientClosed 是一个非标准的 HTTP 状态码，由 Nginx 定义。
	// 表示客户端关闭连接。参考：https://httpstatus.in/499/
	ClientClosed = 499
)

// Converter 是一个状态码转换器接口，用于在 HTTP 状态码和 gRPC 错误码之间进行转换。
type Converter interface {
	// ToGRPCCode 将 HTTP 错误代码转换为对应的 gRPC 响应状态码。
	ToGRPCCode(code int) codes.Code

	// FromGRPCCode 将 gRPC 错误代码转换为对应的 HTTP 响应状态码。
	FromGRPCCode(code codes.Code) int
}

// statusConverter 是 Converter 接口的默认实现。
type statusConverter struct{}

// DefaultConverter 是默认的状态码转换器实例。
var DefaultConverter Converter = statusConverter{}

// ToGRPCCode 将 HTTP 错误代码转换为对应的 gRPC 响应状态码。
// 参考 gRPC 错误码定义：https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func (c statusConverter) ToGRPCCode(code int) codes.Code {
	switch code {
	case http.StatusOK:
		// HTTP 200 OK 对应 gRPC codes.OK
		return codes.OK
	case http.StatusBadRequest:
		// HTTP 400 Bad Request 对应 gRPC codes.InvalidArgument
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		// HTTP 401 Unauthorized 对应 gRPC codes.Unauthenticated
		return codes.Unauthenticated
	case http.StatusForbidden:
		// HTTP 403 Forbidden 对应 gRPC codes.PermissionDenied
		return codes.PermissionDenied
	case http.StatusNotFound:
		// HTTP 404 Not Found 对应 gRPC codes.NotFound
		return codes.NotFound
	case http.StatusConflict:
		// HTTP 409 Conflict 对应 gRPC codes.Aborted
		return codes.Aborted
	case http.StatusTooManyRequests:
		// HTTP 429 Too Many Requests 对应 gRPC codes.ResourceExhausted
		return codes.ResourceExhausted
	case http.StatusInternalServerError:
		// HTTP 500 Internal Server Error 对应 gRPC codes.Internal
		return codes.Internal
	case http.StatusNotImplemented:
		// HTTP 501 Not Implemented 对应 gRPC codes.Unimplemented
		return codes.Unimplemented
	case http.StatusServiceUnavailable:
		// HTTP 503 Service Unavailable 对应 gRPC codes.Unavailable
		return codes.Unavailable
	case http.StatusGatewayTimeout:
		// HTTP 504 Gateway Timeout 对应 gRPC codes.DeadlineExceeded
		return codes.DeadlineExceeded
	case ClientClosed:
		// 自定义状态码 499 Client Closed 对应 gRPC codes.Canceled
		return codes.Canceled
	}
	// 默认返回 gRPC codes.Unknown，表示未知错误
	return codes.Unknown
}

// FromGRPCCode 将 gRPC 错误代码转换为对应的 HTTP 响应状态码。
// 参考 gRPC 错误码定义：https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
func (c statusConverter) FromGRPCCode(code codes.Code) int {
	switch code {
	case codes.OK:
		// gRPC codes.OK 对应 HTTP 200 OK
		return http.StatusOK
	case codes.Canceled:
		// gRPC codes.Canceled 对应自定义状态码 499 Client Closed
		return ClientClosed
	case codes.Unknown:
		// gRPC codes.Unknown 对应 HTTP 500 Internal Server Error
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		// gRPC codes.InvalidArgument 对应 HTTP 400 Bad Request
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		// gRPC codes.DeadlineExceeded 对应 HTTP 504 Gateway Timeout
		return http.StatusGatewayTimeout
	case codes.NotFound:
		// gRPC codes.NotFound 对应 HTTP 404 Not Found
		return http.StatusNotFound
	case codes.AlreadyExists:
		// gRPC codes.AlreadyExists 对应 HTTP 409 Conflict
		return http.StatusConflict
	case codes.PermissionDenied:
		// gRPC codes.PermissionDenied 对应 HTTP 403 Forbidden
		return http.StatusForbidden
	case codes.Unauthenticated:
		// gRPC codes.Unauthenticated 对应 HTTP 401 Unauthorized
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		// gRPC codes.ResourceExhausted 对应 HTTP 429 Too Many Requests
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// gRPC codes.FailedPrecondition 对应 HTTP 400 Bad Request
		return http.StatusBadRequest
	case codes.Aborted:
		// gRPC codes.Aborted 对应 HTTP 409 Conflict
		return http.StatusConflict
	case codes.OutOfRange:
		// gRPC codes.OutOfRange 对应 HTTP 400 Bad Request
		return http.StatusBadRequest
	case codes.Unimplemented:
		// gRPC codes.Unimplemented 对应 HTTP 501 Not Implemented
		return http.StatusNotImplemented
	case codes.Internal:
		// gRPC codes.Internal 对应 HTTP 500 Internal Server Error
		return http.StatusInternalServerError
	case codes.Unavailable:
		// gRPC codes.Unavailable 对应 HTTP 503 Service Unavailable
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		// gRPC codes.DataLoss 对应 HTTP 500 Internal Server Error
		return http.StatusInternalServerError
	}
	// 默认返回 HTTP 500 Internal Server Error
	return http.StatusInternalServerError
}

// ToGRPCCode 是 ToGRPCCode 方法的包装函数，调用默认转换器的实现。
func ToGRPCCode(code int) codes.Code {
	return DefaultConverter.ToGRPCCode(code)
}

// FromGRPCCode 是 FromGRPCCode 方法的包装函数，调用默认转换器的实现。
func FromGRPCCode(code codes.Code) int {
	return DefaultConverter.FromGRPCCode(code)
}
