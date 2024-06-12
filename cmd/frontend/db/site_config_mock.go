package db

import (
	"context"

	"github.com/nxpkg/nxpkg/cmd/frontend/types"
)

type MockSiteConfig struct {
	Get func(ctx context.Context) (*types.SiteConfig, error)
}
