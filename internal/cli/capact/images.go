package capact

import (
	"context"
	"github.com/pkg/errors"
)

func LoadImages(ctx context.Context, images []string, opts Options) error {
	switch opts.Environment {
	case KindEnv:
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
