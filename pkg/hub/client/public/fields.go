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

// AttributeFields for quering the Attributes fields and GenericMetadata.
var AttributeFields = fmt.Sprintf(`
			metadata {
				%s
			}
			revision
			spec {
				additionalRefs
			}
			signature {
				hub
			}
			`, GenericMetadataFields)

// ImplementationFields for quering the Implementation fields with all revisions.
var ImplementationFields = fmt.Sprintf(`
			path
			name
			prefix
			revisions {
				%s
			}
`, ImplementationRevisionFields)

// ImplementationRevisionFields for quering ImplementationRevision fields.
var ImplementationRevisionFields = fmt.Sprintf(`
			metadata {
					%s
					attributes {
						%s
					}
			}
			revision
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
			signature {
				hub
			}
			`, GenericMetadataFields, AttributeFields)

// InterfaceRevisionFields for quering InterfaceRevision fields with GenericMetadata and all revisions.
var InterfaceRevisionFields = fmt.Sprintf(`
      revision
      metadata {
				%s
      }
      spec {
        input {
          parameters {
            name
            jsonSchema
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
`, GenericMetadataFields, ImplementationRevisionFields)

// InterfacesFields for quering Interface with the latest revision only.
var InterfacesFields = fmt.Sprintf(`
		path
		name
		prefix
		latestRevision {
		  %s
		}
`, InterfaceRevisionFields)
