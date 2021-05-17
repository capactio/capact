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
		configPath, err = getDefaultConfigPath()
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
	} else {
		return errors.Wrap(err, "while reading configuration")
	}

	return nil
}

func getDefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(homeDir, ".config", "capact", "config.yaml"), nil
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
