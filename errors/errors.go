package errors

import (
	"errors"
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"

	httpstatus "github.com/cnsync/kratos/transport/http/status"
)

const (
	// UnknownCode 未知错误的默认错误码。
	UnknownCode = 500
	// UnknownReason 未知错误的默认原因。
	UnknownReason = ""
	// SupportPackageIsVersion1 用于标识包版本，禁止其他代码引用此常量。
	SupportPackageIsVersion1 = true
)

// Error 表示一个带有状态信息的错误结构体。
type Error struct {
	Status       // 嵌入的状态信息，包括错误码、原因、消息等。
	cause  error // 错误的实际根因。
}

// Error 实现 `error` 接口，返回错误的字符串表示。
func (e *Error) Error() string {
	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v cause = %v", e.Code, e.Reason, e.Message, e.Metadata, e.cause)
}

// Unwrap 实现 `errors.Unwrap` 接口，用于获取嵌套的错误根因。
func (e *Error) Unwrap() error { return e.cause }

// Is 实现 `errors.Is` 接口，用于判断错误链中是否包含特定错误。
func (e *Error) Is(err error) bool {
	if se := new(Error); errors.As(err, &se) {
		return se.Code == e.Code && se.Reason == e.Reason // 判断 Code 和 Reason 是否一致。
	}
	return false
}

// WithCause 设置错误的根因并返回新的错误对象。
func (e *Error) WithCause(cause error) *Error {
	err := Clone(e)   // 深拷贝当前错误对象。
	err.cause = cause // 设置根因。
	return err
}

// WithMetadata 设置错误的元数据并返回新的错误对象。
func (e *Error) WithMetadata(md map[string]string) *Error {
	err := Clone(e)   // 深拷贝当前错误对象。
	err.Metadata = md // 设置元数据。
	return err
}

// GRPCStatus 将错误转换为 gRPC 的 `status.Status` 对象。
func (e *Error) GRPCStatus() *status.Status {
	s, _ := status.New(httpstatus.ToGRPCCode(int(e.Code)), e.Message).WithDetails(
		&errdetails.ErrorInfo{
			Reason:   e.Reason,
			Metadata: e.Metadata,
		},
	)
	return s
}

// New 创建一个新的错误对象，包含错误码、原因和消息。
func New(code int, reason, message string) *Error {
	return &Error{
		Status: Status{
			Code:    int32(code), // 转换错误码为 int32 类型。
			Message: message,
			Reason:  reason,
		},
	}
}

// Newf 使用格式化字符串创建一个新的错误对象。
func Newf(code int, reason, format string, a ...interface{}) *Error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

// Errorf 创建并返回一个新的错误对象（实现 `error` 接口）。
func Errorf(code int, reason, format string, a ...interface{}) error {
	return New(code, reason, fmt.Sprintf(format, a...))
}

// Code 获取错误的 HTTP 错误码。如果错误为 `nil`，则返回 200。
func Code(err error) int {
	if err == nil {
		return 200 //nolint:mnd
	}
	return int(FromError(err).Code) // 从错误对象中获取错误码。
}

// Reason 获取错误的原因字符串。如果错误为 `nil`，则返回 `UnknownReason`。
func Reason(err error) string {
	if err == nil {
		return UnknownReason
	}
	return FromError(err).Reason
}

// Clone 方法用于创建一个新的 Error 对象，该对象是对传入的 err 对象的深度复制
func Clone(err *Error) *Error {
	// 如果传入的 err 对象为 nil，则返回 nil
	if err == nil {
		return nil
	}
	// 创建一个新的 map，用于存储复制后的元数据
	metadata := make(map[string]string, len(err.Metadata))
	// 遍历传入的 err 对象的元数据，将每个键值对复制到新的 map 中
	for k, v := range err.Metadata {
		metadata[k] = v
	}
	// 创建一个新的 Error 对象，并将传入的 err 对象的各个字段复制到新的对象中
	return &Error{
		// 复制 cause 字段，如果传入的 err 对象的 cause 字段为 nil，则新对象的 cause 字段也为 nil
		cause: err.cause,
		// 复制 Status 结构体中的各个字段
		Status: Status{
			// 复制 Code 字段
			Code: err.Code,
			// 复制 Reason 字段
			Reason: err.Reason,
			// 复制 Message 字段
			Message: err.Message,
			// 复制 Metadata 字段，使用新创建的 metadata map
			Metadata: metadata,
		},
	}
}

// FromError 尝试将一个通用错误转换为 `*Error` 类型。
// 支持嵌套错误。
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if se := new(Error); errors.As(err, &se) {
		return se // 如果错误已经是 `*Error` 类型，则直接返回。
	}
	gs, ok := status.FromError(err)
	if !ok {
		// 如果不是 gRPC 错误，则创建一个默认的 `*Error`。
		return New(UnknownCode, UnknownReason, err.Error())
	}
	ret := New(
		httpstatus.FromGRPCCode(gs.Code()), // 从 gRPC 状态码转换为 HTTP 状态码。
		UnknownReason,
		gs.Message(),
	)
	// 提取 gRPC 错误的详细信息（如 `ErrorInfo`）。
	for _, detail := range gs.Details() {
		switch d := detail.(type) {
		case *errdetails.ErrorInfo:
			ret.Reason = d.Reason
			return ret.WithMetadata(d.Metadata) // 将元数据附加到错误对象。
		}
	}
	return ret
}
