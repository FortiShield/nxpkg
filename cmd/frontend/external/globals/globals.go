package globals

import (
	"net/url"

	"github.com/nxpkg/nxpkg/cmd/frontend/globals"
)

func AppURL() *url.URL {
	return globals.AppURL
}
