package git_test

import (
	"testing"

	"github.com/nxpkg/nxpkg/pkg/gitserver"
	"github.com/nxpkg/nxpkg/pkg/vcs/git"
)

func TestGetObject(t *testing.T) {
	t.Parallel()

	gitCommands := []string{
		"echo x > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo           gitserver.Repo
		objectName     string
		wantOID        string
		wantObjectType git.ObjectType
	}{
		"basic": {
			repo:           makeGitRepository(t, gitCommands...),
			objectName:     "e86b31b62399cfc86199e8b6e21a35e76d0e8b5e^{tree}",
			wantOID:        "a1dffc7a64c0b2d395484bf452e9aeb1da3a18f2",
			wantObjectType: git.ObjectTypeTree,
		},
	}

	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			oid, objectType, err := git.GetObject(ctx, test.repo, test.objectName)
			if err != nil {
				t.Fatal(err)
			}
			if oid.String() != test.wantOID {
				t.Errorf("got OID %q, want %q", oid, test.wantOID)
			}
			if objectType != test.wantObjectType {
				t.Errorf("got object type %q, want %q", objectType, test.wantObjectType)
			}
		})
	}
}
