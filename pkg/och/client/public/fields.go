package public

import "fmt"

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

var AttributeFields = fmt.Sprintf(`
			metadata {
				%s
			}
			revision
			spec {
				additionalRefs
			}
			signature {
				och
			}
			`, GenericMetadataFields)

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
				och
			}
			`, GenericMetadataFields, AttributeFields)

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
