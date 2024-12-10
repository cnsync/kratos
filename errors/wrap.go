package errors

import (
	stderrors "errors"
)

// Is 函数用于检查错误链中是否存在与目标错误相匹配的错误。
//
// 错误链由 err 本身以及通过重复调用 Unwrap 获得的错误序列组成。
//
// 如果错误等于目标错误，或者它实现了一个 Is(error) bool 方法，并且该方法对目标错误返回 true，则认为该错误与目标错误匹配。
func Is(err, target error) bool { return stderrors.Is(err, target) }

// As 函数用于在错误链中查找第一个与目标类型匹配的错误，并将其赋值给目标变量。
//
// 错误链由 err 本身以及通过重复调用 Unwrap 获得的错误序列组成。
//
// 如果错误的具体值可以赋值给目标指针所指向的值，或者错误有一个 As(interface{}) bool 方法，并且该方法对目标返回 true，则认为该错误与目标匹配。
//
// 如果目标不是一个指向实现了 error 接口的类型的非 nil 指针，或者不是一个指向任何接口类型的非 nil 指针，As 函数将 panic。如果 err 为 nil，As 函数返回 false。
func As(err error, target interface{}) bool { return stderrors.As(err, &target) }

// Unwrap 函数返回调用 err 的 Unwrap 方法的结果，如果 err 的类型包含一个返回 error 的 Unwrap 方法。
//
// 否则，Unwrap 函数返回 nil。
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}
