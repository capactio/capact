package public

// FindInterfaceRevisionOptions stores Interface Revision filtering parameters.
type FindInterfaceRevisionOptions struct {
	Fields string
}

// Apply is used to configure the ListImplementationRevisionsOptions.
func (o *FindInterfaceRevisionOptions) Apply(opts ...FindInterfaceRevisionOption) {
	o.Fields = InterfaceRevisionAllFields // default to all fields

	// Apply overrides
	for _, opt := range opts {
		opt(o)
	}
}

// FindInterfaceRevisionOption provides an option to configure the find request for Interface Revision.
type FindInterfaceRevisionOption func(*FindInterfaceRevisionOptions)

// WithInputDataOnly narrows down the request query fields to the `spec.input` property of InterfaceRevision entity.
func WithInputDataOnly(opt *FindInterfaceRevisionOptions) {
	opt.Fields = InterfaceRevisionInputDataFields
}
