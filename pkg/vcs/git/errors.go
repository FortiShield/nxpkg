package git

import (
	"fmt"

	"github.com/nxpkg/nxpkg/pkg/api"
)

// RevisionNotFoundError is an error that reports a revision doesn't exist.
type RevisionNotFoundError struct {
	Repo api.RepoURI
	Spec string
}

func (e *RevisionNotFoundError) Error() string {
	return fmt.Sprintf("revision not found: %s@%s", e.Repo, e.Spec)
}

// IsRevisionNotFound reports if err is a RevisionNotFoundError.
func IsRevisionNotFound(err error) bool {
	_, ok := err.(*RevisionNotFoundError)
	return ok
}
