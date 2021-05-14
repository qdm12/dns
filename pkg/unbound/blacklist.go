package unbound

import "github.com/qdm12/dns/pkg/blacklist"

func convertBlockedToConfigLines(settings blacklist.Settings) (configLines []string) {
	size := len(settings.FqdnHostnames) + len(settings.IPs) + len(settings.IPPrefixes)
	configLines = make([]string, 0, size)

	for _, blockedHostname := range settings.FqdnHostnames {
		configLines = append(configLines, "  local-zone: \""+blockedHostname+"\" static")
	}

	for _, blockedIP := range settings.IPs {
		configLines = append(configLines, "  private-address: "+blockedIP.String())
	}

	for _, blockedIPPrefix := range settings.IPPrefixes {
		configLines = append(configLines, "  private-address: "+blockedIPPrefix.String())
	}

	return configLines
}
