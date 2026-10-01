//go:build windows

package runtimeexecutionevidence

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecontainment"
	"github.com/dsdred/universal-websocket-platform/internal/runtimeidentity"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/sys/windows"
)

const concreteAuthorityHelper = "UWP_TASK073_CONCRETE_AUTHORITY_HELPER"

func TestConcreteAuthorityFreshnessAndFatalFence(t *testing.T) {
	if root := os.Getenv(concreteAuthorityHelper); root != "" {
		exerciseConcreteAuthorityFreshnessAndFatalFence(t, root)
		return
	}
	root, err := os.MkdirTemp("", "uwp-task073-concrete-authority-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	command := exec.Command(os.Args[0], "-test.run=^TestConcreteAuthorityFreshnessAndFatalFence$")
	command.Env = append(os.Environ(), concreteAuthorityHelper+"="+root)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("concrete authority helper: %v\n%s", err, output)
	}
}

func exerciseConcreteAuthorityFreshnessAndFatalFence(t *testing.T, root string) {
	t.Helper()
	domainText, domainBytes := randomIdentity(t)
	domain, err := runtimecontainment.ParseDomain(domainText)
	if err != nil {
		t.Fatal(err)
	}
	config := provisionContainmentLedger(t, root, domainText, domainBytes)
	result := runtimecontainment.AcquireProcessContainment(domain, config)
	authority, ok := result.Authority()
	if !ok {
		t.Fatal("concrete ProcessContainment authority was not acquired")
	}
	generation := runtimeidentity.ExecutionGeneration(authority.CurrentGeneration())
	store := boundStoreForGeneration(t, generation)
	composer := NewComposer(domain, authority, store)

	handle := composer.Query(context.Background(), domain, testInstance, testAttempt, generation, GenerationStatus)
	signalContainmentCapability(t, domainText)
	assertResult(t, handle.Consume(context.Background()), Unknown, Unavailable)
	if authority.IsAuthoritative() {
		t.Fatal("fatal capability fault did not permanently fence concrete authority")
	}

	second := composer.Query(context.Background(), domain, testInstance, testAttempt, generation, GenerationStatus)
	assertResult(t, second.Consume(context.Background()), Unknown, Unavailable)
}

func boundStoreForGeneration(t *testing.T, generation runtimeidentity.ExecutionGeneration) *runtimeidentity.Store {
	t.Helper()
	store := runtimeidentity.NewStore()
	if err := store.CreateRuntimeInstance(11, 22, testInstance); err != nil {
		t.Fatal(err)
	}
	view, err := store.ReadRuntimeInstance(testInstance)
	if err != nil {
		t.Fatal(err)
	}
	claim, err := store.ConditionalClaimLaunchAttempt(testInstance, view.Revision(), testAttempt, 33)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConditionalBindExecutionGeneration(testInstance, claim.Revision(), testAttempt, generation); err != nil {
		t.Fatal(err)
	}
	return store
}

func provisionContainmentLedger(t *testing.T, root, domainText string, domainBytes []byte) runtimecontainment.LocalConfig {
	t.Helper()
	canonicalRoot, physicalRoot := inspectWindowsPath(t, root)
	directory := filepath.Join(root, "runtime-containment")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, domainText+".db")
	db, err := bolt.Open(path, 0o600, &bolt.Options{NoSync: false, NoGrowSync: false})
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	_, ledgerIdentity := inspectWindowsFile(t, file)
	_ = file.Close()
	storageAuthorityText, storageAuthority := randomIdentity(t)
	err = db.Update(func(tx *bolt.Tx) error {
		anchor, err := tx.CreateBucket([]byte("anchor-v1"))
		if err != nil {
			return err
		}
		if _, err = tx.CreateBucket([]byte("generations-v1")); err != nil {
			return err
		}
		if _, err = tx.CreateBucket([]byte("state-v1")); err != nil {
			return err
		}
		values := map[string][]byte{
			"domain":               domainBytes,
			"canonical-root":       []byte(canonicalRoot),
			"physical-root":        []byte(physicalRoot),
			"storage-authority":    storageAuthority,
			"ledger-file-identity": []byte(ledgerIdentity),
		}
		for key, value := range values {
			if err := anchor.Put([]byte(key), value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	return runtimecontainment.LocalConfig{
		ExpectedDomain:               domainText,
		RootDir:                      root,
		ExpectedCanonicalRoot:        canonicalRoot,
		ExpectedPhysicalRootIdentity: physicalRoot,
		StorageAuthorityID:           storageAuthorityText,
	}
}

func randomIdentity(t *testing.T) (string, []byte) {
	t.Helper()
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		t.Fatal(err)
	}
	value[0] |= 1
	return hex.EncodeToString(value), value
}

func signalContainmentCapability(t *testing.T, domainText string) {
	t.Helper()
	name, err := windows.UTF16PtrFromString(`Global\uwp-runtime-containment-` + domainText)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.OpenEvent(windows.EVENT_MODIFY_STATE, false, name)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	if err := windows.SetEvent(handle); err != nil {
		t.Fatal(err)
	}
}

func inspectWindowsPath(t *testing.T, path string) (string, string) {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	handle, err := windows.CreateFile(windows.StringToUTF16Ptr(abs), windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	return inspectWindowsHandle(t, handle)
}

func inspectWindowsFile(t *testing.T, file *os.File) (string, string) {
	t.Helper()
	return inspectWindowsHandle(t, windows.Handle(file.Fd()))
}

func inspectWindowsHandle(t *testing.T, handle windows.Handle) (string, string) {
	t.Helper()
	buffer := make([]uint16, windows.MAX_PATH+1)
	n, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], uint32(len(buffer)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if n >= uint32(len(buffer)) {
		buffer = make([]uint16, n+1)
		n, err = windows.GetFinalPathNameByHandle(handle, &buffer[0], uint32(len(buffer)), 0)
		if err != nil {
			t.Fatal(err)
		}
	}
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		t.Fatal(err)
	}
	if info.VolumeSerialNumber == 0 || (info.FileIndexHigh == 0 && info.FileIndexLow == 0) {
		t.Fatal("Windows path identity is zero")
	}
	canonical := strings.ToLower(windows.UTF16ToString(buffer[:n]))
	identity := strings.ToLower(fmt.Sprintf("%08x:%08x%08x", info.VolumeSerialNumber, info.FileIndexHigh, info.FileIndexLow))
	return canonical, identity
}
