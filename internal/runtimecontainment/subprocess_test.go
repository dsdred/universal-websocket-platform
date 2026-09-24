package runtimecontainment

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

const helperMarker = "UWP_CONTAINMENT_HELPER"

func TestContainmentHelper(t *testing.T) {
	if os.Getenv(helperMarker) != "1" {
		return
	}
	domain, err := ParseDomain(os.Getenv("UWP_DOMAIN"))
	if err != nil {
		fmt.Println("BAD-DOMAIN")
		return
	}
	config := LocalConfig{
		ExpectedDomain: os.Getenv("UWP_EXPECTED_DOMAIN"), RootDir: os.Getenv("UWP_ROOT"), ExpectedCanonicalRoot: os.Getenv("UWP_CANONICAL"),
		ExpectedPhysicalRootIdentity: os.Getenv("UWP_PHYSICAL"), StorageAuthorityID: os.Getenv("UWP_AUTHORITY"),
	}
	if os.Getenv("UWP_CRASH_CUT") != "" {
		capability, acquired, _ := tryAcquire(domain)
		if !acquired {
			os.Exit(24)
		}
		if os.Getenv("UWP_CRASH_CUT") == "candidate" {
			descriptor, _ := parseDescriptor(domain, config)
			if _, err = openExistingProvisioned(domain, descriptor); err != nil {
				os.Exit(25)
			}
			_, _ = newGeneration()
		}
		_ = capability
		os.Exit(23)
	}
	result := AcquireProcessContainment(domain, config)
	if authority, ok := result.Authority(); ok {
		if authority.GuaranteeLevel() != "ProcessContainment" || !authority.IsAuthoritative() {
			fmt.Println("BAD-AUTHORITY")
			return
		}
		if os.Getenv("UWP_CRASH_BEFORE_ACK") == "1" {
			os.Exit(23)
		}
		fmt.Println("ACTIVE " + authority.CurrentGeneration())
		if os.Getenv("UWP_HOLD") == "1" {
			for {
				time.Sleep(time.Second)
			}
		}
		return
	}
	if result.IsFatalFenced() {
		fmt.Println("FATAL")
		return
	}
	fmt.Println("UNAVAILABLE")
}

func TestSubprocessConcurrencyCrashRestartAndDomainIsolation(t *testing.T) {
	requireWindows(t)
	root := t.TempDir()
	domain := testDomain(t)
	config := provisionForTest(t, domain, root)
	commands := make([]*exec.Cmd, 6)
	lines := make([]string, len(commands))
	for i := range commands {
		commands[i] = helperCommand(domain, config, true)
		stdout, err := commands[i].StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err = commands[i].Start(); err != nil {
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(stdout)
		if scanner.Scan() {
			lines[i] = scanner.Text()
		}
	}
	winner := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "ACTIVE ") {
			if winner >= 0 {
				t.Fatalf("multiple live winners: %q", lines)
			}
			winner = i
		} else if line != "UNAVAILABLE" {
			t.Fatalf("helper %d: %q", i, line)
		}
	}
	if winner < 0 {
		t.Fatalf("no winner: %q", lines)
	}
	first := strings.TrimPrefix(lines[winner], "ACTIVE ")
	if err := commands[winner].Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = commands[winner].Wait()
	for i, command := range commands {
		if i != winner {
			_ = command.Wait()
		}
	}
	restarted := runHelper(t, domain, config)
	if !strings.HasPrefix(restarted, "ACTIVE ") || strings.TrimPrefix(restarted, "ACTIVE ") == first {
		t.Fatalf("restart did not append fresh successor: first=%q next=%q", first, restarted)
	}

	other := testDomain(t)
	otherConfig := provisionForTest(t, other, root)
	holder := helperCommand(domain, config, true)
	pipe, _ := holder.StdoutPipe()
	if err := holder.Start(); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(pipe)
	if !scanner.Scan() || !strings.HasPrefix(scanner.Text(), "ACTIVE ") {
		t.Fatal("first domain holder failed")
	}
	if got := runHelper(t, other, otherConfig); !strings.HasPrefix(got, "ACTIVE ") {
		t.Fatalf("independent domain blocked: %q", got)
	}
	_ = holder.Process.Kill()
	_ = holder.Wait()
}

func TestSubprocessCommittedBeforeAcknowledgementConvergesOnRestart(t *testing.T) {
	requireWindows(t)
	domain := testDomain(t)
	config := provisionForTest(t, domain, t.TempDir())
	command := helperCommand(domain, config, false)
	command.Env = append(command.Env, "UWP_CRASH_BEFORE_ACK=1")
	if err := command.Run(); err == nil {
		t.Fatal("crash helper unexpectedly succeeded")
	}
	if got := runHelper(t, domain, config); !strings.HasPrefix(got, "ACTIVE ") {
		t.Fatalf("restart failed after committed-unacknowledged append: %q", got)
	}
	descriptor, _ := parseDescriptor(domain, config)
	db, err := openExistingProvisioned(domain, descriptor)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, empty, err := readTail(db, domain)
	if err != nil || empty || db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(generationBucket).Stats().KeyN != 2 {
			return errLedgerInvalid
		}
		return nil
	}) != nil {
		t.Fatal("restart did not preserve one predecessor and one successor")
	}
}

func TestSubprocessProvisionalCrashCutsLeaveProvisionedEmpty(t *testing.T) {
	requireWindows(t)
	for _, cut := range []string{"guard", "candidate"} {
		t.Run(cut, func(t *testing.T) {
			domain := testDomain(t)
			config := provisionForTest(t, domain, t.TempDir())
			command := helperCommand(domain, config, false)
			command.Env = append(command.Env, "UWP_CRASH_CUT="+cut)
			if err := command.Run(); err == nil {
				t.Fatal("crash helper unexpectedly succeeded")
			}
			descriptor, _ := parseDescriptor(domain, config)
			db, err := openExistingProvisioned(domain, descriptor)
			if err != nil {
				t.Fatal(err)
			}
			_, empty, err := readTail(db, domain)
			_ = db.Close()
			if err != nil || !empty {
				t.Fatalf("provisional cut mutated ledger: empty=%v err=%v", empty, err)
			}
			if got := runHelper(t, domain, config); !strings.HasPrefix(got, "ACTIVE ") {
				t.Fatalf("restart after provisional cut: %q", got)
			}
		})
	}
}

func TestSubprocessFailsClosedForMissingMismatchCopyReplacementAndCorruption(t *testing.T) {
	requireWindows(t)
	t.Run("missing", func(t *testing.T) {
		domain := testDomain(t)
		config := provisionForTest(t, domain, t.TempDir())
		if err := os.Remove(ledgerPath(config.RootDir, domain)); err != nil {
			t.Fatal(err)
		}
		assertFatal(t, runHelper(t, domain, config))
		if _, err := os.Stat(ledgerPath(config.RootDir, domain)); !os.IsNotExist(err) {
			t.Fatal("bootstrap recreated missing ledger")
		}
	})
	t.Run("descriptor-mismatch", func(t *testing.T) {
		domain := testDomain(t)
		config := provisionForTest(t, domain, t.TempDir())
		config.StorageAuthorityID = strings.Repeat("a", identitySize*2)
		assertFatal(t, runHelper(t, domain, config))
	})
	t.Run("descriptor-domain-mismatch", func(t *testing.T) {
		domain := testDomain(t)
		config := provisionForTest(t, domain, t.TempDir())
		config.ExpectedDomain = testDomain(t).text()
		assertUnavailable(t, runHelper(t, domain, config))
	})
	t.Run("physical-root-mismatch", func(t *testing.T) {
		domain := testDomain(t)
		config := provisionForTest(t, domain, t.TempDir())
		config.ExpectedPhysicalRootIdentity = strings.Repeat("f", len(config.ExpectedPhysicalRootIdentity))
		assertFatal(t, runHelper(t, domain, config))
	})
	t.Run("anchor-domain-mismatch", func(t *testing.T) {
		domain := testDomain(t)
		config := provisionForTest(t, domain, t.TempDir())
		db, err := bolt.Open(ledgerPath(config.RootDir, domain), 0o600, nil)
		if err != nil {
			t.Fatal(err)
		}
		other := testDomain(t)
		if err = db.Update(func(tx *bolt.Tx) error { return tx.Bucket(anchorBucket).Put(domainKey, other.value[:]) }); err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
		assertFatal(t, runHelper(t, domain, config))
	})
	t.Run("deleted-anchor", func(t *testing.T) {
		domain := testDomain(t)
		config := provisionForTest(t, domain, t.TempDir())
		db, err := bolt.Open(ledgerPath(config.RootDir, domain), 0o600, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err = db.Update(func(tx *bolt.Tx) error { return tx.DeleteBucket(anchorBucket) }); err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
		assertFatal(t, runHelper(t, domain, config))
	})
	t.Run("alternate-copied-store", func(t *testing.T) {
		domain := testDomain(t)
		original := provisionForTest(t, domain, t.TempDir())
		alternateRoot := t.TempDir()
		if err := os.Mkdir(filepath.Join(alternateRoot, "runtime-containment"), 0o700); err != nil {
			t.Fatal(err)
		}
		copyFile(t, ledgerPath(original.RootDir, domain), ledgerPath(alternateRoot, domain))
		candidate := original
		candidate.RootDir = alternateRoot
		assertFatal(t, runHelper(t, domain, candidate))
		if got := runHelper(t, domain, original); !strings.HasPrefix(got, "ACTIVE ") {
			t.Fatalf("authoritative store rejected: %q", got)
		}
	})
	t.Run("intermediate-reparse-relocation", func(t *testing.T) {
		domain := testDomain(t)
		root := t.TempDir()
		config := provisionForTest(t, domain, root)
		originalDirectory := filepath.Join(root, "runtime-containment")
		relocatedDirectory := filepath.Join(t.TempDir(), "relocated-containment")
		if err := os.Rename(originalDirectory, relocatedDirectory); err != nil {
			t.Fatal(err)
		}
		if output, err := exec.Command("cmd", "/c", "mklink", "/J", originalDirectory, relocatedDirectory).CombinedOutput(); err != nil {
			t.Fatalf("create junction: %v: %s", err, output)
		}
		assertFatal(t, runHelper(t, domain, config))
	})
	t.Run("same-root-replacement", func(t *testing.T) {
		domain := testDomain(t)
		config := provisionForTest(t, domain, t.TempDir())
		path := ledgerPath(config.RootDir, domain)
		copy := path + ".copy"
		copyFile(t, path, copy)
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(copy, path); err != nil {
			t.Fatal(err)
		}
		assertFatal(t, runHelper(t, domain, config))
	})
	t.Run("stale-replacement", func(t *testing.T) {
		domain := testDomain(t)
		config := provisionForTest(t, domain, t.TempDir())
		path := ledgerPath(config.RootDir, domain)
		stale, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := runHelper(t, domain, config); !strings.HasPrefix(got, "ACTIVE ") {
			t.Fatalf("seed generation: %q", got)
		}
		if err = os.Remove(path); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, stale, 0o600); err != nil {
			t.Fatal(err)
		}
		assertFatal(t, runHelper(t, domain, config))
	})
	t.Run("two-candidate-stores", func(t *testing.T) {
		domain := testDomain(t)
		trusted := provisionForTest(t, domain, t.TempDir())
		other := provisionForTest(t, domain, t.TempDir())
		candidate := trusted
		candidate.RootDir = other.RootDir
		assertFatal(t, runHelper(t, domain, candidate))
		if got := runHelper(t, domain, trusted); !strings.HasPrefix(got, "ACTIVE ") {
			t.Fatalf("trusted candidate rejected: %q", got)
		}
	})
	t.Run("self-consistent-copy-cannot-rewrite-trust", func(t *testing.T) {
		domain := testDomain(t)
		trusted := provisionForTest(t, domain, t.TempDir())
		alternateRoot := t.TempDir()
		if err := os.Mkdir(filepath.Join(alternateRoot, "runtime-containment"), 0o700); err != nil {
			t.Fatal(err)
		}
		path := ledgerPath(alternateRoot, domain)
		copyFile(t, ledgerPath(trusted.RootDir, domain), path)
		rewriteCandidateAnchor(t, path, alternateRoot)
		candidate := trusted
		candidate.RootDir = alternateRoot
		assertFatal(t, runHelper(t, domain, candidate))
	})
	t.Run("corruption", func(t *testing.T) {
		domain := testDomain(t)
		config := provisionForTest(t, domain, t.TempDir())
		db, err := bolt.Open(ledgerPath(config.RootDir, domain), 0o600, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err = db.Update(func(tx *bolt.Tx) error { _, e := tx.CreateBucket([]byte("unexpected")); return e }); err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
		assertFatal(t, runHelper(t, domain, config))
	})
}

func helperCommand(domain Domain, config LocalConfig, hold bool) *exec.Cmd {
	command := exec.Command(os.Args[0], "-test.run=^TestContainmentHelper$")
	command.Env = append(os.Environ(), helperMarker+"=1", "UWP_DOMAIN="+domain.text(), "UWP_EXPECTED_DOMAIN="+config.ExpectedDomain, "UWP_ROOT="+config.RootDir,
		"UWP_CANONICAL="+config.ExpectedCanonicalRoot, "UWP_PHYSICAL="+config.ExpectedPhysicalRootIdentity,
		"UWP_AUTHORITY="+config.StorageAuthorityID)
	if hold {
		command.Env = append(command.Env, "UWP_HOLD=1")
	}
	return command
}

func runHelper(t *testing.T, domain Domain, config LocalConfig) string {
	t.Helper()
	output, err := helperCommand(domain, config, false).CombinedOutput()
	if err != nil {
		t.Fatalf("helper: %v\n%s", err, output)
	}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ACTIVE ") || line == "FATAL" || line == "UNAVAILABLE" {
			return line
		}
	}
	t.Fatalf("missing helper result: %s", output)
	return ""
}

func assertFatal(t *testing.T, got string) {
	t.Helper()
	if got != "FATAL" {
		t.Fatalf("got %q, want FATAL", got)
	}
}

func assertUnavailable(t *testing.T, got string) {
	t.Helper()
	if got != "UNAVAILABLE" {
		t.Fatalf("got %q, want UNAVAILABLE", got)
	}
}

func copyFile(t *testing.T, source, target string) {
	t.Helper()
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(target, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func rewriteCandidateAnchor(t *testing.T, path, root string) {
	t.Helper()
	canonical, physical, err := inspectRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_, identity, err := inspectFile(file)
	_ = file.Close()
	if err != nil {
		t.Fatal(err)
	}
	authority, _ := newGeneration()
	err = db.Update(func(tx *bolt.Tx) error {
		anchor := tx.Bucket(anchorBucket)
		for key, value := range map[string][]byte{
			string(canonicalRootKey): []byte(canonical), string(physicalRootKey): []byte(physical),
			string(ledgerIdentityKey): []byte(identity), string(storageAuthorityKey): authority.value[:],
		} {
			if err := anchor.Put([]byte(key), value); err != nil {
				return err
			}
		}
		return nil
	})
	_ = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	if runtime.GOOS != "windows" && os.Getenv(helperMarker) == "1" {
		os.Exit(2)
	}
	os.Exit(m.Run())
}
