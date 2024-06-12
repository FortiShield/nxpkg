package auth

import (
	"net/http"

	"github.com/nxpkg/nxpkg/pkg/conf"
)

// ForbidAllRequestsMiddleware forbids all requests. It is used when no auth provider is configured (as
// a safer default than "server is 100% public, no auth required").
func ForbidAllRequestsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(conf.AuthProviders()) == 0 {
			const msg = "Access to Nxpkg is forbidden because no authentication provider is set in site configuration."
			http.Error(w, msg, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
