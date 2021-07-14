package credstore

import (
	"fmt"

	"capact.io/capact/internal/cli/config"
	"github.com/99designs/keyring"
	"github.com/AlecAivazis/survey/v2"
)

// Name defines Capact local credential store Name.
const Name = "capacthub"

func openStore() (keyring.Keyring, error) {
	cfg := keyring.Config{
		ServiceName:              fmt.Sprintf("%s-vault", Name),
		LibSecretCollectionName:  fmt.Sprintf("%svault", Name),
		KWalletAppID:             fmt.Sprintf("%s-vault", Name),
		KWalletFolder:            fmt.Sprintf("%s-vault", Name),
		WinCredPrefix:            fmt.Sprintf("%s-vault", Name),
		FileDir:                  fmt.Sprintf("~/.config/capact/%s_vault", Name),
		KeychainTrustApplication: true,
		FilePasswordFunc:         fileKeyringPassphrasePrompt,
	}

	backend := config.GetCredentialsStoreBackend()
	if backend != "" {
		cfg.AllowedBackends = []keyring.BackendType{keyring.BackendType(backend)}
	}

	return keyring.Open(cfg)
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
