package config

import (
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	availableContextsKey           = "availableContexts"
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
			return errors.Wrap(err, "while writing default context into config file")
		}
	}

	return nil
}

// GetDefaultContext returns default Hub server URL.
func GetDefaultContext() string {
	return viper.GetString(defaultContextKey)
}

// AddNewContext adds a new context if not exists to the collection of available contexts.
func AddNewContext(server string) error {
	availableContexts := GetAvailableContexts()
	if err := storeAvailableContexts(appendContextIfMissing(availableContexts, server)); err != nil {
		return errors.Wrap(err, "while setting and writing a new context")
	}
	return nil
}

// DeleteContext delete a context from the the collection of available contexts.
func DeleteContext(server string) error {
	availableContexts := GetAvailableContexts()
	for index, context := range availableContexts {
		if context == server {
			availableContexts = append(availableContexts[:index], availableContexts[index+1:]...)
		}
	}
	if err := storeAvailableContexts(availableContexts); err != nil {
		return errors.Wrap(err, "while setting and writing available contexts")
	}
	return nil
}

// GetAvailableContexts return collection of available contexts.
func GetAvailableContexts() []string {
	return viper.GetStringSlice(availableContextsKey)
}

// GetCredentialsStoreBackend returns keyring backend type.
func GetCredentialsStoreBackend() string {
	return viper.GetString(credentialsStoreBackendKey)
}

// GetCredentialsStoreFilePassphrase returns passphrase for file keyring backend type.
func GetCredentialsStoreFilePassphrase() string {
	return viper.GetString(credentialsStoreFilePassphrase)
}

func storeAvailableContexts(contexts []string) error {
	viper.Set(availableContextsKey, contexts)
	if err := viper.WriteConfig(); err != nil {
		return errors.Wrap(err, "while writing available contexts into config file")
	}
	return nil
}

func appendContextIfMissing(contexts []string, newContext string) []string {
	for _, context := range contexts {
		if context == newContext {
			return contexts
		}
	}
	return append(contexts, newContext)
}
