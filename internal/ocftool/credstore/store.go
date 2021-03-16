package credstore

import (
	"encoding/json"

	"github.com/99designs/keyring"
)

var keyringConfigDefaults = keyring.Config{
	ServiceName:              "hub-vault",
	LibSecretCollectionName:  "hubvault",
	KWalletAppID:             "hub-vault",
	KWalletFolder:            "hub-vault",
	KeychainTrustApplication: true,
	WinCredPrefix:            "hub-vault",
}

type Credentials struct {
	Username string
	Secret   string
}

func GetHub(serverURL string) (*Credentials, error) {
	ks, err := keyring.Open(keyringConfigDefaults)
	if err != nil {
		return nil, err
	}

	item, err := ks.Get(serverURL)
	if err != nil {
		return nil, err
	}

	creds := Credentials{}
	if err := json.Unmarshal(item.Data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

func AddHub(serverURL string, creds Credentials) error {
	ks, err := keyring.Open(keyringConfigDefaults)
	if err != nil {
		return err
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	return ks.Set(keyring.Item{
		Key:  serverURL,
		Data: data,
	})
}

func DeleteHub(serverURL string) error {
	ks, err := keyring.Open(keyringConfigDefaults)
	if err != nil {
		return err
	}

	return ks.Remove(serverURL)
}

func ListHubServer() ([]string, error) {
	ks, err := keyring.Open(keyringConfigDefaults)
	if err != nil {
		return nil, err
	}

	return ks.Keys()
}
