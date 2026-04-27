package log

//go:generate mockgen -write_package_comment=false -destination=mocks_test.go -package $GOPACKAGE . Logger
//go:generate mockgen -write_package_comment=false -destination=mocks_dns_test.go -package $GOPACKAGE github.com/miekg/dns ResponseWriter
