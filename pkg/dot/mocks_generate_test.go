package dot

//go:generate mockgen -destination=mocks_test.go -package $GOPACKAGE . Metrics,Logger,Picker
//go:generate mockgen -destination=mocks_local_test.go -package $GOPACKAGE -source interfaces_test.go
