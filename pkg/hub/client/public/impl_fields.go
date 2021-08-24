package public

import "fmt"

// ifaceRevisionFieldsRegistry holds possible fields configuration for ImplementationRevision query.
var implRevisionFieldsRegistry = map[ImplementationRevisionQueryFields]string{
	ImplementationRevisionRootFields: `
		revision`,
	ImplementationRevisionMetadataFields: implRevisionMetadataFields,
	ImplementationRevisionAllFields:      implRevisionAllFields,
}

var implRevisionMetadataFields = fmt.Sprintf(`
      metadata {
        %s
        attributes {
      	%s
        }
      }`, genericMetadataFields, attributeFields)

var implRevisionAllFields = fmt.Sprintf(`
      revision
      %s
      spec {
        appVersion
        implements {
          path
          revision
        }
        requires {
          prefix
          oneOf {
            typeRef {
              path
              revision
            }
            valueConstraints
            alias
          }
          anyOf {
            typeRef {
              path
              revision
            }
            valueConstraints
            alias
          }
          allOf {
            typeRef {
              path
              revision
            }
            valueConstraints
            alias
          }
        }
        imports {
          interfaceGroupPath
          alias
          appVersion
          methods {
            name
            revision
          }
        }
        additionalInput {
          typeInstances {
            name
            typeRef {
              path
              revision
            }
            verbs
          }
          parameters {
            name
            typeRef {
              path
              revision
            }
          }
        }
        additionalOutput {
          typeInstances {
            name
            typeRef {
              path
              revision
            }
          }
        }
        outputTypeInstanceRelations {
          typeInstanceName
          uses
        }
        action {
          runnerInterface
          args
        }
      }
      `, implRevisionMetadataFields)
