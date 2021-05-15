package config

import (
	"github.com/qdm12/dns/pkg/middlewares/log"
	"github.com/qdm12/golibs/params"
)

func getLogSettings(env params.Env) (settings log.Settings, err error) {
	settings.LogRequests, err = env.OnOff("LOG_REQUESTS", params.Default("off"))
	if err != nil {
		return settings, err
	}

	settings.LogResponses, err = env.OnOff("LOG_RESPONSES", params.Default("off"))
	if err != nil {
		return settings, err
	}

	return settings, nil
}
