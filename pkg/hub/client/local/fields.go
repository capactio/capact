package local

import (
	"fmt"
)

var typeInstancesFieldsRegistry = map[TypeInstancesQueryFields]string{
	TypeInstanceRootFields:                 rootFields,
	TypeInstanceTypeRefFields:              typeRefFields,
	TypeInstanceBackendFields:              backendFields,
	TypeInstanceUsesIDField:                usesIDField,
	TypeInstanceUsedByIDField:              usedByIDField,
	TypeInstanceLatestResourceVersionField: latestResourceVersionField,
	TypeInstanceAllFields:                  typeInstanceAllFields,
	TypeInstanceUsesAllFields:              typeInstanceUsesAllFields,
	TypeInstanceUsedByAllFields:            typeInstanceUsedByAllFields,
	// grow the extracted fields if needed
}

var (
	rootFields = `
		id
		lockedBy`

	typeRefFields = `
		typeRef {
			path
			revision
		}`

	backendFields = `
		backend {
			id
			abstract
		}`

	usedByIDField = `
			usedBy {
				id
			}`

	usesIDField = `
			uses {
				id
			}`

	latestResourceVersionField = `
			latestResourceVersion {
				resourceVersion
			}`

	typeInstanceUsesAllFields = fmt.Sprintf(`
		uses {
			%s
		}`, typeInstanceAllFields)

	typeInstanceUsedByAllFields = fmt.Sprintf(`
		usedBy {
			%s
		}`, typeInstanceAllFields)

	typeInstanceAllFields = fmt.Sprintf(`
		%s

		%s

		%s

		latestResourceVersion {
			%s
		}

		firstResourceVersion {
			%s
		}

		previousResourceVersion {
			%s
		}

		resourceVersions {
			%s
		}

		resourceVersion(resourceVersion: 1) {
			%s
		}`, rootFields, typeRefFields, backendFields,
		typeInstanceResourceVersion, typeInstanceResourceVersion, typeInstanceResourceVersion, typeInstanceResourceVersion, typeInstanceResourceVersion)
)

const typeInstanceResourceVersion = `
		resourceVersion
		createdBy
		metadata {
			attributes {
				path
				revision
			}
		}
		spec {
			value
			backend {
			  context
			}
		}
`
