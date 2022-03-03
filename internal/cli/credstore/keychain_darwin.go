//go:build darwin
// +build darwin

package credstore

import (
	"github.com/99designs/keyring"
	"github.com/pkg/errors"
	zkeyring "github.com/zalando/go-keyring"
)

// Keychain is a simple adapter to Zalando go-keyring.
type Keychain struct{}

func (k Keychain) Get(key string) (keyring.Item, error) {
	data, err := zkeyring.Get(serviceName, key)
	if err != nil {
		return keyring.Item{}, errors.Wrap(err, "while getting data from the keyring")
	}
	return keyring.Item{
		Key:  key,
		Data: []byte(data),
	}, nil
}

func (k Keychain) Set(item keyring.Item) error {
	err := zkeyring.Set(serviceName, item.Key, string(item.Data))
	if err != nil {
		return errors.Wrap(err, "while setting data in the keyring")
	}
	return nil
}

func (k Keychain) Remove(key string) error {
	err := zkeyring.Delete(serviceName, key)
	if err != nil {
		return errors.Wrap(err, "while deleting data in the keyring")
	}
	return nil
}
