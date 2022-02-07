package client

import "fmt"

var actionFields = fmt.Sprintf(`
	name
	createdAt
	input {
		parameters
		typeInstances {
			id
			name
		}
		actionPolicy {
			%s
		}
	}
	output {
		typeInstances {
			id
			typeRef {
				path
				revision
			}
		}
	}
	actionRef {
		path
		revision
	}
	cancel
	run
	dryRun
	renderedAction
	renderingAdvancedMode {
		enabled
		typeInstancesForRenderingIteration {
			name
			typeRef {
				path
				revision
			}
		}
	}
	renderedActionOverride
	status {
		phase
		timestamp
		message
		runner {
			status
		}
		canceledBy {
			username
			groups
			extra
		}
		runBy {
			username
			groups
			extra
		}
		createdBy {
			username
			groups
			extra
		}
	}
`, policyFields)

const policyFields = `
	interface {
		rules {
			interface {
				path
				revision
			}
			oneOf {
				implementationConstraints {
					requires {
						path
						revision
					}
					attributes {
						path
						revision
					}
					path
				}
				inject {
					requiredTypeInstances {
						id
						description
					}
					additionalParameters {
						name
						value
					}
					additionalTypeInstances {
						name
						id
					}
				}
			}
		}
	}
	typeInstance {
		rules {
			typeRef {
				path
				revision
			}
			backend {
				id
				description
			}
		}
	}
`
