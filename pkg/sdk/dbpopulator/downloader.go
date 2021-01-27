package dbpopulator

import (
	"context"
	"os"
	"sync"

	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
)

func Download(ctx context.Context, src string, dst string) error {
	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Error getting pwd")
	}
	myContext, cancel := context.WithCancel(ctx)

	// Build the client
	client := &getter.Client{
		Ctx:  myContext,
		Src:  src,
		Dst:  dst,
		Pwd:  pwd,
		Mode: getter.ClientModeDir,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.Get(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-myContext.Done():
		wg.Wait()
	case err := <-errChan:
		wg.Wait()
		return err
	}
	return nil
}
