package runtimerecoverycoordination

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// Fixtures initialize closed owner models before sharing them with any reader
// or writer. This represents detached local test facts, not a production seam.
func fixture(t *testing.T, instance string, f fact, noClaim bool, members []member) (*participant, *aggregateModel, *commandModel) {
	t.Helper()
	s, err := newScope(strings.Repeat("a", 64), 1, 2, instance)
	if err != nil {
		t.Fatal(err)
	}
	p, err := newParticipant(s)
	if err != nil {
		t.Fatal(err)
	}
	a, c, err := attachModels(p)
	if err != nil {
		t.Fatal(err)
	}
	a.fact, c.complete, c.noClaim, c.members = f, true, noClaim, append([]member(nil), members...)
	return p, a, c
}

func read(t *testing.T, p *participant, a *aggregateModel, c *commandModel) snapshot {
	t.Helper()
	s, err := p.read(context.Background(), a, c)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func await[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(3 * time.Second):
		t.Fatal("local operation blocked")
		var zero T
		return zero
	}
}

func openFixture(t *testing.T) (*participant, *aggregateModel, *commandModel) {
	t.Helper()
	p, a, c := fixture(t, "one", clean, true, nil)
	if err := p.observe(context.Background(), a, c, read(t, p, a, c)); err != nil {
		t.Fatal(err)
	}
	if !p.admissionOpen() {
		t.Fatal("fresh fully clean startup stayed closed")
	}
	return p, a, c
}

func TestExactScopeAndOwnerIdentity(t *testing.T) {
	for _, tc := range []struct {
		domain string
		w, c   uint64
		id     string
	}{
		{"", 1, 2, "one"}, {strings.Repeat("A", 64), 1, 2, "one"},
		{strings.Repeat("0", 64), 1, 2, "one"}, {strings.Repeat("a", 64), 0, 2, "one"},
		{strings.Repeat("a", 64), 1, 0, "one"}, {strings.Repeat("a", 64), 1, 2, ""},
	} {
		if _, err := newScope(tc.domain, tc.w, tc.c, tc.id); err == nil {
			t.Fatal("invalid scope accepted")
		}
	}
	if _, err := newParticipant(scope{}); err == nil {
		t.Fatal("zero scope accepted")
	}
	p, a, c := fixture(t, "one", clean, true, nil)
	q, b, d := fixture(t, "one", clean, true, nil)
	if _, _, err := attachModels(p); err == nil {
		t.Fatal("duplicate owners accepted")
	}
	for _, pair := range []struct {
		a *aggregateModel
		c *commandModel
	}{{nil, c}, {a, nil}, {b, c}, {a, d}} {
		if _, err := p.read(context.Background(), pair.a, pair.c); err == nil {
			t.Fatal("foreign/nil owners accepted")
		}
	}
	foreign := read(t, q, b, d)
	if err := p.observe(context.Background(), a, c, foreign); err == nil || p.admissionOpen() {
		t.Fatal("foreign snapshot opened startup")
	}
	var zero participant
	if _, err := zero.read(context.Background(), nil, nil); err == nil {
		t.Fatal("zero participant accepted")
	}
	if err := (*participant)(nil).invalidate(false); err == nil {
		t.Fatal("nil invalidation accepted")
	}
}

func TestStartupAndStickyFence(t *testing.T) {
	for _, tc := range []struct {
		name              string
		f                 fact
		complete, noClaim bool
		wantOpen          bool
	}{
		{"clean", clean, true, true, true}, {"managed", managedLive, true, true, false},
		{"recovery", recoveryRequired, true, true, false}, {"unknown", unknown, true, true, false},
		{"incomplete", clean, false, true, false}, {"claim", clean, true, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, a, c := fixture(t, "one", tc.f, tc.noClaim, nil)
			c.complete = tc.complete
			if p.admissionOpen() {
				t.Fatal("startup open before observation")
			}
			_ = p.observe(context.Background(), a, c, read(t, p, a, c))
			if p.admissionOpen() != tc.wantOpen {
				t.Fatal("wrong startup admission")
			}
		})
	}
	p, a, c := openFixture(t)
	before := read(t, p, a, c)
	if err := p.update(context.Background(), a, c, before, modelUpdate{fact: managedLive, complete: true, noClaim: true}); err != nil {
		t.Fatal(err)
	}
	if err := p.observe(context.Background(), a, c, read(t, p, a, c)); err != nil || !p.admissionOpen() {
		t.Fatal("ordinary managed observation closed admission")
	}
	current := read(t, p, a, c)
	if err := p.update(context.Background(), a, c, current, modelUpdate{fact: recoveryRequired, complete: true, noClaim: true}); err != nil {
		t.Fatal(err)
	}
	// Stale clean classification must not hide current recovery-required truth.
	if err := p.observe(context.Background(), a, c, before); err == nil || p.admissionOpen() {
		t.Fatal("stale clean reopened recovery")
	}
	if err := p.observe(context.Background(), a, c, read(t, p, a, c)); err != nil || p.admissionOpen() {
		t.Fatal("recovery observation failed to stay closed")
	}
	if err := p.update(context.Background(), a, c, read(t, p, a, c), modelUpdate{fact: clean, complete: true, noClaim: true}); err == nil {
		t.Fatal("sticky fence allowed writer")
	}
	if err := p.observe(context.Background(), a, c, before); err == nil || p.admissionOpen() {
		t.Fatal("old clean erased sticky closure")
	}
}

func TestDetachedCompleteMembershipAndRevisionRejection(t *testing.T) {
	members := []member{{kind: primitive, key: "cmd", revision: 1, terminal: true},
		{kind: parent, key: "parent", revision: 2, terminal: true},
		{kind: phase, key: "phase", parentKey: "parent", revision: 3, terminal: true}}
	for _, tc := range []struct {
		name   string
		change func(*snapshot)
	}{
		{"zero", func(s *snapshot) { *s = snapshot{} }},
		{"aggregate revision", func(s *snapshot) { s.aggregateRevision++ }},
		{"command revision", func(s *snapshot) { s.commandRevision++ }},
		{"epoch", func(s *snapshot) { s.epoch++ }},
		{"scope", func(s *snapshot) { s.scope.instance = "other" }},
		{"incomplete", func(s *snapshot) { s.complete = false }},
		{"omission", func(s *snapshot) { s.members = s.members[:2] }},
		{"duplicate", func(s *snapshot) { s.members = append(s.members, s.members[0]) }},
		{"foreign link", func(s *snapshot) { s.members[2].parentKey = "foreign" }},
		{"zero member revision", func(s *snapshot) { s.members[0].revision = 0 }},
		{"member revision", func(s *snapshot) { s.members[0].revision++ }},
		{"unknown kind", func(s *snapshot) { s.members[0].kind = 20 }},
		{"unknown fact", func(s *snapshot) { s.fact = 20 }},
		{"nonterminal clean", func(s *snapshot) { s.members[0].terminal = false }},
		{"over bound", func(s *snapshot) { s.members = make([]member, maxMembers+1) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, a, c := fixture(t, "one", clean, true, members)
			s := read(t, p, a, c)
			tc.change(&s)
			if err := p.observe(context.Background(), a, c, s); err == nil || p.admissionOpen() {
				t.Fatal("invalid snapshot opened admission")
			}
			actual := read(t, p, a, c)
			if actual.members[0] != members[0] || actual.members[2] != members[2] {
				t.Fatal("detached snapshot mutated owner model")
			}
		})
	}
	p, a, c := openFixture(t)
	before := read(t, p, a, c)
	next := modelUpdate{fact: clean, complete: true, noClaim: true, members: members}
	if err := p.update(context.Background(), a, c, before, next); err != nil {
		t.Fatal(err)
	}
	next.members[0].key = "mutated input"
	after := read(t, p, a, c)
	if after.members[0].key != "cmd" || after.aggregateRevision != before.aggregateRevision+1 || after.commandRevision != before.commandRevision+1 {
		t.Fatal("publication failed detached atomic revisions")
	}
	if err := p.update(context.Background(), a, c, before, next); err == nil {
		t.Fatal("stale publication accepted")
	}
}

func TestReservationHasOnlyOneLocalDisposition(t *testing.T) {
	p, a, c := fixture(t, "one", recoveryRequired, true, nil)
	s, err := p.reserve(context.Background(), a, c, read(t, p, a, c))
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	copyReservation := *s.reservation
	for _, invalid := range []*reservation{nil, &reservation{}, &copyReservation} {
		if s.revalidate(invalid) == nil {
			t.Fatal("foreign/nil/zero reservation accepted")
		}
	}
	q, b, d := fixture(t, "other", recoveryRequired, true, nil)
	other, err := q.reserve(context.Background(), b, d, read(t, q, b, d))
	if err != nil {
		t.Fatal(err)
	}
	defer other.close()
	if s.revalidate(other.reservation) == nil {
		t.Fatal("foreign session reservation accepted")
	}
	if err := s.revalidate(s.reservation); err != nil {
		t.Fatal(err)
	}
	if s.revalidate(s.reservation) == nil {
		t.Fatal("duplicate local disposition accepted")
	}
	s.close()
	s.close()
	if s.revalidate(s.reservation) == nil || p.admissionOpen() {
		t.Fatal("disposal reopened or retained usable reservation")
	}
	// Close releases exactly one token; repeated closes neither block nor mint it.
	if _, err := p.read(context.Background(), a, c); err != nil {
		t.Fatal(err)
	}
	if len(p.outer) != 1 {
		t.Fatal("wrong token count")
	}
	var zero session
	zero.close()
	if zero.revalidate(nil) == nil || (*session)(nil).revalidate(nil) == nil {
		t.Fatal("zero/nil session accepted")
	}
}

func TestStaleReservationClosesWithoutOwnerMutation(t *testing.T) {
	p, a, c := openFixture(t)
	cleanBefore := read(t, p, a, c)
	if _, err := p.reserve(context.Background(), a, c, cleanBefore); err == nil || p.admissionOpen() {
		t.Fatal("clean reservation accepted or fence reopened")
	}
	if err := p.observe(context.Background(), a, c, cleanBefore); err != nil || p.admissionOpen() {
		t.Fatal("clean read reopened sticky fence")
	}
	p, a, c = openFixture(t)
	before := read(t, p, a, c)
	if err := p.update(context.Background(), a, c, before, modelUpdate{fact: recoveryRequired, complete: true, noClaim: true}); err != nil {
		t.Fatal(err)
	}
	current := read(t, p, a, c)
	stale := current
	stale.aggregateRevision--
	if _, err := p.reserve(context.Background(), a, c, stale); err == nil || p.admissionOpen() {
		t.Fatal("stale reservation accepted or fence open")
	}
	actual := read(t, p, a, c)
	if actual.aggregateRevision != current.aggregateRevision || actual.commandRevision != current.commandRevision {
		t.Fatal("stale reservation mutated owner revisions")
	}
}

func TestUnresolvedClaimWithCleanAggregateCanReserveButCannotOpen(t *testing.T) {
	p, a, c := fixture(t, "one", clean, false, nil)
	facts := read(t, p, a, c)
	if err := p.observe(context.Background(), a, c, facts); err != nil || p.admissionOpen() {
		t.Fatal("unresolved local claim opened clean aggregate")
	}
	s, err := p.reserve(context.Background(), a, c, facts)
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	if err := s.revalidate(s.reservation); err != nil || p.admissionOpen() {
		t.Fatal("claim candidate validation reopened admission")
	}
}

func TestClosedAdmissionRejectsAllModelWriters(t *testing.T) {
	p, a, c := fixture(t, "one", clean, true, nil)
	before := read(t, p, a, c)
	next := modelUpdate{fact: managedLive, complete: true, noClaim: true}
	if p.update(context.Background(), a, c, before, next) == nil {
		t.Fatal("startup admitted a writer")
	}
	if p.observe(context.Background(), a, c, before) != nil {
		t.Fatal("clean startup failed")
	}
	if err := p.invalidate(true); err != nil {
		t.Fatal(err)
	}
	if p.update(context.Background(), a, c, read(t, p, a, c), next) == nil {
		t.Fatal("fatal fence admitted a writer")
	}
	if p.observe(context.Background(), a, c, read(t, p, a, c)) == nil || p.admissionOpen() {
		t.Fatal("fresh clean read erased fatal fence")
	}
}

func TestMalformedObservationClosesPreviouslyOpenAdmission(t *testing.T) {
	p, a, c := openFixture(t)
	valid := read(t, p, a, c)
	malformed := valid
	malformed.fact = 255
	if p.observe(context.Background(), a, c, malformed) == nil || p.admissionOpen() {
		t.Fatal("unknown enum left normal admission open")
	}
	if p.observe(context.Background(), a, c, valid) != nil || p.admissionOpen() {
		t.Fatal("clean reread erased malformed-observation closure")
	}
}

func TestUnresolvedOwnerPublicationImmediatelyClosesWriterGate(t *testing.T) {
	for _, tc := range []struct {
		name string
		next modelUpdate
	}{
		{"recovery", modelUpdate{fact: recoveryRequired, complete: true, noClaim: true}},
		{"unknown", modelUpdate{fact: unknown, complete: true, noClaim: true}},
		{"claim", modelUpdate{fact: clean, complete: true, noClaim: false}},
		{"incomplete", modelUpdate{fact: clean, complete: false, noClaim: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, a, c := openFixture(t)
			before := read(t, p, a, c)
			if err := p.update(context.Background(), a, c, before, tc.next); err != nil {
				t.Fatal(err)
			}
			if p.admissionOpen() {
				t.Fatal("unresolved publication returned with admission open")
			}
			published := read(t, p, a, c)
			if p.update(context.Background(), a, c, published, modelUpdate{fact: clean, complete: true, noClaim: true}) == nil {
				t.Fatal("second writer bypassed unresolved publication without observation")
			}
			after := read(t, p, a, c)
			if after.aggregateRevision != published.aggregateRevision || after.commandRevision != published.commandRevision {
				t.Fatal("closed writer mutated owner revisions")
			}
		})
	}
}

func TestWriterGateClosesOnMalformedCurrentOwnerFacts(t *testing.T) {
	p, a, c := openFixture(t)
	before := read(t, p, a, c)
	// Inject corruption into the closed test owner, rather than a conforming
	// publication, to exercise a fresh current-fact check at the writer gate.
	c.mu.Lock()
	c.members = []member{{kind: primitive, key: "bad", revision: 0, terminal: true}}
	c.mu.Unlock()
	if p.update(context.Background(), a, c, before, modelUpdate{fact: clean, complete: true, noClaim: true}) == nil || p.admissionOpen() {
		t.Fatal("malformed current truth left a writer gate open")
	}
	after := read(t, p, a, c)
	if after.aggregateRevision != before.aggregateRevision || after.commandRevision != before.commandRevision {
		t.Fatal("corrupt owner rejection changed revisions")
	}
}

func TestMalformedNextPublicationRejectsWithoutMutation(t *testing.T) {
	p, a, c := openFixture(t)
	before := read(t, p, a, c)
	next := modelUpdate{fact: clean, complete: true, noClaim: true,
		members: []member{{kind: phase, key: "orphan", parentKey: "missing", revision: 1, terminal: true}}}
	if p.update(context.Background(), a, c, before, next) == nil {
		t.Fatal("malformed next publication accepted")
	}
	after := read(t, p, a, c)
	if after.aggregateRevision != before.aggregateRevision || after.commandRevision != before.commandRevision || len(after.members) != 0 {
		t.Fatal("malformed proposal mutated owners")
	}
}

func TestInvalidationProgressesDuringRetainedOuterIO(t *testing.T) {
	for _, fatal := range []bool{false, true} {
		t.Run(fmt.Sprint("fatal=", fatal), func(t *testing.T) {
			p, a, c := fixture(t, "one", recoveryRequired, true, nil)
			before := read(t, p, a, c)
			s, err := p.reserve(context.Background(), a, c, before)
			if err != nil {
				t.Fatal(err)
			}
			defer s.close()
			ioEntered, unblockIO := make(chan struct{}), make(chan struct{})
			result := make(chan error, 1)
			go func() { close(ioEntered); <-unblockIO; result <- s.revalidate(s.reservation) }()
			await(t, ioEntered)
			// The external wait is solely in this harness. Each local lock is free.
			locksDone := make(chan struct{})
			go func() { a.mu.Lock(); a.mu.Unlock(); c.mu.Lock(); c.mu.Unlock(); close(locksDone) }()
			await(t, locksDone)
			invalidated := make(chan error, 1)
			go func() { invalidated <- p.invalidate(fatal) }()
			if err := await(t, invalidated); err != nil {
				t.Fatal(err)
			}
			if p.admissionOpen() {
				t.Fatal("invalidation left admission open")
			}
			close(unblockIO)
			if err := await(t, result); err == nil {
				t.Fatal("old candidate survived invalidation")
			}
			s.close()
			after := read(t, p, a, c)
			if after.aggregateRevision != before.aggregateRevision || after.commandRevision != before.commandRevision || after.epoch != before.epoch+1 {
				t.Fatal("invalidation rewrote owners or failed epoch advance")
			}
			if s.revalidate(s.reservation) == nil {
				t.Fatal("disposed stale candidate revived")
			}
		})
	}
}

func TestSameParticipantSerializesIndependentInstanceProgresses(t *testing.T) {
	p, a, c := fixture(t, "one", recoveryRequired, true, nil)
	s, err := p.reserve(context.Background(), a, c, read(t, p, a, c))
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	q, b, d := fixture(t, "two", clean, true, nil)
	independent := make(chan error, 1)
	go func() {
		facts, err := q.read(context.Background(), b, d)
		if err == nil {
			err = q.observe(context.Background(), b, d, facts)
		}
		independent <- err
	}()
	if err := await(t, independent); err != nil || !q.admissionOpen() {
		t.Fatal("independent Instance blocked")
	}
	// Enter synchronously with a live context: rejection must occur only when
	// its deadline expires while the other session still owns the outer token.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if ctx.Err() != nil {
		t.Fatal("contention probe did not start with a live context")
	}
	if _, err := p.read(ctx, a, c); err == nil || ctx.Err() != context.DeadlineExceeded {
		t.Fatal("same participant bypassed retained outer or rejected before deadline")
	}
	s.close()
	if _, err := p.read(context.Background(), a, c); err != nil {
		t.Fatal("released participant remained locked")
	}
}

func TestConcurrentOwnerPublicationsUseOneRevisionPair(t *testing.T) {
	p, a, c := openFixture(t)
	before := read(t, p, a, c)
	start, result := make(chan struct{}), make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			result <- p.update(context.Background(), a, c, before, modelUpdate{fact: managedLive, complete: true, noClaim: true})
		}()
	}
	close(start)
	success := 0
	for i := 0; i < 2; i++ {
		if await(t, result) == nil {
			success++
		}
	}
	after := read(t, p, a, c)
	if success != 1 || after.aggregateRevision != before.aggregateRevision+1 || after.commandRevision != before.commandRevision+1 {
		t.Fatal("writers did not atomically serialize revisions")
	}
}

func TestConcurrentCloseAndValidationNeverReleaseTwice(t *testing.T) {
	p, a, c := fixture(t, "one", recoveryRequired, true, nil)
	s, err := p.reserve(context.Background(), a, c, read(t, p, a, c))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = s.revalidate(s.reservation); s.close() }()
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	await(t, done)
	if len(p.outer) != 1 || p.admissionOpen() {
		t.Fatal("concurrent disposal minted token/opened fence")
	}
}
