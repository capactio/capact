//go:build darwin
// +build darwin

package credstore

import (
	"fmt"

	"capact.io/capact/internal/cli/config"
	"github.com/99designs/keyring"
	"github.com/pkg/errors"
)

var (
	serviceName = fmt.Sprintf("%s-vault", Name)
)

func openStore() (Keyring, error) {
	backend := config.GetCredentialsStoreBackend()

	cfg := keyring.Config{
		ServiceName:      serviceName,
		WinCredPrefix:    serviceName,
		FileDir:          fmt.Sprintf("~/.config/capact/%s_vault", Name),
		FilePasswordFunc: fileKeyringPassphrasePrompt,
	}

	switch backend {
	case "", "keychain":
		return &Keychain{}, nil
	case "file", "pass":
		return keyring.Open(cfg)
	default:
		return nil, errors.New("backend not supported")
	}
}
