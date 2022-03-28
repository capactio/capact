//go:build localhub
// +build localhub

package localhub

import (
	"context"
	"os"
	"regexp"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFindAndDeleteTypeInstances(t *testing.T) {
	ctx := context.Background()
	cli := getLocalClient(t)
	builtinStorage := getBuiltinStorageTypeInstance(ctx, t, cli)

	createdTypeInstanceID, err := cli.CreateTypeInstance(ctx, &gqllocalapi.CreateTypeInstanceInput{
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.capactio.capact.validation.single-key",
			Revision: "0.1.0",
		},
		Attributes: []*gqllocalapi.AttributeReferenceInput{
			{
				Path:     "cap.type.capactio.capact.attribute1",
				Revision: "0.1.0",
			},
		},
		Value: map[string]interface{}{
			"key": "bar",
		},
	})
	require.NoError(t, err)

	typeInstance, err := cli.FindTypeInstance(ctx, createdTypeInstanceID)
	require.NoError(t, err)

	rev := &gqllocalapi.TypeInstanceResourceVersion{
		ResourceVersion: 1,
		Metadata: &gqllocalapi.TypeInstanceResourceVersionMetadata{
			Attributes: []*gqllocalapi.AttributeReference{
				{
					Path:     "cap.type.capactio.capact.attribute1",
					Revision: "0.1.0",
				},
			},
		},
		Spec: &gqllocalapi.TypeInstanceResourceVersionSpec{
			Value: map[string]interface{}{
				"key": "bar",
			},
			Backend: &gqllocalapi.TypeInstanceResourceVersionSpecBackend{},
		},
	}

	assert.Equal(t, typeInstance, &gqllocalapi.TypeInstance{
		ID: createdTypeInstanceID,
		TypeRef: &gqllocalapi.TypeInstanceTypeReference{
			Path:     "cap.type.capactio.capact.validation.single-key",
			Revision: "0.1.0",
		},
		Backend: &gqllocalapi.TypeInstanceBackendReference{
			ID:       builtinStorage.ID,
			Abstract: true,
		},
		Uses:                    []*gqllocalapi.TypeInstance{&builtinStorage},
		UsedBy:                  []*gqllocalapi.TypeInstance{},
		LatestResourceVersion:   rev,
		FirstResourceVersion:    rev,
		PreviousResourceVersion: nil,
		ResourceVersion:         rev,
		ResourceVersions:        []*gqllocalapi.TypeInstanceResourceVersion{rev},
	})

	err = cli.DeleteTypeInstance(ctx, createdTypeInstanceID)
	require.NoError(t, err)

	got, err := cli.FindTypeInstance(ctx, createdTypeInstanceID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestCreatesMultipleTypeInstancesWithUsesRelations(t *testing.T) {
	ctx := context.Background()
	cli := getLocalClient(t)
	builtinStorage := getBuiltinStorageTypeInstance(ctx, t, cli)

	createdTypeInstanceIDs, err := cli.CreateTypeInstances(ctx, createTypeInstancesInputForUsesRelations())
	require.NoError(t, err)
	defer func() {
		var child string
		for _, ti := range createdTypeInstanceIDs {
			if ti.Alias == "child" {
				child = ti.ID
				continue
			}
			// Delete parent first, child TypeInstance is protected as it is used by parent.
			deleteTypeInstance(ctx, t, cli, ti.ID)
		}

		// Delete child TypeInstance as there are no "users".
		deleteTypeInstance(ctx, t, cli, child)
	}()

	parentTiID := findCreatedTypeInstanceID("parent", createdTypeInstanceIDs)
	assert.NotNil(t, parentTiID)

	childTiID := findCreatedTypeInstanceID("child", createdTypeInstanceIDs)
	assert.NotNil(t, childTiID)

	expectedChild := expectedChildTypeInstance(*childTiID, builtinStorage.ID)
	expectedParent := expectedParentTypeInstance(*parentTiID, builtinStorage.ID)
	expectedChild.UsedBy = []*gqllocalapi.TypeInstance{expectedParentTypeInstance(*parentTiID, builtinStorage.ID)}
	expectedChild.Uses = []*gqllocalapi.TypeInstance{&builtinStorage}
	expectedParent.Uses = []*gqllocalapi.TypeInstance{&builtinStorage, expectedChildTypeInstance(*childTiID, builtinStorage.ID)}
	expectedParent.UsedBy = []*gqllocalapi.TypeInstance{}

	assertTypeInstance(ctx, t, cli, *childTiID, expectedChild)
	assertTypeInstance(ctx, t, cli, *parentTiID, expectedParent)
}
func TestUpdateTypeInstances(t *testing.T) {
	const (
		fooOwnerID = "namespace/Foo"
		barOwnerID = "namespace/Bar"
	)
	ctx := context.Background()
	cli := getLocalClient(t)

	var createdTIIDs []string

	t.Log("given id1 and id2 are not locked")
	for _, ver := range []string{"id1", "id2"} {
		outID, err := cli.CreateTypeInstance(ctx, typeInstance(ver))
		require.NoError(t, err)
		createdTIIDs = append(createdTIIDs, outID)
	}
	defer func() {
		for _, id := range createdTIIDs {
			_ = cli.DeleteTypeInstance(ctx, id)
		}
	}()

	expUpdateTI := &gqllocalapi.UpdateTypeInstanceInput{
		Attributes: []*gqllocalapi.AttributeReferenceInput{
			{Path: "cap.update.not.locked", Revision: "0.0.1"},
		},
	}

	t.Log("when try to update them")
	updatedTI, err := cli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
		{
			ID:           createdTIIDs[0],
			TypeInstance: expUpdateTI,
		},
		{
			ID:           createdTIIDs[1],
			TypeInstance: expUpdateTI,
		},
	})

	t.Log("then should success")
	require.NoError(t, err)
	for _, instance := range updatedTI {
		assert.Equal(t, len(instance.LatestResourceVersion.Metadata.Attributes), 1)
		assert.EqualValues(t, instance.LatestResourceVersion.Metadata.Attributes[0], expUpdateTI.Attributes[0])
	}

	t.Log("when id1 and id2 are locked by Foo")
	expUpdateTI = &gqllocalapi.UpdateTypeInstanceInput{
		Attributes: []*gqllocalapi.AttributeReferenceInput{
			{Path: "cap.update.locked.by.foo", Revision: "0.0.1"},
		},
	}

	t.Log("then should success")
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     createdTIIDs,
		OwnerID: fooOwnerID,
	})
	require.NoError(t, err)

	t.Log("when update them as Foo owner")
	updatedTI, err = cli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
		{
			ID:           createdTIIDs[0],
			OwnerID:      ptr.String(fooOwnerID),
			TypeInstance: expUpdateTI,
		},
		{
			ID:           createdTIIDs[1],
			OwnerID:      ptr.String(fooOwnerID),
			TypeInstance: expUpdateTI,
		},
	})

	t.Log("then should success")
	require.NoError(t, err)
	for _, instance := range updatedTI {
		assert.Equal(t, len(instance.LatestResourceVersion.Metadata.Attributes), 1)
		assert.EqualValues(t, instance.LatestResourceVersion.Metadata.Attributes[0], expUpdateTI.Attributes[0])
	}

	t.Log("when update them as Bar owner")
	_, err = cli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
		{
			ID:           createdTIIDs[0],
			OwnerID:      ptr.String(barOwnerID),
			TypeInstance: expUpdateTI,
		},
		{
			ID:           createdTIIDs[1],
			OwnerID:      ptr.String(barOwnerID),
			TypeInstance: expUpdateTI,
		},
	})

	t.Log("then should failed with error id1,id2 already locked by different owner")
	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(heredoc.Docf(`while executing mutation to update TypeInstances: All attempts fail:
	#1: graphql: failed to update TypeInstances: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))), err.Error())

	t.Log("when update them without owner")
	_, err = cli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
		{
			ID:           createdTIIDs[0],
			TypeInstance: expUpdateTI,
		},
		{
			ID:           createdTIIDs[1],
			TypeInstance: expUpdateTI,
		},
	})

	t.Log("then should failed with error id1,id2 already locked by different owner")
	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(heredoc.Docf(`while executing mutation to update TypeInstances: All attempts fail:
	#1: graphql: failed to update TypeInstances: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))), err.Error())

	t.Log("when update one property with Foo owner, and second without owner")
	_, err = cli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
		{
			ID:           createdTIIDs[0],
			OwnerID:      ptr.String(fooOwnerID),
			TypeInstance: expUpdateTI,
		},
		{
			ID:           createdTIIDs[1],
			TypeInstance: expUpdateTI,
		},
	})

	t.Log("then should failed with error id2 already locked by different owner")
	require.Error(t, err)
	require.Equal(t, err.Error(), heredoc.Docf(`while executing mutation to update TypeInstances: All attempts fail:
	     				#1: graphql: failed to update TypeInstances: TypeInstances with IDs "%s" are locked by different owner`, createdTIIDs[1]))

	t.Log("given id3 does not exist")
	t.Log("when try to update it")
	_, err = cli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
		{
			ID:           "id3",
			TypeInstance: expUpdateTI,
		},
	})

	t.Log("then should failed with error id3 not found")
	require.Error(t, err)
	require.Equal(t, err.Error(), heredoc.Doc(`while executing mutation to update TypeInstances: All attempts fail:
	     			#1: graphql: failed to update TypeInstances: TypeInstances with IDs "id3" were not found`))

	t.Log("then should unlock id1,id2,id3")
	err = cli.UnlockTypeInstances(ctx, &gqllocalapi.UnlockTypeInstancesInput{
		Ids:     createdTIIDs,
		OwnerID: fooOwnerID,
	})
	require.NoError(t, err)
}

func getLocalClient(t *testing.T) *local.Client {
	localhubAddrEnv := "LOCALHUB_ADDR"
	localhubAddr := os.Getenv(localhubAddrEnv)
	if localhubAddr == "" {
		t.Skipf("skipping running example test as the env %s is not provided", localhubAddrEnv)
	}
	return local.NewDefaultClient(localhubAddr)
}

func getBuiltinStorageTypeInstance(ctx context.Context, t *testing.T, cli *local.Client) gqllocalapi.TypeInstance {
	coreStorage, err := cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		TypeRef: &gqllocalapi.TypeRefFilterInput{
			Path:     "cap.core.type.hub.storage.neo4j",
			Revision: ptr.String("0.1.0"),
		},
	}, local.WithFields(local.TypeInstanceAllFields))
	require.NoError(t, err)
	assert.Equal(t, len(coreStorage), 1)

	return coreStorage[0]
}

func createTypeInstancesInputForUsesRelations() *gqllocalapi.CreateTypeInstancesInput {
	return &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: []*gqllocalapi.CreateTypeInstanceInput{
			{
				Alias: ptr.String("parent"),
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
					Path:     "com.parent",
					Revision: "0.1.0",
				},
				Attributes: []*gqllocalapi.AttributeReferenceInput{
					{
						Path:     "com.attr",
						Revision: "0.1.0",
					},
				},
				Value: map[string]interface{}{
					"parent": true,
				},
			},
			{
				Alias: ptr.String("child"),
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
					Path:     "com.child",
					Revision: "0.1.0",
				},
				Attributes: []*gqllocalapi.AttributeReferenceInput{
					{
						Path:     "com.attr",
						Revision: "0.1.0",
					},
				},
				Value: map[string]interface{}{
					"child": true,
				},
			},
		},
		UsesRelations: []*gqllocalapi.TypeInstanceUsesRelationInput{
			{
				From: "parent",
				To:   "child",
			},
		},
	}
}

func deleteTypeInstance(ctx context.Context, t *testing.T, cli *local.Client, ID string) {
	err := cli.DeleteTypeInstance(ctx, ID)
	require.NoError(t, err)
}

func findCreatedTypeInstanceID(alias string, instances []gqllocalapi.CreateTypeInstanceOutput) *string {
	for _, el := range instances {
		if el.Alias != alias {
			continue
		}
		return &el.ID
	}

	return nil
}

func expectedChildTypeInstance(tiID, backendID string) *gqllocalapi.TypeInstance {
	tiRev := &gqllocalapi.TypeInstanceResourceVersion{
		ResourceVersion: 1,
		Metadata: &gqllocalapi.TypeInstanceResourceVersionMetadata{
			Attributes: []*gqllocalapi.AttributeReference{
				{
					Path:     "com.attr",
					Revision: "0.1.0",
				},
			},
		},
		Spec: &gqllocalapi.TypeInstanceResourceVersionSpec{
			Value: map[string]interface{}{
				"child": true,
			},
			Backend: &gqllocalapi.TypeInstanceResourceVersionSpecBackend{},
		},
	}

	return &gqllocalapi.TypeInstance{
		ID: tiID,
		TypeRef: &gqllocalapi.TypeInstanceTypeReference{
			Path:     "com.child",
			Revision: "0.1.0",
		},

		Backend: &gqllocalapi.TypeInstanceBackendReference{
			ID:       backendID,
			Abstract: true,
		},
		LatestResourceVersion:   tiRev,
		FirstResourceVersion:    tiRev,
		PreviousResourceVersion: nil,
		ResourceVersion:         tiRev,
		ResourceVersions:        []*gqllocalapi.TypeInstanceResourceVersion{tiRev},
		UsedBy:                  nil,
		Uses:                    nil,
	}
}

func expectedParentTypeInstance(tiID, backendID string) *gqllocalapi.TypeInstance {
	tiRev := &gqllocalapi.TypeInstanceResourceVersion{
		ResourceVersion: 1,
		Metadata: &gqllocalapi.TypeInstanceResourceVersionMetadata{
			Attributes: []*gqllocalapi.AttributeReference{
				{
					Path:     "com.attr",
					Revision: "0.1.0",
				},
			},
		},
		Spec: &gqllocalapi.TypeInstanceResourceVersionSpec{
			Value: map[string]interface{}{
				"parent": true,
			},
			Backend: &gqllocalapi.TypeInstanceResourceVersionSpecBackend{},
		},
	}

	return &gqllocalapi.TypeInstance{
		ID: tiID,
		TypeRef: &gqllocalapi.TypeInstanceTypeReference{
			Path:     "com.parent",
			Revision: "0.1.0",
		},

		Backend: &gqllocalapi.TypeInstanceBackendReference{
			ID:       backendID,
			Abstract: true,
		},
		LatestResourceVersion:   tiRev,
		FirstResourceVersion:    tiRev,
		PreviousResourceVersion: nil,
		ResourceVersion:         tiRev,
		ResourceVersions:        []*gqllocalapi.TypeInstanceResourceVersion{tiRev},
		UsedBy:                  nil,
		Uses:                    nil,
	}
}

func assertTypeInstance(ctx context.Context, t *testing.T, cli *local.Client, ID string, expected *gqllocalapi.TypeInstance) {
	actual, err := cli.FindTypeInstance(ctx, ID)
	require.NoError(t, err)
	assert.NotNil(t, actual)
	assert.NotNil(t, expected)
	assert.Equal(t, *actual, *expected)
}
