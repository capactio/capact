// +build example

package test

import (
	"context"
	"github.com/MakeNowJust/heredoc"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"projectvoltron.dev/voltron/pkg/httputil"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	"projectvoltron.dev/voltron/pkg/och/client/local"
	"strings"
	"testing"
	"time"
)

func TestLockTypeInstances(t *testing.T) {
	// given
	ctx := context.Background()
	fooOwnerID := "namespace/Foo"
	barOwnerID := "namespace/Bar"
	localCli := NewOCHLocalClient("http://localhost:8080/graphql")

	var createdTIIDs []string
	for _, ver := range []string{"id1", "id2", "id3"} {
		out, err := localCli.CreateTypeInstance(ctx, typeInstance(ver))
		require.NoError(t, err)
		createdTIIDs = append(createdTIIDs, out.ID)
	}

	firstTwoInstances := createdTIIDs[:2]
	defer func() {
		for _, id := range createdTIIDs {
			_ = localCli.DeleteTypeInstance(ctx, id)
		}
	}()

	t.Run("Scenario: id1, id2: not locked",
		func(t *testing.T) {
			// when Foo tries to locks them
			err := localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     firstTwoInstances,
				OwnerID: fooOwnerID,
			})
			require.NoError(t, err)

			// then success
			got, err := localCli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{})
			require.NoError(t, err)

			for _, instance := range got {
				if includes(firstTwoInstances, instance.ID) {
					require.NotNil(t, instance.LockedBy)
					require.Equal(t, fooOwnerID, *instance.LockedBy)
				} else {
					require.Nil(t, instance.LockedBy)
				}
			}
		})

	t.Run("Scenario: id1, id2 locked by Foo, id3: not locked",
		func(t *testing.T) {
			// when Foo tries to locks them
			err := localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     createdTIIDs, // lock all 3 instances, when the first two are already locked
				OwnerID: fooOwnerID,
			})
			require.NoError(t, err)

			// then success
			got, err := localCli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{})
			require.NoError(t, err)

			for _, instance := range got {
				require.NotNil(t, instance.LockedBy)
				require.Equal(t, fooOwnerID, *instance.LockedBy)
			}
		})

	t.Run("Scenario: id1, id2, id3 locked by Foo, id4: not found",
		func(t *testing.T) {
			// when Foo tries to locks id1,id2,id3,id4
			lockingIDs := createdTIIDs
			lockingIDs = append(lockingIDs, "123-not-found")
			err := localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     lockingIDs,
				OwnerID: fooOwnerID,
			})

			// then failure - reason: id4 not found
			require.Error(t, err)
			require.Contains(t, err.Error(), heredoc.Doc(`while executing mutation to lock TypeInstances: All attempts fail:
							#1: graphql: 1 error occurred: TypeInstances with IDs 123-not-found were not found`))
		})

	t.Run("Scenario: id1, id2, id3 are locked by Foo, id4: not found",
		func(t *testing.T) {
			// when Bar tries to locks id1,id2,id3,id4
			lockingIDs := createdTIIDs
			lockingIDs = append(lockingIDs, "123-not-found")
			err := localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     lockingIDs,
				OwnerID: barOwnerID,
			})

			// then failure - reason: id4 not found and already locked by Foo
			require.Error(t, err)
			require.Contains(t, err.Error(), heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
				#1: graphql: 2 errors occurred: [TypeInstances with IDs 123-not-found were not found, TypeInstances with IDs %s are already locked by other owner]`, strings.Join(createdTIIDs, ", ")))

		})

	t.Run("Scenario: id1, id2, id3 are locked by Foo, id4: not locked",
		func(t *testing.T) {
			// when Bar tries to locks id1,id2,id3,id4
			id4, err := localCli.CreateTypeInstance(ctx, typeInstance("id4"))
			require.NoError(t, err)
			defer localCli.DeleteTypeInstance(ctx, id4.ID)

			lockingIDs := createdTIIDs
			lockingIDs = append(lockingIDs, id4.ID)
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     lockingIDs,
				OwnerID: barOwnerID,
			})

			// then failure - reason: id1,id2,id3 already locked by Foo
			require.Error(t, err)
			require.Contains(t, err.Error(), heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
				#1: graphql: 1 error occurred: TypeInstances with IDs %s are already locked by other owner`, strings.Join(createdTIIDs, ", ")))

		})
}

func includes(ids []string, expID string) bool {
	for _, i := range ids {
		if i == expID {
			return true
		}
	}
	return false
}

func typeInstance(ver string) *gqllocalapi.CreateTypeInstanceInput {
	return &gqllocalapi.CreateTypeInstanceInput{
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.sample-v" + ver,
			Revision: "0.1.0",
		},
		Attributes: []*gqllocalapi.AttributeReferenceInput{
			{
				Path:     "cap.type.sample-v" + ver,
				Revision: "0.1.0",
			},
		},
		Value: map[string]interface{}{
			"sample-v" + ver: "true",
		},
	}
}

func NewOCHLocalClient(endpoint string) *local.Client {
	httpClient := httputil.NewClient(
		30*time.Second,
		true,
	)

	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return local.NewClient(client)
}
