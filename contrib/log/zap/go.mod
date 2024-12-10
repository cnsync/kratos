module github.com/cnsync/kratos/contrib/log/zap

go 1.23.3

require (
	github.com/cnsync/kratos v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.26.0
)

require go.uber.org/multierr v1.11.0 // indirect

replace github.com/cnsync/kratos => ../../../
