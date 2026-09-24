//go:build windows

package runtimecontainment

import (
	"os"
	"os/exec"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/sys/windows"
)

func TestWindowsEventIsExclusiveNonInheritableAndLossFences(t *testing.T) {
	domain := testDomain(t)
	config := provisionForTest(t, domain, t.TempDir())
	descriptor, _ := parseDescriptor(domain, config)
	db, err := openExistingProvisioned(domain, descriptor)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	first, acquired, valid := tryAcquire(domain)
	if !acquired || !valid {
		t.Fatal("first Global Event acquisition failed")
	}
	defer windows.CloseHandle(first.handle)
	second, secondAcquired, _ := tryAcquire(domain)
	if secondAcquired {
		_ = windows.CloseHandle(second.handle)
		t.Fatal("second live holder acquired the same domain")
	}
	authority := &ActiveAuthority{keeper: &keeper{capability: first, db: db, domain: domain, descriptor: descriptor}}
	if !authority.IsAuthoritative() {
		t.Fatal("fresh capability is not authoritative")
	}
	if err := windows.SetEvent(first.handle); err != nil {
		t.Fatal(err)
	}
	if authority.IsAuthoritative() || !authority.keeper.fenced.Load() {
		t.Fatal("capability loss did not permanently fence authority")
	}
}

func TestWindowsFileIdentityRejectsUnverifiableZeroParts(t *testing.T) {
	for name, info := range map[string]windows.ByHandleFileInformation{
		"all-zero":        {},
		"zero-volume":     {FileIndexLow: 1},
		"zero-file-index": {VolumeSerialNumber: 1},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := formatHandleIdentity(info); err == nil {
				t.Fatal("unverifiable file identity accepted")
			}
		})
	}
	if got, err := formatHandleIdentity(windows.ByHandleFileInformation{VolumeSerialNumber: 1, FileIndexLow: 2}); err != nil || got != "00000001:0000000000000002" {
		t.Fatalf("valid identity: got=%q err=%v", got, err)
	}
}

func TestWindowsEventHandleIsNotInherited(t *testing.T) {
	if os.Getenv("UWP_HANDLE_CHILD") == "1" {
		time.Sleep(30 * time.Second)
		return
	}
	domain := testDomain(t)
	capability, acquired, valid := tryAcquire(domain)
	if !acquired || !valid {
		t.Fatal("Global Event acquisition failed")
	}
	child := exec.Command(os.Args[0], "-test.run=^TestWindowsEventHandleIsNotInherited$")
	child.Env = append(os.Environ(), "UWP_HANDLE_CHILD=1")
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = child.Process.Kill()
		_ = child.Wait()
	}()
	if err := windows.CloseHandle(capability.handle); err != nil {
		t.Fatal(err)
	}
	reacquired, ok, valid := tryAcquire(domain)
	if !ok || !valid {
		t.Fatal("child inherited the private Event handle")
	}
	_ = windows.CloseHandle(reacquired.handle)
}

func TestLiveStorageCorruptionPermanentlyFencesAuthority(t *testing.T) {
	domain := testDomain(t)
	config := provisionForTest(t, domain, t.TempDir())
	descriptor, _ := parseDescriptor(domain, config)
	db, err := openExistingProvisioned(domain, descriptor)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	capability, acquired, valid := tryAcquire(domain)
	if !acquired || !valid {
		t.Fatal("Global Event acquisition failed")
	}
	defer windows.CloseHandle(capability.handle)
	authority := &ActiveAuthority{keeper: &keeper{capability: capability, db: db, domain: domain, descriptor: descriptor}}
	if !authority.IsAuthoritative() {
		t.Fatal("fresh authority rejected")
	}
	if err := db.Update(func(tx *bolt.Tx) error { return tx.DeleteBucket(anchorBucket) }); err != nil {
		t.Fatal(err)
	}
	if authority.IsAuthoritative() || !authority.keeper.fenced.Load() {
		t.Fatal("live storage corruption did not permanently fence authority")
	}
}
