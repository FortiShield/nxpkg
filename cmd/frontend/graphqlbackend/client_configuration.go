package graphqlbackend

import (
	"context"

	"github.com/nxpkg/nxpkg/pkg/conf"
)

type clientConfigurationResolver struct {
	contentScriptUrls []string
	parentNxpkg *parentNxpkgResolver
}

type parentNxpkgResolver struct {
	url string
}

func (r *clientConfigurationResolver) ContentScriptURLs() []string {
	return r.contentScriptUrls
}

func (r *clientConfigurationResolver) ParentNxpkg() *parentNxpkgResolver {
	return r.parentNxpkg
}

func (r *parentNxpkgResolver) URL() string {
	return r.url
}

func (r *schemaResolver) ClientConfiguration(ctx context.Context) (*clientConfigurationResolver, error) {
	cfg := conf.Get()
	var contentScriptUrls []string
	for _, gh := range cfg.Github {
		contentScriptUrls = append(contentScriptUrls, gh.Url)
	}
	for _, bb := range cfg.BitbucketServer {
		contentScriptUrls = append(contentScriptUrls, bb.Url)
	}
	for _, gl := range cfg.Gitlab {
		contentScriptUrls = append(contentScriptUrls, gl.Url)
	}
	for _, ph := range cfg.Phabricator {
		contentScriptUrls = append(contentScriptUrls, ph.Url)
	}
	for _, rb := range cfg.ReviewBoard {
		contentScriptUrls = append(contentScriptUrls, rb.Url)
	}

	var parentNxpkg parentNxpkgResolver
	if cfg.ParentNxpkg != nil {
		parentNxpkg.url = cfg.ParentNxpkg.Url
	}

	return &clientConfigurationResolver{
		contentScriptUrls: contentScriptUrls,
		parentNxpkg: &parentNxpkg,
	}, nil
}
