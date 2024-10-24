package localdns

//go:generate mockgen -destination=mocks_test.go -package $GOPACKAGE . Logger
//go:generate mockgen -destination=mocks_dns_test.go -package $GOPACKAGE github.com/miekg/dns ResponseWriter
//go:generate mockgen -destination=mocks_local_test.go -package $GOPACKAGE -source interfaces_local.go
