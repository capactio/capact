package credstore

import (
	"fmt"

	"capact.io/capact/internal/cli/config"
	"github.com/99designs/keyring"
	"github.com/AlecAivazis/survey/v2"
)

const (
	CredStoreName = "capacthub"
)

func Config(prefix string) keyring.Config {
	cfg := keyring.Config{
		ServiceName:              fmt.Sprintf("%s-vault", prefix),
		LibSecretCollectionName:  fmt.Sprintf("%svault", prefix),
		KWalletAppID:             fmt.Sprintf("%s-vault", prefix),
		KWalletFolder:            fmt.Sprintf("%s-vault", prefix),
		WinCredPrefix:            fmt.Sprintf("%s-vault", prefix),
		FileDir:                  fmt.Sprintf("~/.config/capact/%s_vault", prefix),
		KeychainTrustApplication: true,
		FilePasswordFunc:         FileKeyringPassphrasePrompt,
	}

	backend := config.GetCredentialsStoreBackend()
	if backend != "" {
		cfg.AllowedBackends = []keyring.BackendType{keyring.BackendType(backend)}
	}

	return cfg
}

func FileKeyringPassphrasePrompt(promptMessage string) (string, error) {
	password := config.GetCredentialsStoreFilePassphrase()
	if password != "" {
		return password, nil
	}

	err := survey.AskOne(&survey.Password{
		Message: promptMessage,
	}, &password)
	return password, err
}
