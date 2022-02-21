//go:build darwin
// +build darwin

package credstore

import (
	"capact.io/capact/internal/cli/config"
	"fmt"
	"github.com/99designs/keyring"
	"github.com/pkg/errors"
)

func openStore() (Keyring, error) {
	backend := config.GetCredentialsStoreBackend()

	cfg := keyring.Config{
		ServiceName:      fmt.Sprintf("%s-vault", Name),
		WinCredPrefix:    fmt.Sprintf("%s-vault", Name),
		FileDir:          fmt.Sprintf("~/.config/capact/%s_vault", Name),
		FilePasswordFunc: fileKeyringPassphrasePrompt,
	}

	switch backend {
	case "keychain":
		return &Keychain{}, nil
	case "file":
		return keyring.Open(cfg)
	case "pass":
		return keyring.Open(cfg)
	default:
		return nil, errors.New("not supported")
	}

	return nil, nil
}
