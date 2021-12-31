package config

import (
	"github.com/qdm12/gotree"
)

func (s *Settings) String() string {
	return s.ToLinesNode().String()
}

func (s *Settings) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Settings summary:")

	node.Appendf("Upstream type: %s", s.UpstreamType)

	switch s.UpstreamType {
	case DoT:
		node.AppendNode(s.DoT.ToLinesNode())
	case DoH:
		node.AppendNode(s.DoH.ToLinesNode())
	}

	node.AppendNode(s.Cache.ToLinesNode())
	node.AppendNode(s.Log.ToLinesNode())
	node.AppendNode(s.Metrics.ToLinesNode())
	node.AppendNode(s.BlockBuilder.ToLinesNode())

	const disabled, enabled = "disabled", "enabled"
	checkDNS := disabled
	if s.CheckDNS {
		checkDNS = enabled
	}
	node.Appendf("Check DNS: %s", checkDNS)

	update := disabled
	if s.UpdatePeriod > 0 {
		update = "every " + s.UpdatePeriod.String()
	}
	node.Appendf("Update: %s", update)

	return node
}
