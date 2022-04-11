package storagebackend

// TypeInstanceValue defines properties for TypeInstance value for every Storage Backend.
type TypeInstanceValue struct {
	URL           string      `json:"url"`
	AcceptValue   bool        `json:"acceptValue"`
	ContextSchema interface{} `json:"contextSchema"`
}
