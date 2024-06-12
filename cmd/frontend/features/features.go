package features

import "context"

// CanWhitelistExtensions checks the current product plan to see if it can
// whitelist Nxpkg extensions.
func CanWhitelistExtensions(ctx context.Context) bool {
	// TODO(sqs): Add back in feature logic.
	return true
}
