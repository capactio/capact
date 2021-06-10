package graphql

import "github.com/mindstand/gogm"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	sess *gogm.Session
}

func NewResolver(sess *gogm.Session) *Resolver {
	return &Resolver{
		sess: sess,
	}
}
