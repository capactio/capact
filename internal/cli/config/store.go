package config

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	ConfigPath = "$HOME/.config/capact"
)

const (
	defaultContextKey              = "defaultContext"
	credentialsStoreBackendKey     = "credentialsStore.backend"
	credentialsStoreFilePassphrase = "credentialsStore.filePassphrase"
)

func Init(configPath string) error {
	err := viper.BindEnv(credentialsStoreBackendKey, "CAPACT_CREDENTIALS_STORE_BACKEND")
	if err != nil {
		return errors.Wrapf(err, "while binding %s key", credentialsStoreBackendKey)
	}

	err = viper.BindEnv(credentialsStoreFilePassphrase, "CAPACT_CREDENTIALS_STORE_FILE_PASSPHRASE")
	if err != nil {
		return errors.Wrapf(err, "while binding %s key", credentialsStoreFilePassphrase)
	}

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		path, err := getConfigPath()
		if err != nil {
			return errors.Wrap(err, "while getting default config path")
		}
		viper.AddConfigPath(path)
	}

	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		err = os.MkdirAll(configPath, 0700)
		if err != nil {
			return errors.Wrap(err, "while creating directory for config file")
		}

		err = viper.WriteConfig()
		if err != nil {
			return errors.Wrap(err, "while writing config file")
		}
	} else if os.IsNotExist(err) {
		err = viper.WriteConfig()
		if err != nil {
			return errors.Wrap(err, "while writing config file")
		}
	} else {
		return errors.Wrap(err, "while reading configuration")
	}

	return nil
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, ".config", "capact"), nil
}

func SetAsDefaultContext(server string, override bool) error {
	currentDefaultContext := GetDefaultContext()

	if currentDefaultContext == "" || override {
		viper.Set(defaultContextKey, server)

		if err := viper.WriteConfig(); err != nil {
			return errors.Wrap(err, "while writing config file")
		}
	}

	return nil
}

func GetDefaultContext() string {
	return viper.GetString(defaultContextKey)
}

func GetCredentialsStoreBackend() string {
	return viper.GetString(credentialsStoreBackendKey)
}

func GetCredentialsStoreFilePassphrase() string {
	return viper.GetString(credentialsStoreFilePassphrase)
}
