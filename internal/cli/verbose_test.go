package cli_test

import (
	"testing"

	"capact.io/capact/internal/cli"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerboseModeFlag(t *testing.T) {
	tests := map[string]struct {
		givenRawOption string
		expectedMode   cli.VerboseModeFlag
	}{
		"Should parse simple verbose just by -v": {
			givenRawOption: "-v",
			expectedMode:   cli.VerboseModeSimple,
		},
		"Should parse simple verbose by id": {
			givenRawOption: "-v=1",
			expectedMode:   cli.VerboseModeSimple,
		},
		"Should parse tracing verbose by id": {
			givenRawOption: "-v=2",
			expectedMode:   cli.VerboseModeTracing,
		},
		"Should parse simple verbose by human name": {
			givenRawOption: "-v=simple",
			expectedMode:   cli.VerboseModeSimple,
		},
		"Should parse tracing verbose by human name": {
			givenRawOption: "-v=tracing",
			expectedMode:   cli.VerboseModeTracing,
		},
		"Should parse disable verbose by human name": {
			givenRawOption: "-v=disable",
			expectedMode:   cli.VerboseModeDisabled,
		},
		"Should parse disable if flag not provided": {
			givenRawOption: "",
			expectedMode:   cli.VerboseModeDisabled,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			old := cli.VerboseMode
			defer func() { cli.VerboseMode = old }()

			flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
			cli.RegisterVerboseModeFlag(flags)

			var args []string
			if tc.givenRawOption != "" {
				args = append(args, tc.givenRawOption)
			}

			// when
			err := flags.Parse(args)
			require.NoError(t, err)

			// then
			assert.Equal(t, tc.expectedMode, cli.VerboseMode)
		})
	}
}
