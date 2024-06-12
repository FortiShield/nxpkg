package jscontext

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/csrf"

	"github.com/nxpkg/nxpkg/cmd/frontend/db"
	"github.com/nxpkg/nxpkg/cmd/frontend/globals"
	"github.com/nxpkg/nxpkg/cmd/frontend/graphqlbackend"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/app/assetsutil"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/app/envvar"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/auth"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/auth/userpasswd"
	"github.com/nxpkg/nxpkg/cmd/frontend/internal/pkg/siteid"
	"github.com/nxpkg/nxpkg/pkg/actor"
	"github.com/nxpkg/nxpkg/pkg/conf"
	"github.com/nxpkg/nxpkg/pkg/env"
	"github.com/nxpkg/nxpkg/schema"
)

var sentryDSNFrontend = env.Get("SENTRY_DSN_FRONTEND", "", "Sentry/Raven DSN used for tracking of JavaScript errors")

type authProviderInfo struct {
	IsBuiltin         bool   `json:"isBuiltin"`
	DisplayName       string `json:"displayName"`
	AuthenticationURL string `json:"authenticationURL"`
}

// JSContext is made available to JavaScript code via the
// "nxpkg/app/context" module.
//
// 🚨 SECURITY: This struct is sent to all users regardless of whether or
// not they are logged in, for example on an auth.public=false private
// server. Including secret fields here is OK if it is based on the user's
// authentication above, but do not include e.g. hard-coded secrets about
// the server instance here as they would be sent to anonymous users.
type JSContext struct {
	AppRoot        string            `json:"appRoot,omitempty"`
	AppURL         string            `json:"appURL,omitempty"`
	XHRHeaders     map[string]string `json:"xhrHeaders"`
	CSRFToken      string            `json:"csrfToken"`
	UserAgentIsBot bool              `json:"userAgentIsBot"`
	AssetsRoot     string            `json:"assetsRoot"`
	Version        string            `json:"version"`

	IsAuthenticatedUser bool `json:"isAuthenticatedUser"`

	SentryDSN      string `json:"sentryDSN"`
	SiteID         string `json:"siteID"`
	SiteGQLID      string `json:"siteGQLID"`
	Debug          bool   `json:"debug"`
	ShowOnboarding bool   `json:"showOnboarding"`
	EmailEnabled   bool   `json:"emailEnabled"`

	Site                schema.SiteConfiguration `json:"site"` // public subset of site configuration
	LikelyDockerOnMac   bool                     `json:"likelyDockerOnMac"`
	NeedServerRestart   bool                     `json:"needServerRestart"`
	IsRunningDataCenter bool                     `json:"isRunningDataCenter"`

	NxpkgDotComMode bool `json:"nxpkgDotComMode"`

	AccessTokensAllow conf.AccessTokAllow `json:"accessTokensAllow"`

	AllowSignup bool `json:"allowSignup"`

	ResetPasswordEnabled bool `json:"resetPasswordEnabled"`

	AuthProviders []authProviderInfo `json:"authProviders"`
}

// NewJSContextFromRequest populates a JSContext struct from the HTTP
// request.
func NewJSContextFromRequest(req *http.Request) JSContext {
	actor := actor.FromContext(req.Context())

	headers := make(map[string]string)
	headers["x-nxpkg-client"] = globals.AppURL.String()
	headers["X-Requested-With"] = "Nxpkg" // required for httpapi to use cookie auth

	// -- currently we don't associate XHR calls with the parent page's span --
	// if span := opentracing.SpanFromContext(req.Context()); span != nil {
	// 	if err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.TextMapCarrier(headers)); err != nil {
	// 		return JSContext{}, err
	// 	}
	// }

	// Propagate Cache-Control no-cache and max-age=0 directives
	// to the requests made by our client-side JavaScript. This is
	// not a perfect parser, but it catches the important cases.
	if cc := req.Header.Get("cache-control"); strings.Contains(cc, "no-cache") || strings.Contains(cc, "max-age=0") {
		headers["Cache-Control"] = "no-cache"
	}

	csrfToken := csrf.Token(req)
	headers["X-Csrf-Token"] = csrfToken

	siteID := siteid.Get()

	// Show the site init screen?
	siteConfig, err := db.SiteConfig.Get(req.Context())
	showOnboarding := err == nil && !siteConfig.Initialized

	// Auth providers
	var authProviders []authProviderInfo
	for _, p := range auth.Providers() {
		info := p.CachedInfo()
		if info != nil {
			authProviders = append(authProviders, authProviderInfo{
				IsBuiltin:         p.Config().Builtin != nil,
				DisplayName:       info.DisplayName,
				AuthenticationURL: info.AuthenticationURL,
			})
		}
	}

	// 🚨 SECURITY: This struct is sent to all users regardless of whether or
	// not they are logged in, for example on an auth.public=false private
	// server. Including secret fields here is OK if it is based on the user's
	// authentication above, but do not include e.g. hard-coded secrets about
	// the server instance here as they would be sent to anonymous users.
	return JSContext{
		AppURL:              globals.AppURL.String(),
		XHRHeaders:          headers,
		CSRFToken:           csrfToken,
		UserAgentIsBot:      isBot(req.UserAgent()),
		AssetsRoot:          assetsutil.URL("").String(),
		Version:             env.Version,
		IsAuthenticatedUser: actor.IsAuthenticated(),
		SentryDSN:           sentryDSNFrontend,
		Debug:               envvar.InsecureDevMode(),
		SiteID:              siteID,

		SiteGQLID: string(graphqlbackend.SiteGQLID()),

		ShowOnboarding:      showOnboarding,
		EmailEnabled:        conf.CanSendEmail(),
		Site:                publicSiteConfiguration(),
		LikelyDockerOnMac:   likelyDockerOnMac(),
		NeedServerRestart:   conf.NeedServerRestart(),
		IsRunningDataCenter: conf.IsDataCenter(conf.DeployType()),

		NxpkgDotComMode: envvar.NxpkgDotComMode(),

		// Experiments. We pass these through explicitly so we can
		// do the default behavior only in Go land.
		AccessTokensAllow: conf.AccessTokensAllow(),

		ResetPasswordEnabled: userpasswd.ResetPasswordEnabled(),

		AllowSignup: conf.AuthAllowSignup(),

		AuthProviders: authProviders,
	}
}

// publicSiteConfiguration is the subset of the site.schema.json site configuration
// that is necessary for the web app and is not sensitive/secret.
func publicSiteConfiguration() schema.SiteConfiguration {
	c := conf.Get()
	return schema.SiteConfiguration{
		AuthPublic: c.AuthPublic,
	}
}

var isBotPat = regexp.MustCompile(`(?i:googlecloudmonitoring|pingdom.com|go .* package http|nxpkg e2etest|bot|crawl|slurp|spider|feed|rss|camo asset proxy|http-client|nxpkg-client)`)

func isBot(userAgent string) bool {
	return isBotPat.MatchString(userAgent)
}

func likelyDockerOnMac() bool {
	data, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return false // permission errors, or maybe not a Linux OS, etc. Assume we're not docker for mac.
	}
	return bytes.Contains(data, []byte("mac")) || bytes.Contains(data, []byte("osx"))
}
