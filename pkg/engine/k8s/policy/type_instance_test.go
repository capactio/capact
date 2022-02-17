package policy_test

import (
	"strings"
	"testing"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"github.com/stretchr/testify/assert"
)

func TestTypeInstanceBackendCollection_GetByTypeRef(t *testing.T) {
	data := policy.TypeInstanceBackendCollection{}

	data.SetByTypeRef(fixTypeRef("cap.type.capactio.examples.message:0.1.0"), fixTypeInstanceBackend("ID1"))
	data.SetByTypeRef(fixTypeRef("cap.type.capactio.examples.*"), fixTypeInstanceBackend("ID2"))
	data.SetByTypeRef(fixTypeRef("cap.type.capactio.examples.*:0.2.0"), fixTypeInstanceBackend("ID3"))
	data.SetByTypeRef(fixTypeRef("cap.*"), fixTypeInstanceBackend("ID4"))
	data.SetByTypeRef(fixTypeRef("cap.*:0.1.0"), fixTypeInstanceBackend("ID5"))

	data.SetByAlias("aws-secret-manager", fixTypeInstanceBackend("ID333")) // ensure that Alias does not affect proper selection

	tests := map[string]struct {
		givenTypeRef types.TypeRef
		expBackend   policy.TypeInstanceBackend
		expFound     bool
	}{
		"Should match exact type ref": {
			givenTypeRef: types.TypeRef{
				Path:     "cap.type.capactio.examples.message",
				Revision: "0.1.0",
			},
			expFound:   true,
			expBackend: fixTypeInstanceBackend("ID1"),
		},
		"Should match pattern cap.type.capactio.examples.*": {
			givenTypeRef: types.TypeRef{
				Path:     "cap.type.capactio.examples.other-that-message",
				Revision: "0.1.0",
			},
			expFound:   true,
			expBackend: fixTypeInstanceBackend("ID2"),
		},
		"Should match pattern cap.type.capactio.examples.*:0.2.0": {
			givenTypeRef: types.TypeRef{
				Path:     "cap.type.capactio.examples.other-that-message",
				Revision: "0.2.0",
			},
			expFound:   true,
			expBackend: fixTypeInstanceBackend("ID3"),
		},
		"Should match generic cap.* with revision": {
			givenTypeRef: types.TypeRef{
				Path:     "cap.type.aws.examples",
				Revision: "0.1.0",
			},
			expFound:   true,
			expBackend: fixTypeInstanceBackend("ID5"),
		},
		"Should match generic cap.* with different revision": {
			givenTypeRef: types.TypeRef{
				Path:     "cap.type.aws.examples",
				Revision: "0.2.0",
			},
			expFound:   true,
			expBackend: fixTypeInstanceBackend("ID4"),
		},

		"Should not found backend for unknown path": {
			givenTypeRef: types.TypeRef{
				Path:     "some.not.registered.type.ref",
				Revision: "0.2.0",
			},
			expFound: false,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			gotBackend, gotFound := data.GetByTypeRef(tc.givenTypeRef)

			assert.Equal(t, tc.expBackend, gotBackend)
			assert.Equal(t, tc.expFound, gotFound)
		})
	}
}

func TestTypeInstanceBackendCollection_GetByAlias(t *testing.T) {
	data := policy.TypeInstanceBackendCollection{}

	data.SetByTypeRef(fixTypeRef("cap.type.capactio.examples.message:0.1.0"), fixTypeInstanceBackend("ID1")) // ensure that TypeRef does not affect proper selection
	data.SetByAlias("helm-storage", fixTypeInstanceBackend("ID2"))
	data.SetByAlias("aws-secret-manager", fixTypeInstanceBackend("ID3"))

	tests := map[string]struct {
		givenAlias string

		expFound   bool
		expBackend policy.TypeInstanceBackend
	}{
		"Should match helm-storage": {
			givenAlias: "helm-storage",

			expFound:   true,
			expBackend: fixTypeInstanceBackend("ID2"),
		},
		"Should match aws-secret-manager": {
			givenAlias: "aws-secret-manager",

			expFound:   true,
			expBackend: fixTypeInstanceBackend("ID3"),
		},
		"Should not found valut": {
			givenAlias: "vault",

			expFound: false,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			gotBackend, found := data.GetByAlias(tc.givenAlias)
			assert.Equal(t, tc.expFound, found)
			assert.Equal(t, tc.expBackend, gotBackend)
		})
	}
}

func fixTypeRef(in string) types.ManifestRefWithOptRevision {
	var rev *string
	parts := strings.Split(in, ":")
	if len(parts) > 1 {
		rev = ptr.String(parts[1])
	}
	return types.ManifestRefWithOptRevision{
		Path:     parts[0],
		Revision: rev,
	}
}

func fixTypeInstanceBackend(id string) policy.TypeInstanceBackend {
	return policy.TypeInstanceBackend{
		TypeInstanceReference: policy.TypeInstanceReference{
			ID: id,
		},
	}
}
