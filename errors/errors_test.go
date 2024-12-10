package errors

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestError struct{ message string }

func (e *TestError) Error() string { return e.message }

func TestErrors(t *testing.T) {
	// 定义一个指向 Error 类型的指针 base，初始值为 nil
	var base *Error
	// 创建一个新的 Error 对象 err，状态码为 http.StatusBadRequest，原因是 "reason"，消息是 "message"
	err := Newf(http.StatusBadRequest, "reason", "message")
	// 创建另一个新的 Error 对象 err2，状态码为 http.StatusBadRequest，原因是 "reason"，消息是 "message"
	err2 := Newf(http.StatusBadRequest, "reason", "message")
	// 创建一个新的 Error 对象 err3，它是 err 的副本，并添加了一个元数据项 "foo"，其值为 "bar"
	err3 := err.WithMetadata(map[string]string{
		"foo": "bar",
	})
	// 创建一个新的错误对象 werr，它是通过将 err 包装在 fmt.Errorf 函数中创建的
	werr := fmt.Errorf("wrap %w", err)

	// 检查 err 是否等于 new(Error)，如果相等，则测试失败
	if errors.Is(err, new(Error)) {
		t.Errorf("should not be equal: %v", err)
	}
	// 检查 werr 是否等于 err，如果不相等，则测试失败
	if !errors.Is(werr, err) {
		t.Errorf("should be equal: %v", err)
	}
	// 检查 werr 是否等于 err2，如果不相等，则测试失败
	if !errors.Is(werr, err2) {
		t.Errorf("should be equal: %v", err)
	}

	// 检查 err 是否可以被赋值给 base，如果不可以，则测试失败
	if !errors.As(err, &base) {
		t.Errorf("should be matches: %v", err)
	}
	// 检查 err 是否是一个 HTTP 400 错误，如果不是，则测试失败
	if !IsBadRequest(err) {
		t.Errorf("should be matches: %v", err)
	}

	// 检查 err 的原因是否与 err3 的原因相同，如果不同，则测试失败
	if reason := Reason(err); reason != err3.Reason {
		t.Errorf("got %s want: %s", reason, err)
	}

	// 检查 err3 的元数据中是否包含 "foo" 键，并且其值是否为 "bar"，如果不是，则测试失败
	if err3.Metadata["foo"] != "bar" {
		t.Error("not expected metadata")
	}

	// 获取 err 的 gRPC 状态
	gs := err.GRPCStatus()
	// 从 gRPC 状态中创建一个新的 Error 对象 se
	se := FromError(gs.Err())
	// 检查 se 的原因是否与 "reason" 相同，如果不同，则测试失败
	if se.Reason != "reason" {
		t.Errorf("got %+v want %+v", se, err)
	}

	// 创建一个新的 gRPC 状态 gs2，其代码为 codes.InvalidArgument，消息为 "bad request"
	gs2 := status.New(codes.InvalidArgument, "bad request")
	// 从 gs2 中创建一个新的 Error 对象 se2
	se2 := FromError(gs2.Err())
	// 检查 se2 的 HTTP 状态码是否为 http.StatusBadRequest，如果不是，则测试失败
	if se2.Code != http.StatusBadRequest {
		t.Errorf("convert code err, got %d want %d", UnknownCode, http.StatusBadRequest)
	}
	// 检查 FromError(nil) 是否返回 nil，如果不是，则测试失败
	if FromError(nil) != nil {
		t.Errorf("FromError(nil) should be nil")
	}
	// 创建一个新的错误对象 e，它是通过将一个新的错误 "test" 转换为 Error 对象创建的
	e := FromError(errors.New("test"))
	// 检查 e 的代码是否与 UnknownCode 相同，如果不同，则测试失败
	if !reflect.DeepEqual(e.Code, int32(UnknownCode)) {
		t.Errorf("no expect value: %v, but got: %v", e.Code, int32(UnknownCode))
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name string
		e    *Error
		err  error
		want bool
	}{
		{
			name: "true",
			e:    New(404, "test", ""),
			err:  New(http.StatusNotFound, "test", ""),
			want: true,
		},
		{
			name: "false",
			e:    New(0, "test", ""),
			err:  errors.New("test"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用 tt.e.Is(tt.err) 方法，检查返回值是否与 tt.want 相等
			if ok := tt.e.Is(tt.err); ok != tt.want {
				// 如果不相等，打印错误信息
				t.Errorf("Error.Error() = %v, want %v", ok, tt.want)
			}
		})
	}
}

// TestCause 测试 WithCause 方法是否正确设置了错误的原因。
func TestCause(t *testing.T) {
	// 创建一个自定义错误 testError，其消息为 "test"。
	testError := &TestError{message: "test"}
	// 使用 BadRequest 函数创建一个新的错误 err，其原因是 testError。
	err := BadRequest("foo", "bar").WithCause(testError)
	// 检查 err 是否是 testError 的实例。
	if !errors.Is(err, testError) {
		// 如果不是，打印错误信息并终止测试。
		t.Fatalf("want %v but got %v", testError, err)
	}
	// 创建一个新的 TestError 实例 te，用于接收 errors.As 方法的返回值。
	te := new(TestError)
	// 使用 errors.As 方法从 err 中提取 TestError 类型的错误。
	if errors.As(err, &te) {
		// 如果提取成功，检查提取的错误消息是否与原始错误消息一致。
		if te.message != testError.message {
			// 如果不一致，打印错误信息并终止测试。
			t.Fatalf("want %s but got %s", testError.message, te.message)
		}
	}
}

func TestOther(t *testing.T) {
	// 定义一个错误对象 err，其状态码为 10001，原因是 "test code 10001"，消息是 "message"
	err := Errorf(10001, "test code 10001", "message")

	// 测试 Code 方法
	// 检查 Code(nil) 是否返回 200
	if !reflect.DeepEqual(Code(nil), 200) {
		t.Errorf("Code(nil) = %v, want %v", Code(nil), 200)
	}
	// 检查 Code(errors.New("test")) 是否返回 UnknownCode
	if !reflect.DeepEqual(Code(errors.New("test")), UnknownCode) {
		t.Errorf(`Code(errors.New("test")) = %v, want %v`, Code(nil), 200)
	}
	// 检查 Code(err) 是否返回 10001
	if !reflect.DeepEqual(Code(err), 10001) {
		t.Errorf(`Code(err) = %v, want %v`, Code(err), 10001)
	}

	// 测试 Reason 方法
	// 检查 Reason(nil) 是否返回 UnknownReason
	if !reflect.DeepEqual(Reason(nil), UnknownReason) {
		t.Errorf(`Reason(nil) = %v, want %v`, Reason(nil), UnknownReason)
	}
	// 检查 Reason(errors.New("test")) 是否返回 UnknownReason
	if !reflect.DeepEqual(Reason(errors.New("test")), UnknownReason) {
		t.Errorf(`Reason(errors.New("test")) = %v, want %v`, Reason(nil), UnknownReason)
	}
	// 检查 Reason(err) 是否返回 "test code 10001"
	if !reflect.DeepEqual(Reason(err), "test code 10001") {
		t.Errorf(`Reason(err) = %v, want %v`, Reason(err), "test code 10001")
	}

	// 测试 Clone 方法
	// 定义一个 HTTP 400 错误对象 err400，其原因是 "BAD_REQUEST"，消息是 "param invalid"
	err400 := Newf(http.StatusBadRequest, "BAD_REQUEST", "param invalid")
	// 为 err400 添加元数据
	err400.Metadata = map[string]string{
		"key1": "val1",
		"key2": "val2",
	}
	// 检查 Clone(err400) 是否返回一个非 nil 的错误对象，且其错误信息与 err400 相同
	if cerr := Clone(err400); cerr == nil || cerr.Error() != err400.Error() {
		t.Errorf("Clone(err) = %v, want %v", Clone(err400), err400)
	}
	// 检查 Clone(nil) 是否返回 nil
	if cerr := Clone(nil); cerr != nil {
		t.Errorf("Clone(nil) = %v, want %v", Clone(err400), err400)
	}
}
