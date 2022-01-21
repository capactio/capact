package dbpopulator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupManifestsMerging(t *testing.T) {
	tests := map[string]struct {
		groupA GroupManifests
		groupB GroupManifests

		expResult GroupManifests
	}{
		"Merge empty data": {
			groupA:    GroupManifests{},
			groupB:    GroupManifests{},
			expResult: GroupManifests{},
		},
		"Merge with same groups names": {
			groupA: GroupManifests{
				"group-1": manifestPathFixtures("foo"),
			},
			groupB: GroupManifests{
				"group-1": manifestPathFixtures("bar"),
			},
			expResult: GroupManifests{
				"group-1": manifestPathFixtures("foo", "bar"),
			},
		},
		"Merge with different groups names": {
			groupA: GroupManifests{
				"group-1": manifestPathFixtures("foo"),
			},
			groupB: GroupManifests{
				"group-2": manifestPathFixtures("bar"),
			},
			expResult: GroupManifests{
				"group-1": manifestPathFixtures("foo"),
				"group-2": manifestPathFixtures("bar"),
			},
		},
		"Merge with mixed groups names": {
			groupA: GroupManifests{
				"group-1": manifestPathFixtures("foo"),
				"group-2": manifestPathFixtures("bar"),
			},
			groupB: GroupManifests{
				"group-1": manifestPathFixtures("baz"),
				"group-3": manifestPathFixtures("xyz"),
			},
			expResult: GroupManifests{
				"group-1": manifestPathFixtures("foo", "baz"),
				"group-2": manifestPathFixtures("bar"),
				"group-3": manifestPathFixtures("xyz"),
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			gotMergedGroupedManifests := GroupManifests{}

			// when
			gotMergedGroupedManifests.MergeWith(tc.groupA)
			gotMergedGroupedManifests.MergeWith(tc.groupB)

			// then
			assert.Equal(t, tc.expResult, gotMergedGroupedManifests)
		})
	}
}

func manifestPathFixtures(names ...string) []manifestPath {
	var out []manifestPath
	for _, n := range names {
		out = append(out, manifestPath{
			path:   n,
			prefix: fmt.Sprintf("fix-prefix-%s", n),
		})
	}

	return out
}
