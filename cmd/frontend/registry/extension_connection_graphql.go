package registry

import (
	"context"
	"sort"
	"sync"

	"github.com/nxpkg/nxpkg/cmd/frontend/graphqlbackend"
	"github.com/nxpkg/nxpkg/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/nxpkg/nxpkg/pkg/registry"
)

// makePrioritizeExtensionIDsSet returns a set whose values are the elements of
// args.PrioritizeExtensionIDs.
func makePrioritizeExtensionIDsSet(args graphqlbackend.RegistryExtensionConnectionArgs) map[string]struct{} {
	if args.PrioritizeExtensionIDs == nil {
		return nil
	}
	set := make(map[string]struct{}, len(*args.PrioritizeExtensionIDs))
	for _, id := range *args.PrioritizeExtensionIDs {
		set[id] = struct{}{}
	}
	return set
}

func (r *extensionRegistryResolver) Extensions(ctx context.Context, args *graphqlbackend.RegistryExtensionConnectionArgs) (graphqlbackend.RegistryExtensionConnection, error) {
	return &registryExtensionConnectionResolver{args: *args}, nil
}

// registryExtensionConnectionResolver resolves a list of registry extensions.
type registryExtensionConnectionResolver struct {
	args graphqlbackend.RegistryExtensionConnectionArgs

	// cache results because they are used by multiple fields
	once               sync.Once
	registryExtensions []graphqlbackend.RegistryExtension
	err                error
}

var (
	// ListLocalRegistryExtensions lists and returns local registry extensions according to the args. If
	// there is no local extension registry, it is not implemented.
	ListLocalRegistryExtensions func(context.Context, graphqlbackend.RegistryExtensionConnectionArgs) ([]graphqlbackend.RegistryExtension, error)

	// CountLocalRegistryExtensions returns the count of local registry extensions according to the
	// args. Pagination-related args are ignored. If there is no local extension registry, it is not
	// implemented.
	CountLocalRegistryExtensions func(context.Context, graphqlbackend.RegistryExtensionConnectionArgs) (int, error)
)

func (r *registryExtensionConnectionResolver) compute(ctx context.Context) ([]graphqlbackend.RegistryExtension, error) {
	r.once.Do(func() {
		args2 := r.args
		if args2.First != nil {
			*args2.First++ // so we can detect if there is a next page
		}

		var query string
		if args2.Query != nil {
			query = *args2.Query
		}

		// Query local registry extensions.
		var local []graphqlbackend.RegistryExtension
		if r.args.Local && ListLocalRegistryExtensions != nil {
			local, r.err = ListLocalRegistryExtensions(ctx, r.args)
			if r.err != nil {
				return
			}
		}

		var remote []*registry.Extension

		// BACKCOMPAT: Include synthesized extensions for known language servers.
		if r.args.Local {
			remote = append(remote, listSynthesizedRegistryExtensions(ctx, query)...)
		}

		// Query remote registry extensions, if filters would match any.
		if args2.Publisher == nil && r.args.Remote {
			xs, err := listRemoteRegistryExtensions(ctx, query)
			if err != nil {
				// Continue execution even if r.err != nil so that partial (local) results are returned
				// even when the remote registry is inaccessible.
				r.err = err
			}
			remote = append(remote, xs...)
		}

		r.registryExtensions = make([]graphqlbackend.RegistryExtension, len(local)+len(remote))
		copy(r.registryExtensions, local)
		for i, x := range remote {
			r.registryExtensions[len(local)+i] = &registryExtensionRemoteResolver{v: x}
		}

		if r.args.PrioritizeExtensionIDs != nil && len(*r.args.PrioritizeExtensionIDs) > 0 {
			// Sort prioritized extension IDs first.
			set := makePrioritizeExtensionIDsSet(r.args)
			sort.SliceStable(r.registryExtensions, func(i, j int) bool {
				_, pi := set[r.registryExtensions[i].ExtensionID()]
				_, pj := set[r.registryExtensions[j].ExtensionID()]
				return pi && !pj
			})
		}
	})
	return r.registryExtensions, r.err
}

func (r *registryExtensionConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.RegistryExtension, error) {
	// See (*registryExtensionConnectionResolver).Error for why we ignore the error.
	xs, _ := r.compute(ctx)
	if r.args.First != nil && len(xs) > int(*r.args.First) {
		xs = xs[:int(*r.args.First)]
	}
	return xs, nil
}

func (r *registryExtensionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	var total int

	if r.args.Local && CountLocalRegistryExtensions != nil {
		dbCount, err := CountLocalRegistryExtensions(ctx, r.args)
		if err != nil {
			return 0, err
		}
		total += dbCount
	}

	// Count remote extensions. Performing an actual fetch is necessary.
	//
	// See (*registryExtensionConnectionResolver).Error for why we ignore the error.
	xs, _ := r.compute(ctx)
	for _, x := range xs {
		if _, isRemote := x.(*registryExtensionRemoteResolver); isRemote {
			total++
		}
	}

	return int32(total), nil
}

func (r *registryExtensionConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	// See (*registryExtensionConnectionResolver).Error for why we ignore the error.
	registryExtensions, _ := r.compute(ctx)
	return graphqlutil.HasNextPage(r.args.First != nil && len(registryExtensions) > int(*r.args.First)), nil
}

func (r *registryExtensionConnectionResolver) URL(ctx context.Context) (*string, error) {
	if r.args.Publisher == nil || RegistryPublisherByID == nil {
		return nil, nil
	}

	publisher, err := RegistryPublisherByID(ctx, *r.args.Publisher)
	if err != nil {
		return nil, err
	}
	return publisher.RegistryExtensionConnectionURL()
}

func (r *registryExtensionConnectionResolver) Error(ctx context.Context) *string {
	// See the GraphQL API schema documentation for this field for an explanation of why we return
	// errors in this way.
	//
	// TODO(sqs): When https://github.com/graph-gophers/graphql-go/pull/219 or similar is merged, we
	// can make the other fields return data *and* an error, instead of using this separate error
	// field.
	_, err := r.compute(ctx)
	if err == nil {
		return nil
	}
	return strptr(err.Error())
}

func strptr(s string) *string { return &s }
