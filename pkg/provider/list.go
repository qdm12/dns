package provider

func All() []Provider {
	return []Provider{
		Cloudflare(),
	}
}
