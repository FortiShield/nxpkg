package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/gqltesting"
	"github.com/nxpkg/nxpkg/cmd/frontend/db"
	"github.com/nxpkg/nxpkg/cmd/frontend/types"
)

func TestOrganization(t *testing.T) {
	resetMocks()
	db.Mocks.Orgs.GetByName = func(context.Context, string) (*types.Org, error) {
		return &types.Org{ID: 1, Name: "acme"}, nil
	}

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					organization(name: "acme") {
						name
					}
				}
			`,
			ExpectedResult: `
				{
					"organization": {
						"name": "acme"
					}
				}
			`,
		},
	})
}

func TestNode_Org(t *testing.T) {
	resetMocks()
	db.Mocks.Orgs.MockGetByID_Return(t, &types.Org{ID: 1, Name: "acme"}, nil)

	gqltesting.RunTests(t, []*gqltesting.Test{
		{
			Schema: GraphQLSchema,
			Query: `
				{
					node(id: "T3JnOjE=") {
						id
						... on Org {
							name
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"node": {
						"id": "T3JnOjE=",
						"name": "acme"
					}
				}
			`,
		},
	})
}
