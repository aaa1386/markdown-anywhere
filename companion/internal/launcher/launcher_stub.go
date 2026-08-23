//go:build !windows

package launcher

import "errors"

var ErrUnsupported = errors.New("Obsidian launching is only available on Windows")

func Open(uri string) error {
	if err := ValidateURI(uri); err != nil {
		return err
	}
	return ErrUnsupported
}
