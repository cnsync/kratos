module github.com/cnsync/kratos/contrib/log/logrus

go 1.23.3

require (
	github.com/cnsync/kratos v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
)

require golang.org/x/sys v0.27.0 // indirect

replace github.com/cnsync/kratos => ../../../
