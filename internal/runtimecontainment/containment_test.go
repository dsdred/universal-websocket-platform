package runtimecontainment

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func TestParseDomainRejectsNonCanonicalAndZero(t *testing.T) {
	valid := strings.Repeat("01", identitySize)
	if _, err := ParseDomain(valid); err != nil {
		t.Fatalf("valid domain: %v", err)
	}
	for _, value := range []string{"", strings.Repeat("0", identitySize*2), strings.Repeat("AB", identitySize), valid[:63]} {
		if _, err := ParseDomain(value); err == nil {
			t.Fatalf("ParseDomain(%q) succeeded", value)
		}
	}
}

func TestDescriptorRejectsWrongDomainAndRelativeRoot(t *testing.T) {
	domain, _ := ParseDomain(strings.Repeat("01", identitySize))
	valid := LocalConfig{ExpectedDomain: domain.text(), RootDir: t.TempDir(),
		ExpectedCanonicalRoot: "canonical", ExpectedPhysicalRootIdentity: "physical",
		StorageAuthorityID: strings.Repeat("02", identitySize)}
	wrongDomain := valid
	wrongDomain.ExpectedDomain = strings.Repeat("03", identitySize)
	if _, ok := parseDescriptor(domain, wrongDomain); ok {
		t.Fatal("descriptor bound to another domain accepted")
	}
	relative := valid
	relative.RootDir = filepath.Join("relative", "root")
	if _, ok := parseDescriptor(domain, relative); ok {
		t.Fatal("relative RootDir accepted")
	}
}

func TestPublicTypesExposeNoReleaseOrReacquireSurface(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeOf(Result{}), reflect.TypeOf(&ActiveAuthority{})} {
		for _, forbidden := range []string{"Close", "Release", "Unlock", "Transfer", "Renew", "Reacquire"} {
			if _, ok := typ.MethodByName(forbidden); ok {
				t.Fatalf("%v exposes %s", typ, forbidden)
			}
		}
	}
}

func TestUnsupportedPlatformIsUnavailable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows has the real adapter")
	}
	domain, _ := ParseDomain(strings.Repeat("01", identitySize))
	result := AcquireProcessContainment(domain, LocalConfig{ExpectedDomain: domain.text(), RootDir: t.TempDir(), ExpectedCanonicalRoot: "unused",
		ExpectedPhysicalRootIdentity: "unused", StorageAuthorityID: strings.Repeat("01", identitySize)})
	if !result.IsUnavailable() || result.IsFatalFenced() {
		t.Fatalf("unexpected unsupported result: %+v", result)
	}
}

func TestLedgerExpectedTailRetryAndCommittedInspection(t *testing.T) {
	requireWindows(t)
	domain := testDomain(t)
	config := provisionForTest(t, domain, t.TempDir())
	descriptor, ok := parseDescriptor(domain, config)
	if !ok {
		t.Fatal("test descriptor invalid")
	}
	db, err := openExistingProvisioned(domain, descriptor)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	candidate, _ := newGeneration()
	calls := 0
	err = bootstrapWithAppend(db, domain, candidate, func(db *bolt.DB, d Domain, expected generation, empty bool, got generation) error {
		calls++
		if got != candidate {
			t.Fatal("candidate changed across retry")
		}
		if calls == 1 {
			return errors.New("definite absence")
		}
		return appendExpected(db, d, expected, empty, got)
	})
	if err != nil || calls != 2 {
		t.Fatalf("exact retry: calls=%d err=%v", calls, err)
	}

	next, _ := newGeneration()
	calls = 0
	err = bootstrapWithAppend(db, domain, next, func(db *bolt.DB, d Domain, expected generation, empty bool, got generation) error {
		calls++
		if err := appendExpected(db, d, expected, empty, got); err != nil {
			return err
		}
		return errors.New("indeterminate acknowledgement")
	})
	if err != nil || calls != 1 {
		t.Fatalf("committed convergence: calls=%d err=%v", calls, err)
	}
}

func TestLedgerRejectsReuseAndCorruption(t *testing.T) {
	requireWindows(t)
	domain := testDomain(t)
	config := provisionForTest(t, domain, t.TempDir())
	descriptor, _ := parseDescriptor(domain, config)
	db, err := openExistingProvisioned(domain, descriptor)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	candidate, _ := newGeneration()
	if err := bootstrapSuccessor(db, domain, candidate); err != nil {
		t.Fatal(err)
	}
	if err := appendExpected(db, domain, candidate, false, candidate); err == nil {
		t.Fatal("candidate reuse accepted")
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(stateBucket).Put(tailKey, bytes.Repeat([]byte{0xff}, identitySize))
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := readTail(db, domain); err == nil {
		t.Fatal("corrupt tail accepted")
	}
}

func TestExactInspectionRejectsCandidateWithDifferentPredecessor(t *testing.T) {
	requireWindows(t)
	domain := testDomain(t)
	config := provisionForTest(t, domain, t.TempDir())
	descriptor, _ := parseDescriptor(domain, config)
	db, err := openExistingProvisioned(domain, descriptor)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	first, _ := newGeneration()
	second, _ := newGeneration()
	candidate, _ := newGeneration()
	if err = appendExpected(db, domain, generation{}, true, first); err != nil {
		t.Fatal(err)
	}
	if err = appendExpected(db, domain, first, false, second); err != nil {
		t.Fatal(err)
	}
	if err = appendExpected(db, domain, second, false, candidate); err != nil {
		t.Fatal(err)
	}
	if _, err = inspectExactAppend(db, domain, first, false, candidate); err == nil {
		t.Fatal("candidate with a different predecessor was accepted as exact append")
	}
}

func requireWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only adapter")
	}
}

func testDomain(t *testing.T) Domain {
	t.Helper()
	candidate, err := newGeneration()
	if err != nil {
		t.Fatal(err)
	}
	return Domain{value: candidate.value}
}

func provisionForTest(t *testing.T, domain Domain, root string) LocalConfig {
	t.Helper()
	canonical, physical, err := inspectRoot(root)
	if err != nil {
		t.Fatalf("inspect root: %v", err)
	}
	authority, _ := newGeneration()
	directory := filepath.Join(root, "runtime-containment")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := ledgerPath(root, domain)
	db, err := bolt.Open(path, 0o600, &bolt.Options{NoSync: false, NoGrowSync: false})
	if err != nil {
		t.Fatalf("provision open: %v", err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_, fileIdentity, err := inspectFile(file)
	_ = file.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		anchor, err := tx.CreateBucket(anchorBucket)
		if err != nil {
			return err
		}
		if _, err = tx.CreateBucket(generationBucket); err != nil {
			return err
		}
		if _, err = tx.CreateBucket(stateBucket); err != nil {
			return err
		}
		values := map[string][]byte{
			string(domainKey): domain.value[:], string(canonicalRootKey): []byte(canonical),
			string(physicalRootKey): []byte(physical), string(storageAuthorityKey): authority.value[:],
			string(ledgerIdentityKey): []byte(fileIdentity),
		}
		for key, value := range values {
			if err = anchor.Put([]byte(key), value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return LocalConfig{ExpectedDomain: domain.text(), RootDir: root, ExpectedCanonicalRoot: canonical,
		ExpectedPhysicalRootIdentity: physical, StorageAuthorityID: authorityText(authority)}
}

func authorityText(value generation) string {
	const digits = "0123456789abcdef"
	result := make([]byte, identitySize*2)
	for i, b := range value.value {
		result[i*2], result[i*2+1] = digits[b>>4], digits[b&15]
	}
	return string(result)
}
