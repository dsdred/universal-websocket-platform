package runtimecontainment

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	anchorBucket        = []byte("anchor-v1")
	generationBucket    = []byte("generations-v1")
	stateBucket         = []byte("state-v1")
	domainKey           = []byte("domain")
	canonicalRootKey    = []byte("canonical-root")
	physicalRootKey     = []byte("physical-root")
	storageAuthorityKey = []byte("storage-authority")
	ledgerIdentityKey   = []byte("ledger-file-identity")
	tailKey             = []byte("tail")
	errLedgerInvalid    = errors.New("invalid containment ledger")
)

func openExistingProvisioned(domain Domain, d descriptor) (*bolt.DB, error) {
	canonical, physical, err := inspectRoot(d.root)
	if err != nil || !strings.EqualFold(canonical, d.canonicalRoot) || physical != d.physicalRoot {
		return nil, errLedgerInvalid
	}
	path := ledgerPath(d.root, domain)
	expectedPath, err := validateLedgerLocation(path, canonical)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errLedgerInvalid
	}
	var openedCanonical, openedIdentity string
	options := &bolt.Options{Timeout: time.Second, NoSync: false, NoGrowSync: false}
	options.OpenFile = func(name string, _ int, _ os.FileMode) (*os.File, error) {
		file, openErr := os.OpenFile(name, os.O_RDWR, 0)
		if openErr != nil {
			return nil, openErr
		}
		openedCanonical, openedIdentity, openErr = inspectFile(file)
		if openErr != nil {
			_ = file.Close()
			return nil, openErr
		}
		return file, nil
	}
	db, err := bolt.Open(path, 0o600, options)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(openedCanonical, expectedPath) {
		return db, errLedgerInvalid
	}
	err = db.View(func(tx *bolt.Tx) error {
		return validateAnchor(tx, domain, d, canonical, physical, openedIdentity)
	})
	if err != nil {
		return db, err
	}
	return db, nil
}

func revalidateExistingProvisioned(db *bolt.DB, domain Domain, d descriptor) error {
	canonical, physical, err := inspectRoot(d.root)
	if err != nil || !strings.EqualFold(canonical, d.canonicalRoot) || physical != d.physicalRoot {
		return errLedgerInvalid
	}
	path := ledgerPath(d.root, domain)
	expectedPath, err := validateLedgerLocation(path, canonical)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()
	openedCanonical, identity, err := inspectFile(file)
	if err != nil {
		return err
	}
	if !strings.EqualFold(openedCanonical, expectedPath) {
		return errLedgerInvalid
	}
	return db.View(func(tx *bolt.Tx) error {
		return validateAnchor(tx, domain, d, canonical, physical, identity)
	})
}

func validateLedgerLocation(path, canonicalRoot string) (string, error) {
	directoryCanonical, _, err := inspectRoot(filepath.Dir(path))
	expectedDirectory := filepath.Join(canonicalRoot, "runtime-containment")
	if err != nil || !strings.EqualFold(directoryCanonical, expectedDirectory) {
		return "", errLedgerInvalid
	}
	return filepath.Join(expectedDirectory, filepath.Base(path)), nil
}

func validateAnchor(tx *bolt.Tx, domain Domain, d descriptor, canonical, physical, fileIdentity string) error {
	anchor := tx.Bucket(anchorBucket)
	if anchor == nil || tx.Bucket(generationBucket) == nil || tx.Bucket(stateBucket) == nil ||
		!onlyTopLevelBuckets(tx, anchorBucket, generationBucket, stateBucket) {
		return errLedgerInvalid
	}
	expected := map[string][]byte{
		string(domainKey): domain.value[:], string(canonicalRootKey): []byte(canonical),
		string(physicalRootKey): []byte(physical), string(storageAuthorityKey): d.storageAuthority[:],
		string(ledgerIdentityKey): []byte(fileIdentity),
	}
	if !exactValues(anchor, expected) || !bytes.Equal(anchor.Get(canonicalRootKey), []byte(d.canonicalRoot)) {
		return errLedgerInvalid
	}
	_, _, err := readTailTx(tx, domain)
	return err
}

func onlyTopLevelBuckets(tx *bolt.Tx, names ...[]byte) bool {
	want := make(map[string]bool, len(names))
	for _, name := range names {
		want[string(name)] = true
	}
	count := 0
	err := tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
		if !want[string(name)] {
			return errLedgerInvalid
		}
		count++
		return nil
	})
	return err == nil && count == len(want)
}

func exactValues(bucket *bolt.Bucket, expected map[string][]byte) bool {
	count := 0
	err := bucket.ForEach(func(key, value []byte) error {
		want, ok := expected[string(key)]
		if !ok || value == nil || !bytes.Equal(value, want) {
			return errLedgerInvalid
		}
		count++
		return nil
	})
	return err == nil && count == len(expected)
}

func readTail(db *bolt.DB, domain Domain) (generation, bool, error) {
	var tail generation
	var empty bool
	err := db.View(func(tx *bolt.Tx) error {
		var err error
		tail, empty, err = readTailTx(tx, domain)
		return err
	})
	return tail, empty, err
}

func readTailTx(tx *bolt.Tx, domain Domain) (generation, bool, error) {
	entries, state := tx.Bucket(generationBucket), tx.Bucket(stateBucket)
	if entries == nil || state == nil {
		return generation{}, false, errLedgerInvalid
	}
	rawTail := state.Get(tailKey)
	if rawTail == nil {
		if state.Stats().KeyN != 0 || entries.Stats().KeyN != 0 {
			return generation{}, false, errLedgerInvalid
		}
		return generation{}, true, nil
	}
	if len(rawTail) != identitySize || state.Stats().KeyN != 1 {
		return generation{}, false, errLedgerInvalid
	}
	var tail generation
	copy(tail.value[:], rawTail)
	if zeroIdentity(tail.value) {
		return generation{}, false, errLedgerInvalid
	}
	seen := map[[identitySize]byte]bool{}
	current := tail
	for {
		if seen[current.value] {
			return generation{}, false, errLedgerInvalid
		}
		seen[current.value] = true
		record := entries.Get(current.value[:])
		if len(record) != identitySize*2+1 || !bytes.Equal(record[:identitySize], domain.value[:]) {
			return generation{}, false, errLedgerInvalid
		}
		if record[identitySize] == 0 {
			if !zeroBytes(record[identitySize+1:]) {
				return generation{}, false, errLedgerInvalid
			}
			break
		}
		if record[identitySize] != 1 {
			return generation{}, false, errLedgerInvalid
		}
		copy(current.value[:], record[identitySize+1:])
		if zeroIdentity(current.value) {
			return generation{}, false, errLedgerInvalid
		}
	}
	if len(seen) != entries.Stats().KeyN {
		return generation{}, false, errLedgerInvalid
	}
	return tail, false, nil
}

func appendExpected(db *bolt.DB, domain Domain, expected generation, expectedEmpty bool, candidate generation) error {
	if zeroIdentity(candidate.value) {
		return errLedgerInvalid
	}
	return db.Update(func(tx *bolt.Tx) error {
		actual, empty, err := readTailTx(tx, domain)
		if err != nil || empty != expectedEmpty || (!empty && actual != expected) {
			return errLedgerInvalid
		}
		entries, state := tx.Bucket(generationBucket), tx.Bucket(stateBucket)
		if entries.Get(candidate.value[:]) != nil {
			return errLedgerInvalid
		}
		record := successorRecord(domain, expected, expectedEmpty)
		if err := entries.Put(candidate.value[:], record); err != nil {
			return err
		}
		return state.Put(tailKey, candidate.value[:])
	})
}

type inspection uint8

const (
	inspectionCommitted inspection = iota + 1
	inspectionAbsentExpected
)

func inspectExactAppend(db *bolt.DB, domain Domain, expected generation, expectedEmpty bool, candidate generation) (inspection, error) {
	actual, empty, err := readTail(db, domain)
	if err != nil {
		return 0, err
	}
	if !empty && actual == candidate {
		err = db.View(func(tx *bolt.Tx) error {
			if !bytes.Equal(tx.Bucket(generationBucket).Get(candidate.value[:]), successorRecord(domain, expected, expectedEmpty)) {
				return errLedgerInvalid
			}
			return nil
		})
		if err == nil {
			return inspectionCommitted, nil
		}
		return 0, err
	}
	if empty == expectedEmpty && (empty || actual == expected) {
		var exists bool
		err = db.View(func(tx *bolt.Tx) error {
			exists = tx.Bucket(generationBucket).Get(candidate.value[:]) != nil
			return nil
		})
		if err == nil && !exists {
			return inspectionAbsentExpected, nil
		}
	}
	return 0, errLedgerInvalid
}

func successorRecord(domain Domain, expected generation, expectedEmpty bool) []byte {
	record := make([]byte, identitySize*2+1)
	copy(record, domain.value[:])
	if !expectedEmpty {
		record[identitySize] = 1
		copy(record[identitySize+1:], expected.value[:])
	}
	return record
}

func bootstrapSuccessor(db *bolt.DB, domain Domain, candidate generation) error {
	return bootstrapWithAppend(db, domain, candidate, appendExpected)
}

func bootstrapWithAppend(db *bolt.DB, domain Domain, candidate generation,
	appendFn func(*bolt.DB, Domain, generation, bool, generation) error,
) error {
	expected, empty, err := readTail(db, domain)
	if err != nil {
		return err
	}
	for attempts := 0; attempts < 2; attempts++ {
		appendErr := appendFn(db, domain, expected, empty, candidate)
		outcome, inspectErr := inspectExactAppend(db, domain, expected, empty, candidate)
		if inspectErr != nil {
			return inspectErr
		}
		if outcome == inspectionCommitted {
			return nil
		}
		if appendErr == nil || outcome != inspectionAbsentExpected {
			return errLedgerInvalid
		}
	}
	return errLedgerInvalid
}

func zeroBytes(value []byte) bool {
	for _, b := range value {
		if b != 0 {
			return false
		}
	}
	return true
}
