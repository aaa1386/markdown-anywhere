//go:build windows

package registry

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	hkeyCurrentUser   = 0x80000001
	keyQueryValue     = 0x0001
	keySetValue       = 0x0002
	keyCreateSubKey   = 0x0004
	keyEnumerate      = 0x0008
	keyDelete         = 0x00010000
	keyRead           = 0x20019
	keyWrite          = 0x20006
	keyAllAccess      = 0xF003F
	regOptionNonVol   = 0x00000000
	regSZ             = 1
	regExpandSZ       = 2
	errorSuccess      = 0
	errorFileNotFound = 2
)

var (
	advapi32            = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyExW   = advapi32.NewProc("RegOpenKeyExW")
	procRegCreateKeyExW = advapi32.NewProc("RegCreateKeyExW")
	procRegCloseKey     = advapi32.NewProc("RegCloseKey")
	procRegQueryValueEx = advapi32.NewProc("RegQueryValueExW")
	procRegSetValueEx   = advapi32.NewProc("RegSetValueExW")
	procRegDeleteValue  = advapi32.NewProc("RegDeleteValueW")
	procRegDeleteTree   = advapi32.NewProc("RegDeleteTreeW")
)

type win32Error struct {
	operation string
	code      uintptr
}

func (e win32Error) Error() string {
	return fmt.Sprintf("%s failed with Windows error %d", e.operation, e.code)
}

type windowsStore struct{}

func newStore() (Store, error) {
	return windowsStore{}, nil
}

func (windowsStore) Read(keyPath, valueName string) (Value, error) {
	key, err := openKey(keyPath, keyQueryValue)
	if err != nil {
		if isCode(err, errorFileNotFound) {
			return Value{}, nil
		}
		return Value{}, err
	}
	defer closeKey(key)

	valuePtr, err := valueNamePtr(valueName)
	if err != nil {
		return Value{}, err
	}
	var valueType uint32
	var size uint32
	code, _, _ := procRegQueryValueEx.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(valuePtr)),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		0,
		uintptr(unsafe.Pointer(&size)),
	)
	if code == errorFileNotFound {
		return Value{KeyExists: true}, nil
	}
	if code != errorSuccess {
		return Value{}, win32Error{operation: "RegQueryValueExW", code: code}
	}
	if valueType != regSZ && valueType != regExpandSZ {
		return Value{}, fmt.Errorf("registry value %s has unsupported type %d", keyPath, valueType)
	}

	buffer := make([]uint16, (size+1)/2)
	if len(buffer) == 0 {
		buffer = make([]uint16, 1)
	}
	code, _, _ = procRegQueryValueEx.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(valuePtr)),
		0,
		uintptr(unsafe.Pointer(&valueType)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if code != errorSuccess {
		return Value{}, win32Error{operation: "RegQueryValueExW", code: code}
	}
	return Value{Exists: true, KeyExists: true, Data: syscall.UTF16ToString(buffer)}, nil
}

func (windowsStore) Set(keyPath, valueName, value string) error {
	key, err := createKey(keyPath)
	if err != nil {
		return err
	}
	defer closeKey(key)

	valuePtr, err := valueNamePtr(valueName)
	if err != nil {
		return err
	}
	data, err := syscall.UTF16FromString(value)
	if err != nil {
		return fmt.Errorf("encode registry value: %w", err)
	}
	code, _, _ := procRegSetValueEx.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(valuePtr)),
		0,
		regSZ,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)*2),
	)
	if code != errorSuccess {
		return win32Error{operation: "RegSetValueExW", code: code}
	}
	return nil
}

func (windowsStore) DeleteValue(keyPath, valueName string) error {
	key, err := openKey(keyPath, keySetValue)
	if err != nil {
		if isCode(err, errorFileNotFound) {
			return nil
		}
		return err
	}
	defer closeKey(key)

	valuePtr, err := valueNamePtr(valueName)
	if err != nil {
		return err
	}
	code, _, _ := procRegDeleteValue.Call(uintptr(key), uintptr(unsafe.Pointer(valuePtr)))
	if code == errorFileNotFound || code == errorSuccess {
		return nil
	}
	return win32Error{operation: "RegDeleteValueW", code: code}
}

func (windowsStore) DeleteTree(keyPath string) error {
	parentPath, childName := splitKey(keyPath)
	var parent syscall.Handle
	var closeParent bool
	if parentPath == "" {
		parent = syscall.Handle(hkeyCurrentUser)
	} else {
		var err error
		parent, err = openKey(parentPath, keyAllAccess|keyEnumerate|keyDelete)
		if err != nil {
			if isCode(err, errorFileNotFound) {
				return nil
			}
			return err
		}
		closeParent = true
	}
	if closeParent {
		defer closeKey(parent)
	}
	childPtr, err := syscall.UTF16PtrFromString(childName)
	if err != nil {
		return err
	}
	code, _, _ := procRegDeleteTree.Call(uintptr(parent), uintptr(unsafe.Pointer(childPtr)))
	if code == errorFileNotFound || code == errorSuccess {
		return nil
	}
	return win32Error{operation: "RegDeleteTreeW", code: code}
}

func openKey(path string, access uint32) (syscall.Handle, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var key syscall.Handle
	code, _, _ := procRegOpenKeyExW.Call(
		uintptr(hkeyCurrentUser),
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		uintptr(access),
		uintptr(unsafe.Pointer(&key)),
	)
	if code != errorSuccess {
		return 0, win32Error{operation: "RegOpenKeyExW", code: code}
	}
	return key, nil
}

func createKey(path string) (syscall.Handle, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var key syscall.Handle
	code, _, _ := procRegCreateKeyExW.Call(
		uintptr(hkeyCurrentUser),
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		0,
		regOptionNonVol,
		uintptr(keyRead|keyWrite),
		0,
		uintptr(unsafe.Pointer(&key)),
		0,
	)
	if code != errorSuccess {
		return 0, win32Error{operation: "RegCreateKeyExW", code: code}
	}
	return key, nil
}

func closeKey(key syscall.Handle) {
	_, _, _ = procRegCloseKey.Call(uintptr(key))
}

func valueNamePtr(valueName string) (*uint16, error) {
	if valueName == "" {
		return nil, nil
	}
	return syscall.UTF16PtrFromString(valueName)
}

func splitKey(path string) (string, string) {
	index := strings.LastIndex(path, `\`)
	if index < 0 {
		return "", path
	}
	return path[:index], path[index+1:]
}

func isCode(err error, code uintptr) bool {
	var winErr win32Error
	return errors.As(err, &winErr) && winErr.code == code
}
