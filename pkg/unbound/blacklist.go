package unbound

import (
	"net"
)

func convertBlockedToConfigLines(blockedHostnames []string,
	blockedIPs []net.IP, blockedIPNets []*net.IPNet) (configLines []string) {
	size := len(blockedHostnames) + len(blockedIPs) + len(blockedIPNets)
	configLines = make([]string, 0, size)

	for _, blockedHostname := range blockedHostnames {
		configLines = append(configLines, "  local-zone: \""+blockedHostname+"\" static")
	}

	for _, blockedIP := range blockedIPs {
		configLines = append(configLines, "  private-address: "+blockedIP.String())
	}

	for _, blockedIPNet := range blockedIPNets {
		configLines = append(configLines, "  private-address: "+blockedIPNet.String())
	}

	return configLines
}
