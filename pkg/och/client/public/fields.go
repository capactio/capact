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
					}
					anyOf {
						typeRef {
							path
							revision
						}
						valueConstraints
					}
					allOf {
						typeRef {
							path
							revision
						}
						valueConstraints
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
					typeInstanceRelations {
						typeInstanceName
						uses
					}
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
