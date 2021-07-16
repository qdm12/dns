package config

import (
	"errors"
	"fmt"

	"inet.af/netaddr"
)

var errPrivateIPInvalid = errors.New("invalid private IP address string")

func getPrivateAddresses(reader *reader) (privateIPs []netaddr.IP,
	privateIPPrefixes []netaddr.IPPrefix, err error) {
	values, err := reader.env.CSV("PRIVATE_ADDRESS")
	if err != nil {
		return nil, nil, err
	}
	privateIPs, privateIPPrefixes, err = convertStringsToIPs(values)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", errPrivateIPInvalid, err)
	}
	return privateIPs, privateIPPrefixes, nil
}
