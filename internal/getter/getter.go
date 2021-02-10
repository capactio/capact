package getter

import (
	"context"
	"os"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
)

func Download(ctx context.Context, src string, dst string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Error getting pwd")
	}

	// Build the client
	client := &getter.Client{
		Ctx:  ctx,
		Src:  src,
		Dst:  dst,
		Pwd:  pwd,
		Mode: getter.ClientModeDir,
	}

	return client.Get()
}
