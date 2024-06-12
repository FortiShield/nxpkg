package db

import (
	"context"

	"github.com/nxpkg/nxpkg/pkg/api"
)

type MockGlobalDeps struct {
	Dependencies func(context.Context, DependenciesOptions) ([]*api.DependencyReference, error)
}
