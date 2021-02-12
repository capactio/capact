package public

import (
	"regexp"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

func filterImplementationRevisions(revs []gqlpublicapi.ImplementationRevision, filter *getImplementationOptions) []gqlpublicapi.ImplementationRevision {
	if filter == nil {
		return revs
	}

	revs = filterImplementationRevisionsByPattern(revs, filter.implPrefixPattern)
	revs = filterImplementationRevisionsByAttr(revs, filter.attrFilter)
	revs = filterImplementationRevisionsByRequirements(revs, filter.requirementsSatisfiedBy)

	return revs
}

func filterImplementationRevisionsByPattern(revs []gqlpublicapi.ImplementationRevision, pattern *string) []gqlpublicapi.ImplementationRevision {
	if pattern == nil {
		return revs
	}

	var out []gqlpublicapi.ImplementationRevision

	for _, impl := range revs {
		if impl.Metadata.Prefix == nil {
			continue
		}
		matched, err := regexp.Match(*pattern, []byte(*impl.Metadata.Prefix))
		if err != nil || !matched {
			continue
		}
		out = append(out, impl)
	}
	return out
}

func filterImplementationRevisionsByRequirements(revs []gqlpublicapi.ImplementationRevision, requirementsSatisfiedBy map[string]*string) []gqlpublicapi.ImplementationRevision {
	if len(requirementsSatisfiedBy) == 0 {
		return revs
	}

	var out []gqlpublicapi.ImplementationRevision

requirements:
	for _, impl := range revs {
		if impl.Spec == nil || len(impl.Spec.Requires) == 0 {
			out = append(out, impl)
			continue
		}

		for _, req := range impl.Spec.Requires {
			satisfied := allRequirementsAreSatisfied(req.AllOf, requirementsSatisfiedBy)
			if !satisfied {
				continue requirements
			}

			atLeastOneSatisfied := atLeastOneRequirementIsSatisfied(req.AnyOf, requirementsSatisfiedBy)
			if !atLeastOneSatisfied {
				continue requirements
			}

			onlyOneSatisfied := onlyOneRequirementIsSatisfied(req.OneOf, requirementsSatisfiedBy)
			if !onlyOneSatisfied {
				continue requirements
			}
		}

		out = append(out, impl)
	}
	return out
}

func onlyOneRequirementIsSatisfied(implReq []*gqlpublicapi.ImplementationRequirementItem, availableReq map[string]*string) bool {
	if len(implReq) == 0 {
		return true
	}

	satisfiedCnt := 0
	for _, all := range implReq {
		if all.TypeRef == nil {
			continue
		}
		satisfied := contains(availableReq, all.TypeRef.Path, all.TypeRef.Revision)
		if satisfied {
			satisfiedCnt++
		}
		if satisfiedCnt > 1 {
			return false
		}
	}

	return satisfiedCnt == 1
}

func atLeastOneRequirementIsSatisfied(implReq []*gqlpublicapi.ImplementationRequirementItem, availableReq map[string]*string) bool {
	if len(implReq) == 0 {
		return true
	}

	for _, all := range implReq {
		if all.TypeRef == nil {
			continue
		}
		satisfied := contains(availableReq, all.TypeRef.Path, all.TypeRef.Revision)
		if satisfied {
			return true
		}
	}
	return false
}

func allRequirementsAreSatisfied(implReq []*gqlpublicapi.ImplementationRequirementItem, availableReq map[string]*string) bool {
	if len(implReq) == 0 {
		return true
	}

	for _, all := range implReq {
		if all.TypeRef == nil {
			continue
		}
		satisfied := contains(availableReq, all.TypeRef.Path, all.TypeRef.Revision)
		if !satisfied { // all needs to be satisfied so we can already give up
			return false
		}
	}
	return true
}

func filterImplementationRevisionsByAttr(revs []gqlpublicapi.ImplementationRevision, attrFilter map[gqlpublicapi.FilterRule]map[string]*string) []gqlpublicapi.ImplementationRevision {
	includedAttr := attrFilter[gqlpublicapi.FilterRuleInclude]
	excludedAttr := attrFilter[gqlpublicapi.FilterRuleExclude]

	if len(includedAttr) == 0 && len(excludedAttr) == 0 {
		return revs
	}

	var out []gqlpublicapi.ImplementationRevision

	for _, impl := range revs {
		isExclude := containsAtLeastOne(impl.Metadata.Attributes, excludedAttr)
		if isExclude {
			continue
		}

		isIncluded := containsAll(impl.Metadata.Attributes, includedAttr)
		if !isIncluded {
			continue
		}
		out = append(out, impl)
	}

	return out
}

//  contains returns true if all items from expAtr are defined in implAtr. Duplicates are skipped.
func containsAtLeastOne(attr []*gqlpublicapi.AttributeRevision, expAttr map[string]*string) bool {
	for _, atr := range attr {
		if atr == nil || atr.Metadata == nil || atr.Metadata.Path == nil {
			continue
		}

		if contains(expAttr, *atr.Metadata.Path, atr.Revision) {
			return true
		}
	}

	return false
}

//  contains returns true if all items from expAtr are defined in implAtr. Duplicates are skipped.
func containsAll(attr []*gqlpublicapi.AttributeRevision, expAttr map[string]*string) bool {
	matchedEntities := 0
	for _, atr := range attr {
		if atr == nil || atr.Metadata == nil || atr.Metadata.Path == nil {
			continue
		}

		if contains(expAttr, *atr.Metadata.Path, atr.Revision) {
			matchedEntities++
		}
	}

	return len(expAttr) == matchedEntities
}

func contains(attr map[string]*string, path, rev string) bool {
	revision, found := attr[path]
	if !found {
		return false
	}

	if revision != nil && *revision != rev {
		return false
	}

	return true
}
