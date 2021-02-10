package runner

import (
	"io/ioutil"

	"github.com/pkg/errors"
)

const DefaultFilePermissions = 0644

func SaveToFile(path string, bytes []byte) error {
	err := ioutil.WriteFile(path, bytes, DefaultFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "while writing file to %q", path)
	}
	return nil
}
