package git_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/nxpkg/nxpkg/pkg/gitserver"
	"github.com/nxpkg/nxpkg/pkg/vcs/git"
)

func TestRepository_BlameFile(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"echo line2 >> f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	gitWantHunks := []*git.Hunk{
		{
			StartLine: 1, EndLine: 2, StartByte: 0, EndByte: 6, CommitID: "e6093374dcf5725d8517db0dccbbf69df65dbde0",
			Message: "foo", Author: git.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
		},
		{
			StartLine: 2, EndLine: 3, StartByte: 6, EndByte: 12, CommitID: "fad406f4fe02c358a09df0d03ec7a36c2c8a20f1",
			Message: "foo", Author: git.Signature{Name: "a", Email: "a@a.com", Date: mustParseTime(time.RFC3339, "2006-01-02T15:04:05Z")},
		},
	}
	tests := map[string]struct {
		repo gitserver.Repo
		path string
		opt  *git.BlameOptions

		wantHunks []*git.Hunk
	}{
		"git cmd": {
			repo: makeGitRepository(t, gitCommands...),
			path: "f",
			opt: &git.BlameOptions{
				NewestCommit: "master",
			},
			wantHunks: gitWantHunks,
		},
	}

	for label, test := range tests {
		newestCommitID, err := git.ResolveRevision(ctx, test.repo, nil, string(test.opt.NewestCommit), nil)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on base: %s", label, test.opt.NewestCommit, err)
			continue
		}

		test.opt.NewestCommit = newestCommitID
		hunks, err := git.BlameFile(ctx, test.repo, test.path, test.opt)
		if err != nil {
			t.Errorf("%s: BlameFile(%s, %+v): %s", label, test.path, test.opt, err)
			continue
		}

		if !reflect.DeepEqual(hunks, test.wantHunks) {
			t.Errorf("%s: hunks != wantHunks\n\nhunks ==========\n%s\n\nwantHunks ==========\n%s", label, asJSON(hunks), asJSON(test.wantHunks))
		}
	}
}
