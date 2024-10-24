package server

//go:generate mockgen -destination=mocks_test.go -package $GOPACKAGE . Dialer,Logger,Middleware
//go:generate mockgen -destination=mocks_integration_test.go -package $GOPACKAGE -source interfaces_integration_test.go
//go:generate mockgen -destination=mocks_local_test.go -package $GOPACKAGE -source interfaces_local.go
