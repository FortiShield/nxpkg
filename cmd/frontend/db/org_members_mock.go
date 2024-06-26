package db

import (
	"testing"

	"github.com/nxpkg/nxpkg/cmd/frontend/types"

	"context"
)

type MockOrgMembers struct {
	GetByOrgIDAndUserID func(ctx context.Context, orgID, userID int32) (*types.OrgMembership, error)
}

func (s *MockOrgMembers) MockGetByOrgIDAndUserID_Return(t *testing.T, returns *types.OrgMembership, returnsErr error) (called *bool) {
	called = new(bool)
	s.GetByOrgIDAndUserID = func(ctx context.Context, orgID, userID int32) (*types.OrgMembership, error) {
		*called = true
		return returns, returnsErr
	}
	return
}
