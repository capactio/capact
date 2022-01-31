package public

import "fmt"

// typeRevisionFieldsRegistry holds possible fields configuration for InterfaceRevision query.
var typeRevisionFieldsRegistry = map[TypeRevisionQueryFields]string{
	TypeRevisionRootFields: `
		revision`,
	TypeRevisionMetadataFields:          typeRevisionMetadataFields,
	TypeRevisionSpecFields:              typeRevisionSpecFields,
	TypeRevisionSpecAdditionalRefsField: typeRevisionSpecAdditionalRefsField,
}

// typeRevisionMetadataFields specifies TypeRevision's Metadata fields.
var typeRevisionMetadataFields = fmt.Sprintf(`
      metadata {
        %s
      }`, genericMetadataFields)

// typeRevisionSpecFields specifies TypeRevision's spec fields only.
const typeRevisionSpecFields = `
      spec {
        jsonSchema
        additionalRefs
      }
`

// typeRevisionSpecAdditionalRefsField specifies TypeRevision's spec.additionalRefs field only.
const typeRevisionSpecAdditionalRefsField = `
      spec {
        additionalRefs
      }
`
