package runtimecontainment

import (
	"bytes"
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimecoordination"
	bolt "go.etcd.io/bbolt"
)

func associatedFixture(t *testing.T) (*ActiveAuthority, runtimecoordination.Scope, *runtimecoordination.Participant, *runtimecoordination.ClientEpoch) {
	t.Helper()
	d := testDomain(t)
	a := acquiredAuthority(t, d, provisionForTest(t, d, t.TempDir()))
	s, _ := runtimecoordination.NewScope(d.text(), 1, 2, "instance")
	p, err := runtimecoordination.NewSourceParticipant(a.keeper.root, s)
	if err != nil {
		t.Fatal(err)
	}
	c, err := runtimecoordination.NewClientEpoch(a.keeper.root)
	if err != nil {
		t.Fatal(err)
	}
	_, r, err := c.RegisterCommand(p, s)
	if err != nil {
		t.Fatal(err)
	}
	r.Commit()
	r.Close()
	return a, s, p, c
}

func assertSourceFatal(t *testing.T, a *ActiveAuthority, s runtimecoordination.Scope, p *runtimecoordination.Participant, c *runtimecoordination.ClientEpoch) {
	t.Helper()
	if !a.keeper.fenced.Load() {
		t.Fatal("keeper flag absent")
	}
	if _, err := runtimecoordination.NewClientEpoch(a.keeper.root); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal("root escaped", err)
	}
	if _, err := p.AttachAggregate(s); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal("participant escaped", err)
	}
	if _, err := c.BeginReplacement(); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal("client escaped", err)
	}
}

func TestActualSourceFatalInventory(t *testing.T) {
	// Keep branch inventory tied to executable production calls, excluding all
	// direct fixture stores. Any new source publication requires a new proof row.
	want := map[string]int{"isAuthoritative": 1, "finishAcquisitionWithGeneration": 4, "ReadGeneration": 1, "validateForEvidence": 1, "fenceEvidenceError": 1, "InspectCoordination": 2}
	got := map[string]int{}
	for _, name := range []string{"containment.go", "evidence.go", "coordination.go"} {
		f, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if ok && sel.Sel.Name == "publishFatal" {
					got[fn.Name.Name]++
				}
				if ok && sel.Sel.Name == "Store" && fn.Name.Name != "publishFatal" {
					t.Errorf("uncoupled flag store in %s", fn.Name.Name)
				}
				return true
			})
		}
	}
	if len(got) != len(want) {
		t.Fatal(got)
	}
	for fn, count := range want {
		if got[fn] != count {
			t.Fatalf("%s fatal sites=%d want%d", fn, got[fn], count)
		}
	}
}

func TestActualSourceAuthorityAndEvidenceFatalPaths(t *testing.T) {
	requireWindows(t)
	for _, name := range []string{"authority", "evidence-unavailable", "evidence-tail", "evidence-error"} {
		t.Run(name, func(t *testing.T) {
			a, s, p, c := associatedFixture(t)
			switch name {
			case "authority":
				db := a.keeper.db
				a.keeper.db = nil
				if a.IsAuthoritative() {
					t.Fatal("nil database accepted")
				}
				a.keeper.db = db
			case "evidence-unavailable":
				db := a.keeper.db
				a.keeper.db = nil
				got := a.ReadGeneration(context.Background(), a.keeper.domain, a.CurrentGeneration())
				a.keeper.db = db
				if got.Reason() != UnknownUnavailable {
					t.Fatal(got)
				}
			case "evidence-tail":
				next, _ := newGeneration()
				if err := appendExpected(a.keeper.db, a.keeper.domain, a.generation, false, next); err != nil {
					t.Fatal(err)
				}
				if got := a.ReadGeneration(context.Background(), a.keeper.domain, a.CurrentGeneration()); got.Reason() != UnknownContradictory {
					t.Fatal(got)
				}
			case "evidence-error":
				if err := a.keeper.db.Update(func(tx *bolt.Tx) error {
					return tx.Bucket(stateBucket).Put(tailKey, bytes.Repeat([]byte{255}, identitySize))
				}); err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				if got := a.ReadGeneration(ctx, a.keeper.domain, a.CurrentGeneration()); got.Reason() != UnknownContradictory {
					t.Fatal("fatal lost cancellation precedence", got)
				}
			}
			assertSourceFatal(t, a, s, p, c)
		})
	}
}

func TestActualSourceBootstrapFatalBeforeExposure(t *testing.T) {
	requireWindows(t)
	for _, name := range []string{"acquired-invalid", "existing-store", "successor"} {
		t.Run(name, func(t *testing.T) {
			d := testDomain(t)
			config := provisionForTest(t, d, t.TempDir())
			desc, _ := parseDescriptor(d, config)
			var result Result
			if name == "existing-store" {
				if err := os.Remove(ledgerPath(desc.root, d)); err != nil {
					t.Fatal(err)
				}
				result = AcquireProcessContainment(d, config)
			} else {
				cap, acquired, valid := tryAcquire(d)
				if !acquired || !valid {
					t.Fatal("fixture guard unavailable")
				}
				if name == "acquired-invalid" {
					result = finishAcquisition(d, desc, cap, false)
				} else {
					result = finishAcquisitionWithGeneration(d, desc, cap, true, func() (generation, error) { return generation{}, errors.New("injected generation failure") })
				}
			}
			if _, ok := result.Authority(); ok || !result.IsFatalFenced() {
				t.Fatal("bootstrap exposed authority", result)
			}
			retained.Lock()
			k := retained.keepers[len(retained.keepers)-1]
			retained.Unlock()
			if k.domain != d || k.root == nil || !k.fenced.Load() {
				t.Fatal("source was not retained before fatal")
			}
			if _, err := runtimecoordination.NewClientEpoch(k.root); !errors.Is(err, runtimecoordination.ErrClosed) {
				t.Fatal("bootstrap root not terminal", err)
			}
			if k.db != nil {
				if err := k.db.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestActualSourcePausedPublicationJoinsRepeatedAndFinalReaders(t *testing.T) {
	requireWindows(t)
	a, s, p, c := associatedFixture(t)
	// Existing replacement gate pauses actual publisher after its flag store,
	// without any callback or test hook in the opaque fatal region.
	replacement, err := c.BeginReplacement()
	if err != nil {
		t.Fatal(err)
	}
	entered, inspectionDone := make(chan struct{}), make(chan struct{})
	final := make(chan bool, 1)
	go func() { final <- a.isAuthoritative(func() bool { close(entered); <-inspectionDone; return true }) }()
	<-entered
	published := make(chan struct{})
	go func() { a.keeper.publishFatal(); close(published) }()
	deadline := time.After(time.Second)
	for !a.keeper.fenced.Load() {
		select {
		case <-deadline:
			t.Fatal("publisher did not reach flag")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	close(inspectionDone)
	var wg sync.WaitGroup
	returned := make(chan struct{}, 16)
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if a.IsAuthoritative() {
				t.Error("fenced source authoritative")
			}
			if got := a.ReadGeneration(context.Background(), a.keeper.domain, a.CurrentGeneration()); got.Reason() != UnknownUnavailable {
				t.Error(got)
			}
			returned <- struct{}{}
		}()
	}
	select {
	case <-returned:
		t.Fatal("flag-only reader returned before fanout")
	case <-final:
		t.Fatal("final false returned before fanout")
	case <-published:
		t.Fatal("publisher bypassed client gate")
	case <-time.After(30 * time.Millisecond):
	}
	replacement.Close()
	<-published
	wg.Wait()
	if <-final {
		t.Fatal("final observation accepted concurrent fatal")
	}
	assertSourceFatal(t, a, s, p, c)
}

func TestSourceRootExactDomainAndNoLegacyAdoption(t *testing.T) {
	requireWindows(t)
	a, _, _, _ := associatedFixture(t)
	root, err := a.CoordinationRoot(a.keeper.domain)
	if err != nil || root != a.keeper.root {
		t.Fatal(root, err)
	}
	if _, err := a.CoordinationRoot(testDomain(t)); !errors.Is(err, runtimecoordination.ErrForeignScope) {
		t.Fatal(err)
	}
	legacy := *a
	legacy.keeper = &keeper{capability: a.keeper.capability, db: a.keeper.db, domain: a.keeper.domain, descriptor: a.keeper.descriptor}
	if _, err := legacy.CoordinationRoot(a.keeper.domain); !errors.Is(err, runtimecoordination.ErrClosed) {
		t.Fatal("fabricated root", err)
	}
	// Copy ActiveAuthority itself, not its private atomic keeper.
	copy := *a
	copy.keeper.publishFatal()
	if a.IsAuthoritative() {
		t.Fatal("copy lost source")
	}
}

func TestCoordinationInspectionHasNoGateOrCallbackSurface(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "coordination.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "InspectCoordination" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if _, ok := n.(*ast.FuncLit); ok {
				t.Error("inspection callback surface")
			}
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					calls = append(calls, sel.Sel.Name)
					if sel.Sel.Name == "Lock" || sel.Sel.Name == "RLock" {
						t.Error("inspection holds gate")
					}
				}
			}
			return true
		})
	}
	positions := map[string]int{}
	for i, name := range calls {
		positions[name] = i
	}
	for _, name := range []string{"ReadGeneration", "IsAuthoritative", "Observe", "RevalidateCurrentObservation"} {
		if _, ok := positions[name]; !ok {
			t.Fatal("missing actual boundary", name)
		}
	}
	if positions["ReadGeneration"] >= positions["Observe"] || positions["IsAuthoritative"] >= positions["Observe"] || positions["Observe"] >= positions["RevalidateCurrentObservation"] {
		t.Fatal("external inspection after local acceptance", calls)
	}
}
