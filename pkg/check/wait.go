package check

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

var ErrDNSFailure = errors.New("DNS is not working")

func WaitForDNS(ctx context.Context, resolver *net.Resolver) (err error) {
	const maxTries = 10
	const hostToResolve = "github.com"
	const waitTime = 300 * time.Millisecond
	timer := time.NewTimer(waitTime)
	select {
	case <-timer.C:
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return ctx.Err()
	}
	for try := 1; try <= maxTries; try++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err = resolver.LookupIP(ctx, "ip", hostToResolve)
		if err == nil {
			return nil
		}
		const msStep = 50
		waitTime := maxTries * msStep * time.Millisecond
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
	return fmt.Errorf("%w: after %d tries: %s", ErrDNSFailure, maxTries, err)
}
