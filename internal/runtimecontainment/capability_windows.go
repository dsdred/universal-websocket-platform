//go:build windows

package runtimecontainment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

type heldCapability struct{ handle windows.Handle }

func tryAcquire(domain Domain) (heldCapability, bool, bool) {
	name, err := windows.UTF16PtrFromString(`Global\uwp-runtime-containment-` + domain.text())
	if err != nil {
		return heldCapability{}, false, false
	}
	handle, err := windows.CreateEvent(nil, 1, 0, name)
	if err != nil {
		if handle != 0 {
			_ = windows.CloseHandle(handle)
		}
		return heldCapability{}, false, false
	}
	capability := heldCapability{handle: handle}
	if windows.SetHandleInformation(handle, windows.HANDLE_FLAG_INHERIT, 0) != nil {
		return capability, true, false
	}
	return capability, true, validateHeld(capability)
}

func validateHeld(capability heldCapability) bool {
	if capability.handle == 0 {
		return false
	}
	event, err := windows.WaitForSingleObject(capability.handle, 0)
	return err == nil && event == uint32(windows.WAIT_TIMEOUT)
}

func inspectRoot(path string) (canonical, physical string, err error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", "", err
	}
	attributes, err := windows.GetFileAttributes(windows.StringToUTF16Ptr(abs))
	if err != nil || attributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 || attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return "", "", errInvalidDescriptor
	}
	volume := make([]uint16, windows.MAX_PATH+1)
	if windows.GetVolumePathName(windows.StringToUTF16Ptr(abs), &volume[0], uint32(len(volume))) != nil ||
		windows.GetDriveType(&volume[0]) != windows.DRIVE_FIXED {
		return "", "", errInvalidDescriptor
	}
	handle, err := windows.CreateFile(windows.StringToUTF16Ptr(abs), windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", "", err
	}
	defer windows.CloseHandle(handle)
	canonical, err = finalPath(handle)
	if err != nil {
		return "", "", err
	}
	physical, err = handleIdentity(handle)
	return strings.ToLower(canonical), physical, err
}

func inspectFile(file *os.File) (string, string, error) {
	canonical, err := finalPath(windows.Handle(file.Fd()))
	if err != nil {
		return "", "", err
	}
	identity, err := handleIdentity(windows.Handle(file.Fd()))
	return strings.ToLower(canonical), identity, err
}

func finalPath(handle windows.Handle) (string, error) {
	buffer := make([]uint16, windows.MAX_PATH+1)
	n, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], uint32(len(buffer)), 0)
	if err != nil {
		return "", err
	}
	if n >= uint32(len(buffer)) {
		buffer = make([]uint16, n+1)
		n, err = windows.GetFinalPathNameByHandle(handle, &buffer[0], uint32(len(buffer)), 0)
	}
	if err != nil {
		return "", err
	}
	return windows.UTF16ToString(buffer[:n]), nil
}

func handleIdentity(handle windows.Handle) (string, error) {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return "", err
	}
	return formatHandleIdentity(info)
}

func formatHandleIdentity(info windows.ByHandleFileInformation) (string, error) {
	if info.VolumeSerialNumber == 0 || (info.FileIndexHigh == 0 && info.FileIndexLow == 0) {
		return "", errInvalidDescriptor
	}
	return strings.ToLower(fmt.Sprintf("%08x:%08x%08x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow)), nil
}
