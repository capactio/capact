package public

import "fmt"

// ifaceRevisionFieldsRegistry holds possible fields configuration for InterfaceRevision query.
var ifaceRevisionFieldsRegistry = map[InterfaceRevisionQueryFields]string{
	InterfaceRevisionRootFields: `
		revision`,
	InterfaceRevisionMetadataFields:                  ifaceRevisionMetadataFields,
	InterfaceRevisionInputFields:                     ifaceRevisionInputDataFields,
	InterfaceRevisionImplementationRevisionsMetadata: ifaceRevisionImplRevisionsMetadata,
	InterfaceRevisionAllFields:                       ifaceRevisionAllFields,
}

// ifaceRevisionAllFields for querying InterfaceRevision fields with GenericMetadata and all revisions.
var ifaceRevisionAllFields = fmt.Sprintf(`
      revision
      %s
      spec {
        input {
          parameters {
            name
            jsonSchema
            typeRef {
              path
              revision
            }
          }
          typeInstances {
            name
            typeRef {
              path
              revision
            }
            verbs
          }
        }
        output {
          typeInstances {
            name
            typeRef {
              path
              revision
            }
          }
        }
      }
      implementationRevisions {
          %s
      }
`, ifaceRevisionMetadataFields, implRevisionAllFields)

// ifaceRevisionMetadataFields for querying InterfaceRevision's implementationRevisions fields.
var ifaceRevisionImplRevisionsMetadata = fmt.Sprintf(`implementationRevisions {
          %s
      }`, implRevisionMetadataFields)

// ifaceRevisionMetadataFields for querying InterfaceRevision's Metadata fields.
var ifaceRevisionMetadataFields = fmt.Sprintf(`
      metadata {
        %s
      }`, genericMetadataFields)

// ifaceRevisionInputDataFields for fetching InterfaceRevision's input data fields only.
var ifaceRevisionInputDataFields = `
      spec {
        input {
          parameters {
            name
            jsonSchema
            typeRef {
              path
              revision
            }
          }
          typeInstances {
            name
            typeRef {
              path
              revision
            }
            verbs
          }
        }
      }
`
