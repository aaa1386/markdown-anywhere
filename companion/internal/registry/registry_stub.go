//go:build !windows

package registry

import "errors"

var ErrUnsupported = errors.New("Windows registry integration is only available on Windows")

func newStore() (Store, error) {
	return nil, ErrUnsupported
}
