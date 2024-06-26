package app

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/app/envvar"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/auth/userpasswd"
	"github.com/nxpkg/nxpkg/cmd/frontend/registry"

	"github.com/nxpkg/nxpkg/cmd/frontend/globals"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/app/errorutil"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/app/router"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/app/ui"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/session"
	"github.com/nxpkg/nxpkg/pkg/trace"
)

// NewHandler returns a new app handler that uses the app router.
//
// 🚨 SECURITY: The caller MUST wrap the returned handler in middleware that checks authentication
// and sets the actor in the request context.
func NewHandler() http.Handler {
	session.SetSessionStore(session.NewRedisStore(globals.AppURL.Scheme == "https"))

	r := router.Router()

	m := http.NewServeMux()

	m.Handle("/", r)

	r.Get(router.RobotsTxt).Handler(trace.TraceRoute(http.HandlerFunc(robotsTxt)))
	r.Get(router.Favicon).Handler(trace.TraceRoute(http.HandlerFunc(favicon)))
	r.Get(router.OpenSearch).Handler(trace.TraceRoute(http.HandlerFunc(openSearch)))

	r.Get(router.RepoBadge).Handler(trace.TraceRoute(errorutil.Handler(serveRepoBadge)))

	// Redirects
	r.Get(router.OldToolsRedirect).Handler(trace.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/beta", 301)
	})))

	r.Get(router.GopherconLiveBlog).Handler(trace.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://about.nxpkg.com/go", 302)
	})))

	if envvar.NxpkgDotComMode() {
		r.Get(router.GoSymbolURL).Handler(trace.TraceRoute(errorutil.Handler(serveGoSymbolURL)))
	}

	r.Get(router.UI).Handler(ui.Router())

	r.Get(router.SignUp).Handler(trace.TraceRoute(http.HandlerFunc(userpasswd.HandleSignUp)))
	r.Get(router.SiteInit).Handler(trace.TraceRoute(http.HandlerFunc(userpasswd.HandleSiteInit)))
	r.Get(router.SignIn).Handler(trace.TraceRoute(http.HandlerFunc(userpasswd.HandleSignIn)))
	r.Get(router.SignOut).Handler(trace.TraceRoute(http.HandlerFunc(serveSignOut)))
	r.Get(router.VerifyEmail).Handler(trace.TraceRoute(http.HandlerFunc(serveVerifyEmail)))
	r.Get(router.ResetPasswordInit).Handler(trace.TraceRoute(http.HandlerFunc(userpasswd.HandleResetPasswordInit)))
	r.Get(router.ResetPasswordCode).Handler(trace.TraceRoute(http.HandlerFunc(userpasswd.HandleResetPasswordCode)))

	r.Get(router.RegistryExtensionBundle).Handler(trace.TraceRoute(gziphandler.GzipHandler(http.HandlerFunc(registry.HandleRegistryExtensionBundle))))

	r.Get(router.GDDORefs).Handler(trace.TraceRoute(errorutil.Handler(serveGDDORefs)))
	r.Get(router.Editor).Handler(trace.TraceRoute(errorutil.Handler(serveEditor)))

	r.Get(router.DebugHeaders).Handler(trace.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("Cookie")
		r.Header.Write(w)
	})))
	addDebugHandlers(r.Get(router.Debug).Subrouter())

	return m
}
