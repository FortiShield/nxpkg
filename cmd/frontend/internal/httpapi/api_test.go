package httpapi

import (
	"github.com/gorilla/mux"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/httpapi/router"
	"github.com/nxpkg/nxpkg/pkg/httptestutil"
	"github.com/nxpkg/nxpkg/pkg/txemail"
)

func init() {
	txemail.DisableSilently()
}

func newTest() *httptestutil.Client {
	mux := NewHandler(router.New(mux.NewRouter()))
	return httptestutil.NewTest(mux)
}
