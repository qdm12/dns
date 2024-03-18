package dnssec

import (
	"fmt"

	"github.com/miekg/dns"
)

func mustRRToDS(rr dns.RR) *dns.DS {
	ds, ok := rr.(*dns.DS)
	if !ok {
		panic(fmt.Sprintf("RR is of type %T and not of type *dns.DS", rr))
	}
	return ds
}
