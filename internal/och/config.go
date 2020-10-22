package och

import "fmt"

// Mode represents the possible modes for OCH. It works with the envconfig library.
type Mode string

const (
	// PublicMode is the read-only mode for public OCH. Manages following entities:
	//  - RepoMetadata
	//  - InterfaceGroup
	//  - Interface
	//  - Type
	//  - Implementation
	//  - Tag
	PublicMode Mode = "public"
	// LocalMode is the read-write mode for local OCH. Manages TypeInstance entity.
	LocalMode Mode = "local"
)

// Validate returns errors if Mode is unknown.
func (h Mode) Validate() error {
	switch h {
	case PublicMode, LocalMode:
		return nil
	}
	return fmt.Errorf("OCH mode needs to be specified. Possible options: %s and %s", PublicMode, LocalMode)
}

// Unmarshal fulfils the envconfig interface for unmarshaling.
func (h *Mode) Unmarshal(s string) error {
	hub := Mode(s)
	if err := hub.Validate(); err != nil {
		return err
	}
	*h = hub
	return nil
}

// String fulfils the core interface for string printing.
func (h *Mode) String() string {
	if h == nil {
		return "Hub - unknown"
	}
	return "Hub - " + string(*h)
}
