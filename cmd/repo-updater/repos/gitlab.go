package repos

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/nxpkg/nxpkg/cmd/repo-updater/internal/externalservice/gitlab"
	"github.com/nxpkg/nxpkg/pkg/api"
	"github.com/nxpkg/nxpkg/pkg/atomicvalue"
	"github.com/nxpkg/nxpkg/pkg/conf"
	"github.com/nxpkg/nxpkg/pkg/conf/reposource"
	"github.com/nxpkg/nxpkg/pkg/repoupdater/protocol"
	"github.com/nxpkg/nxpkg/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// GitLabServiceType is the (api.ExternalRepoSpec).ServiceType value for GitLab projects. The ServiceID value is
// the base URL to the GitLab instance (https://gitlab.com or self-hosted GitLab URL).
const GitLabServiceType = "gitlab"

// GitLabExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified GitLab project.
func GitLabExternalRepoSpec(proj *gitlab.Project, baseURL url.URL) *api.ExternalRepoSpec {
	return &api.ExternalRepoSpec{
		ID:          strconv.Itoa(proj.ID),
		ServiceType: GitLabServiceType,
		ServiceID:   NormalizeBaseURL(&baseURL).String(),
	}
}

var gitlabConnections = atomicvalue.New()

func init() {
	conf.Watch(func() {
		gitlabConnections.Set(func() interface{} {
			gitlabConf := conf.Get().Gitlab

			var hasGitLabDotComConnection bool
			for _, c := range gitlabConf {
				u, _ := url.Parse(c.Url)
				if u != nil && (u.Hostname() == "gitlab.com" || u.Hostname() == "www.gitlab.com") {
					hasGitLabDotComConnection = true
					break
				}
			}
			if !hasGitLabDotComConnection {
				// Add a GitLab.com entry by default, to support navigating to URL paths like
				// /gitlab.com/foo/bar to auto-add that project.
				gitlabConf = append(gitlabConf, &schema.GitLabConnection{
					ProjectQuery:                []string{"none"}, // don't try to list all repositories during syncs
					Url:                         "https://gitlab.com",
					InitialRepositoryEnablement: true,
				})
			}

			var conns []*gitlabConnection
			for _, c := range gitlabConf {
				conn, err := newGitLabConnection(c)
				if err != nil {
					log15.Error("Error processing configured GitLab connection. Skipping it.", "url", c.Url, "error", err)
					continue
				}
				conns = append(conns, conn)
			}
			return conns
		})
		gitLabRepositorySyncWorker.restart()
	})
}

// getGitLabConnection returns the GitLab connection (config + API client) that is responsible for
// the repository specified by the args.
func getGitLabConnection(args protocol.RepoLookupArgs) (*gitlabConnection, error) {
	gitlabConnections := gitlabConnections.Get().([]*gitlabConnection)
	if args.ExternalRepo != nil && args.ExternalRepo.ServiceType == GitLabServiceType {
		// Look up by external repository spec.
		for _, conn := range gitlabConnections {
			if args.ExternalRepo.ServiceID == conn.baseURL.String() {
				return conn, nil
			}
		}
		return nil, errors.Wrap(gitlab.ErrNotFound, fmt.Sprintf("no configured GitLab connection with URL: %q", args.ExternalRepo.ServiceID))
	}

	if args.Repo != "" {
		// Look up by repository URI.
		repo := strings.ToLower(string(args.Repo))
		for _, conn := range gitlabConnections {
			if strings.HasPrefix(repo, conn.baseURL.Hostname()+"/") {
				return conn, nil
			}
		}
	}

	return nil, nil
}

// GetGitLabRepositoryMock is set by tests that need to mock GetGitLabRepository.
var GetGitLabRepositoryMock func(args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error)

// GetGitLabRepository queries a configured GitLab connection endpoint for information about the
// specified repository (a.k.a. project in GitLab's naming scheme).
//
// If args.Repo refers to a repository that is not known to be on a configured GitLab connection's
// host, it returns authoritative == false.
func GetGitLabRepository(ctx context.Context, args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
	if GetGitLabRepositoryMock != nil {
		return GetGitLabRepositoryMock(args)
	}

	ghrepoToRepoInfo := func(proj *gitlab.Project, conn *gitlabConnection) *protocol.RepoInfo {
		return &protocol.RepoInfo{
			URI:          gitlabProjectToRepoPath(conn, proj),
			ExternalRepo: GitLabExternalRepoSpec(proj, *conn.baseURL),
			Description:  proj.Description,
			Fork:         proj.ForkedFromProject != nil,
			Archived:     proj.Archived,
			VCS: protocol.VCSInfo{
				URL: conn.authenticatedRemoteURL(proj),
			},
			Links: &protocol.RepoLinks{
				Root:   proj.WebURL,
				Tree:   proj.WebURL + "/tree/{rev}/{path}",
				Blob:   proj.WebURL + "/blob/{rev}/{path}",
				Commit: proj.WebURL + "/commit/{commit}",
			},
		}
	}

	conn, err := getGitLabConnection(args)
	if err != nil {
		return nil, true, err // refers to a GitLab repo but the host is not configured
	}
	if conn == nil {
		return nil, false, nil // refers to a non-GitLab repo
	}

	if args.ExternalRepo != nil && args.ExternalRepo.ServiceType == GitLabServiceType {
		// Look up by external repository spec.
		id, err := strconv.Atoi(args.ExternalRepo.ID)
		if err != nil {
			return nil, true, err
		}
		proj, err := conn.client.GetProject(ctx, id, "")
		if proj != nil {
			repo = ghrepoToRepoInfo(proj, conn)
		}
		return repo, true, err
	}

	if args.Repo != "" {
		// Look up by repository URI.
		pathWithNamespace := strings.TrimPrefix(strings.ToLower(string(args.Repo)), conn.baseURL.Hostname()+"/")
		proj, err := conn.client.GetProject(ctx, 0, pathWithNamespace)
		if proj != nil {
			repo = ghrepoToRepoInfo(proj, conn)
		}
		return repo, true, err
	}

	return nil, true, fmt.Errorf("unable to look up GitLab repository (%+v)", args)
}

var gitLabRepositorySyncWorker = &worker{
	work: func(ctx context.Context, shutdown chan struct{}) {
		gitlabConnections := gitlabConnections.Get().([]*gitlabConnection)
		if len(gitlabConnections) == 0 {
			return
		}
		for _, c := range gitlabConnections {
			go func(c *gitlabConnection) {
				for {
					if rateLimitRemaining, rateLimitReset, ok := c.client.RateLimit.Get(); ok && rateLimitRemaining < 50 {
						wait := rateLimitReset + 10*time.Second
						log15.Warn("GitLab API rate limit is almost exhausted. Waiting until rate limit is reset.", "wait", rateLimitReset, "rateLimitRemaining", rateLimitRemaining)
						time.Sleep(wait)
					}
					updateGitLabProjects(ctx, c)
					gitlabUpdateTime.WithLabelValues(c.baseURL.String()).Set(float64(time.Now().Unix()))
					select {
					case <-shutdown:
						return
					case <-time.After(getUpdateInterval()):
					}
				}
			}(c)
		}
	},
}

// RunGitLabRepositorySyncWorker runs the worker that syncs projects from configured GitLab instances to
// Nxpkg.
func RunGitLabRepositorySyncWorker(ctx context.Context) {
	gitLabRepositorySyncWorker.start(ctx)
}

func gitlabProjectToRepoPath(conn *gitlabConnection, proj *gitlab.Project) api.RepoURI {
	return reposource.GitLabRepoURI(conn.config.RepositoryPathPattern, conn.baseURL.Hostname(), proj.PathWithNamespace)
}

// updateGitLabProjects ensures that all provided repositories exist in the repository table.
func updateGitLabProjects(ctx context.Context, conn *gitlabConnection) {
	projs := conn.listAllProjects(ctx)

	repoChan := make(chan repoCreateOrUpdateRequest)
	defer close(repoChan)
	go createEnableUpdateRepos(ctx, fmt.Sprintf("gitlab:%s", conn.config.Token), repoChan)
	for proj := range projs {
		repoChan <- repoCreateOrUpdateRequest{
			RepoCreateOrUpdateRequest: api.RepoCreateOrUpdateRequest{
				RepoURI:      gitlabProjectToRepoPath(conn, proj),
				ExternalRepo: GitLabExternalRepoSpec(proj, *conn.baseURL),
				Description:  proj.Description,
				Fork:         proj.ForkedFromProject != nil,
				Archived:     proj.Archived,
				Enabled:      conn.config.InitialRepositoryEnablement,
			},
			URL: conn.authenticatedRemoteURL(proj),
		}
	}
}

func newGitLabConnection(config *schema.GitLabConnection) (*gitlabConnection, error) {
	baseURL, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	baseURL = NormalizeBaseURL(baseURL)

	transport, err := cachedTransportWithCertTrusted(config.Certificate)
	if err != nil {
		return nil, err
	}

	return &gitlabConnection{
		config:  config,
		baseURL: baseURL,
		client:  gitlab.NewClient(baseURL, config.Token, transport),
	}, nil
}

type gitlabConnection struct {
	config  *schema.GitLabConnection
	baseURL *url.URL // URL with path /api/v4 (no trailing slash)
	client  *gitlab.Client
}

// authenticatedRemoteURL returns the GitLab projects's Git remote URL with the configured GitLab personal access
// token inserted in the URL userinfo, for repositories needing authentication.
func (c *gitlabConnection) authenticatedRemoteURL(proj *gitlab.Project) string {
	if c.config.GitURLType == "ssh" {
		return proj.SSHURLToRepo // SSH authentication must be provided out-of-band
	}
	if c.config.Token == "" || !proj.RequiresAuthentication() {
		return proj.HTTPURLToRepo
	}
	u, err := url.Parse(proj.HTTPURLToRepo)
	if err != nil {
		log15.Warn("Error adding authentication to GitLab repository Git remote URL.", "url", proj.HTTPURLToRepo, "error", err)
		return proj.HTTPURLToRepo
	}
	// Any username works; "git" is not special.
	u.User = url.UserPassword("git", c.config.Token)
	return u.String()
}

func (c *gitlabConnection) listAllProjects(ctx context.Context) <-chan *gitlab.Project {
	if len(c.config.ProjectQuery) == 0 {
		c.config.ProjectQuery = []string{"?membership=true"}
	}

	normalizeQuery := func(projectQuery string) (url.Values, error) {
		q, err := url.ParseQuery(strings.TrimPrefix(projectQuery, "?"))
		if err != nil {
			return nil, err
		}
		if q.Get("order_by") == "" && q.Get("sort") == "" {
			// Apply default ordering to get the likely more relevant projects first.
			q.Set("order_by", "last_activity_at")
		}
		return q, nil
	}

	const perPage = 100 // max GitLab API per_page parameter
	ch := make(chan *gitlab.Project, perPage)
	go func() {
	projectsQueries:
		for _, projectQuery := range c.config.ProjectQuery {
			if projectQuery == "none" {
				continue
			}
			q, err := normalizeQuery(projectQuery)
			if err != nil {
				log15.Error("Skipping invalid GitLab projectQuery", "projectQuery", projectQuery, "error", err)
				continue
			}
			q.Set("per_page", strconv.Itoa(perPage))

			url := "projects?" + q.Encode() // first page URL
			for {
				projects, nextPageURL, err := c.client.ListProjects(ctx, url)
				if err != nil {
					log15.Error("Error listing GitLab projects", "url", url, "error", err)
					continue projectsQueries
				}
				for _, p := range projects {
					ch <- p
				}
				if nextPageURL == nil {
					break
				}
				url = *nextPageURL
			}
		}
		close(ch)
	}()

	return ch
}
