package runtimecontainment

import (
	"encoding/hex"
	"errors"
	"path/filepath"
	"strings"
)

var errInvalidDescriptor = errors.New("invalid containment descriptor")

type descriptor struct {
	root, canonicalRoot, physicalRoot string
	storageAuthority                  [identitySize]byte
}

func parseDescriptor(domain Domain, config LocalConfig) (descriptor, bool) {
	var d descriptor
	if zeroIdentity(domain.value) || config.ExpectedDomain != domain.text() || !filepath.IsAbs(config.RootDir) || config.ExpectedCanonicalRoot == "" ||
		config.ExpectedPhysicalRootIdentity == "" || len(config.StorageAuthorityID) != identitySize*2 ||
		config.StorageAuthorityID != strings.ToLower(config.StorageAuthorityID) {
		return d, false
	}
	decoded, err := hex.DecodeString(config.StorageAuthorityID)
	if err != nil {
		return d, false
	}
	copy(d.storageAuthority[:], decoded)
	if zeroIdentity(d.storageAuthority) {
		return descriptor{}, false
	}
	d.root = filepath.Clean(config.RootDir)
	d.canonicalRoot = config.ExpectedCanonicalRoot
	d.physicalRoot = strings.ToLower(config.ExpectedPhysicalRootIdentity)
	return d, true
}

func ledgerPath(root string, domain Domain) string {
	return filepath.Join(root, "runtime-containment", domain.text()+".db")
}
