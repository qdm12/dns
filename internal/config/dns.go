package config

import (
	"fmt"
	"net"
)

func getPrivateAddresses(reader *reader) (privateIPs []net.IP,
	privateIPNets []*net.IPNet, err error) {
	values, err := reader.env.CSV("PRIVATE_ADDRESS")
	if err != nil {
		return nil, nil, err
	}
	privateIPs, privateIPNets, err = convertStringsToIPs(values)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid private IP string: %s", err)
	}
	return privateIPs, privateIPNets, nil
}
