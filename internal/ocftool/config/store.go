package config

import (
	"projectvoltron.dev/voltron/internal/ocftool/credstore"

	"github.com/99designs/keyring"
)

func SetAsDefaultContext(server string, override bool) error {
	ks, err := keyring.Open(credstore.Config(credstore.ConfigStoreName))
	if err != nil {
		return err
	}

	currentServer, err := getDefaultContext(ks)
	if err != nil {
		return err
	}
	if currentServer == "" || override {
		return ks.Set(keyring.Item{
			Key:  credstore.ConfigStoreName,
			Data: []byte(server),
		})
	}

	return nil
}

func GetDefaultContext() (string, error) {
	ks, err := keyring.Open(credstore.Config(credstore.ConfigStoreName))
	if err != nil {
		return "", err
	}

	return getDefaultContext(ks)
}

func getDefaultContext(ks keyring.Keyring) (string, error) {
	item, err := ks.Get(credstore.ConfigStoreName)
	switch {
	case err == nil:
		return string(item.Data), nil
	case err == keyring.ErrKeyNotFound:
		return "", nil
	default:
		return "", err
	}
}
