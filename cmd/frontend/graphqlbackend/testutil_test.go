package graphqlbackend

import (
	"github.com/nxpkg/nxpkg/cmd/frontend/backend"
	"github.com/nxpkg/nxpkg/cmd/frontend/db"
)

func init() {
	skipRefresh = true
}

func resetMocks() {
	db.Mocks = db.MockStores{}
	backend.Mocks = backend.MockServices{}
}
