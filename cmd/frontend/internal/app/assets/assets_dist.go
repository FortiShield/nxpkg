// +build dist

package assets

import "github.com/nxpkg/nxpkg/cmd/frontend/assets"

func init() {
	assets.Assets = DistAssets
}
