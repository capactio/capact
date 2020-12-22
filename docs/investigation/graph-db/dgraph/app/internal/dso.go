package internal

type InterfaceGroup struct {
	Uid        string      `json:"uid,omitempty"`
	DType      []string    `json:"dgraph.type,omitempty"`
	Interfaces []Interface `json:"InterfaceGroup.interfaces"`
}

type Interface struct {
	Uid            string                   `json:"uid,omitempty"`
	Path           string                   `json:"Interface.path"`
	Name           string                   `json:"Interface.name"`
	Prefix         string                   `json:"Interface.prefix"`
	LatestRevision map[string]interface{}   `json:"Interface.latestRevision,omitempty"`
	Revisions      []map[string]interface{} `json:"Interface.revisions"`
	DType          []string                 `json:"dgraph.type,omitempty"`
}

type InterfaceRevision struct {
	Metadata GenericMetadata `mapstructure:"InterfaceRevision.metadata"`
	Revision string          `mapstructure:"InterfaceRevision.revision"`
}

type GenericMetadata struct {
	Path   string `mapstructure:"path"`
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

// Implementation

type Implementation struct {
	Uid            string                   `json:"uid,omitempty"`
	Path           string                   `json:"Implementation.path"`
	Name           string                   `json:"Implementation.name"`
	Prefix         string                   `json:"Implementation.prefix"`
	LatestRevision map[string]interface{}   `json:"Implementation.latestRevision,omitempty"`
	Revisions      []map[string]interface{} `json:"Implementation.revisions"`
	DType          []string                 `json:"dgraph.type,omitempty"`
}

type ImplementationRevision struct {
	Metadata GenericMetadata    `mapstructure:"ImplementationRevision.metadata"`
	Revision string             `mapstructure:"ImplementationRevision.revision"`
	Spec     ImplementationSpec `mapstructure:"ImplementationRevision.spec"`
}

type ImplementationSpec struct {
	Implements []InterfaceReference `mapstructure:"ImplementationSpec.implements"`
	Requires [] TypeReference `mapstructure:"ImplementationSpec.requires"`
}

type TypeReference struct {
	Path     string `mapstructure:"TypeReference.path"`
	Revision string `mapstructure:"TypeReference.revision"`
}
type InterfaceReference struct {
	Path     string `mapstructure:"InterfaceReference.path"`
	Revision string `mapstructure:"InterfaceReference.revision"`
}

// Decode objects from queries

type DecodeInterfaceQuery struct {
	Path      string
	Revisions []struct {
		Uid       string
		Rev string
	}
}
