package credstore

import (
	"github.com/docker/docker-credential-helpers/credentials"
)

const (
	ochLabel    = "och-store"
	configLabel = "config-store"
)

// Helper is the interface a credentials store helper must implement.
type Store interface {
	// Add appends credentials to the store.
	Add(credentials *credentials.Credentials) error
	// Delete removes credentials from the store.
	Delete(serverURL string) error
	// Get retrieves credentials from the store.
	// It returns username and secret as strings.
	Get(serverURL string) (string, string, error)
	// List returns the stored serverURLs and their associated usernames.
	List() (map[string]string, error)
}

// TODO: It is not thread save because docker uses global variable for labeling
// Mutex needs to be added
func NewOCH() Store {
	credentials.SetCredsLabel(ochLabel)
	return nativeStore
}

// TODO: It is not thread save because docker uses global variable for labeling
// Mutex needs to be added
func NewConfig() Store {
	credentials.SetCredsLabel(configLabel)
	return nativeStore
}
