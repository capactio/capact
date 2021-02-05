package public

import gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

type getImplementationOptions struct {
	attrFilter              map[gqlpublicapi.FilterRule]map[string]*string
	implPrefixPattern       *string
	requirementsSatisfiedBy map[string]*string
}

func (o *getImplementationOptions) Apply(opts ...GetImplementationOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// ListOption is some configuration that modifies options for a list request.
type GetImplementationOption func(*getImplementationOptions)

func WithImplementationFilter(filter gqlpublicapi.ImplementationRevisionFilter) GetImplementationOption {
	return func(opt *getImplementationOptions) {
		// 1. Process attributes
		opt.attrFilter = map[gqlpublicapi.FilterRule]map[string]*string{}

		for _, attr := range filter.Attributes {
			if attr == nil || attr.Rule == nil {
				continue
			}

			if opt.attrFilter[*attr.Rule] == nil {
				opt.attrFilter[*attr.Rule] = map[string]*string{}
			}

			opt.attrFilter[*attr.Rule][attr.Path] = attr.Revision
		}

		// 2. Process prefix pattern
		opt.implPrefixPattern = filter.PrefixPattern

		// 3. Process TypeInstances
		if len(filter.RequirementsSatisfiedBy) > 0 {
			opt.requirementsSatisfiedBy = map[string]*string{}
			for _, req := range filter.RequirementsSatisfiedBy {
				if req.TypeRef == nil {
					continue
				}
				opt.requirementsSatisfiedBy[req.TypeRef.Path] = req.TypeRef.Revision
			}
		}
	}
}
