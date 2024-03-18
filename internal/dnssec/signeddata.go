package dnssec

type signedData struct {
	zone string
	// TODO do we need this class field? Maybe for caching??
	class          uint16
	dnsKeyResponse dnssecResponse
	dsResponse     dnssecResponse
}
