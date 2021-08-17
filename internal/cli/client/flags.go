package client

import (
	"time"

	"github.com/spf13/pflag"
)

var timeout = 30 * time.Second

// RegisterFlags registers client terminal flags.
// TODO: consider adding skip TLS verification for the HTTP server.
func RegisterFlags(flags *pflag.FlagSet) {
	flags.DurationVar(&timeout, "timeout", timeout, "Timeout for HTTP request")
}
