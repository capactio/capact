package public

import (
	"sort"

	"github.com/Masterminds/semver/v3"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

func SortImplementationRevisions(revs []gqlpublicapi.ImplementationRevision, opts *GetImplementationOptions) []gqlpublicapi.ImplementationRevision {
	if opts == nil {
		return revs
	}

	revs = sortImplementationRevisionsByPathAscAndRevisionDesc(revs, opts.sortByPathAscAndRevisionDesc)

	return revs
}

// sortImplementationRevisionsByPathAscAndRevisionDesc sorts by Path ascending, and Revision descending.
func sortImplementationRevisionsByPathAscAndRevisionDesc(revs []gqlpublicapi.ImplementationRevision, shouldSort bool) []gqlpublicapi.ImplementationRevision {
	if !shouldSort {
		return revs
	}

	sort.Slice(revs, func(i, j int) bool {
		if revs[i].Metadata == nil {
			return false
		}

		if revs[j].Metadata == nil {
			return true
		}

		if revs[i].Metadata.Path < revs[j].Metadata.Path {
			return true
		}

		if revs[i].Metadata.Path > revs[j].Metadata.Path {
			return false
		}

		vi, erri := semver.NewVersion(revs[i].Revision)
		vj, errj := semver.NewVersion(revs[j].Revision)
		if erri != nil || errj != nil {
			// fallback to string comparison
			return revs[i].Revision > revs[j].Revision
		}

		return vi.GreaterThan(vj)
	})

	return revs
}
