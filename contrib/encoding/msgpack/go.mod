module github.com/cnsync/kratos/contrib/encoding/msgpack

go 1.23.3

require (
	github.com/cnsync/kratos v0.0.0-00010101000000-000000000000
	github.com/vmihailenco/msgpack/v5 v5.4.1
)

require github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect

replace github.com/cnsync/kratos => ../../../
