module github.com/go-kratos/kratos/contrib/log/zap/v2

go 1.23.3

require (
	github.com/cnsync/kratos v2.8.2
	go.uber.org/zap v1.26.0
)

require go.uber.org/multierr v1.11.0 // indirect

replace github.com/cnsync/kratos => ../../../
