package main

import (
	"context"
	"log"

	"github.com/qdm12/dns/pkg/dot"
)

func main() {
	ctx := context.Background()
	resolver := dot.NewResolver(dot.ResolverSettings{})
	ips, err := resolver.LookupIPAddr(ctx, "github.com")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("IP addresses resolved: ", ips)
}
