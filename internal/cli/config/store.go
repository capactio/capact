package config

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	defaultContextKey              = "defaultContext"
	credentialsStoreBackendKey     = "credentialsStore.backend"
	credentialsStoreFilePassphrase = "credentialsStore.filePassphrase"
)

// Init initializes config store for Capact CLI.
func Init(configPath string) error {
	err := viper.BindEnv(credentialsStoreBackendKey, "CAPACT_CREDENTIALS_STORE_BACKEND")
	if err != nil {
		return errors.Wrapf(err, "while binding %s key", credentialsStoreBackendKey)
	}

	err = viper.BindEnv(credentialsStoreFilePassphrase, "CAPACT_CREDENTIALS_STORE_FILE_PASSPHRASE")
	if err != nil {
		return errors.Wrapf(err, "while binding %s key", credentialsStoreFilePassphrase)
	}

	if configPath == "" {
		configPath, err = GetDefaultConfigPath("config.yaml")
		if err != nil {
			return errors.Wrap(err, "while getting default config path")
		}
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok || os.IsNotExist(err) {
		dir := path.Dir(configPath)

		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return errors.Wrap(err, "while creating directory for config file")
		}

		err = viper.WriteConfig()
		if err != nil {
			return errors.Wrap(err, "while writing config file")
		}
	} else if err != nil {
		return errors.Wrap(err, "while reading configuration")
	}

	return nil
}

// GetDefaultConfigPath returns Capact location for a given config file
func GetDefaultConfigPath(fileName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, ".config", "capact", fileName), nil
}

// SetAsDefaultContext sets default Hub server which is used for all executed operations.
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

// GetDefaultContext returns default Hub server URL.
func GetDefaultContext() string {
	return viper.GetString(defaultContextKey)
}

// GetCredentialsStoreBackend returns keyring backend type.
func GetCredentialsStoreBackend() string {
	return viper.GetString(credentialsStoreBackendKey)
}

// GetCredentialsStoreFilePassphrase returns passphrase for file keyring backend type.
func GetCredentialsStoreFilePassphrase() string {
	return viper.GetString(credentialsStoreFilePassphrase)
}
