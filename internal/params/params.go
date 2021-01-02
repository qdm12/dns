package params

import (
	"time"

	"github.com/qdm12/golibs/logging"
	libparams "github.com/qdm12/golibs/params"
	"github.com/qdm12/golibs/verification"
)

// Reader contains methods to obtain parameters.
type Reader interface {
	// DNS getters
	GetProviders() (providers []string, err error)
	GetPrivateAddresses() (privateAddresses []string, err error)

	// Unbound getters
	GetListeningPort() (listeningPort uint16, err error)
	GetCaching() (caching bool, err error)
	GetVerbosity() (verbosityLevel uint8, err error)
	GetVerbosityDetails() (verbosityDetailsLevel uint8, err error)
	GetValidationLogLevel() (validationLogLevel uint8, err error)
	GetCheckUnbound() (check bool, err error)
	GetIPv4() (doIPv4 bool, err error)
	GetIPv6() (doIPv6 bool, err error)

	// Blocking getters
	GetMaliciousBlocking() (blocking bool, err error)
	GetSurveillanceBlocking() (blocking bool, err error)
	GetAdsBlocking() (blocking bool, err error)
	GetUnblockedHostnames() (hostnames []string, err error)
	GetBlockedHostnames() (hostnames []string, err error)
	GetBlockedIPs() (IPs []string, err error)

	// Update getters
	GetUpdatePeriod() (period time.Duration, err error)
}

type reader struct {
	envParams libparams.EnvParams
	logger    logging.Logger
	verifier  verification.Verifier
}

// NewParamsReader returns a paramsReadeer object to read parameters from
// environment variables.
func NewParamsReader(logger logging.Logger) Reader {
	return &reader{
		envParams: libparams.NewEnvParams(),
		logger:    logger,
		verifier:  verification.NewVerifier(),
	}
}
