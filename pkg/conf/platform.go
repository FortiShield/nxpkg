package conf

import "github.com/nxpkg/nxpkg/schema"

// PlatformConfiguration contains site configuration for the Nxpkg platform.
type PlatformConfiguration struct {
	RemoteRegistryURL string
}

// Extensions returns the configuration for the Nxpkg platform, or nil if it is disabled.
func Extensions() *PlatformConfiguration {
	cfg := Get()

	x := cfg.Extensions
	if x == nil {
		x = &schema.Extensions{}
	}

	if x.Disabled != nil && *x.Disabled {
		return nil
	}

	var pc PlatformConfiguration

	// If the "remoteRegistry" value is a string, use that. If false, then keep it empty. Otherwise
	// use the default.
	const defaultRemoteRegistry = "https://nxpkg.com/.api/registry"
	if s, ok := x.RemoteRegistry.(string); ok {
		pc.RemoteRegistryURL = s
	} else if b, ok := x.RemoteRegistry.(bool); ok && !b {
		// Nothing to do.
	} else {
		pc.RemoteRegistryURL = defaultRemoteRegistry
	}

	return &pc
}
