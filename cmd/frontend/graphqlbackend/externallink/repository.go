package externallink

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/nxpkg/nxpkg/cmd/frontend/backend"
	"github.com/nxpkg/nxpkg/cmd/frontend/db"
	"github.com/nxpkg/nxpkg/cmd/frontend/types"
	"github.com/nxpkg/nxpkg/pkg/api"
	"github.com/nxpkg/nxpkg/pkg/errcode"
	"github.com/nxpkg/nxpkg/pkg/repoupdater"
	"github.com/nxpkg/nxpkg/pkg/repoupdater/protocol"
	"github.com/nxpkg/nxpkg/pkg/vcs/git"
)

// Repository returns the external links for a repository.
//
// For example, a repository might have 2 external links, one to its origin repository on GitHub.com
// and one to the repository on Phabricator.
func Repository(ctx context.Context, repo *types.Repo) (links []*Resolver, err error) {
	phabRepo, link, serviceType := linksForRepository(ctx, repo)
	if phabRepo != nil {
		links = append(links, &Resolver{
			url:         strings.TrimSuffix(phabRepo.URL, "/") + "/diffusion/" + phabRepo.Callsign,
			serviceType: "phabricator",
		})
	}
	if link != nil && link.Root != "" {
		links = append(links, &Resolver{url: link.Root, serviceType: serviceType})
	}
	return links, nil
}

// FileOrDir returns the external links for a file or directory in a repository.
func FileOrDir(ctx context.Context, repo *types.Repo, rev, path string, isDir bool) (links []*Resolver, err error) {
	rev = url.PathEscape(rev)

	phabRepo, link, serviceType := linksForRepository(ctx, repo)
	if phabRepo != nil {
		// We need a branch name to construct the Phabricator URL.
		branchName, _, _, err := git.ExecSafe(ctx, backend.CachedGitRepo(repo), []string{"symbolic-ref", "--short", "HEAD"})
		branchName = bytes.TrimSpace(branchName)
		if err == nil && string(branchName) != "" {
			links = append(links, &Resolver{
				url:         fmt.Sprintf("%s/source/%s/browse/%s/%s;%s", strings.TrimSuffix(phabRepo.URL, "/"), phabRepo.Callsign, url.PathEscape(string(branchName)), path, rev),
				serviceType: "phabricator",
			})
		}
	}

	if link != nil {
		var url string
		if isDir {
			url = link.Tree
		} else {
			url = link.Blob
		}
		if url != "" {
			url = strings.NewReplacer("{rev}", rev, "{path}", path).Replace(url)
			links = append(links, &Resolver{url: url, serviceType: serviceType})
		}
	}

	return links, nil
}

// Commit returns the external links for a commit in a repository.
func Commit(ctx context.Context, repo *types.Repo, commitID api.CommitID) (links []*Resolver, err error) {
	commitStr := url.PathEscape(string(commitID))

	phabRepo, link, serviceType := linksForRepository(ctx, repo)
	if phabRepo != nil {
		links = append(links, &Resolver{
			url:         fmt.Sprintf("%s/r%s%s", strings.TrimSuffix(phabRepo.URL, "/"), phabRepo.Callsign, commitStr),
			serviceType: "phabricator",
		})
	}

	if link != nil && link.Commit != "" {
		links = append(links, &Resolver{
			url:         strings.Replace(link.Commit, "{commit}", commitStr, -1),
			serviceType: serviceType,
		})
	}

	return links, nil
}

// linksForRepository gets the information necessary to construct links to resources within this
// repository.
//
// It logs errors to the trace but does not return errors, because external links are not worth
// failing any request for.
func linksForRepository(ctx context.Context, repo *types.Repo) (phabRepo *types.PhabricatorRepo, link *protocol.RepoLinks, serviceType string) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "externallink.linksForRepository")
	defer span.Finish()
	span.SetTag("Repo", repo.URI)
	if repo.ExternalRepo != nil {
		span.SetTag("ExternalRepo", repo.ExternalRepo)
	}

	var err error
	phabRepo, err = db.Phabricator.GetByURI(ctx, repo.URI)
	if err != nil && !errcode.IsNotFound(err) {
		ext.Error.Set(span, true)
		span.SetTag("phabErr", err.Error())
	}

	// Look up repo links in the repo-updater. This supplies links from code host APIs as well as
	// explicitly configured links for repos.list repos.
	info, err := repoupdater.DefaultClient.RepoLookup(ctx, protocol.RepoLookupArgs{
		Repo:         repo.URI,
		ExternalRepo: repo.ExternalRepo,
	})
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("repoUpdaterErr", err.Error())
	}
	if info != nil && info.Repo != nil {
		link = info.Repo.Links
		if info.Repo.ExternalRepo != nil {
			serviceType = info.Repo.ExternalRepo.ServiceType
		}
	}

	return phabRepo, link, serviceType
}
