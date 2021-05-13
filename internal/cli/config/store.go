package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	ConfigPath = "$HOME/.config/capact"
)

const (
	defaultContextKey = "defaultContext"
)

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
