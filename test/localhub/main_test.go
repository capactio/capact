package localhub

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	"capact.io/capact/internal/regexutil"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"

	prmt "github.com/gitchander/permutation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	typeInstanceTypeRef = "cap.type.testing:0.1.0"
	hubJSBackendAddr    = "HUBJS_ADDR"
)

type StorageSpec struct {
	URL           *string `json:"url,omitempty"`
	AcceptValue   *bool   `json:"acceptValue,omitempty"`
	ContextSchema *string `json:"contextSchema,omitempty"`
}

func TestExternalStorageInputValidation(t *testing.T) {
	hubJSAddr := os.Getenv(hubJSBackendAddr)
	if hubJSAddr == "" {
		t.Skipf("skipping running example test as the env %s is not provided", hubJSAddr)
	}
	ctx := context.Background()
	cli := local.NewDefaultClient(hubJSAddr)

	tests := map[string]struct {
		// given
		storageSpec interface{}
		value       map[string]interface{}
		context     interface{}

		// then
		expErrMsg string
	}{
		"Should rejected value": {
			storageSpec: StorageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": input value not allowed
              	* Error: rollback externally stored values: External backend "MOCKED_ID": input value not allowed`),
		},
		"Should rejected context": {
			storageSpec: StorageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
			},
			context: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": input context not allowed
              	* Error: rollback externally stored values: External backend "MOCKED_ID": input context not allowed`),
		},
		"Should return error that context is not an object": {
			storageSpec: StorageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
				ContextSchema: ptr.String(heredoc.Doc(`
					   {
					   	"$id": "#/properties/contextSchema",
					   	"type": "object",
					   	"properties": {
					   		"provider": {
					   			"$id": "#/properties/contextSchema/properties/name",
					   			"type": "string"
					   		}
					   	},
					   	"additionalProperties": false
					   }`)),
			},
			context: "Luke Skywalker",
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": invalid input: context must be object
              	* Error: rollback externally stored values: External backend "MOCKED_ID": invalid input: context must be object`),
		},
		"Should return validation error for context": {
			storageSpec: StorageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
				ContextSchema: ptr.String(heredoc.Doc(`
					   {
					   	"$id": "#/properties/contextSchema",
					   	"type": "object",
					   	"properties": {
					   		"provider": {
					   			"$id": "#/properties/contextSchema/properties/name",
					   			"type": "string"
					   		}
					   	},
					   	"additionalProperties": false
					   }`)),
			},
			context: map[string]interface{}{
				"provider": true,
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": invalid input: context/provider must be string
              	* Error: rollback externally stored values: External backend "MOCKED_ID": invalid input: context/provider must be string`),
		},
		"Should reject value and context": {
			storageSpec: StorageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			context: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			// TODO(review): currently it's a an early return, is it sufficient?
			// if not, we will need to an support for throwing multierr to print aggregated data in higher layer.
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: External backend "MOCKED_ID": input value not allowed
              	* Error: rollback externally stored values: External backend "MOCKED_ID": input value not allowed`),
		},

		// Invalid Storage TypeInstance
		"Should reject usage of backend without URL field": {
			storageSpec: StorageSpec{
				AcceptValue: ptr.Bool(false),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'url'
              	* Error: rollback externally stored values: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'url'`),
		},
		"Should reject usage of backend without AcceptValue field": {
			storageSpec: StorageSpec{
				URL: ptr.String("http://localhost:5000/fake"),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'acceptValue'
              	* Error: rollback externally stored values: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'acceptValue'`),
		},
		"Should reject usage of backend without URL and AcceptValue fields": {
			storageSpec: map[string]interface{}{
				"other-data": true,
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'url', spec.value must have required property 'acceptValue'
              	* Error: rollback externally stored values: failed to resolve the TypeInstance's backend "MOCKED_ID": spec.value must have required property 'url', spec.value must have required property 'acceptValue'`),
		},
		"Should reject usage of backend with wrong context schema": {
			storageSpec: StorageSpec{
				URL:         ptr.String("http://localhost:5000/fake"),
				AcceptValue: ptr.Bool(false),
				ContextSchema: ptr.String(heredoc.Doc(`
					   yaml: true`)),
			},
			value: map[string]interface{}{
				"name": "Luke Skywalker",
			},
			expErrMsg: heredoc.Doc(`
              while executing mutation to create TypeInstance: All attempts fail:
              #1: graphql: failed to create TypeInstance: failed to create the TypeInstances: 2 error occurred:
              	* Error: failed to process the TypeInstance's backend "MOCKED_ID": invalid spec.context: Unexpected token y in JSON at position 0
              	* Error: rollback externally stored values: failed to process the TypeInstance's backend "MOCKED_ID": invalid spec.context: Unexpected token y in JSON at position 0`),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			externalStorageID, cleanup := registerExternalStorage(ctx, t, cli, tc.storageSpec)
			defer cleanup()

			// when
			_, err := cli.CreateTypeInstance(ctx, &gqllocalapi.CreateTypeInstanceInput{
				TypeRef: typeRef(typeInstanceTypeRef),
				Value:   tc.value,
				Backend: &gqllocalapi.TypeInstanceBackendInput{
					ID:      externalStorageID,
					Context: tc.context,
				},
			})

			require.Error(t, err)

			regex := regexp.MustCompile(`\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`)
			gotErr := regex.ReplaceAllString(err.Error(), "MOCKED_ID")

			// then
			assert.Equal(t, tc.expErrMsg, gotErr)
		})
	}

	// sanity check
	familyDetails, err := cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{
		TypeRef: &gqllocalapi.TypeRefFilterInput{
			Path:     typeRef(typeInstanceTypeRef).Path,
			Revision: ptr.String(typeRef(typeInstanceTypeRef).Revision),
		},
	}, local.WithFields(local.TypeInstanceRootFields))
	require.NoError(t, err)
	assert.Len(t, familyDetails, 0)
}

func TestCreateFindAndDeleteTypeInstances(t *testing.T) {
	hubJSAddr := os.Getenv(hubJSBackendAddr)
	if hubJSAddr == "" {
		t.Skipf("skipping running example test as the env %s is not provided", hubJSAddr)
	}
	ctx := context.Background()
	cli := local.NewDefaultClient(hubJSAddr)
	builtinStorage := getBuiltinStorageTypeInstance(ctx, t, cli)

	// create TypeInstance
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

	// check delete TypeInstance
	err = cli.DeleteTypeInstance(ctx, createdTypeInstanceID)
	require.NoError(t, err)

	got, err := cli.FindTypeInstance(ctx, createdTypeInstanceID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestCreatesMultipleTypeInstancesWithUsesRelations(t *testing.T) {
	hubJSAddr := os.Getenv(hubJSBackendAddr)
	if hubJSAddr == "" {
		t.Skipf("skipping running example test as the env %s is not provided", hubJSAddr)
	}
	ctx := context.Background()
	cli := local.NewDefaultClient(hubJSAddr)
	builtinStorage := getBuiltinStorageTypeInstance(ctx, t, cli)

	createdTypeInstanceIDs, err := cli.CreateTypeInstances(ctx, createTypeInstancesInput())
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

func TestLockTypeInstances(t *testing.T) {
	const (
		fooOwnerID = "namespace/Foo"
		barOwnerID = "namespace/Bar"
	)
	hubJSAddr := os.Getenv(hubJSBackendAddr)
	if hubJSAddr == "" {
		t.Skipf("skipping running example test as the env %s is not provided", hubJSAddr)
	}
	ctx := context.Background()
	cli := local.NewDefaultClient(hubJSAddr)

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

	// given id1 and id2 are not locked
	firstTwoInstances := createdTIIDs[:2]
	lastInstances := createdTIIDs[2:]

	// when Foo tries to locks them
	err := cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     firstTwoInstances,
		OwnerID: fooOwnerID,
	})

	// then should success
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

	// when Foo tries to locks them
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     createdTIIDs, // lock all 3 instances, when the first two are already locked
		OwnerID: fooOwnerID,
	})
	require.NoError(t, err)

	// then should success
	got, err = cli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{})
	require.NoError(t, err)

	for _, instance := range got {
		if !includes(createdTIIDs, instance.ID) {
			continue
		}
		assert.NotNil(t, instance.LockedBy)
		assert.Equal(t, *instance.LockedBy, fooOwnerID)
	}

	// given id1, id2, id3 are locked by Foo, id4: not found
	lockingIDs := createdTIIDs
	lockingIDs = append(lockingIDs, "123-not-found")

	// when Foo tries to locks id1,id2,id3,id4
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     lockingIDs,
		OwnerID: fooOwnerID,
	})

	// then should failed with id4 not found error
	require.Error(t, err)
	require.Equal(t, err.Error(), heredoc.Doc(`while executing mutation to lock TypeInstances: All attempts fail:
	#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs "123-not-found" were not found`))

	// when Bar tries to locks id1,id2,id3,id4
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     lockingIDs,
		OwnerID: barOwnerID,
	})

	// then should failed with id4 not found and already locked error for id1,id2,id3
	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
	#1: graphql: failed to lock TypeInstances: 2 errors occurred: \[TypeInstances with IDs "123-not-found" were not found, TypeInstances with IDs %s are locked by different owner\]`, allPermutations(createdTIIDs))), err.Error())

	// given id1, id2, id3 are locked by Foo, id4: not locked
	id4, err := cli.CreateTypeInstance(ctx, typeInstance("id4"))
	require.NoError(t, err)

	defer cli.DeleteTypeInstance(ctx, id4)

	// when Bar tries to locks all of them
	lockingIDs = createdTIIDs
	lockingIDs = append(lockingIDs, id4)
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     lockingIDs,
		OwnerID: barOwnerID,
	})

	// then should failed with error id1,id2,id3 already locked by Foo
	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
	#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))), err.Error())

	// when Bar tries to locks all of them
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     lockingIDs,
		OwnerID: barOwnerID,
	})

	// then should failed with error id1,id2,id3 already locked by Foo
	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
	#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))), err.Error())

	// then should unlock id1,id2,id3
	err = cli.UnlockTypeInstances(ctx, &gqllocalapi.UnlockTypeInstancesInput{
		Ids:     createdTIIDs,
		OwnerID: fooOwnerID,
	})
	require.NoError(t, err)
}

func TestUpdateTypeInstances(t *testing.T) {
	const (
		fooOwnerID = "namespace/Foo"
		barOwnerID = "namespace/Bar"
	)
	hubJSAddr := os.Getenv(hubJSBackendAddr)
	if hubJSAddr == "" {
		t.Skipf("skipping running example test as the env %s is not provided", hubJSAddr)
	}
	ctx := context.Background()
	cli := local.NewDefaultClient(hubJSAddr)

	var createdTIIDs []string

	// given id1 and id2 are not locked
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

	// when try to update them
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

	// then should success
	require.NoError(t, err)
	for _, instance := range updatedTI {
		assert.Equal(t, len(instance.LatestResourceVersion.Metadata.Attributes), 1)
		assert.EqualValues(t, instance.LatestResourceVersion.Metadata.Attributes[0], expUpdateTI.Attributes[0])
	}

	// when id1 and id2 are locked by Foo
	expUpdateTI = &gqllocalapi.UpdateTypeInstanceInput{
		Attributes: []*gqllocalapi.AttributeReferenceInput{
			{Path: "cap.update.locked.by.foo", Revision: "0.0.1"},
		},
	}

	// then should success
	err = cli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
		Ids:     createdTIIDs,
		OwnerID: fooOwnerID,
	})
	require.NoError(t, err)

	// when update them as Foo owner
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

	// then should success
	require.NoError(t, err)
	for _, instance := range updatedTI {
		assert.Equal(t, len(instance.LatestResourceVersion.Metadata.Attributes), 1)
		assert.EqualValues(t, instance.LatestResourceVersion.Metadata.Attributes[0], expUpdateTI.Attributes[0])
	}

	// when update them as Bar owner
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

	// then should failed with error id1,id2 already locked by different owner
	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(heredoc.Docf(`while executing mutation to update TypeInstances: All attempts fail:
	#1: graphql: failed to update TypeInstances: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))), err.Error())

	// when update them without owner
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

	// then should failed with error id1,id2 already locked by different owner
	require.Error(t, err)
	assert.Regexp(t, regexp.MustCompile(heredoc.Docf(`while executing mutation to update TypeInstances: All attempts fail:
	#1: graphql: failed to update TypeInstances: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))), err.Error())

	// when update one property with Foo owner, and second without owner
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

	// then should failed with error id2 already locked by different owner
	require.Error(t, err)
	require.Equal(t, err.Error(), heredoc.Docf(`while executing mutation to update TypeInstances: All attempts fail:
	     				#1: graphql: failed to update TypeInstances: TypeInstances with IDs "%s" are locked by different owner`, createdTIIDs[1]))

	// given id3 does not exist
	// when try to update it
	_, err = cli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
		{
			ID:           "id3",
			TypeInstance: expUpdateTI,
		},
	})

	// then should failed with error id3 not found
	require.Error(t, err)
	require.Equal(t, err.Error(), heredoc.Doc(`while executing mutation to update TypeInstances: All attempts fail:
	     			#1: graphql: failed to update TypeInstances: TypeInstances with IDs "id3" were not found`))

	// then should unlock id1,id2,id3"
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

func registerExternalStorage(ctx context.Context, t *testing.T, cli *local.Client, value interface{}) (string, func()) {
	t.Helper()

	externalStorageID, err := cli.CreateTypeInstance(ctx, fixExternalDotenvStorage(value))
	require.NoError(t, err)
	require.NotEmpty(t, externalStorageID)

	return externalStorageID, func() {
		_ = cli.DeleteTypeInstance(ctx, externalStorageID)
	}
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

func fixExternalDotenvStorage(value interface{}) *gqllocalapi.CreateTypeInstanceInput {
	return &gqllocalapi.CreateTypeInstanceInput{
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.example.filesystem.storage",
			Revision: "0.1.0",
		},
		Value: value,
	}
}

func typeRef(in string) *gqllocalapi.TypeInstanceTypeReferenceInput {
	out := strings.Split(in, ":")
	return &gqllocalapi.TypeInstanceTypeReferenceInput{Path: out[0], Revision: out[1]}
}

func createTypeInstancesInput() *gqllocalapi.CreateTypeInstancesInput {
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
