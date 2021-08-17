package public

import "fmt"

// genericMetadataFields for querying the GenericMetadata fields.
var genericMetadataFields = `
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

// attributeFields for querying the Attributes fields and GenericMetadata.
var attributeFields = fmt.Sprintf(`
      metadata {
        %s
      }
      revision
      spec {
        additionalRefs
      }
      `, genericMetadataFields)
