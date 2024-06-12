package envvar

import (
	"strconv"

	"github.com/nxpkg/nxpkg/cmd/frontend/globals"

	"github.com/nxpkg/nxpkg/pkg/conf"
	"github.com/nxpkg/nxpkg/pkg/env"
)

var nxpkgDotComMode, _ = strconv.ParseBool(env.Get("NXPKGDOTCOM_MODE", "false", "run as Nxpkg.com, with add'l marketing and redirects"))

// NxpkgDotComMode is true if this server is running Nxpkg.com. It shows
// add'l marketing and sets up some add'l redirects.
func NxpkgDotComMode() bool {
	u := globals.AppURL.String()
	return nxpkgDotComMode || u == "https://nxpkg.com" || u == "https://nxpkg.com/"
}

var insecureDevMode, _ = strconv.ParseBool(env.Get("INSECURE_DEV", "false", "development mode, for showing more diagnostics (INSECURE: only use on local dev servers)"))

// InsecureDevMode is true if and only if the application is running in local development mode. In
// this mode, the application displays more verbose and informative errors in the UI. It should also
// show all features (as possible). Dev mode should NEVER be true in production.
func InsecureDevMode() bool { return insecureDevMode }

// HasCodeIntelligence reports whether the site has enabled code intelligence.
func HasCodeIntelligence() bool {
	return len(conf.EnabledLangservers()) > 0
}
