package setup

import "github.com/qdm12/dns/v2/pkg/server"

func toServerMiddlewares(middlewares []Middleware) (serverMiddlewares []server.Middleware) {
	serverMiddlewares = make([]server.Middleware, len(middlewares))
	for i, middleware := range middlewares {
		serverMiddlewares[i] = server.Middleware(middleware)
	}
	return serverMiddlewares
}
