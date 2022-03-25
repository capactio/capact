package localhub

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/regexutil"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	prmt "github.com/gitchander/permutation"
)

func TestLockTypeInstances(t *testing.T) {
	const (
		fooOwnerID = "namespace/Foo"
		barOwnerID = "namespace/Bar"
	)
	ctx := context.Background()
	cli := getLocalClient(t)

	var createdTIIDs []string

	for _, ver := range []string{"id1", "id2", "id3"} {
		outID, err := cli.CreateTypeInstance(ctx, typeInstance(ver))
		require.NoError(t, err)
		createdTIIDs = append(createdTIIDs, outID)
	}
	defer func() {
		for _, id := range createdTIIDs {
			_ = cli.DeleteTypeInstance(ctx, id)
		}
	}()

	t.Log("given id1 and id2 are not locked")
	firstTwoInstances := createdTIIDs[:2]
	lastInstances := createdTIIDs[2:]

	t.Log("when Foo tries to locks them")
	err := cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     firstTwoInstances,
		OwnerID: fooOwnerID,
	})

	t.Log("then should success")
	require.NoError(t, err)

	got, err := cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{})
	require.NoError(t, err)

	for _, instance := range got {
		if includes(firstTwoInstances, instance.ID) {
			assert.NotNil(t, instance.LockedBy)
			assert.Equal(t, *instance.LockedBy, fooOwnerID)
		} else if includes(lastInstances, instance.ID) {
			assert.Nil(t, instance.LockedBy)
		}
	}

	t.Log("when Foo tries to locks them")
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     createdTIIDs, // lock all 3 instances, when the first two are already locked
		OwnerID: fooOwnerID,
	})
	require.NoError(t, err)

	t.Log("then should success")
	got, err = cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{})
	require.NoError(t, err)

	for _, instance := range got {
		if !includes(createdTIIDs, instance.ID) {
			continue
		}
		assert.NotNil(t, instance.LockedBy)
		assert.Equal(t, *instance.LockedBy, fooOwnerID)
	}

	t.Log("given id1, id2, id3 are locked by Foo, id4: not found")
	lockingIDs := createdTIIDs
	lockingIDs = append(lockingIDs, "123-not-found")

	t.Log("when Foo tries to locks id1,id2,id3,id4")
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     lockingIDs,
		OwnerID: fooOwnerID,
	})

	t.Log("then should failed with id4 not found error")
	require.Error(t, err)
	require.Equal(t, err.Error(), heredoc.Doc(`while executing mutation to lock TypeInstances: All attempts fail:
	#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs "123-not-found" were not found`))

	t.Log("when Bar tries to locks id1,id2,id3,id4")
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     lockingIDs,
		OwnerID: barOwnerID,
	})

	t.Log("then should failed with id4 not found and already locked error for id1,id2,id3")
	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
	#1: graphql: failed to lock TypeInstances: 2 errors occurred: \[TypeInstances with IDs "123-not-found" were not found, TypeInstances with IDs %s are locked by different owner\]`, allPermutations(createdTIIDs))), err.Error())

	t.Log("given id1, id2, id3 are locked by Foo, id4: not locked")
	id4, err := cli.CreateTypeInstance(ctx, typeInstance("id4"))
	require.NoError(t, err)

	defer func() {
		_ = cli.DeleteTypeInstance(ctx, id4)
	}()

	t.Log("when Bar tries to locks all of them")
	lockingIDs = createdTIIDs
	lockingIDs = append(lockingIDs, id4)
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     lockingIDs,
		OwnerID: barOwnerID,
	})

	t.Log("then should failed with error id1,id2,id3 already locked by Foo")
	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
	#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))), err.Error())

	t.Log("should unlock id1,id2,id3")
	err = cli.UnlockTypeInstances(ctx, &gqllocalapi.UnlockTypeInstancesInput{
		Ids:     createdTIIDs,
		OwnerID: fooOwnerID,
	})
	require.NoError(t, err)
}

func typeInstance(ver string) *gqllocalapi.CreateTypeInstanceInput {
	return &gqllocalapi.CreateTypeInstanceInput{
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.capactio.capact.validation.single-key",
			Revision: "0.1.0",
		},
		Attributes: []*gqllocalapi.AttributeReferenceInput{
			{
				Path:     "cap.type.sample-v" + ver,
				Revision: "0.1.0",
			},
		},
		Value: map[string]interface{}{
			"key": "sample-v" + ver,
		},
	}
}

func includes(ids []string, expID string) bool {
	for _, i := range ids {
		if i == expID {
			return true
		}
	}

	return false
}

// allPermutations returns all possible permutations in regex format.
// For such input
//	a := []string{"alpha", "beta"}
// returns
//	("alpha", "beta"|"beta", "alpha")
//
// This function allows you to match list of words in any order using regex.
func allPermutations(in []string) string {
	p := prmt.New(prmt.StringSlice(in))
	var opts []string
	for p.Next() {
		opts = append(opts, fmt.Sprintf(`"%s"`, strings.Join(in, `", "`)))
	}
	return regexutil.OrStringSlice(opts)
}
