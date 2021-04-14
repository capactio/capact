package public

import (
	"sort"

	gqlpublicapi "capact.io/capact/pkg/och/api/graphql/public"
	"github.com/Masterminds/semver/v3"
)

func SortImplementationRevisions(revs []gqlpublicapi.ImplementationRevision, opts *ListImplementationRevisionsOptions) []gqlpublicapi.ImplementationRevision {
	if opts == nil {
		return revs
	}

	if opts.sortByPathAscAndRevisionDesc {
		sort.Sort(implRevsByPathAscAndRevisionDesc(revs))
	}

	return revs
}

type implRevsByPathAscAndRevisionDesc []gqlpublicapi.ImplementationRevision

func (revs implRevsByPathAscAndRevisionDesc) Len() int {
	return len(revs)
}

func (revs implRevsByPathAscAndRevisionDesc) Swap(i, j int) {
	revs[i], revs[j] = revs[j], revs[i]
}

func (revs implRevsByPathAscAndRevisionDesc) Less(i, j int) bool {
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
}
