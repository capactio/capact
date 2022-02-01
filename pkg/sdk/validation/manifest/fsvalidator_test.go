package manifest_test

import (
	"context"
	"testing"

	"capact.io/capact/internal/cli/schema"
	graphql "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/validation/manifest"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystemValidator_ValidateFile(t *testing.T) {
	// given
	sampleAttr := manifestRef("cap.core.sample.attr")
	tests := map[string]struct {
		manifestPath                string
		expectedValidationErrorMsgs []string
		expectedGeneralErrMsg       string
		hubCli                      manifest.Hub
	}{
		"Valid Implementation": {
			manifestPath:                "testdata/valid-implementation.yaml",
			expectedValidationErrorMsgs: []string{},
			hubCli: fixHubForManifestsExistence(t, map[graphql.ManifestReference]bool{
				manifestRef("cap.sample.attribute"):                          true,
				manifestRef("cap.type.mattermost.helm.install-input"):        true,
				manifestRef("cap.type.database.postgresql.config"):           true,
				manifestRef("cap.interface.productivity.mattermost.install"): true,
				manifestRef("cap.core.type.platform.kubernetes"):             true,
				manifestRef("cap.interface.runner.helm.install"):             true,
				manifestRef("cap.interface.runner.argo.run"):                 true,
				manifestRef("cap.interface.templating.jinja2.template"):      true,
				manifestRef("cap.interface.database.postgresql.install"):     true,
				manifestRef("cap.interface.database.postgresql.create-db"):   true,
				manifestRef("cap.interface.database.postgresql.create-user"): true,
			}, nil),
		},
		"Valid Interface": {
			manifestPath:                "testdata/valid-interface.yaml",
			expectedValidationErrorMsgs: []string{},
			hubCli: fixHubForManifestsExistence(t, map[graphql.ManifestReference]bool{
				manifestRef("cap.type.productivity.mattermost.config"):        true,
				manifestRef("cap.type.productivity.mattermost.install-input"): true,
			}, nil),
		},
		"Valid Type": {
			manifestPath:                "testdata/valid-type.yaml",
			expectedValidationErrorMsgs: []string{},
			hubCli:                      fixHubForManifestsExistence(t, map[graphql.ManifestReference]bool{sampleAttr: true}, nil),
		},
		"Invalid Implementation": {
			manifestPath: "testdata/invalid-implementation.yaml",
			expectedValidationErrorMsgs: []string{
				"OCFSchemaValidator: spec: appVersion is required",
				`RemoteImplementationValidator: Type "cap.type.database.postgresql.config:0.1.0" is not attached to "cap.core.type.platform" parent node`,
				"RemoteImplementationValidator: manifest revision 'cap.interface.cms.wordpress:0.1.0' doesn't exist in Hub",
			},
			hubCli: fixHub(t, []*graphql.Type{
				fixGQLType("cap.type.database.postgresql.config", "0.1.0", ""),
			}, map[graphql.ManifestReference]bool{
				manifestRef("cap.interface.cms.wordpress"):         false,
				manifestRef("cap.type.database.postgresql.config"): true,
			}, nil),
		},
		"Invalid Interface": {
			manifestPath: "testdata/invalid-interface.yaml",
			expectedValidationErrorMsgs: []string{
				"RemoteInterfaceValidator: manifest revision 'cap.type.productivity.mattermost.install-input:0.1.0' doesn't exist in Hub",
			},
			hubCli: fixHubForManifestsExistence(t, map[graphql.ManifestReference]bool{
				manifestRef("cap.type.productivity.mattermost.install-input"): false,
				manifestRef("cap.type.productivity.mattermost.config"):        true,
			}, nil),
		},
		"Invalid JSON Schema in Interface": {
			manifestPath: "testdata/invalid-interface_json-schema.yaml",
			expectedValidationErrorMsgs: []string{
				"InterfaceValidator: properties.config.type: Must validate at least one schema (anyOf)",
				`InterfaceValidator: properties.config.type: properties.config.type must be one of the following: "array", "boolean", "integer", "null", "number", "object", "string"`,
			},
		},
		"Invalid JSON Schema in Type": {
			manifestPath: "testdata/invalid-type_json-schema.yaml",
			expectedValidationErrorMsgs: []string{
				"TypeValidator: type: Must validate at least one schema (anyOf)",
				`TypeValidator: type: type must be one of the following: "array", "boolean", "integer", "null", "number", "object", "string"`,
			},
		},
		"Invalid additionalRefs in Type": {
			manifestPath: "testdata/invalid-type_additionalRefs.yaml",
			expectedValidationErrorMsgs: []string{
				`TypeValidator: spec.additionalRefs: "cap.interface.postgresql" is not allowed. It can refer only to a parent node under "cap.core.type." or "cap.type."`,
				`RemoteTypeValidator: "cap.core.type.platform.kubernetes" cannot be used as parent node as it resolves to concrete Type`,
			},
			hubCli: fixHub(t, []*graphql.Type{
				fixGQLType("cap.core.type.platform.kubernetes", "0.1.0", ""),
			}, nil, nil),
		},
		"Invalid Type": {
			manifestPath: "testdata/invalid-type.yaml",
			expectedValidationErrorMsgs: []string{
				`TypeValidator: spec.jsonSchema.value: invalid JSON: invalid character '}' looking for beginning of object key string`,
				"RemoteTypeValidator: manifest revision 'cap.core.sample.attr:0.1.0' doesn't exist in Hub",
			},
			hubCli: fixHubForManifestsExistence(t, map[graphql.ManifestReference]bool{
				sampleAttr: false,
			}, nil),
		},
		"Error from Hub": {
			manifestPath: "testdata/valid-interface.yaml",
			expectedValidationErrorMsgs: []string{
				"RemoteInterfaceValidator: internal: while checking if manifest revisions exist: test error",
			},
			hubCli: fixHubForManifestsExistence(t, map[graphql.ManifestReference]bool{
				manifestRef("cap.type.productivity.mattermost.config"):        true,
				manifestRef("cap.type.productivity.mattermost.install-input"): true,
			}, errors.New("test error")),
		},
		"Cannot load file": {
			manifestPath:          "testdata/no-file.yaml",
			expectedGeneralErrMsg: "open testdata/no-file.yaml: no such file or directory",
		},
		"Invalid manifest": {
			manifestPath: "testdata/invalid-manifest.yaml",
			expectedValidationErrorMsgs: []string{
				"failed to read manifest metadata: OCFVersion and Kind must not be empty",
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			var opts []manifest.ValidatorOption
			if tc.hubCli != nil {
				opts = append(opts, manifest.WithRemoteChecks(tc.hubCli))
			}

			validator := manifest.NewDefaultFilesystemValidator(
				&schema.LocalFileSystem{},
				"../../../../ocf-spec",
				opts...,
			)

			// when
			result, err := validator.Do(context.Background(), tc.manifestPath)

			// then
			if tc.expectedGeneralErrMsg != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tc.expectedGeneralErrMsg)
			} else {
				require.Nil(t, err)
			}

			require.Len(t, result.Errors, len(tc.expectedValidationErrorMsgs))

			if len(result.Errors) > 0 {
				var errMsgs []string
				for _, err := range result.Errors {
					errMsgs = append(errMsgs, err.Error())
				}
				assert.ElementsMatch(t, tc.expectedValidationErrorMsgs, errMsgs)
			}
		})
	}
}

func manifestRef(path string) graphql.ManifestReference {
	return graphql.ManifestReference{
		Path:     path,
		Revision: "0.1.0",
	}
}
