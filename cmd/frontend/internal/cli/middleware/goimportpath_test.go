package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nxpkg/nxpkg/cmd/frontend/internal/cli/middleware"
)

func TestGoImportPath(t *testing.T) {
	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			path:       "/nxpkg/nxpkg/usercontent",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/nxpkg/nxpkg git https://github.com/nxpkg/nxpkg">`,
		},
		{
			path:       "/nxpkg/srclib/ann",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/nxpkg/srclib git https://github.com/nxpkg/srclib">`,
		},
		{
			path:       "/nxpkg/srclib-go",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/nxpkg/srclib-go git https://github.com/nxpkg/srclib-go">`,
		},
		{
			path:       "/nxpkg/doesntexist/foobar",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/nxpkg/doesntexist git https://github.com/nxpkg/doesntexist">`,
		},
		{
			path:       "/sqs/pbtypes",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/sqs/pbtypes git https://github.com/sqs/pbtypes">`,
		},
		{
			path:       "/gorilla/mux",
			wantStatus: http.StatusNotFound,
		},
		{
			path:       "/github.com/gorilla/mux",
			wantStatus: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		rw := httptest.NewRecorder()

		req, err := http.NewRequest("GET", test.path+"?go-get=1", nil)
		if err != nil {
			panic(err)
		}

		middleware.NxpkgComGoGetHandler(nil).ServeHTTP(rw, req)

		if got, want := rw.Code, test.wantStatus; got != want {
			t.Errorf("%s:\ngot  %#v\nwant %#v", test.path, got, want)
		}

		if test.wantBody != "" && !strings.Contains(rw.Body.String(), test.wantBody) {
			t.Errorf("response body %q doesn't contain expected substring %q", rw.Body.String(), test.wantBody)
		}
	}
}
