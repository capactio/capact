//go:build !darwin
// +build !darwin

package credstore

import (
	"fmt"

	"capact.io/capact/internal/cli/config"
	"github.com/99designs/keyring"
)

func openStore() (Keyring, error) {
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
		cfg.AllowedBackends = []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KWalletBackend,
			keyring.WinCredBackend,
			keyring.FileBackend,
			keyring.PassBackend,
		}
	}

	return keyring.Open(cfg)
}
