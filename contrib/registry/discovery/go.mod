module github.com/go-kratos/kratos/contrib/registry/discovery/v2

go 1.23.3

require (
	github.com/cnsync/kratos v2.8.2
	github.com/go-resty/resty/v2 v2.11.0
	github.com/pkg/errors v0.9.1
)

require golang.org/x/net v0.29.0 // indirect

replace github.com/cnsync/kratos => ../../../
