package check

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrDNSMalfunction = errors.New("DNS is not working")

func WaitForDNS(ctx context.Context, settings Settings) (err error) {
	settings.SetDefaults()

	waitTime := settings.WaitTime

	timer := time.NewTimer(waitTime)
	select {
	case <-timer.C:
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return ctx.Err()
	}
	for try := 1; try <= settings.MaxTries; try++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err = settings.Resolver.LookupIP(ctx, "ip", settings.HostToResolve)
		if err == nil {
			return nil
		}

		waitTime += settings.AddWaitTime
		timer := time.NewTimer(waitTime)
		select {
		case <-timer.C:
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		}
	}
	return fmt.Errorf("%w: after %d tries: %s", ErrDNSMalfunction, settings.MaxTries, err)
}
