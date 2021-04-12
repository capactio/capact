package credstore

import (
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"github.com/AlecAivazis/survey/v2"
)

const (
	// TODO: current hack to do not play with `.config` directory. Needs to be fixed!
	// defined here to avoid import cycle
	ConfigStoreName = "capact-config"

	CredStoreName   = "capact-hub"
	OverrideBackend = "CAPACT_CREDENTIALS_STORE_BACKEND"
	// #nosec G101
	FileBackendPassphrase = "CAPACT_FILE_PASSPHRASE"
)

func Config(prefix string) keyring.Config {
	cfg := keyring.Config{
		ServiceName:              fmt.Sprintf("%s-vault", prefix),
		LibSecretCollectionName:  fmt.Sprintf("%svault", prefix),
		KWalletAppID:             fmt.Sprintf("%s-vault", prefix),
		KWalletFolder:            fmt.Sprintf("%s-vault", prefix),
		WinCredPrefix:            fmt.Sprintf("%s-vault", prefix),
		FileDir:                  fmt.Sprintf("~/.capectl/%s_vault", prefix),
		KeychainTrustApplication: true,
		FilePasswordFunc:         FileKeyringPassphrasePrompt,
	}
	if backend := os.Getenv(OverrideBackend); backend != "" {
		cfg.AllowedBackends = []keyring.BackendType{keyring.BackendType(backend)}
	}
	return cfg
}

func FileKeyringPassphrasePrompt(promptMessage string) (string, error) {
	password := os.Getenv(FileBackendPassphrase)
	if password != "" {
		return password, nil
	}

	err := survey.AskOne(&survey.Password{
		Message: promptMessage,
	}, &password)
	return password, err
}
