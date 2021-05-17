package credstore

import (
	b64 "encoding/base64"
	"encoding/json"

	"github.com/99designs/keyring"
)

type Credentials struct {
	Username string
	Secret   string
}

func GetHub(serverURL string) (*Credentials, error) {
	ks, err := keyring.Open(Config(CredStoreName))
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
	ks, err := keyring.Open(Config(CredStoreName))
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
	ks, err := keyring.Open(Config(CredStoreName))
	if err != nil {
		return err
	}

	return ks.Remove(b64.StdEncoding.EncodeToString([]byte(serverURL)))
}

func ListHubServer() ([]string, error) {
	ks, err := keyring.Open(Config(CredStoreName))
	if err != nil {
		return nil, err
	}

	keys, err := ks.Keys()
	if err != nil {
		return nil, err
	}

	var out []string
	for _, k := range keys {
		dec, err := b64.StdEncoding.DecodeString(k)
		if err != nil {
			return nil, err
		}
		out = append(out, string(dec))
	}
	return out, nil
}
