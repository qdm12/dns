package config

import (
	"fmt"

	"inet.af/netaddr"
)

func getPrivateAddresses(reader *reader) (privateIPs []netaddr.IP,
	privateIPPrefixes []netaddr.IPPrefix, err error) {
	values, err := reader.env.CSV("PRIVATE_ADDRESS")
	if err != nil {
		return nil, nil, err
	}
	privateIPs, privateIPPrefixes, err = convertStringsToIPs(values)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid private IP string: %s", err)
	}
	return privateIPs, privateIPPrefixes, nil
}
