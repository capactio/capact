package credstore

import (
	b64 "encoding/base64"
	"encoding/json"
	"os"

	configstore "projectvoltron.dev/voltron/internal/ocftool/config"

	"github.com/99designs/keyring"
)

type Credentials struct {
	Username string
	Secret   string
}

func GetHub(serverURL string) (*Credentials, error) {
	ks, err := keyring.Open(config())
	if err != nil {
		return nil, err
	}

	item, err := ks.Get(b64.StdEncoding.EncodeToString([]byte(serverURL)))
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
	ks, err := keyring.Open(config())
	if err != nil {
		return err
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	return ks.Set(keyring.Item{
		Key:  b64.StdEncoding.EncodeToString([]byte(serverURL)),
		Data: data,
	})
}

func DeleteHub(serverURL string) error {
	ks, err := keyring.Open(config())
	if err != nil {
		return err
	}

	return ks.Remove(b64.StdEncoding.EncodeToString([]byte(serverURL)))
}

func ListHubServer() ([]string, error) {
	ks, err := keyring.Open(config())
	if err != nil {
		return nil, err
	}

	keys, err := ks.Keys()
	if err != nil {
		return nil, err
	}

	var out []string
	for _, k := range keys {
		if k == configstore.StoreName {
			continue
		}
		dec, err := b64.StdEncoding.DecodeString(k)
		if err != nil {
			return nil, err
		}
		out = append(out, string(dec))
	}
	return out, nil
}

const overrideBackend = "CAPECTL_CREDENTIALS_STORE_BACKEND"

var cfg = keyring.Config{
	ServiceName:              "hub-vault",
	LibSecretCollectionName:  "hubvault",
	KWalletAppID:             "hub-vault",
	KWalletFolder:            "hub-vault",
	KeychainTrustApplication: true,
	WinCredPrefix:            "hub-vault",
}

func config() keyring.Config {
	if backend := os.Getenv(overrideBackend); backend != "" {
		cfg.AllowedBackends = []keyring.BackendType{keyring.BackendType(backend)}
	}
	return cfg
}
