package provider

func All() []Provider {
	return []Provider{
		CiraFamily(),
		CiraPrivate(),
		CiraProtected(),
		CleanBrowsingAdult(),
		CleanBrowsingFamily(),
		CleanBrowsingSecurity(),
		Cloudflare(),
		CloudflareFamily(),
		CloudflareSecurity(),
		Google(),
		LibreDNS(),
		Quad9(),
		Quad9Secured(),
		Quad9Unsecured(),
		Quadrant(),
	}
}
