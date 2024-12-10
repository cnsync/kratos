module github.com/go-kratos/kratos/contrib/log/logrus/v2

go 1.23.3

require (
	github.com/cnsync/kratos v2.8.2
	github.com/sirupsen/logrus v1.8.1
)

require golang.org/x/sys v0.27.0 // indirect

replace github.com/cnsync/kratos => ../../../
