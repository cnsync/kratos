module github.com/cnsync/kratos/contrib/registry/zookeeper

go 1.23.3

require (
	github.com/cnsync/kratos v0.0.0-00010101000000-000000000000
	github.com/go-zookeeper/zk v1.0.3
	golang.org/x/sync v0.10.0
)

replace github.com/cnsync/kratos => ../../../
