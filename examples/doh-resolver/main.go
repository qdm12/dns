package main

import (
	"context"
	"log"
	"net"

	"github.com/qdm12/dns/v2/pkg/doh"
)

func main() {
	ctx := context.Background()
	dohDialer, err := doh.New(doh.Settings{})
	if err != nil {
		log.Fatal(err)
	}

	resolver := &net.Resolver{
		PreferGo: true,
		Dial:     dohDialer.Dial,
	}

	ips, err := resolver.LookupIPAddr(ctx, "github.com")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("IP addresses resolved: ", ips)
}
