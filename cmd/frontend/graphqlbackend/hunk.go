package graphqlbackend

import (
	"context"

	"github.com/nxpkg/nxpkg/cmd/frontend/backend"
	"github.com/nxpkg/nxpkg/pkg/vcs/git"
)

type hunkResolver struct {
	repo *repositoryResolver
	hunk *git.Hunk
}

func (r *hunkResolver) Author() signatureResolver {
	return signatureResolver{
		person: &personResolver{
			name:  r.hunk.Author.Name,
			email: r.hunk.Author.Email,
		},
		date: r.hunk.Author.Date,
	}
}

func (r *hunkResolver) StartLine() int32 {
	return int32(r.hunk.StartLine)
}

func (r *hunkResolver) EndLine() int32 {
	return int32(r.hunk.EndLine)
}

func (r *hunkResolver) StartByte() int32 {
	return int32(r.hunk.EndLine)
}

func (r *hunkResolver) EndByte() int32 {
	return int32(r.hunk.EndByte)
}

func (r *hunkResolver) Rev() string {
	return string(r.hunk.CommitID)
}

func (r *hunkResolver) Message() string {
	return r.hunk.Message
}

func (r *hunkResolver) Commit(ctx context.Context) (*gitCommitResolver, error) {
	commit, err := git.GetCommit(ctx, backend.CachedGitRepo(r.repo.repo), r.hunk.CommitID)
	if err != nil {
		return nil, err
	}
	return toGitCommitResolver(r.repo, commit), nil
}
