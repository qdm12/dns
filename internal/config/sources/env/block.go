package env

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func (r *Reader) readBlock() (settings settings.Block, err error) {
	settings.BlockMalicious, err = r.reader.BoolPtr("BLOCK_MALICIOUS")
	if err != nil {
		return settings, fmt.Errorf("environment variable BLOCK_MALICIOUS: %w", err)
	}

	settings.BlockSurveillance, err = r.reader.BoolPtr("BLOCK_SURVEILLANCE")
	if err != nil {
		return settings, fmt.Errorf("environment variable BLOCK_SURVEILLANCE: %w", err)
	}

	settings.BlockAds, err = r.reader.BoolPtr("BLOCK_ADS")
	if err != nil {
		return settings, fmt.Errorf("environment variable BLOCK_ADS: %w", err)
	}

	settings.RebindingProtection, err = r.reader.BoolPtr("REBINDING_PROTECTION")
	if err != nil {
		return settings, fmt.Errorf("environment variable REBINDING_PROTECTION: %w", err)
	}

	settings.AllowedHosts = r.reader.CSV("ALLOWED_HOSTNAMES")
	settings.AddBlockedHosts = r.reader.CSV("BLOCK_HOSTNAMES")

	settings.AllowedIPs, err = r.reader.CSVNetipAddresses("ALLOWED_IPS")
	if err != nil {
		return settings, err
	}
	settings.AddBlockedIPs, err = r.reader.CSVNetipAddresses("BLOCK_IPS")
	if err != nil {
		return settings, err
	}

	settings.AllowedIPPrefixes, err = r.reader.CSVNetipPrefixes("ALLOWED_CIDRS")
	if err != nil {
		return settings, err
	}
	settings.AddBlockedIPPrefixes, err = r.reader.CSVNetipPrefixes("BLOCK_CIDRS")
	if err != nil {
		return settings, err
	}

	settings.RebindingProtection, err = r.reader.BoolPtr("REBINDING_PROTECTION")
	if err != nil {
		return settings, fmt.Errorf("environment variable REBINDING_PROTECTION: %w", err)
	}

	return settings, nil
}
