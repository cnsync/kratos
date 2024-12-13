module github.com/cnsync/kratos/contrib/registry/servicecomb

go 1.23.3

require (
	github.com/cnsync/kratos v0.0.0-00010101000000-000000000000
	github.com/go-chassis/cari v0.6.0
	github.com/go-chassis/sc-client v0.6.1-0.20210615014358-a45e9090c751
	github.com/gofrs/uuid v4.2.0+incompatible
)

require (
	github.com/cenkalti/backoff v2.0.0+incompatible // indirect
	github.com/go-chassis/foundation v0.4.0 // indirect
	github.com/go-chassis/openlog v1.1.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gorilla/websocket v1.4.3-0.20210424162022-e8629af678b7 // indirect
	golang.org/x/net v0.32.0 // indirect
)

replace github.com/cnsync/kratos => ../../../
