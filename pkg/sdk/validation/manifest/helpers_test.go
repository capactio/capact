package manifest_test

import (
	"context"
	"errors"
	"testing"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"

	"github.com/stretchr/testify/assert"
)

type fakeHub struct {
	fn                      func(ctx context.Context, manifestRefs []gqlpublicapi.ManifestReference) (map[gqlpublicapi.ManifestReference]bool, error)
	knownTypesByPathPattern map[string][]*gqlpublicapi.Type
}

func (h *fakeHub) ListTypes(_ context.Context, opts ...public.TypeOption) ([]*gqlpublicapi.Type, error) {
	if h.knownTypesByPathPattern == nil {
		return nil, nil
	}

	typeOpts := &public.TypeOptions{}
	typeOpts.Apply(opts...)

	if typeOpts.Filter.PathPattern == nil {
		return nil, nil
	}
	return h.knownTypesByPathPattern[*typeOpts.Filter.PathPattern], nil
}

func (h *fakeHub) CheckManifestRevisionsExist(ctx context.Context, manifestRefs []gqlpublicapi.ManifestReference) (map[gqlpublicapi.ManifestReference]bool, error) {
	return h.fn(ctx, manifestRefs)
}

func fixHubForManifestsExistence(t *testing.T, result map[gqlpublicapi.ManifestReference]bool, err error) *fakeHub {
	t.Helper()

	hub := &fakeHub{
		fn: func(ctx context.Context, manifestRefs []gqlpublicapi.ManifestReference) (map[gqlpublicapi.ManifestReference]bool, error) {
			var resultManifestRefs []gqlpublicapi.ManifestReference
			for key := range result {
				resultManifestRefs = append(resultManifestRefs, key)
			}
			ok := assert.ElementsMatch(t, manifestRefs, resultManifestRefs)
			if !ok {
				return nil, errors.New("manifest references don't match")
			}

			return result, err
		},
	}
	return hub
}
