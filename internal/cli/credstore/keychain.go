//go:build darwin
// +build darwin

package credstore

import (
	"github.com/99designs/keyring"
	zkeyring "github.com/zalando/go-keyring"
	"fmt"
)

var (
	serviceName = fmt.Sprintf("%s-vault", Name)
)

// Keychain is a simple adapter to zalando go-keyring.
type Keychain struct{}

func (k Keychain) Get(key string) (keyring.Item, error) {
	data, err := zkeyring.Get(serviceName, key)
	if err != nil {
		return keyring.Item{}, err
	}
	return keyring.Item{
		Key:  key,
		Data: []byte(data),
	}, nil
}

func (k Keychain) Set(item keyring.Item) error {
	return zkeyring.Set(serviceName, item.Key, string(item.Data))
}

func (k Keychain) Remove(key string) error {
	return zkeyring.Delete(serviceName, key)
}
