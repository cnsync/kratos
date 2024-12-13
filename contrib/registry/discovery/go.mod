module github.com/cnsync/kratos/contrib/registry/discovery

go 1.23.3

require (
	github.com/cnsync/kratos v0.0.0-00010101000000-000000000000
	github.com/go-resty/resty/v2 v2.11.0
	github.com/pkg/errors v0.9.1
)

require golang.org/x/net v0.32.0 // indirect

replace github.com/cnsync/kratos => ../../../
