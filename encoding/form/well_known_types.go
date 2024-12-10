package form

import (
	"encoding/base64"
	"fmt"
	"math"
	"strings"
	"time"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// timestamp
	// 定义了 google.protobuf.Timestamp 消息的全名
	timestampMessageFullname protoreflect.FullName = "google.protobuf.Timestamp"
	// 定义了最大的时间戳秒数
	maxTimestampSeconds = 253402300799
	// 定义了最小的时间戳秒数
	minTimestampSeconds = -6213559680013
	// 定义了时间戳消息中秒字段的编号
	timestampSecondsFieldNumber protoreflect.FieldNumber = 1
	// 定义了时间戳消息中纳秒字段的编号
	timestampNanosFieldNumber protoreflect.FieldNumber = 2

	// duration
	// 定义了 google.protobuf.Duration 消息的全名
	durationMessageFullname protoreflect.FullName = "google.protobuf.Duration"
	// 定义了一秒钟的纳秒数
	secondsInNanos = 999999999
	// 定义了持续时间消息中秒字段的编号
	durationSecondsFieldNumber protoreflect.FieldNumber = 1
	// 定义了持续时间消息中纳秒字段的编号
	durationNanosFieldNumber protoreflect.FieldNumber = 2

	// bytes
	// 定义了 google.protobuf.BytesValue 消息的全名
	bytesMessageFullname protoreflect.FullName = "google.protobuf.BytesValue"
	// 定义了字节消息中字节字段的编号
	bytesValueFieldNumber protoreflect.FieldNumber = 1

	// google.protobuf.Struct.
	// 定义了 google.protobuf.Struct 消息的全名
	structMessageFullname protoreflect.FullName = "google.protobuf.Struct"
	// 定义了结构体消息中字段字段的编号
	structFieldsFieldNumber protoreflect.FieldNumber = 1

	// 定义了 google.protobuf.FieldMask 消息的全名
	fieldMaskFullName protoreflect.FullName = "google.protobuf.FieldMask"
)

// marshalTimestamp 函数用于将一个 protobuf 消息中的时间戳字段编码为 URL 查询字符串格式。
// 参数：
//   - m：要编码的时间戳消息。
//
// 返回值：
//   - string：编码后的 URL 查询字符串。
//   - error：如果编码过程中发生错误，返回该错误。
func marshalTimestamp(m protoreflect.Message) (string, error) {
	// 获取消息的字段描述符
	fds := m.Descriptor().Fields()
	// 获取秒字段的描述符
	fdSeconds := fds.ByNumber(timestampSecondsFieldNumber)
	// 获取纳秒字段的描述符
	fdNanos := fds.ByNumber(timestampNanosFieldNumber)

	// 获取秒字段的值
	secsVal := m.Get(fdSeconds)
	// 获取纳秒字段的值
	nanosVal := m.Get(fdNanos)
	// 将秒字段的值转换为整数
	secs := secsVal.Int()
	// 将纳秒字段的值转换为整数
	nanos := nanosVal.Int()
	// 检查秒字段的值是否在有效范围内
	if secs < minTimestampSeconds || secs > maxTimestampSeconds {
		// 如果不在有效范围内，返回错误
		return "", fmt.Errorf("%s: seconds out of range %v", timestampMessageFullname, secs)
	}
	// 检查纳秒字段的值是否在有效范围内
	if nanos < 0 || nanos > secondsInNanos {
		// 如果不在有效范围内，返回错误
		return "", fmt.Errorf("%s: nanos out of range %v", timestampMessageFullname, nanos)
	}
	// 使用 RFC 3339 格式，生成的输出将是 Z 标准化的，并使用 0、3、6 或 9 位小数。
	t := time.Unix(secs, nanos).Local()
	// 格式化时间戳
	x := t.Format("2006-01-02T15:04:05.000000000")
	// 去除末尾的多余 0
	x = strings.TrimSuffix(x, "000")
	// 去除末尾的多余 0
	x = strings.TrimSuffix(x, "000")
	// 去除末尾的多余.000
	x = strings.TrimSuffix(x, ".000")
	// 添加 Z 表示 UTC 时间
	return x + "Z", nil
}

// marshalDuration 函数用于将一个 protobuf 消息中的持续时间字段编码为 URL 查询字符串格式。
// 参数：
//   - m：要编码的持续时间消息。
//
// 返回值：
//   - string：编码后的 URL 查询字符串。
//   - error：如果编码过程中发生错误，返回该错误。
func marshalDuration(m protoreflect.Message) (string, error) {
	// 获取消息的字段描述符
	fds := m.Descriptor().Fields()
	// 获取秒字段的描述符
	fdSeconds := fds.ByNumber(durationSecondsFieldNumber)
	// 获取纳秒字段的描述符
	fdNanos := fds.ByNumber(durationNanosFieldNumber)

	// 获取秒字段的值
	secsVal := m.Get(fdSeconds)
	// 获取纳秒字段的值
	nanosVal := m.Get(fdNanos)
	// 将秒字段的值转换为整数
	secs := secsVal.Int()
	// 将纳秒字段的值转换为整数
	nanos := nanosVal.Int()
	// 创建一个 time.Duration 类型的变量 d，其值为秒数乘以秒
	d := time.Duration(secs) * time.Second
	// 检查是否发生了溢出
	overflow := d/time.Second != time.Duration(secs)
	// 将纳秒数加到 d 上
	d += time.Duration(nanos) * time.Nanosecond
	// 再次检查是否发生了溢出
	overflow = overflow || (secs < 0 && nanos < 0 && d > 0)
	overflow = overflow || (secs > 0 && nanos > 0 && d < 0)
	// 如果发生了溢出
	if overflow {
		// 根据秒数的正负情况，返回最大或最小的 int64 持续时间
		switch {
		case secs < 0:
			return time.Duration(math.MinInt64).String(), nil
		case secs > 0:
			return time.Duration(math.MaxInt64).String(), nil
		}
	}
	// 返回编码后的持续时间字符串
	return d.String(), nil
}

// marshalBytes 函数用于将一个 protobuf 消息中的字节字段编码为 URL 查询字符串格式。
// 参数：
//   - m：要编码的字节消息。
//
// 返回值：
//   - string：编码后的 URL 查询字符串。
//   - error：如果编码过程中发生错误，返回该错误。
func marshalBytes(m protoreflect.Message) (string, error) {
	// 获取消息的字段描述符
	fds := m.Descriptor().Fields()
	// 获取字节字段的描述符
	fdBytes := fds.ByNumber(bytesValueFieldNumber)
	// 获取字节字段的值
	bytesVal := m.Get(fdBytes)
	// 获取字节字段的值
	val := bytesVal.Bytes()
	// 将字节字段的值编码为 base64 字符串
	return base64.StdEncoding.EncodeToString(val), nil
}
