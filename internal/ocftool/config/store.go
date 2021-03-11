package config

import (
	"github.com/99designs/keyring"
	"os"
)

// TODO: current hack to do not play with `.config` directory. Needs to be fixed!
var keyringConfigDefaults = keyring.Config{
	ServiceName:              "config-vault",
	LibSecretCollectionName:  "configvault",
	KWalletAppID:             "config-vault",
	KWalletFolder:            "config-vault",
	KeychainTrustApplication: true,
	WinCredPrefix:            "config-vault",
}

const configStoreName = "voltron-config"

func SetAsDefaultContext(server string, override bool) error {
	ks, err := keyring.Open(config())
	if err != nil {
		return err
	}

	currentServer, err := getDefaultContext(ks)
	if err != nil {
		return err
	}
	if currentServer == "" || override {
		return ks.Set(keyring.Item{
			Key:  configStoreName,
			Data: []byte(server),
		})
	}

	return nil
}

func GetDefaultContext() (string, error) {
	ks, err := keyring.Open(config())
	if err != nil {
		return "", err
	}

	return getDefaultContext(ks)
}

func getDefaultContext(ks keyring.Keyring) (string, error) {
	item, err := ks.Get(configStoreName)
	switch {
	case err == nil:
		return string(item.Data), nil
	case err == keyring.ErrKeyNotFound:
		return "", nil
	default:
		return "", err
	}
}

func config() keyring.Config {
	if backend := os.Getenv("CAPECTL_CREDENTIALS_STORE_BACKEND"); backend != "" {
		keyringConfigDefaults.AllowedBackends = []keyring.BackendType{keyring.BackendType(backend)}
	}
	return keyringConfigDefaults
}
