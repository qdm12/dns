package config

import (
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gotree"
)

type DNSSEC struct {
	Enabled *bool
}

func (d *DNSSEC) setDefaults() {
	d.Enabled = gosettings.DefaultPointer(d.Enabled, true)
}

func (d DNSSEC) validate() (err error) {
	return nil
}

func (d *DNSSEC) String() string {
	return d.ToLinesNode().String()
}

func (d *DNSSEC) ToLinesNode() (node *gotree.Node) {
	if !*d.Enabled {
		return gotree.New("DNSSEC validation: disabled")
	}
	return gotree.New("DNSSEC validation: enabled")
}

func (d *DNSSEC) read(reader *reader.Reader) (err error) {
	d.Enabled, err = reader.BoolPtr("DNSSEC_VALIDATION")
	if err != nil {
		return err
	}

	return nil
}
