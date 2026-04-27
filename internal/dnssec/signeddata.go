package dnssec

type signedData struct {
	zone           string
	dnsKeyResponse dnssecResponse
	dsResponse     dnssecResponse
}
