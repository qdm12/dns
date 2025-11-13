package health

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/qdm12/dns/v2/internal/config"
	"github.com/qdm12/gosettings/reader"
)

func IsClientMode(args []string) bool {
	return len(args) > 1 && args[1] == "healthcheck"
}

type Client struct {
	*http.Client
}

func NewClient() *Client {
	const timeout = 5 * time.Second
	return &Client{
		Client: &http.Client{Timeout: timeout},
	}
}

var ErrUnhealthy = errors.New("program is unhealthy")

// Query sends an HTTP request to the other instance of
// the program, and to its internal healthcheck server.
func (c *Client) Query(ctx context.Context, settingsReader *reader.Reader) error {
	port, err := readServerPort(settingsReader)
	if err != nil {
		return fmt.Errorf("reading server port from config: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:"+port, nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode == http.StatusOK {
		return nil
	}

	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	return fmt.Errorf("%w: %s", ErrUnhealthy, string(b))
}

func readServerPort(reader *reader.Reader) (port string, err error) {
	// Extract the health server port from the configuration.
	var config config.Health
	config.Read(reader)
	config.SetDefaults()
	err = config.Validate()
	if err != nil {
		return "", err
	}
	_, port, err = net.SplitHostPort(config.ServerAddress)
	if err != nil {
		return "", err
	}
	return port, nil
}
