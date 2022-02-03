package public

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"capact.io/capact/pkg/httputil"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"github.com/avast/retry-go"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const retryAttempts = 1

// Client used to communicate with the Capact Public Hub GraphQL APIs
type Client struct {
	client *graphql.Client
}

// NewClient creates a public client with a given GraphQL custom client instance.
func NewClient(cli *graphql.Client) *Client {
	return &Client{client: cli}
}

// NewDefaultClient creates ready to use client with default values.
func NewDefaultClient(endpoint string, opts ...httputil.ClientOption) *Client {
	httpClient := httputil.NewClient(opts...)
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return NewClient(client)
}

// FindInterfaceRevision returns the InterfaceRevision for the given InterfaceReference.
// It will return nil, if the InterfaceRevision is not found.
func (c *Client) FindInterfaceRevision(ctx context.Context, ref gqlpublicapi.InterfaceReference, opts ...InterfaceRevisionOption) (*gqlpublicapi.InterfaceRevision, error) {
	findOpts := &InterfaceRevisionOptions{}
	findOpts.Apply(opts...)

	query, params := c.interfaceQueryForRef(findOpts.fields, ref)
	req := graphql.NewRequest(fmt.Sprintf(`query FindInterfaceRevision($interfacePath: NodePath!, %s) {
		  interface(path: $interfacePath) {
				%s
		  }
		}`, params.Query(), query))

	req.Var("interfacePath", ref.Path)
	params.PopulateVars(req)

	var resp struct {
		Interface struct {
			Revision *gqlpublicapi.InterfaceRevision `json:"rev"`
		} `json:"interface"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))

	if err != nil {
		return nil, errors.Wrap(err, "while executing query to fetch Hub Interface Revision")
	}

	return resp.Interface.Revision, nil
}

// ListTypes returns all requested Types. By default, only root fields are populated.
// Use options to add latestRevision fields or apply additional filtering.
func (c *Client) ListTypes(ctx context.Context, opts ...TypeOption) ([]*gqlpublicapi.Type, error) {
	typeOpts := &TypeOptions{}
	typeOpts.Apply(opts...)

	queryFields := fmt.Sprintf(`
			path
			name
			prefix
			%s`, typeOpts.additionalFields)

	req := graphql.NewRequest(fmt.Sprintf(`query ListTypes($typeFilter: TypeFilter!)  {
		  types(filter: $typeFilter) {
			  %s
		  }
		}`, queryFields))

	req.Var("typeFilter", typeOpts.Filter)

	var resp struct {
		Types []*gqlpublicapi.Type `json:"types"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))

	if err != nil {
		return nil, errors.Wrap(err, "while executing query to list Types")
	}

	return resp.Types, nil
}

func (c *Client) FindTypeRevision(ctx context.Context, ref gqlpublicapi.TypeReference, opts ...TypeRevisionOption) (*gqlpublicapi.TypeRevision, error) {
	findOpts := &TypeRevisionOptions{}
	findOpts.Apply(opts...)

	query, params := c.typeQueryForRef(findOpts.fields, ref)
	req := graphql.NewRequest(fmt.Sprintf(`query FindTypeRevision($typePath: NodePath!, %s) {
		  type(path: $typePath) {
				%s
		  }
		}`, params.Query(), query))

	req.Var("typePath", ref.Path)
	params.PopulateVars(req)

	var resp struct {
		Type struct {
			Revision *gqlpublicapi.TypeRevision `json:"rev"`
		} `json:"type"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))

	if err != nil {
		return nil, errors.Wrap(err, "while executing query to fetch Hub Type Revision")
	}

	return resp.Type.Revision, nil
}

// ListInterfaces returns all Interfaces. By default, only root fields are populated. Use options to add
// latestRevision fields or apply additional filtering.
func (c *Client) ListInterfaces(ctx context.Context, opts ...InterfaceOption) ([]*gqlpublicapi.Interface, error) {
	ifaceOpts := &InterfaceOptions{}
	ifaceOpts.Apply(opts...)

	queryFields := fmt.Sprintf(`
			path
			name
			prefix
			%s`, ifaceOpts.additionalFields)

	var req *graphql.Request
	if ifaceOpts.filter.PathPattern != nil {
		// Send query with filter only if defined.
		// Sending without `interfaceFilter` or with
		//   {
		//    "interfaceFilter": {
		//      "pathPattern": null
		//    }
		//   }
		// always results in empty response and no error.
		req = graphql.NewRequest(fmt.Sprintf(`query ListInterfaces($interfaceFilter: InterfaceFilter!)  {
		  interfaces(filter: $interfaceFilter) {
		  	%s
		  }
		}`, queryFields))
		req.Var("interfaceFilter", ifaceOpts.filter)
	} else {
		req = graphql.NewRequest(fmt.Sprintf(`query ListInterfaces{
		  interfaces {
			%s
		  }
		}`, queryFields))
	}

	var resp struct {
		Interfaces []*gqlpublicapi.Interface `json:"interfaces"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))

	if err != nil {
		return nil, errors.Wrap(err, "while executing query to list Hub Interfaces")
	}

	return resp.Interfaces, nil
}

// GetInterfaceLatestRevisionString returns the latest revision of the available Interfaces.
// Semantic versioning is used to determine the latest revision.
func (c *Client) GetInterfaceLatestRevisionString(ctx context.Context, ref gqlpublicapi.InterfaceReference) (string, error) {
	req := graphql.NewRequest(`query GetInterfaceLatestRevisionString($interfacePath: NodePath!) {
		interface(path: $interfacePath) {
			latestRevision {
				revision
			}
		}
	}`)

	req.Var("interfacePath", ref.Path)

	var resp struct {
		Interface struct {
			LatestRevision *struct {
				Revision string `json:"revision"`
			} `json:"latestRevision"`
		} `json:"interface"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return "", errors.Wrap(err, "while executing query to fetch Interface latest revision string")
	}

	if resp.Interface.LatestRevision == nil {
		return "", fmt.Errorf("cannot find latest revision for Interface %q", ref.Path)
	}

	return resp.Interface.LatestRevision.Revision, nil
}

// ListImplementationRevisions returns ImplementationRevisions. Use options to apply additional filtering.
func (c *Client) ListImplementationRevisions(ctx context.Context, opts ...ListImplementationRevisionsOption) ([]*gqlpublicapi.ImplementationRevision, error) {
	getOpts := &ListImplementationRevisionsOptions{}
	getOpts.Apply(opts...)

	req := graphql.NewRequest(fmt.Sprintf(`query ListImplementationRevisions{
		implementations {
			revisions {
				%s
			}
		}
	}`, getOpts.fields))

	var resp struct {
		Implementations []gqlpublicapi.Implementation `json:"implementations"`
	}

	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))

	if err != nil {
		return nil, errors.Wrap(err, "while executing query to fetch Hub Implementations")
	}

	var revs []*gqlpublicapi.ImplementationRevision

	for _, impl := range resp.Implementations {
		revs = append(revs, impl.Revisions...)
	}

	return revs, nil
}

// ListImplementationRevisionsForInterface returns ImplementationRevisions for the given Interface.
func (c *Client) ListImplementationRevisionsForInterface(ctx context.Context, ref gqlpublicapi.InterfaceReference, opts ...ListImplementationRevisionsForInterfaceOption) ([]gqlpublicapi.ImplementationRevision, error) {
	getOpts := &ListImplementationRevisionsForInterfaceOptions{}
	getOpts.Apply(opts...)

	query, params := c.interfaceQueryForRef(ifaceRevisionAllFields, ref)
	req := graphql.NewRequest(fmt.Sprintf(`query ListImplementationRevisionsForInterface($interfacePath: NodePath!, %s) {
		  interface(path: $interfacePath) {
				%s
		  }
		}`, params.Query(), query))

	req.Var("interfacePath", ref.Path)
	params.PopulateVars(req)

	var resp struct {
		Interface struct {
			LatestRevision struct {
				ImplementationRevisions []gqlpublicapi.ImplementationRevision `json:"implementationRevisions"`
			} `json:"rev"`
		} `json:"interface"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))

	if err != nil {
		return nil, errors.Wrap(err, "while executing query to fetch Hub Implementation")
	}

	result := FilterImplementationRevisions(resp.Interface.LatestRevision.ImplementationRevisions, getOpts)

	result = SortImplementationRevisions(result, getOpts)

	return result, nil
}

// CheckManifestRevisionsExist checks if manifests with provided manifest references exist.
func (c *Client) CheckManifestRevisionsExist(ctx context.Context, manifestRefs []gqlpublicapi.ManifestReference) (map[gqlpublicapi.ManifestReference]bool, error) {
	if len(manifestRefs) == 0 {
		return map[gqlpublicapi.ManifestReference]bool{}, nil
	}

	getAlias := func(i int) string {
		return fmt.Sprintf("partial%d", i)
	}

	strBuilder := strings.Builder{}
	for i, manifestRef := range manifestRefs {
		alias := getAlias(i)
		queryName, err := manifestRef.GQLQueryName()
		if err != nil {
			return nil, errors.Wrap(err, "while getting GraphQL query name for a given manifest")
		}

		partialQuery := fmt.Sprintf(`
			%s: %s(path:"%s") {
				revision(revision:"%s") {
					revision
				}
			}
		`, alias, queryName, manifestRef.Path, manifestRef.Revision)
		strBuilder.WriteString(partialQuery)
	}

	req := graphql.NewRequest(fmt.Sprintf(`
		query CheckManifestRevisionsExist {
			%s
		}`,
		strBuilder.String(),
	))

	var resp map[string]struct {
		Revision struct {
			Revision *string `json:"revision"`
		} `json:"revision"`
	}

	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))

	if err != nil {
		return nil, errors.Wrap(err, "while executing query to check Type Revisions exist")
	}

	result := map[gqlpublicapi.ManifestReference]bool{}
	for i, manifestRef := range manifestRefs {
		alias := getAlias(i)
		result[manifestRef] = resp[alias].Revision.Revision != nil
	}

	return result, nil
}

var key = regexp.MustCompile(`\$(\w+):`)

// Args is used to store arguments to GraphQL queries.
type Args map[string]interface{}

// Query returns the definition for the arguments
// stored in this Args, which has to be put in the
// GraphQL query.
func (a Args) Query() string {
	var out []string
	for k := range a {
		out = append(out, k)
	}
	return strings.Join(out, ",")
}

// PopulateVars fills the variables stores in this Args
// in the provided *graphql.Request.
func (a Args) PopulateVars(req *graphql.Request) {
	for k, v := range a {
		name := key.FindStringSubmatch(k)
		req.Var(name[1], v)
	}
}

func (c *Client) interfaceQueryForRef(fields string, ref gqlpublicapi.InterfaceReference) (string, Args) {
	if ref.Revision == "" {
		return c.latestInterfaceRevision(fields)
	}

	return c.specificInterfaceRevision(fields, ref.Revision)
}

func (c *Client) latestInterfaceRevision(fields string) (string, Args) {
	latestRevision := fmt.Sprintf(`
			rev: latestRevision {
				%s
			}`, fields)

	return latestRevision, Args{}
}

func (c *Client) specificInterfaceRevision(fields string, rev string) (string, Args) {
	specificRevision := fmt.Sprintf(`
			rev: revision(revision: $interfaceRev) {
				%s
			}`, fields)

	return specificRevision, Args{
		"$interfaceRev: Version!": rev,
	}
}

func (c *Client) typeQueryForRef(fields string, ref gqlpublicapi.TypeReference) (string, Args) {
	if ref.Revision == "" {
		return c.latestTypeRevision(fields)
	}

	return c.specificTypeRevision(fields, ref.Revision)
}

func (c *Client) latestTypeRevision(fields string) (string, Args) {
	latestRevision := fmt.Sprintf(`
			rev: latestRevision {
				%s
			}`, fields)

	return latestRevision, Args{}
}

func (c *Client) specificTypeRevision(fields string, rev string) (string, Args) {
	specificRevision := fmt.Sprintf(`
			rev: revision(revision: $typeRev) {
				%s
			}`, fields)

	return specificRevision, Args{
		"$typeRev: Version!": rev,
	}
}
