// Code generated by github.com/Yamashou/gqlgenc, DO NOT EDIT.

package graphql

import (
	"context"
	"net/http"

	"github.com/Yamashou/gqlgenc/client"
	graphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type Client struct {
	Client *client.Client
}

func NewClient(cli *http.Client, baseURL string, options ...client.HTTPRequestOption) *Client {
	return &Client{Client: client.NewClient(cli, baseURL, options...)}
}

type Query struct {
	RepoMetadata    *graphql.RepoMetadata     "json:\"repoMetadata\" graphql:\"repoMetadata\""
	InterfaceGroups []*graphql.InterfaceGroup "json:\"interfaceGroups\" graphql:\"interfaceGroups\""
	InterfaceGroup  *graphql.InterfaceGroup   "json:\"interfaceGroup\" graphql:\"interfaceGroup\""
	Interfaces      []*graphql.Interface      "json:\"interfaces\" graphql:\"interfaces\""
	Interface       *graphql.Interface        "json:\"interface\" graphql:\"interface\""
	Types           []*graphql.Type           "json:\"types\" graphql:\"types\""
	Type            *graphql.Type             "json:\"type\" graphql:\"type\""
	Implementations []*graphql.Implementation "json:\"implementations\" graphql:\"implementations\""
	Implementation  *graphql.Implementation   "json:\"implementation\" graphql:\"implementation\""
	Attributes      []*graphql.Attribute      "json:\"attributes\" graphql:\"attributes\""
	Attribute       *graphql.Attribute        "json:\"attribute\" graphql:\"attribute\""
}
type InterfaceRevisionFragment struct {
	Metadata struct {
		Prefix      *string "json:\"prefix\" graphql:\"prefix\""
		Path        string  "json:\"path\" graphql:\"path\""
		Name        string  "json:\"name\" graphql:\"name\""
		DisplayName *string "json:\"displayName\" graphql:\"displayName\""
		Description string  "json:\"description\" graphql:\"description\""
		Maintainers []*struct {
			Name  *string "json:\"name\" graphql:\"name\""
			Email string  "json:\"email\" graphql:\"email\""
		} "json:\"maintainers\" graphql:\"maintainers\""
		IconURL *string "json:\"iconURL\" graphql:\"iconURL\""
	} "json:\"metadata\" graphql:\"metadata\""
	Revision string "json:\"revision\" graphql:\"revision\""
	Spec     struct {
		Input struct {
			Parameters []*struct {
				Name       string      "json:\"name\" graphql:\"name\""
				JSONSchema interface{} "json:\"jsonSchema\" graphql:\"jsonSchema\""
			} "json:\"parameters\" graphql:\"parameters\""
			TypeInstances []*struct {
				Name    string "json:\"name\" graphql:\"name\""
				TypeRef struct {
					Path     string "json:\"path\" graphql:\"path\""
					Revision string "json:\"revision\" graphql:\"revision\""
				} "json:\"typeRef\" graphql:\"typeRef\""
				Verbs []graphql.TypeInstanceOperationVerb "json:\"verbs\" graphql:\"verbs\""
			} "json:\"typeInstances\" graphql:\"typeInstances\""
		} "json:\"input\" graphql:\"input\""
		Output struct {
			TypeInstances []*struct {
				Name    string "json:\"name\" graphql:\"name\""
				TypeRef struct {
					Path     string "json:\"path\" graphql:\"path\""
					Revision string "json:\"revision\" graphql:\"revision\""
				} "json:\"typeRef\" graphql:\"typeRef\""
			} "json:\"typeInstances\" graphql:\"typeInstances\""
		} "json:\"output\" graphql:\"output\""
	} "json:\"spec\" graphql:\"spec\""
	Signature struct {
		Och string "json:\"och\" graphql:\"och\""
	} "json:\"signature\" graphql:\"signature\""
}
type InterfacesWithPrefixFilter struct {
	Interfaces []*struct {
		Name           string                       "json:\"name\" graphql:\"name\""
		Prefix         string                       "json:\"prefix\" graphql:\"prefix\""
		Path           string                       "json:\"path\" graphql:\"path\""
		LatestRevision *InterfaceRevisionFragment   "json:\"latestRevision\" graphql:\"latestRevision\""
		Revisions      []*InterfaceRevisionFragment "json:\"revisions\" graphql:\"revisions\""
	} "json:\"interfaces\" graphql:\"interfaces\""
}
type InterfaceLatestRevision struct {
	Interface *struct {
		LatestRevision *InterfaceRevisionFragment "json:\"latestRevision\" graphql:\"latestRevision\""
	} "json:\"interface\" graphql:\"interface\""
}

const InterfacesWithPrefixFilterQuery = `query InterfacesWithPrefixFilter ($pathPattern: NodePathPattern!) {
	interfaces(filter: {pathPattern:$pathPattern}) {
		name
		prefix
		path
		latestRevision {
			... InterfaceRevisionFragment
		}
		revisions {
			... InterfaceRevisionFragment
		}
	}
}
fragment InterfaceRevisionFragment on InterfaceRevision {
	metadata {
		prefix
		path
		name
		displayName
		description
		maintainers {
			name
			email
		}
		iconURL
	}
	revision
	spec {
		input {
			parameters {
				name
				jsonSchema
			}
			typeInstances {
				name
				typeRef {
					path
					revision
				}
				verbs
			}
		}
		output {
			typeInstances {
				name
				typeRef {
					path
					revision
				}
			}
		}
	}
	signature {
		och
	}
}
`

func (c *Client) InterfacesWithPrefixFilter(ctx context.Context, pathPattern string, httpRequestOptions ...client.HTTPRequestOption) (*InterfacesWithPrefixFilter, error) {
	vars := map[string]interface{}{
		"pathPattern": pathPattern,
	}

	var res InterfacesWithPrefixFilter
	if err := c.Client.Post(ctx, "InterfacesWithPrefixFilter", InterfacesWithPrefixFilterQuery, &res, vars, httpRequestOptions...); err != nil {
		return nil, err
	}

	return &res, nil
}

const InterfaceLatestRevisionQuery = `query InterfaceLatestRevision ($interfacePath: NodePath!) {
	interface(path: $interfacePath) {
		latestRevision {
			... InterfaceRevisionFragment
		}
	}
}
fragment InterfaceRevisionFragment on InterfaceRevision {
	metadata {
		prefix
		path
		name
		displayName
		description
		maintainers {
			name
			email
		}
		iconURL
	}
	revision
	spec {
		input {
			parameters {
				name
				jsonSchema
			}
			typeInstances {
				name
				typeRef {
					path
					revision
				}
				verbs
			}
		}
		output {
			typeInstances {
				name
				typeRef {
					path
					revision
				}
			}
		}
	}
	signature {
		och
	}
}
`

func (c *Client) InterfaceLatestRevision(ctx context.Context, interfacePath string, httpRequestOptions ...client.HTTPRequestOption) (*InterfaceLatestRevision, error) {
	vars := map[string]interface{}{
		"interfacePath": interfacePath,
	}

	var res InterfaceLatestRevision
	if err := c.Client.Post(ctx, "InterfaceLatestRevision", InterfaceLatestRevisionQuery, &res, vars, httpRequestOptions...); err != nil {
		return nil, err
	}

	return &res, nil
}
