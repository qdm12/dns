package params

import (
	"strconv"

	libparams "github.com/qdm12/golibs/params"
)

// GetListeningPort obtains the port Unbound should listen on
// from the environment variable LISTENINGPORT
func (p *paramsReader) GetListeningPort() (listeningPort uint16, err error) {
	s, err := p.envParams.GetEnv("LISTENINGPORT", libparams.Default("53"))
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(n), nil
}

// GetCaching obtains if Unbound caching should be enable or not
// from the environment variable CACHING
func (p *paramsReader) GetCaching() (caching bool, err error) {
	return p.envParams.GetOnOff("CACHING")
}

// GetVerbosity obtains the verbosity level to use for Unbound
// from the environment variable VERBOSITY
func (p *paramsReader) GetVerbosity() (verbosityLevel uint8, err error) {
	n, err := p.envParams.GetEnvIntRange("VERBOSITY", 0, 5, libparams.Default("1"))
	return uint8(n), err
}

// GetVerbosityDetails obtains the verbosity details level to use for Unbound
// from the environment variable VERBOSITY_DETAILS
func (p *paramsReader) GetVerbosityDetails() (verbosityDetailsLevel uint8, err error) {
	n, err := p.envParams.GetEnvIntRange("VERBOSITY_DETAILS", 0, 4, libparams.Default("0"))
	return uint8(n), err
}

// GetValidationLogLevel obtains the log level to use for Unbound DOT validation
// from the environment variable VALIDATION_LOGLEVEL
func (p *paramsReader) GetValidationLogLevel() (validationLogLevel uint8, err error) {
	n, err := p.envParams.GetEnvIntRange("VALIDATION_LOGLEVEL", 0, 2, libparams.Default("0"))
	return uint8(n), err
}

// GetCheckUnbound obtains if the program should check Unbound is running correctly
// at 127.0.0.1:53 from the environment variable CHECK_UNBOUND
func (p *paramsReader) GetCheckUnbound() (check bool, err error) {
	return p.envParams.GetOnOff("CHECK_UNBOUND", libparams.Default("on"))
}
