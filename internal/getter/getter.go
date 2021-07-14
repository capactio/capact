package getter

import (
	"context"
	"os"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
)

// Download downloads data from a given source to local file system under a given destination path.
func Download(ctx context.Context, src string, dst string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "while getting pwd")
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
