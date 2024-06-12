package app

import "github.com/nxpkg/nxpkg/pkg/txemail"

func init() {
	txemail.DisableSilently()
}
