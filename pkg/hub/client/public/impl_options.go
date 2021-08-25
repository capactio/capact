package public

import (
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
)

// ListImplementationRevisionsForInterfaceOption provides an option to configure the get request for Implementations.
type ListImplementationRevisionsForInterfaceOption func(*ListImplementationRevisionsForInterfaceOptions)

// ListImplementationRevisionsForInterfaceOptions stores Implementation Revision filtering parameters.
type ListImplementationRevisionsForInterfaceOptions struct {
	attrFilter                     map[gqlpublicapi.FilterRule]map[string]*string
	implPathPattern                *string
	requirementsSatisfiedBy        map[gqlpublicapi.TypeReference]struct{}
	requiredTIInjectionSatisfiedBy map[gqlpublicapi.TypeReference]struct{}
	requires                       map[string]*string
	sortByPathAscAndRevisionDesc   bool
}

// Apply is used to configure the ListImplementationRevisionsForInterfaceOptions.
func (o *ListImplementationRevisionsForInterfaceOptions) Apply(opts ...ListImplementationRevisionsForInterfaceOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithFilter returns an options, which adds a filter for ImplementationRevisions.
func WithFilter(filter gqlpublicapi.ImplementationRevisionFilter) ListImplementationRevisionsForInterfaceOption {
	return func(opt *ListImplementationRevisionsForInterfaceOptions) {
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

		// 2. Process path pattern
		opt.implPathPattern = filter.PathPattern

		// 3. Process TypeInstances, which should satisfy requirements
		if len(filter.RequirementsSatisfiedBy) > 0 {
			opt.requirementsSatisfiedBy = make(map[gqlpublicapi.TypeReference]struct{})
			for _, req := range filter.RequirementsSatisfiedBy {
				if req.TypeRef == nil {
					continue
				}
				typeRef := gqlpublicapi.TypeReference{
					Path:     req.TypeRef.Path,
					Revision: req.TypeRef.Revision,
				}
				opt.requirementsSatisfiedBy[typeRef] = struct{}{}
			}
		}

		// 4. Process TypeInstances which should satisfy required TypeInstances injection
		if len(filter.RequiredTypeInstancesInjectionSatisfiedBy) > 0 {
			opt.requiredTIInjectionSatisfiedBy = make(map[gqlpublicapi.TypeReference]struct{})
			for _, req := range filter.RequiredTypeInstancesInjectionSatisfiedBy {
				if req.TypeRef == nil {
					continue
				}

				typeRef := gqlpublicapi.TypeReference{
					Path:     req.TypeRef.Path,
					Revision: req.TypeRef.Revision,
				}
				opt.requiredTIInjectionSatisfiedBy[typeRef] = struct{}{}

				// append to RequirementsSatisfiedBy as well
				opt.requirementsSatisfiedBy[typeRef] = struct{}{}
			}
		}

		// 5. Process TypeInstances, which should be defined in `requires` section
		if len(filter.Requires) > 0 {
			opt.requires = map[string]*string{}
			for _, req := range filter.Requires {
				if req == nil {
					continue
				}
				opt.requires[req.Path] = req.Revision
			}
		}
	}
}

// WithSortingByPathAscAndRevisionDesc returns an options, which ensures
// that the returned ImplementationRevision slice will be sorted
// in ascending order by the Implementation path
// and descending by the Implementation revision.
func WithSortingByPathAscAndRevisionDesc(options *ListImplementationRevisionsForInterfaceOptions) {
	options.sortByPathAscAndRevisionDesc = true
}
