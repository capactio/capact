package archiveimages

import (
	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/capact"
	"github.com/docker/cli/cli/command"
	"github.com/pkg/errors"
)

// HelmArchiveImagesOptions holds options for Helm images archiver
type HelmArchiveImagesOptions struct {
	CapactOpts capact.Options
	Output     struct {
		Path     string
		ToStdout bool
	}
	Compress string
}

// Resolve resolves HelmArchiveImagesOptions to final form.
func (h *HelmArchiveImagesOptions) Resolve() error {
	err := h.CapactOpts.Parameters.ResolveVersion()
	if err != nil {
		return errors.Wrap(err, "while resolving version")
	}
	return nil
}

// Validate validates the HelmArchiveImagesOptions fields.
func (h *HelmArchiveImagesOptions) Validate() error {
	if !h.Output.ToStdout && h.Output.Path == "" {
		return errors.New("use either '--output' or '--output-stdout'")
	}
	if h.Output.ToStdout && cli.VerboseMode.IsEnabled() {
		return errors.New("cannot use verbose mode with '--output-stdout'")
	}

	if h.Output.Path != "" {
		if err := command.ValidateOutputPath(h.Output.Path); err != nil {
			return errors.Wrap(err, "invalid '--output' flag")
		}
	}

	return nil
}
