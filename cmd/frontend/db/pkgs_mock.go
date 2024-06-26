package db

import (
	"context"

	"github.com/nxpkg/nxpkg/pkg/api"
)

type MockPkgs struct {
	ListPackages func(context.Context, *api.ListPackagesOp) ([]*api.PackageInfo, error)
	Delete       func(ctx context.Context, repo api.RepoID) error
}
