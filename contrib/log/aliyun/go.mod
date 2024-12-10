module github.com/cnsync/kratos/contrib/log/aliyun

go 1.23.3

require (
	github.com/aliyun/aliyun-log-go-sdk v0.1.75
	github.com/cnsync/kratos v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.35.2
)

require (
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/go-logfmt/logfmt v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/atomic v1.5.0 // indirect
	golang.org/x/lint v0.0.0-20190930215403-16217165b5de // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/tools v0.0.0-20210106214847-113979e3529a // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

replace (
	github.com/cnsync/kratos => ../../../
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
)
