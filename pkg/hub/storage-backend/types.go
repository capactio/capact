package storagebackend

// TypeValue defines properties for TypeInstance value for every Storage Backend.
type TypeValue struct {
	URL           string      `json:"url"`
	AcceptValue   bool        `json:"acceptValue"`
	ContextSchema interface{} `json:"contextSchema"`
}
