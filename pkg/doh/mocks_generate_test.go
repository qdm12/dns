package doh

//go:generate mockgen -destination=mocks_test.go -package $GOPACKAGE . Metrics,Logger
//go:generate mockgen -destination=mocks_local_test.go -package $GOPACKAGE -source interfaces_test.go
