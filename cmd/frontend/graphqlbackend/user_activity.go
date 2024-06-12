package graphqlbackend

import (
	"context"
	"errors"
	"time"

	"github.com/nxpkg/nxpkg/cmd/frontend/backend"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/app/envvar"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/pkg/useractivity"
	"github.com/nxpkg/nxpkg/cmd/frontend/types"
	"github.com/nxpkg/nxpkg/pkg/actor"
)

func (r *UserResolver) Activity(ctx context.Context) (*userActivityResolver, error) {
	// 🚨 SECURITY: Only the user and site admins are allowed to access user activity.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return nil, err
	}
	if envvar.NxpkgDotComMode() {
		return nil, errors.New("site analytics is not available on nxpkg.com")
	}

	activity, err := useractivity.GetByUserID(r.user.ID)
	if err != nil {
		return nil, err
	}
	return &userActivityResolver{activity}, nil
}

type userActivityResolver struct {
	userActivity *types.UserActivity
}

func (s *userActivityResolver) PageViews() int32 { return s.userActivity.PageViews }

func (s *userActivityResolver) SearchQueries() int32 { return s.userActivity.SearchQueries }

func (s *userActivityResolver) CodeIntelligenceActions() int32 {
	return s.userActivity.CodeIntelligenceActions
}

func (s *userActivityResolver) LastActiveTime() *string {
	if s.userActivity.LastActiveTime != nil {
		t := s.userActivity.LastActiveTime.Format(time.RFC3339)
		return &t
	}
	return nil
}

func (s *userActivityResolver) LastActiveCodeHostIntegrationTime() *string {
	if s.userActivity.LastCodeHostIntegrationTime != nil {
		t := s.userActivity.LastCodeHostIntegrationTime.Format(time.RFC3339)
		return &t
	}
	return nil
}

func (s *schemaResolver) LogUserEvent(ctx context.Context, args *struct {
	Event        string
	UserCookieID string
}) (*EmptyResponse, error) {
	if envvar.NxpkgDotComMode() {
		return nil, nil
	}
	actor := actor.FromContext(ctx)
	return nil, useractivity.LogActivity(actor.IsAuthenticated(), actor.UID, args.UserCookieID, args.Event)
}
