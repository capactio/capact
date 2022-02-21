package credstore

import "github.com/99designs/keyring"

// Keyring provides the uniform interface over the underlying backends.
type Keyring interface {
	// Get returns an Item matching the key or ErrKeyNotFound
	Get(key string) (keyring.Item, error)
	// Set stores an Item on the keyring
	Set(item keyring.Item) error
	// Remove removes the item with matching key
	Remove(key string) error
}
