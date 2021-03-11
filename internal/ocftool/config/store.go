package config

import (
	"github.com/docker/docker-credential-helpers/credentials"
	"projectvoltron.dev/voltron/internal/ocftool/credstore"
)

// TODO: current hack to do not play with `.config` directory. Needs to be fixed!

const configStoreName = "voltron-config"

func SetAsDefaultContext(server string, override bool) error {
	store := credstore.NewConfig()
	currentDefault, _, _ := store.Get(configStoreName)

	if currentDefault == "" || override {
		return store.Add(&credentials.Credentials{
			ServerURL: configStoreName,
			Username:  server,
			Secret:    server,
		})
	}

	return nil
}

func GetDefaultContext() string {
	store := credstore.NewConfig()
	currentDefault, _, _ := store.Get(configStoreName)
	return currentDefault
}
