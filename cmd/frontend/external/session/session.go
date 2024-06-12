package session

import "github.com/nxpkg/nxpkg/cmd/frontend/internal/session"

var (
	ResetMockSessionStore = session.ResetMockSessionStore
	SetActor              = session.SetActor
	SetData               = session.SetData
	GetData               = session.GetData
)
