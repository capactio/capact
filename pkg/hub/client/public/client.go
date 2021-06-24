package public

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

// ListInterfacesMetadata returns only name, prefix and path. Rest fields have zero value.
func (c *Client) ListInterfacesMetadata(ctx context.Context) ([]gqlpublicapi.Interface, error) {
	req := graphql.NewRequest(`query {
		interfaces {
			name
			prefix
			path
		}		
	}`)

	var resp struct {
		Interfaces []gqlpublicapi.Interface `json:"interfaces"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to fetch Hub Implementation")
	}

	return resp.Interfaces, nil
}

// FindInterfaceRevision returns the InterfaceRevision for the given InterfaceReference.
// It will return nil, if the InterfaceRevision is not found.
func (c *Client) FindInterfaceRevision(ctx context.Context, ref gqlpublicapi.InterfaceReference) (*gqlpublicapi.InterfaceRevision, error) {
	query, params := c.interfaceQueryForRef(ref)
	req := graphql.NewRequest(fmt.Sprintf(`query($interfacePath: NodePath!, %s) {
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

// ListInterfacesWithLatestRevision returns the latest revision of the Interfaces,
// which match the provided filter.
func (c *Client) ListInterfacesWithLatestRevision(ctx context.Context, filter gqlpublicapi.InterfaceFilter) ([]*gqlpublicapi.Interface, error) {
	req := graphql.NewRequest(fmt.Sprintf(`query ListInterface($interfaceFilter: InterfaceFilter!)  {
		  interfaces(filter: $interfaceFilter) {
			%s
		  }
		}`, InterfacesFields))

	req.Var("interfaceFilter", filter)

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
	req := graphql.NewRequest(`query ($interfacePath: NodePath!) {
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

// ListImplementationRevisions returns ImplementationRevisions,
// which match the given filter.
func (c *Client) ListImplementationRevisions(ctx context.Context, filter *gqlpublicapi.ImplementationRevisionFilter) ([]*gqlpublicapi.ImplementationRevision, error) {
	req := graphql.NewRequest(fmt.Sprintf(`query {
		implementations {
			%s
		}		
	}`, ImplementationFields))

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
func (c *Client) ListImplementationRevisionsForInterface(ctx context.Context, ref gqlpublicapi.InterfaceReference, opts ...GetImplementationOption) ([]gqlpublicapi.ImplementationRevision, error) {
	getOpts := &ListImplementationRevisionsOptions{}
	getOpts.Apply(opts...)

	query, params := c.interfaceQueryForRef(ref)
	req := graphql.NewRequest(fmt.Sprintf(`query($interfacePath: NodePath!, %s) {
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

func (c *Client) interfaceQueryForRef(ref gqlpublicapi.InterfaceReference) (string, Args) {
	if ref.Revision == "" {
		return c.latestInterfaceRevision()
	}

	return c.specificInterfaceRevision(ref.Revision)
}

func (c *Client) latestInterfaceRevision() (string, Args) {
	latestRevision := fmt.Sprintf(`
			rev: latestRevision {
				%s
			}`, InterfaceRevisionFields)

	return latestRevision, Args{}
}

func (c *Client) specificInterfaceRevision(rev string) (string, Args) {
	specificRevision := fmt.Sprintf(`
			rev: revision(revision: $interfaceRev) {
				%s
			}`, InterfaceRevisionFields)

	return specificRevision, Args{
		"$interfaceRev: Version!": rev,
	}
}
