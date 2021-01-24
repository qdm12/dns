package params

import (
	libparams "github.com/qdm12/golibs/params"
)

// GetListeningPort obtains the port Unbound should listen on
// from the environment variable LISTENINGPORT.
func (r *reader) GetListeningPort() (listeningPort uint16, err error) {
	return r.envParams.Port("LISTENINGPORT", libparams.Default("53"))
}

// GetCaching obtains if Unbound caching should be enable or not
// from the environment variable CACHING.
func (r *reader) GetCaching() (caching bool, err error) {
	return r.envParams.OnOff("CACHING")
}

// GetVerbosity obtains the verbosity level to use for Unbound
// from the environment variable VERBOSITY.
func (r *reader) GetVerbosity() (verbosityLevel uint8, err error) {
	n, err := r.envParams.IntRange("VERBOSITY", 0, 5, libparams.Default("1"))
	return uint8(n), err
}

// GetVerbosityDetails obtains the verbosity details level to use for Unbound
// from the environment variable VERBOSITY_DETAILS.
func (r *reader) GetVerbosityDetails() (verbosityDetailsLevel uint8, err error) {
	n, err := r.envParams.IntRange("VERBOSITY_DETAILS", 0, 4, libparams.Default("0"))
	return uint8(n), err
}

// GetValidationLogLevel obtains the log level to use for Unbound DOT validation
// from the environment variable VALIDATION_LOGLEVEL.
func (r *reader) GetValidationLogLevel() (validationLogLevel uint8, err error) {
	n, err := r.envParams.IntRange("VALIDATION_LOGLEVEL", 0, 2, libparams.Default("0"))
	return uint8(n), err
}

// GetCheckUnbound obtains if the program should check Unbound is running correctly
// at 127.0.0.1:53 from the environment variable CHECK_UNBOUND.
func (r *reader) GetCheckUnbound() (check bool, err error) {
	return r.envParams.OnOff("CHECK_UNBOUND", libparams.Default("on"))
}

func (r *reader) GetIPv4() (doIPv4 bool, err error) {
	return r.envParams.OnOff("IPV4", libparams.Default("on"))
}

func (r *reader) GetIPv6() (doIPv6 bool, err error) {
	return r.envParams.OnOff("IPV6", libparams.Default("off"))
}
