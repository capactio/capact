package credstore

import (
	b64 "encoding/base64"
	"encoding/json"

	"capact.io/capact/internal/cli/config"
	"github.com/99designs/keyring"
	"github.com/AlecAivazis/survey/v2"
)

// Name defines Capact local credential store Name.
const Name = "capacthub"

// Credentials holds credentials details.
type Credentials struct {
	Username string
	Secret   string
}

// GetHub returns Public Hub credentials associated with a given URL.
func GetHub(serverURL string) (*Credentials, error) {
	ks, err := openStore()
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

// AddHub saves and associates Public Hub credentials with a given URL.
func AddHub(serverURL string, creds Credentials) error {
	ks, err := openStore()
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

// DeleteHub removes credentials associates with a given Public Hub URL.
func DeleteHub(serverURL string) error {
	ks, err := openStore()
	if err != nil {
		return err
	}

	return ks.Remove(b64.StdEncoding.EncodeToString([]byte(serverURL)))
}

func fileKeyringPassphrasePrompt(promptMessage string) (string, error) {
	password := config.GetCredentialsStoreFilePassphrase()
	if password != "" {
		return password, nil
	}

	err := survey.AskOne(&survey.Password{
		Message: promptMessage,
	}, &password)
	return password, err
}
