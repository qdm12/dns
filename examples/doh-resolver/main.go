package main

import (
	"context"
	"log"

	"github.com/qdm12/dns/pkg/doh"
)

func main() {
	ctx := context.Background()
	resolver, err := doh.NewResolver(doh.ResolverSettings{})
	if err != nil {
		log.Fatal(err)
	}

	ips, err := resolver.LookupIPAddr(ctx, "github.com")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("IP addresses resolved: ", ips)
}
