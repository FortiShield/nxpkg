// +build !dist

package assets

import (
	"net/http"

	"github.com/nxpkg/nxpkg/cmd/frontend/assets"
)

func init() {
	assets.Assets = http.Dir("./ui/assets")
}
