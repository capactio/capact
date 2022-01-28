package manifest_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client/public"

	"github.com/stretchr/testify/assert"
)

type fakeHub struct {
	checkManifestsFn func(ctx context.Context, manifestRefs []gqlpublicapi.ManifestReference) (map[gqlpublicapi.ManifestReference]bool, error)
	knownTypes       []*gqlpublicapi.Type
}

func (h *fakeHub) ListTypes(_ context.Context, opts ...public.TypeOption) ([]*gqlpublicapi.Type, error) {
	if h.knownTypes == nil {
		return nil, nil
	}

	typeOpts := &public.TypeOptions{}
	typeOpts.Apply(opts...)

	if typeOpts.Filter.PathPattern == nil {
		return h.knownTypes, nil
	}
	var out []*gqlpublicapi.Type
	for _, item := range h.knownTypes {
		matched, err := regexp.MatchString(*typeOpts.Filter.PathPattern, item.Path)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}
		out = append(out, item)
	}

	return out, nil
}

func (h *fakeHub) CheckManifestRevisionsExist(ctx context.Context, manifestRefs []gqlpublicapi.ManifestReference) (map[gqlpublicapi.ManifestReference]bool, error) {
	return h.checkManifestsFn(ctx, manifestRefs)
}

func fixHubForManifestsExistence(t *testing.T, result map[gqlpublicapi.ManifestReference]bool, err error) *fakeHub {
	t.Helper()

	hub := &fakeHub{
		checkManifestsFn: func(ctx context.Context, manifestRefs []gqlpublicapi.ManifestReference) (map[gqlpublicapi.ManifestReference]bool, error) {
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
