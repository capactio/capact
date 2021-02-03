package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

// current hack to do not take into account gcp solutions
var ignoreImplementationsWithAttributes = map[string]struct{}{
	"cap.attribute.cloud.provider.gcp": {},
}

// Client used to communicate with the Voltron OCH GraphQL APIs
// TODO this should be split into public and local OCH clients and composed together here
type Client struct {
	client *graphql.Client
}

func NewClient(endpoint string, httpClient *http.Client) *Client {
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return &Client{
		client: client,
	}
}

func (c *Client) ListInterfacesMetadata(ctx context.Context) ([]ochpublicgraphql.Interface, error) {
	req := graphql.NewRequest(`query {
		interfaces {
			name
			prefix
			path
		}		
	}`)

	var resp struct {
		Interfaces []ochpublicgraphql.Interface `json:"interfaces"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to fetch OCH Implementation")
	}

	return resp.Interfaces, nil
}

// TODO(SV-206): handle that properly and take into account the ref.Revision - default to latest if not present.
func (c *Client) GetImplementationForInterface(ctx context.Context, ref ochpublicgraphql.TypeReference) (*ochpublicgraphql.ImplementationRevision, error) {
	req := graphql.NewRequest(`query($interfacePath: NodePath!) {
		  interface(path: $interfacePath) {
			latestRevision {
			  implementationRevisions {
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
					documentationURL
					supportURL
					iconURL
					attributes {
					  metadata {
						path
					  }
					}
				}
				revision
				spec {
					appVersion
					implements {
						path
						revision
					}
					requires {
						prefix
						oneOf {
							typeRef {
								path
								revision
							}
							valueConstraints
						}
						anyOf {
							typeRef {
								path
								revision
							}
							valueConstraints
						}
						allOf {
							typeRef {
								path
								revision
							}
							valueConstraints
						}
					}
					imports {
						interfaceGroupPath
						alias
						appVersion
						methods {
							name
							revision
						}
					}
					additionalInput {
						typeInstances {
							name
							typeRef {
								path
								revision
							}
							verbs
						}
					}
					additionalOutput {
						typeInstances {
							name
							typeRef {
								path
								revision
							}
						}
						typeInstanceRelations {
							typeInstanceName
							uses
						}
					}
					action {
						runnerInterface
						args
					}
				}
				signature {
					och
				}
			  }
			}
		  }
		}`)

	req.Var("interfacePath", ref.Path)
	var resp struct {
		Interface struct {
			LatestRevision struct {
				ImplementationRevisions []ochpublicgraphql.ImplementationRevision `json:"implementationRevisions"`
			} `json:"latestRevision"`
		} `json:"interface"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to fetch OCH Implementation")
	}

	impls := c.filterOutIgnoredImpl(resp.Interface.LatestRevision.ImplementationRevisions)

	if len(impls) == 0 {
		return nil, errors.Errorf("No implementation found for %q", ref.Path)
	}

	return &impls[0], nil
}

func (c *Client) filterOutIgnoredImpl(revs []ochpublicgraphql.ImplementationRevision) []ochpublicgraphql.ImplementationRevision {
	var out []ochpublicgraphql.ImplementationRevision

revCheck:
	for _, impl := range revs {
		for _, atr := range impl.Metadata.Attributes {
			if atr != nil && atr.Metadata != nil && atr.Metadata.Path != nil {
				_, found := ignoreImplementationsWithAttributes[*atr.Metadata.Path]
				if found {
					continue revCheck
				}
			}
		}
		out = append(out, impl)
	}

	return out
}

func (c *Client) CreateTypeInstance(ctx context.Context, in *ochlocalgraphql.CreateTypeInstanceInput) (*ochlocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`mutation($in: CreateTypeInstanceInput!) {
		createTypeInstance(
			in: $in
		) {
			%s
		}
	}`, typeInstanceFields)

	req := graphql.NewRequest(query)
	req.Var("in", in)

	var resp struct {
		TypeInstance ochlocalgraphql.TypeInstance `json:"createTypeInstance"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to create TypeInstance")
	}

	return &resp.TypeInstance, nil
}

func (c *Client) GetTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`query($id: ID!) {
		typeInstance(id: $id) {
			%s	
		}
	}`, typeInstanceFields)

	req := graphql.NewRequest(query)
	req.Var("id", id)

	var resp struct {
		TypeInstance ochlocalgraphql.TypeInstance `json:"typeInstance"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to get TypeInstance")
	}

	return &resp.TypeInstance, nil
}

func (c *Client) DeleteTypeInstance(ctx context.Context, id string) error {
	req := graphql.NewRequest(`mutation ($id: ID!) {
	  deleteTypeInstance(
	    id: $id
	  )
	}`)
	req.Var("id", id)

	var resp struct{}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing query to get TypeInstance")
	}

	return nil
}

const typeInstanceFields = `
	resourceVersion
	metadata {
	  id
	  attributes {
	    path
	    revision
	  }
	}
	spec {
	  typeRef {
	    path
	    revision
	  }
	  value
	}
`
