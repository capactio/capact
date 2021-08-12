package public

import "fmt"

// ifaceRevisionFieldsRegistry holds possible fields configuration for ImplementationRevision query.
var implRevisionFieldsRegistry = map[ImplementationRevisionQueryFields]string{
	ImplRevRootFields: `
		revision`,
	ImplRevMetadataFields: implRevisionMetadataFields,
	ImplRevAllFields:      implRevisionAllFields,
}

var implRevisionMetadataFields = fmt.Sprintf(`
      metadata {
        %s
        attributes {
      	%s
        }
      }`, GenericMetadataFields, AttributeFields)

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
