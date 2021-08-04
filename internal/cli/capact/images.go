package capact

import (
	"context"
)

func LoadImages(ctx context.Context, envType string, envName string, images []string) error {
	switch envType {
	case KindEnv:
	//	for _, img := range images {
	//		if err := LoadKindImage(envName, img); err != nil {
	//			return errors.Wrap(err, "while loading images into kind environment")
	//		}
	//	}
	//case K3dEnv:
		return LoadK3dImages(ctx, envName, images)
	}

	return nil
}
