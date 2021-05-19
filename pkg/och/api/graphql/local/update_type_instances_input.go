package graphql

import "encoding/json"

// The types had to be moved out from generated models to add `omitempty` tags.
type UpdateTypeInstancesInput struct {
	// Allows you to update TypeInstances which are locked by a given ownerID. If not provided,
	// you can update only those TypeInstances which are not locked.
	OwnerID      *string                  `json:"ownerID,omitempty"`
	ID           string                   `json:"id"`
	CreatedBy    *string                  `json:"createdBy,omitempty"`
	TypeInstance *UpdateTypeInstanceInput `json:"typeInstance"`
}

// At least one property needs to be specified.
type UpdateTypeInstanceInput struct {
	// The attributes property is optional. If not provided, previous value is used.
	Attributes []*AttributeReferenceInput `json:"attributes"`
	// The value property is optional. If not provided, previous value is used.
	Value interface{} `json:"value,omitempty"`
}

type NativeUpdateTypeInstanceInput UpdateTypeInstanceInput

func (u *UpdateTypeInstanceInput) MarshalJSON() ([]byte, error) {
	if u.Attributes != nil {
		return json.Marshal(NativeUpdateTypeInstanceInput(*u))
	}

	// we do not want to take into account `attributes = nil`
	// cause OCH doesn't handle "null" properties
	a := struct {
		// The value property is optional. If not provided, previous value is used.
		Value interface{} `json:"value,omitempty"`
	}{
		Value: u.Value,
	}
	return json.Marshal(a)
}
