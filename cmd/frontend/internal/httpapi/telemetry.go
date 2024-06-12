package httpapi

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"

	"github.com/nxpkg/nxpkg/cmd/frontend/internal/app/envvar"
	"github.com/nxpkg/nxpkg/pkg/env"
)

var telemetryHandler http.Handler

func init() {
	if envvar.NxpkgDotComMode() {
		telemetryHandler = &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "https"
				req.URL.Host = "nxpkg-logging.telligentdata.com"
				req.Host = "nxpkg-logging.telligentdata.com"
				req.URL.Path = "/" + mux.Vars(req)["TelemetryPath"]
			},
			ErrorLog: log.New(env.DebugOut, "telemetry proxy: ", log.LstdFlags),
		}
	} else {
		telemetryHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "event-level telemetry is disabled")
			w.WriteHeader(http.StatusNoContent)
		})
	}
}
