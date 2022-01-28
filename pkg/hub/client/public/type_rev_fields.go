package public

import "fmt"

// typeRevisionFieldsRegistry holds possible fields configuration for InterfaceRevision query.
var typeRevisionFieldsRegistry = map[TypeRevisionQueryFields]string{
	TypeRevisionRootFields: `
		revision`,
	TypeRevisionMetadataFields: typeRevisionMetadataFields,
	TypeRevisionSpecFields:     typeRevisionSpecFields,
}

// typeRevisionMetadataFields for querying TypeRevision's Metadata fields.
var typeRevisionMetadataFields = fmt.Sprintf(`
      metadata {
        %s
      }`, genericMetadataFields)

// typeRevisionSpecFields for fetching TypeRevision's spec fields only.
var typeRevisionSpecFields = `
      spec {
        jsonSchema
				additionalRefs
      }
`
