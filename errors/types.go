// nolint:mnd
package errors

// BadRequest 用于创建一个表示“错误请求”的错误，该错误对应于 HTTP 400 状态码。
func BadRequest(reason, message string) *Error {
	return New(400, reason, message)
}

// IsBadRequest 用于判断给定的错误是否为“错误请求”类型的错误。
func IsBadRequest(err error) bool {
	return Code(err) == 400
}

// Unauthorized 用于创建一个表示“未授权”的错误，该错误对应于 HTTP 401 状态码。
func Unauthorized(reason, message string) *Error {
	return New(401, reason, message)
}

// IsUnauthorized 用于判断给定的错误是否为“未授权”类型的错误。
func IsUnauthorized(err error) bool {
	return Code(err) == 401
}

// Forbidden 用于创建一个表示“禁止”的错误，该错误对应于 HTTP 403 状态码。
func Forbidden(reason, message string) *Error {
	return New(403, reason, message)
}

// IsForbidden 用于判断给定的错误是否为“禁止”类型的错误。
func IsForbidden(err error) bool {
	return Code(err) == 403
}

// NotFound 用于创建一个表示“未找到”的错误，该错误对应于 HTTP 404 状态码。
func NotFound(reason, message string) *Error {
	return New(404, reason, message)
}

// IsNotFound 用于判断给定的错误是否为“未找到”类型的错误。
func IsNotFound(err error) bool {
	return Code(err) == 404
}

// Conflict 用于创建一个表示“冲突”的错误，该错误对应于 HTTP 409 状态码。
func Conflict(reason, message string) *Error {
	return New(409, reason, message)
}

// IsConflict 用于判断给定的错误是否为“冲突”类型的错误。
func IsConflict(err error) bool {
	return Code(err) == 409
}

// InternalServer 用于创建一个表示“内部服务器错误”的错误，该错误对应于 HTTP 500 状态码。
func InternalServer(reason, message string) *Error {
	return New(500, reason, message)
}

// IsInternalServer 用于判断给定的错误是否为“内部服务器错误”类型的错误。
func IsInternalServer(err error) bool {
	return Code(err) == 500
}

// ServiceUnavailable 用于创建一个表示“服务不可用”的错误，该错误对应于 HTTP 503 状态码。
func ServiceUnavailable(reason, message string) *Error {
	return New(503, reason, message)
}

// IsServiceUnavailable 用于判断给定的错误是否为“服务不可用”类型的错误。
func IsServiceUnavailable(err error) bool {
	return Code(err) == 503
}

// GatewayTimeout 用于创建一个表示“网关超时”的错误，该错误对应于 HTTP 504 状态码。
func GatewayTimeout(reason, message string) *Error {
	return New(504, reason, message)
}

// IsGatewayTimeout 用于判断给定的错误是否为“网关超时”类型的错误。
func IsGatewayTimeout(err error) bool {
	return Code(err) == 504
}

// ClientClosed 用于创建一个表示“客户端关闭”的错误，该错误对应于 HTTP 499 状态码。
func ClientClosed(reason, message string) *Error {
	return New(499, reason, message)
}

// IsClientClosed 用于判断给定的错误是否为“客户端关闭”类型的错误。
func IsClientClosed(err error) bool {
	return Code(err) == 499
}
