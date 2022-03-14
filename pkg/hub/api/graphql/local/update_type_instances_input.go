package graphql

import "encoding/json"

// The types had to be moved out from generated models to add `omitempty` tags.

// UpdateTypeInstancesInput defines input for update TypeInstances mutation.
type UpdateTypeInstancesInput struct {
	// Allows you to update TypeInstances which are locked by a given ownerID. If not provided,
	// you can update only those TypeInstances which are not locked.
	OwnerID      *string                  `json:"ownerID,omitempty"`
	ID           string                   `json:"id"`
	CreatedBy    *string                  `json:"createdBy,omitempty"`
	TypeInstance *UpdateTypeInstanceInput `json:"typeInstance"`
}

// UpdateTypeInstanceInput defines input for update TypeInstance mutation.
// At least one property needs to be specified.
type UpdateTypeInstanceInput struct {
	// The attributes property is optional. If not provided, previous value is used.
	Attributes []*AttributeReferenceInput `json:"attributes"`
	// The value property is optional. If not provided, previous value is used.
	Value interface{} `json:"value,omitempty"`
	// The backend property is optional. If not provided, previous value is used.
	Backend *UpdateTypeInstanceBackendInput `json:"backend,omitempty"`
}

// NativeUpdateTypeInstanceInput declared to shadow custom MarshalJSON declared on UpdateTypeInstanceInput.
type NativeUpdateTypeInstanceInput UpdateTypeInstanceInput

// MarshalJSON provides custom marshaling to support case when Attributes are not specified.
// TODO: it is a temporary solution and should be fixed directly in Hub server.
func (u *UpdateTypeInstanceInput) MarshalJSON() ([]byte, error) {
	if u.Attributes != nil {
		return json.Marshal(NativeUpdateTypeInstanceInput(*u))
	}

	// we do not want to take into account `attributes = nil`
	// cause Hub doesn't handle "null" properties
	a := struct {
		// The value property is optional. If not provided, previous value is used.
		Value interface{} `json:"value,omitempty"`
		// The backend property is optional. If not provided, previous value is used.
		Backend *UpdateTypeInstanceBackendInput `json:"backend,omitempty"`
	}{
		Value:   u.Value,
		Backend: u.Backend,
	}
	return json.Marshal(a)
}
