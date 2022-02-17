package client

import (
	"github.com/machinebox/graphql"
)

// Client used to communicate with the Capact Engine GraphQL API
type Client struct {
	Action
	Policy
}

// New returns a new Client instance.
func New(gqlClient *graphql.Client) *Client {
	return &Client{
		Action: Action{client: gqlClient},
		Policy: Policy{client: gqlClient},
	}
}
