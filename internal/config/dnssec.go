package config

import (
	"fmt"
	"time"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gotree"
)

type DNSSEC struct {
	Enabled                      *bool
	RootTrustAnchorRefreshPeriod *time.Duration
}

func (d *DNSSEC) setDefaults() {
	d.Enabled = gosettings.DefaultPointer(d.Enabled, true)
	const defaultRefreshPeriod = 7 * 24 * time.Hour
	d.RootTrustAnchorRefreshPeriod = gosettings.DefaultPointer(
		d.RootTrustAnchorRefreshPeriod, defaultRefreshPeriod)
}

func (d DNSSEC) validate() (err error) {
	if *d.RootTrustAnchorRefreshPeriod < 0 {
		return fmt.Errorf("root trust anchor refresh period must be positive: %d",
			*d.RootTrustAnchorRefreshPeriod)
	}
	return nil
}

func (d *DNSSEC) String() string {
	return d.ToLinesNode().String()
}

func (d *DNSSEC) ToLinesNode() (node *gotree.Node) {
	if !*d.Enabled {
		return gotree.New("DNSSEC validation: disabled")
	}
	node = gotree.New("DNSSEC validation: enabled")
	if *d.RootTrustAnchorRefreshPeriod == 0 {
		node.Appendf("Root trust anchor refresh: disabled")
	} else {
		node.Appendf("Root trust anchor refresh: every %s",
			*d.RootTrustAnchorRefreshPeriod)
	}
	return node
}

func (d *DNSSEC) read(reader *reader.Reader) (err error) {
	d.Enabled, err = reader.BoolPtr("DNSSEC_VALIDATION")
	if err != nil {
		return err
	}

	d.RootTrustAnchorRefreshPeriod, err = reader.DurationPtr(
		"DNSSEC_ROOT_TRUST_ANCHOR_REFRESH_PERIOD")
	if err != nil {
		return err
	}

	return nil
}
