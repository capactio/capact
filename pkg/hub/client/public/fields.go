package public

import "fmt"

// GenericMetadataFields for querying the GenericMetadata fields.
var GenericMetadataFields = `
      prefix
      path
      name
      displayName
      description
      maintainers {
        name
        email
      }
      iconURL
      documentationURL
      supportURL
      iconURL
      `

// AttributeFields for querying the Attributes fields and GenericMetadata.
var AttributeFields = fmt.Sprintf(`
      metadata {
        %s
      }
      revision
      spec {
        additionalRefs
      }
      `, GenericMetadataFields)
