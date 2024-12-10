module github.com/go-kratos/kratos/contrib/encoding/msgpack/v2

go 1.23.3

require (
	github.com/cnsync/kratos v2.8.2
	github.com/vmihailenco/msgpack/v5 v5.4.1
)

require github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect

replace github.com/cnsync/kratos => ../../../
