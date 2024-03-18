package dnssec

import (
	"fmt"

	"github.com/miekg/dns"
)

func mustRRToCNAME(rr dns.RR) *dns.CNAME {
	cname, ok := rr.(*dns.CNAME)
	if !ok {
		panic(fmt.Sprintf("RR is of type %T and not of type *dns.CNAME", rr))
	}
	return cname
}

func getCnameTarget(rrSets []dnssecRRSet) (target string) {
	for _, rrSet := range rrSets {
		if rrSet.qtype() == dns.TypeCNAME {
			cname := mustRRToCNAME(rrSet.rrSet[0])
			return cname.Target
		}
	}
	return ""
}
