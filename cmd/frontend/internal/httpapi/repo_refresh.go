package httpapi

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/pkg/handlerutil"
	"github.com/nxpkg/nxpkg/pkg/gitserver"
	"github.com/nxpkg/nxpkg/pkg/repoupdater"
)

func serveRepoRefresh(w http.ResponseWriter, r *http.Request) error {
	repo, err := handlerutil.GetRepo(r.Context(), mux.Vars(r))
	if err != nil {
		return err
	}
	return repoupdater.DefaultClient.EnqueueRepoUpdate(context.Background(), gitserver.Repo{Name: repo.URI})
}
