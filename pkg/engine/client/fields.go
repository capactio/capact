package client

const actionFields = `
    name
    createdAt
    input {
        parameters
        typeInstances {
            id
            name
            optional
            typeRef {
                path
                revision
            }
        }
    }
    output {
        typeInstances {
            name
            typeRef {
                path
                revision
            }
            id
            name
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
`

const policyFields = `
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
			injectTypeInstances {
				id
				typeRef {
					path
					revision
				}
			}
		}
	}
`
