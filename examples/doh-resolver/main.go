package main

import (
	"context"
	"log"

	"github.com/qdm12/dns/pkg/doh"
)

func main() {
	ctx := context.Background()
	resolver := doh.NewResolver(doh.ResolverSettings{})
	ips, err := resolver.LookupIPAddr(ctx, "github.com")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("IP addresses resolved: ", ips)
}
