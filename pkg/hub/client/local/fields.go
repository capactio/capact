package local

import (
	"fmt"
)

var typeInstancesFieldsRegistry = map[TypeInstancesQueryFields]string{
	TypeInstanceRootFields:                 rootFields,
	TypeInstanceTypeRefFields:              typeRefFields,
	TypeInstanceUsesIDField:                usesIDField,
	TypeInstanceUsedByIDField:              usedByIDField,
	TypeInstanceLatestResourceVersionField: latestResourceVersionField,
	TypeInstanceAllFields:                  typeInstanceAllFields,
	TypeInstanceAllFieldsWithUses:          typeInstanceWithUsesFields,
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
	typeInstanceWithUsesFields = fmt.Sprintf(`
		%s
		uses {
			%s
		}
		usedBy {
			%s
		}`, typeInstanceAllFields, typeInstanceAllFields, typeInstanceAllFields)

	typeInstanceAllFields = fmt.Sprintf(`
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
		}`, rootFields, typeRefFields,
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
		}
`
