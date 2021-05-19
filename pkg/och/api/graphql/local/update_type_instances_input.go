package graphql

// The types had to be moved out from generated models to add `omitempty` tags.
type UpdateTypeInstancesInput struct {
	// Allows you to update TypeInstances which are locked by a given ownerID. If not provided,
	// you can update only those TypeInstances which are not locked.
	OwnerID      *string                  `json:"ownerID,omitempty"`
	ID           string                   `json:"id"`
	TypeInstance *UpdateTypeInstanceInput `json:"typeInstance"`
}

// At least one property needs to be specified.
type UpdateTypeInstanceInput struct {
	// The attributes property is optional. If not provided, previous value is used.
	Attributes []*AttributeReferenceInput `json:"attributes,omitempty"`
	// The value property is optional. If not provided, previous value is used.
	Value interface{} `json:"value,omitempty"`
}
