//go:build !windows

package runtimecontainment

import "os"

type heldCapability struct{}

func tryAcquire(Domain) (heldCapability, bool, bool) { return heldCapability{}, false, false }
func validateHeld(heldCapability) bool               { return false }
func inspectRoot(string) (string, string, error) {
	return "", "", errInvalidDescriptor
}
func inspectFile(*os.File) (string, string, error) { return "", "", errInvalidDescriptor }
