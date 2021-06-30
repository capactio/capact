package runner

import (
	"io/ioutil"

	"github.com/pkg/errors"
)

// DefaultFilePermissions are the default file permissions
// of the output artifact files created by the runners.
const DefaultFilePermissions = 0644

// SaveToFile saves the bytes to a file under the path.
func SaveToFile(path string, bytes []byte) error {
	err := ioutil.WriteFile(path, bytes, DefaultFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "while writing file to %q", path)
	}
	return nil
}
