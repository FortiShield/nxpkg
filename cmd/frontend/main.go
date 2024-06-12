//docker:user nxpkg

// Note: All frontend code should be added to shared.Main, not here. See that
// function for details.

package main

import (
	_ "github.com/nxpkg/nxpkg/cmd/frontend/internal/app/assets"
	_ "github.com/nxpkg/nxpkg/cmd/frontend/registry"
	"github.com/nxpkg/nxpkg/cmd/frontend/shared"
)

func main() {
	shared.Main()
}
