package capact

import (
	"context"

	"capact.io/capact/internal/cli/printer"

	"github.com/pkg/errors"
)

// LoadImages loads Docker images into proper environment
func LoadImages(ctx context.Context, status *printer.Status, images []string, opts Options) error {
	switch opts.Environment {
	case KindEnv:
		status.Step("Loading Docker images into kind cluster")
		for _, img := range images {
			if err := LoadKindImage(opts.Name, img); err != nil {
				return errors.Wrap(err, "while loading images into kind environment")
			}
		}
	case K3dEnv:
		return LoadK3dImages(ctx, opts.Name, images)
	}

	return nil
}
