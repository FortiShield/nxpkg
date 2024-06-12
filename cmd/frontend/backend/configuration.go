package backend

import (
	"context"

	"github.com/nxpkg/nxpkg/cmd/frontend/db"
	"github.com/nxpkg/nxpkg/pkg/api"
	"github.com/nxpkg/nxpkg/pkg/jsonc"
	"github.com/nxpkg/nxpkg/schema"
)

// Configuration backend.
var Configuration = &configuration{}

type configuration struct{}

// GetForSubject gets the latest settings for a single configuration subject, without
// performing any cascading (merging configuration from multiple subjects).
func (configuration) GetForSubject(ctx context.Context, subject api.ConfigurationSubject) (*schema.Settings, error) {
	settings, err := db.Settings.GetLatest(ctx, subject)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		// Settings have never been saved for this subject; equivalent to `{}`.
		return &schema.Settings{}, nil
	}

	var v schema.Settings
	if err := jsonc.Unmarshal(settings.Contents, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
