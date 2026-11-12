package runtimemanagement

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dsdred/universal-websocket-platform/internal/runtimeconfigload"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnostics"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticsconsumer"
	"github.com/dsdred/universal-websocket-platform/internal/runtimediagnosticspull"
)

// DiagnosticsConsume is dedicated destructive-consumption authority, never Observe.
const DiagnosticsConsume = "runtime.diagnostics.consume"

// ErrDiagnosticsConfiguration indicates invalid trusted construction input.
var ErrDiagnosticsConfiguration = errors.New("invalid diagnostics configuration")

// DiagnosticsPrincipal is an opaque construction-issued capability. Only trusted
// authentication may select a configured handle; request data cannot construct one.
type DiagnosticsPrincipal struct {
	state *diagnosticsState
	slot  uint8
}

// DiagnosticsRequest contains exact caller-known routing, not discovery input.
type DiagnosticsRequest struct {
	Principal       DiagnosticsPrincipal
	Target          Target
	ExpectedAttempt runtimeconfigload.LaunchAttemptID
	Budget          uint8
}

// DiagnosticsAuthorization is the full input of one current policy evaluation.
type DiagnosticsAuthorization struct {
	DiagnosticsRequest
	Action string
}

// DiagnosticsGrant enables one exact principal/action/Target/attempt tuple.
type DiagnosticsGrant struct {
	Principal       DiagnosticsPrincipal
	Action          string
	Target          Target
	ExpectedAttempt runtimeconfigload.LaunchAttemptID
	Enabled         bool
}

// DiagnosticsSnapshot is a detached bounded current grant snapshot. Providers
// transfers an immutable Grants slice to this invocation: the provider must never
// mutate it after returning. No grant history is retained by the adapter.
type DiagnosticsSnapshot struct {
	Grants []DiagnosticsGrant
}

// DiagnosticsPolicy synchronously supplies current trusted policy. It must be
// prompt, concurrency-safe and nonpanicking, with no remote I/O or hidden worker.
// Error/malformed/missing membership denies; success is not an arbitrary allow.
type DiagnosticsPolicy func(context.Context, DiagnosticsAuthorization) (DiagnosticsSnapshot, error)

// DiagnosticsStatus is a safe core classification, not a transport implementation.
type DiagnosticsStatus uint8

const (
	// DiagnosticsUnavailable normalizes deny, missing/mismatch/retired/busy/failure.
	DiagnosticsUnavailable DiagnosticsStatus = iota
	// DiagnosticsInvalid indicates syntactically invalid input, without lookup.
	DiagnosticsInvalid
	// DiagnosticsLimited is process/principal refusal independent of Target.
	DiagnosticsLimited
	// DiagnosticsClosed indicates process-wide shutdown admission.
	DiagnosticsClosed
	// DiagnosticsOK contains one validated detached batch.
	DiagnosticsOK
)

// DiagnosticsResponse contains only safe bytes. Body is borrowed until output
// returns/unwinds and is then cleared. Output must copy explicitly to retain it.
type DiagnosticsResponse struct {
	Status DiagnosticsStatus
	Body   []byte
}

// DiagnosticsAdapter is copyable; copies share all admission/resource state.
// Exactly one sanctioned adapter per process is a composition proof, not a
// constructor-enforced global registry. This isolated core adds no HTTP or auth.
type DiagnosticsAdapter struct{ state *diagnosticsState }

type diagnosticsCall struct{ cancel context.CancelFunc }
type diagnosticsBucket struct {
	at     time.Duration
	tokens uint8
}
type diagnosticsState struct {
	directory       *Directory
	policy          DiagnosticsPolicy
	clock           func() time.Duration
	principals      uint8
	targets         map[Target]uint8
	closed          atomic.Bool
	calls           [8]atomic.Pointer[diagnosticsCall]
	principalActive [32]atomic.Bool
	targetActive    [64]atomic.Bool
	quota           sync.Mutex
	routing         sync.Mutex
	process         diagnosticsBucket
	sampled         time.Duration
	principal       [32]diagnosticsBucket
	target          [64]diagnosticsBucket
}

// NewDiagnosticsAdapter issues principalCount unique handles over the complete
// immutable Directory (1..64 Targets). Clock supplies trusted nonnegative monotonic
// elapsed time, is prompt/concurrency-safe/nonpanicking, and never wraps. A real
// clock may use time.Since(origin). Invalid later samples fail closed. Construction
// performs no policy evaluation, Poll, I/O, worker, timer or lifecycle operation.
func NewDiagnosticsAdapter(d *Directory, principalCount uint8, policy DiagnosticsPolicy, clock func() time.Duration) (DiagnosticsAdapter, []DiagnosticsPrincipal, error) {
	if d == nil || !d.valid() || len(d.scopes) > 64 || principalCount == 0 || principalCount > 32 || policy == nil || clock == nil {
		return DiagnosticsAdapter{}, nil, ErrDiagnosticsConfiguration
	}
	now := clock()
	if now < 0 {
		return DiagnosticsAdapter{}, nil, ErrDiagnosticsConfiguration
	}
	s := &diagnosticsState{directory: d, policy: policy, clock: clock, principals: principalCount, targets: make(map[Target]uint8, len(d.scopes)), process: diagnosticsBucket{now, 8}, sampled: now}
	owners := make(map[any]bool, len(d.scopes))
	for _, selected := range d.scopes {
		if owners[selected.owner] {
			return DiagnosticsAdapter{}, nil, ErrDiagnosticsConfiguration
		}
		owners[selected.owner] = true
		index := uint8(len(s.targets))
		s.targets[selected.target] = index
		s.target[index] = diagnosticsBucket{now, 1}
	}
	handles := make([]DiagnosticsPrincipal, principalCount)
	for i := range handles {
		handles[i] = DiagnosticsPrincipal{s, uint8(i)}
		s.principal[i] = diagnosticsBucket{now, 2}
	}
	return DiagnosticsAdapter{s}, handles, nil
}

// Shutdown atomically closes admission and cancels at most eight in-flight calls.
// It never joins callbacks, output, Host operations or Manager waits. An invocation
// admitted at the routing cut may reach the cancellation-aware Owner operation;
// route admission does not imply View capture. No flush occurs.
func (a DiagnosticsAdapter) Shutdown() {
	if a.state == nil {
		return
	}
	s := a.state
	s.routing.Lock()
	s.closed.Store(true)
	s.routing.Unlock()
	for i := range s.calls {
		if call := s.calls[i].Load(); call != nil {
			call.cancel()
		}
	}
}

// Execute acquires process capacity and a 2s context BEFORE prepare (for future
// decoding); all acquired slots remain held through output. Both callbacks are
// synchronous, use well-formed contexts and must cooperate with cancellation.
// No callback receives raw Owner/Reader/View or arbitrary error details. Panics
// propagate after cleanup; hostile/nonreturning callbacks have no hard bound.
// Output errors lose only this invocation's detached batch; no retry is performed.
func (a DiagnosticsAdapter) Execute(ctx context.Context, prepare func(context.Context) (DiagnosticsRequest, error), output func(context.Context, DiagnosticsResponse) error) DiagnosticsStatus {
	if a.state == nil || ctx == nil || prepare == nil || output == nil {
		return DiagnosticsInvalid
	}
	s := a.state
	if s.closed.Load() {
		return DiagnosticsClosed
	}
	if ctx.Err() != nil {
		return DiagnosticsUnavailable
	}
	callCtx, cancel := context.WithCancel(ctx)
	call := &diagnosticsCall{cancel}
	slot := -1
	for i := range s.calls {
		if s.calls[i].CompareAndSwap(nil, call) {
			slot = i
			break
		}
	}
	if slot < 0 {
		cancel()
		return DiagnosticsLimited
	}
	requestCtx, deadlineCancel := context.WithTimeout(callCtx, 2*time.Second)
	defer func() { deadlineCancel(); cancel(); s.calls[slot].Store(nil) }()
	if s.closed.Load() {
		return DiagnosticsClosed
	}
	if requestCtx.Err() != nil {
		return DiagnosticsUnavailable
	}
	request, err := prepare(requestCtx)
	status := DiagnosticsInvalid
	var body []byte
	lease := diagnosticsLease{principal: -1, target: -1}
	defer func() {
		if lease.target >= 0 {
			s.targetActive[lease.target].Store(false)
		}
		if lease.principal >= 0 {
			s.principalActive[lease.principal].Store(false)
		}
	}()
	defer func() { clear(body); request = DiagnosticsRequest{} }()
	if err == nil {
		status, body = s.consume(requestCtx, request, &lease)
	}
	if requestCtx.Err() != nil {
		return DiagnosticsUnavailable
	}
	if body == nil {
		if status == DiagnosticsInvalid {
			body = []byte(`{"error":"invalid_diagnostics_request"}`)
		} else {
			body = []byte(`{"error":"diagnostics_unavailable"}`)
		}
	}
	if output(requestCtx, DiagnosticsResponse{status, body}) != nil {
		return DiagnosticsUnavailable
	}
	return status
}

func (s *diagnosticsState) principalIndex(p DiagnosticsPrincipal) (uint8, bool) {
	return p.slot, p.state == s && p.slot < s.principals
}

type diagnosticsLease struct{ principal, target int }

func (s *diagnosticsState) consume(ctx context.Context, r DiagnosticsRequest, lease *diagnosticsLease) (DiagnosticsStatus, []byte) {
	if !r.Target.valid() || r.ExpectedAttempt == "" || r.Budget == 0 || r.Budget > 64 {
		return DiagnosticsInvalid, nil
	}
	p, ok := s.principalIndex(r.Principal)
	if !ok || ctx.Err() != nil {
		return DiagnosticsUnavailable, nil
	}
	if !s.charge(p, false) {
		return DiagnosticsLimited, nil
	}
	if !s.principalActive[p].CompareAndSwap(false, true) {
		return DiagnosticsLimited, nil
	}
	lease.principal = int(p)
	snapshot, err := s.policy(ctx, DiagnosticsAuthorization{r, DiagnosticsConsume})
	allowed := err == nil && s.authorized(snapshot, r)
	snapshot = DiagnosticsSnapshot{}
	if !allowed || ctx.Err() != nil || s.closed.Load() {
		return DiagnosticsUnavailable, nil
	}
	// This is the first request-time Directory lookup. Neither valid/admit/Observe
	// nor a scope traversal can run before exact consume authorization.
	if !s.routing.TryLock() {
		return DiagnosticsUnavailable, nil
	}
	if s.closed.Load() || ctx.Err() != nil {
		s.routing.Unlock()
		return DiagnosticsUnavailable, nil
	}
	selected, exists := s.directory.scopes[r.Target.RuntimeInstanceID()]
	s.routing.Unlock()
	if !exists || selected.target != r.Target {
		return DiagnosticsUnavailable, nil
	}
	t, exists := s.targets[r.Target]
	if !exists || !s.targetActive[t].CompareAndSwap(false, true) {
		return DiagnosticsUnavailable, nil
	}
	lease.target = int(t)
	if !s.charge(t, true) || ctx.Err() != nil || s.closed.Load() {
		return DiagnosticsUnavailable, nil
	}
	result, admitted := selected.owner.PollDiagnosticsExpectedAttempt(ctx, r.ExpectedAttempt, r.Budget)
	defer func() { result = runtimediagnosticspull.PollResult{} }()
	if !admitted || ctx.Err() != nil {
		return DiagnosticsUnavailable, nil
	}
	body := diagnosticsProjection(result, r.Budget)
	if body == nil {
		return DiagnosticsUnavailable, nil
	}
	return DiagnosticsOK, body
}

func (s *diagnosticsState) authorized(snapshot DiagnosticsSnapshot, r DiagnosticsRequest) bool {
	if len(snapshot.Grants) > 2048 {
		return false
	}
	// A bounded stack bitmap enforces one grant per principal/configured Target.
	var seen [32][64]bool
	var targets [64]Target
	targetCount := 0
	allowed := false
	for _, grant := range snapshot.Grants {
		p, valid := s.principalIndex(grant.Principal)
		t := 0
		for t < targetCount && targets[t] != grant.Target {
			t++
		}
		if t == targetCount {
			if targetCount == len(targets) {
				return false
			}
			targets[t] = grant.Target
			targetCount++
		}
		if !valid || !grant.Target.valid() || grant.Action != DiagnosticsConsume || grant.ExpectedAttempt == "" || seen[p][t] {
			return false
		}
		seen[p][t] = true
		if grant.Enabled && grant.Principal == r.Principal && grant.Target == r.Target && grant.ExpectedAttempt == r.ExpectedAttempt {
			allowed = true
		}
	}
	return allowed
}

func (s *diagnosticsState) charge(index uint8, target bool) bool {
	now := s.clock()
	if !s.quota.TryLock() {
		return false
	}
	defer s.quota.Unlock()
	if now < 0 || now < s.sampled {
		return false
	}
	if target {
		next, ok := diagnosticsRefill(s.target[index], now, time.Second, 1)
		if !ok {
			return false
		}
		s.sampled = now
		s.target[index] = next
		if next.tokens == 0 {
			return false
		}
		s.target[index].tokens--
		return true
	}
	process, ok := diagnosticsRefill(s.process, now, 125*time.Millisecond, 8)
	principal, valid := diagnosticsRefill(s.principal[index], now, 500*time.Millisecond, 2)
	if !ok || !valid {
		return false
	}
	s.sampled = now
	s.process = process
	s.principal[index] = principal
	if process.tokens == 0 || principal.tokens == 0 {
		return false
	}
	s.process.tokens--
	s.principal[index].tokens--
	return true
}

func diagnosticsRefill(b diagnosticsBucket, now, interval time.Duration, burst uint8) (diagnosticsBucket, bool) {
	if now < 0 || now < b.at || interval <= 0 || b.at < 0 || b.tokens > burst {
		return b, false
	}
	whole := (now - b.at) / interval
	if whole > 0 {
		b.at = now - (now-b.at)%interval
		if whole >= time.Duration(burst-b.tokens) {
			b.tokens = burst
		} else {
			b.tokens += uint8(whole)
		}
	}
	return b, true
}

type diagnosticsEventJSON struct {
	SchemaVersion uint8                   `json:"schemaVersion"`
	Code          runtimediagnostics.Code `json:"code"`
	Sequence      string                  `json:"sequence"`
}
type diagnosticsReceiptJSON struct {
	ReadAttempts  uint8  `json:"readAttempts"`
	Observed      uint8  `json:"observed"`
	SinkSucceeded uint8  `json:"sinkSucceeded"`
	SinkFailed    uint8  `json:"sinkFailed"`
	SinkPanicked  uint8  `json:"sinkPanicked"`
	Discarded     uint8  `json:"discarded"`
	Stop          string `json:"stop"`
}

func diagnosticsProjection(result runtimediagnosticspull.PollResult, budget uint8) []byte {
	r := result.Receipt
	if budget == 0 || budget > 64 || result.Count > budget || r.SinkSucceeded != result.Count || r.Observed != result.Count || r.SinkFailed != 0 || r.SinkPanicked != 0 || r.Discarded != 0 || r.ReadAttempts > budget {
		return nil
	}
	stop := ""
	switch r.Stop {
	case runtimediagnosticsconsumer.NoItem:
		if int(r.ReadAttempts) != int(result.Count)+1 {
			return nil
		}
		stop = "no_item"
	case runtimediagnosticsconsumer.Budget:
		if r.ReadAttempts != budget || result.Count != budget {
			return nil
		}
		stop = "budget"
	default:
		return nil
	}
	events := make([]diagnosticsEventJSON, result.Count)
	defer clear(events)
	for i, event := range result.Events {
		if i >= int(result.Count) {
			if event != (runtimediagnostics.Event{}) {
				return nil
			}
			continue
		}
		if event.SchemaVersion != 1 || event.Code != runtimediagnostics.ListenerServeFailed && event.Code != runtimediagnostics.AuthenticationProviderFailed || event.Sequence == 0 {
			return nil
		}
		events[i] = diagnosticsEventJSON{event.SchemaVersion, event.Code, strconv.FormatUint(event.Sequence, 10)}
	}
	projection := struct {
		APIVersion uint8                  `json:"apiVersion"`
		Count      uint8                  `json:"count"`
		Events     []diagnosticsEventJSON `json:"events"`
		Receipt    diagnosticsReceiptJSON `json:"receipt"`
	}{1, result.Count, events, diagnosticsReceiptJSON{r.ReadAttempts, r.Observed, r.SinkSucceeded, 0, 0, 0, stop}}
	body, err := json.Marshal(projection)
	if err != nil || len(body) > 16*1024 {
		clear(body)
		return nil
	}
	return body
}
