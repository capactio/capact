package config

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	ConfigPath = "$HOME/.config/capact"
)

const (
	defaultContextKey = "defaultContext"
)

func ReadConfig() error {
	viper.AddConfigPath(ConfigPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfig(); err != nil {
				return err
			}
		} else {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

func WriteConfig() error {
	return viper.WriteConfig()
}

func SetAsDefaultContext(server string, override bool) error {
	currentDefaultContext, _ := GetDefaultContext()

	if currentDefaultContext == "" || override {
		viper.Set(defaultContextKey, server)
	}

	if err := WriteConfig(); err != nil {
		return errors.Wrap(err, "while writing config file")
	}

	return nil
}

func GetDefaultContext() (string, error) {
	return viper.GetString(defaultContextKey), nil
}
