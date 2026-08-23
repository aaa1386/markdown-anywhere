//go:build windows

package launcher

import (
	"fmt"
	"syscall"
	"unsafe"
)

const showNormal = 1

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

// Open asks the Windows shell to open the validated obsidian:// URI.
func Open(uri string) error {
	if err := ValidateURI(uri); err != nil {
		return err
	}
	verb, err := syscall.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	target, err := syscall.UTF16PtrFromString(uri)
	if err != nil {
		return err
	}
	result, _, callErr := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(target)),
		0,
		0,
		showNormal,
	)
	if result > 32 {
		return nil
	}
	if callErr != nil && callErr != syscall.Errno(0) {
		return fmt.Errorf("ShellExecuteW failed: %w", callErr)
	}
	return fmt.Errorf("ShellExecuteW rejected URI with code %d", result)
}
