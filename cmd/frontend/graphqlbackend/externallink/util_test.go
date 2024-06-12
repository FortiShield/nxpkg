package externallink

import (
	"github.com/nxpkg/nxpkg/cmd/frontend/backend"
	"github.com/nxpkg/nxpkg/cmd/frontend/db"
	"github.com/nxpkg/nxpkg/pkg/repoupdater"
)

func resetMocks() {
	repoupdater.MockRepoLookup = nil
	db.Mocks = db.MockStores{}
	backend.Mocks = backend.MockServices{}
}
