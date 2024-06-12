package db

import (
	dbtesting "github.com/nxpkg/nxpkg/cmd/frontend/db/testing"
)

func init() {
	dbtesting.BeforeTest = append(dbtesting.BeforeTest, func() { Mocks = MockStores{} })
}
