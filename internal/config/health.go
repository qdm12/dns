package config

import (
	"fmt"
	"os"

	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

// Health contains settings for the healthcheck and health server.
type Health struct {
	// ServerAddress is the listening address for the health check server.
	// It cannot be the empty string in the internal state.
	ServerAddress string
}

func (h Health) Validate() (err error) {
	err = validate.ListeningAddress(h.ServerAddress, os.Getuid())
	if err != nil {
		return fmt.Errorf("server listening address is not valid: %w", err)
	}

	return nil
}

func (h *Health) copy() (copied Health) {
	return Health{
		ServerAddress: h.ServerAddress,
	}
}

// OverrideWith overrides fields of the receiver
// settings object with any field set in the other
// settings.
func (h *Health) OverrideWith(other Health) {
	h.ServerAddress = gosettings.OverrideWithComparable(h.ServerAddress, other.ServerAddress)
}

func (h *Health) SetDefaults() {
	h.ServerAddress = gosettings.DefaultComparable(h.ServerAddress, "127.0.0.1:9999")
}

func (h Health) String() string {
	return h.ToLinesNode().String()
}

func (h Health) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Health settings:")
	node.Appendf("Server listening address: %s", h.ServerAddress)
	return node
}

func (h *Health) Read(r *reader.Reader) {
	h.ServerAddress = r.String("HEALTH_SERVER_ADDRESS")
}
