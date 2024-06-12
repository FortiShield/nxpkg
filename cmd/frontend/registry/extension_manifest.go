package registry

import (
	"sync"

	"github.com/nxpkg/nxpkg/cmd/frontend/graphqlbackend"
	"github.com/nxpkg/nxpkg/pkg/jsonc"
	"github.com/nxpkg/nxpkg/schema"
)

// extensionManifest implements the GraphQL type ExtensionManifest.
type extensionManifest struct {
	raw string

	// cache result because it is used by multiple fields
	once   sync.Once
	result *schema.NxpkgExtensionManifest
	err    error
}

// NewExtensionManifest creates a new resolver for the GraphQL type ExtensionManifest with the given
// raw contents of an extension manifest.
func NewExtensionManifest(raw *string) graphqlbackend.ExtensionManifest {
	if raw == nil {
		return nil
	}
	return &extensionManifest{raw: *raw}
}

func (r *extensionManifest) parse() (*schema.NxpkgExtensionManifest, error) {
	r.once.Do(func() {
		r.err = jsonc.Unmarshal(r.raw, &r.result)
	})
	return r.result, r.err
}

func (r *extensionManifest) Raw() string { return r.raw }

func (r *extensionManifest) Title() (*string, error) {
	parsed, err := r.parse()
	if parsed == nil || err != nil {
		return nil, err
	}
	if parsed.Title == "" {
		return nil, nil
	}
	return &parsed.Title, nil
}

func (r *extensionManifest) Description() (*string, error) {
	parsed, err := r.parse()
	if parsed == nil || err != nil {
		return nil, err
	}
	if parsed.Description == "" {
		return nil, nil
	}
	return &parsed.Description, nil
}

func (r *extensionManifest) BundleURL() (*string, error) {
	parsed, err := r.parse()
	if parsed == nil || err != nil {
		return nil, err
	}
	if parsed.Url == "" {
		return nil, nil
	}
	return &parsed.Url, nil
}
