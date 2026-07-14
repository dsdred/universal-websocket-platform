package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/dsdred/universal-websocket-platform/internal/authentication"
	"github.com/dsdred/universal-websocket-platform/internal/message"
)

func TestDefaultSessionImplementsSession(t *testing.T) {
	var _ Session = (*DefaultSession)(nil)
}

func TestNewSession(t *testing.T) {
	serverConnection, _ := testWebSocketPair(t)
	principal := validPrincipal()
	createdBefore := time.Now().UTC()
	session, err := New(serverConnection, principal, " 192.0.2.1:4321 ")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if session.ID() == "" {
		t.Fatal("ID() is empty")
	}
	if session.RemoteAddress() != "192.0.2.1:4321" {
		t.Fatalf("RemoteAddress() = %q", session.RemoteAddress())
	}
	if session.CreatedAt().Before(createdBefore) || session.CreatedAt().Location() != time.UTC {
		t.Fatalf("CreatedAt() = %v", session.CreatedAt())
	}
	if session.Running() {
		t.Fatal("Running() = true before Start")
	}
}

func TestNewSessionRejectsNilConnectionAndInvalidPrincipal(t *testing.T) {
	if session, err := New(nil, validPrincipal(), ""); session != nil || !errors.Is(err, ErrNilConnection) {
		t.Fatalf("New(nil connection) = (%v, %v)", session, err)
	}
	serverConnection, _ := testWebSocketPair(t)
	for _, principal := range []authentication.Principal{
		{},
		{Authenticated: true, Anonymous: true},
	} {
		if session, err := New(serverConnection, principal, ""); session != nil || !errors.Is(err, ErrInvalidPrincipal) {
			t.Fatalf("New(invalid Principal) = (%v, %v)", session, err)
		}
	}
}

func TestNewSessionReturnsIDGenerationError(t *testing.T) {
	serverConnection, _ := testWebSocketPair(t)
	wantErr := errors.New("random source unavailable")
	session, err := newWithIDGenerator(serverConnection, validPrincipal(), "", func() (string, error) {
		return "", wantErr
	})
	if session != nil || !errors.Is(err, wantErr) {
		t.Fatalf("newWithIDGenerator() = (%v, %v)", session, err)
	}
}

func TestSessionIDsAreUniqueForConcurrentCreation(t *testing.T) {
	serverConnection, _ := testWebSocketPair(t)
	const sessions = 64
	ids := make(chan string, sessions)
	errorsChannel := make(chan error, sessions)
	var waitGroup sync.WaitGroup
	for range sessions {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			runtimeSession, err := New(serverConnection, validPrincipal(), "")
			if err != nil {
				errorsChannel <- err
				return
			}
			ids <- runtimeSession.ID()
		}()
	}
	waitGroup.Wait()
	close(ids)
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("New() error = %v", err)
	}
	unique := make(map[string]struct{}, sessions)
	for id := range ids {
		if _, exists := unique[id]; exists {
			t.Fatalf("duplicate Session ID %q", id)
		}
		unique[id] = struct{}{}
	}
	if len(unique) != sessions {
		t.Fatalf("unique Session IDs = %d, want %d", len(unique), sessions)
	}
}

func TestSessionPrincipalDeepCopy(t *testing.T) {
	serverConnection, _ := testWebSocketPair(t)
	principal := validPrincipal()
	session, err := New(serverConnection, principal, "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	principal.Claims["tenant"] = "changed"
	principal.Roles[0] = "changed"
	principal.Attributes["region"] = "changed"
	principal.Metadata["provider"] = "changed"
	firstCopy := session.Principal()
	firstCopy.Claims["tenant"] = "changed-again"
	firstCopy.Roles[0] = "changed-again"
	firstCopy.Attributes["region"] = "changed-again"
	firstCopy.Metadata["provider"] = "changed-again"

	unchanged := session.Principal()
	if unchanged.Claims["tenant"] != "alpha" || unchanged.Roles[0] != "admin" ||
		unchanged.Attributes["region"] != "eu" || unchanged.Metadata["provider"] != "api-key" {
		t.Fatalf("Principal() = %+v", unchanged)
	}
}

func TestSessionStoresOnlySafeTransportMetadata(t *testing.T) {
	sessionType := reflect.TypeOf(DefaultSession{})
	requestType := reflect.TypeOf((*http.Request)(nil))
	for index := 0; index < sessionType.NumField(); index++ {
		field := sessionType.Field(index)
		fieldName := strings.ToLower(field.Name)
		if field.Type == requestType || strings.Contains(fieldName, "header") || strings.Contains(fieldName, "query") ||
			strings.Contains(fieldName, "request") || strings.Contains(fieldName, "cookie") ||
			strings.Contains(fieldName, "credential") {
			t.Fatalf("DefaultSession contains unsafe transport field %q (%s)", field.Name, field.Type)
		}
	}
}

func TestSessionStartAndDoubleStart(t *testing.T) {
	session := newTestSession(t)
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if !session.Running() {
		t.Fatal("Running() = false after Start")
	}
	if err := session.Start(context.Background()); !errors.Is(err, ErrSessionAlreadyRunning) {
		t.Fatalf("second Start() error = %v", err)
	}
}

func TestSessionCanceledStartDoesNotChangeState(t *testing.T) {
	session := newTestSession(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := session.Start(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Start() error = %v, want context.Canceled", err)
	}
	if session.Running() {
		t.Fatal("Running() = true after canceled Start")
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() after cancellation error = %v", err)
	}
}

func TestSessionStopSendsNormalClosure(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	assertNormalSessionStop(t, session, clientConnection)
	if session.Running() {
		t.Fatal("Running() = true after Stop")
	}
}

func TestSessionStopBeforeStartAndRestartForbidden(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	assertNormalSessionStop(t, session, clientConnection)
	if err := session.Start(context.Background()); !errors.Is(err, ErrSessionAlreadyRunning) {
		t.Fatalf("Start() after Stop error = %v, want ErrSessionAlreadyRunning", err)
	}
	if err := session.Stop(context.Background()); err != nil {
		t.Fatalf("second Stop() error = %v", err)
	}
}

func TestSessionConcurrentStart(t *testing.T) {
	session := newTestSession(t)
	const goroutines = 64
	var successes atomic.Uint64
	var alreadyRunning atomic.Uint64
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			err := session.Start(context.Background())
			switch {
			case err == nil:
				successes.Add(1)
			case errors.Is(err, ErrSessionAlreadyRunning):
				alreadyRunning.Add(1)
			default:
				t.Errorf("Start() error = %v", err)
			}
		}()
	}
	waitGroup.Wait()
	if successes.Load() != 1 || alreadyRunning.Load() != goroutines-1 {
		t.Fatalf("Start outcomes = (%d success, %d already running)", successes.Load(), alreadyRunning.Load())
	}
}

func TestSessionConcurrentStop(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	const goroutines = 64
	errorsChannel := make(chan error, goroutines)
	var waitGroup sync.WaitGroup
	for range goroutines {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if stopErr := session.Stop(context.Background()); stopErr != nil {
				errorsChannel <- stopErr
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, readErr := clientConnection.Read(ctx)
	if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d, want %d", status, websocket.StatusNormalClosure)
	}
	waitGroup.Wait()
	close(errorsChannel)
	for stopErr := range errorsChannel {
		t.Errorf("Stop() error = %v", stopErr)
	}
	if session.Running() {
		t.Fatal("Running() = true after concurrent Stop")
	}
}

func TestSessionRunBeforeStart(t *testing.T) {
	session := newTestSession(t)
	if err := session.Run(context.Background()); !errors.Is(err, ErrSessionNotRunning) {
		t.Fatalf("Run() error = %v, want ErrSessionNotRunning", err)
	}
}

func TestSessionRunReadsTextBinaryAndMultipleMessages(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	received := make(chan message.Message, 3)
	session, err := newWithObserver(serverConnection, validPrincipal(), "", func(runtimeMessage message.Message) {
		received <- runtimeMessage
	})
	if err != nil {
		t.Fatalf("newWithObserver() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runResult := make(chan error, 1)
	go func() { runResult <- session.Run(context.Background()) }()

	expected := []struct {
		websocketType websocket.MessageType
		messageType   message.Type
		payload       string
	}{
		{websocket.MessageText, message.TypeText, "first text"},
		{websocket.MessageBinary, message.TypeBinary, "binary data"},
		{websocket.MessageText, message.TypeText, "second text"},
	}
	for _, expectedMessage := range expected {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := clientConnection.Write(ctx, expectedMessage.websocketType, []byte(expectedMessage.payload))
		cancel()
		if err != nil {
			t.Fatalf("Write() error = %v", err)
		}
		select {
		case actual := <-received:
			if actual.Type() != expectedMessage.messageType || string(actual.Data()) != expectedMessage.payload {
				t.Fatalf("Message = (%q, %q), want (%q, %q)", actual.Type(), actual.Data(), expectedMessage.messageType, expectedMessage.payload)
			}
		case <-time.After(time.Second):
			t.Fatal("Run() did not observe message")
		}
		select {
		case runErr := <-runResult:
			t.Fatalf("Run() returned while connection remained open: %v", runErr)
		default:
		}
	}

	closeClientAndWaitForRun(t, clientConnection, websocket.StatusNormalClosure, runResult)
	if err := session.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSessionRunWithNilHandlerKeepsDiscardBehavior(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	received := make(chan message.Message, 1)
	session, err := newWithObserver(serverConnection, validPrincipal(), "", func(runtimeMessage message.Message) {
		received <- runtimeMessage
	})
	if err != nil {
		t.Fatalf("newWithObserver() error = %v", err)
	}
	if session.handler != nil {
		t.Fatal("Handler is not nil")
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runResult := make(chan error, 1)
	go func() { runResult <- session.Run(context.Background()) }()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	if err := clientConnection.Write(ctx, websocket.MessageText, []byte("discarded")); err != nil {
		cancel()
		t.Fatalf("client Write() error = %v", err)
	}
	cancel()
	select {
	case runtimeMessage := <-received:
		if string(runtimeMessage.Data()) != "discarded" {
			t.Fatalf("observed payload = %q", runtimeMessage.Data())
		}
	case <-time.After(time.Second):
		t.Fatal("nil-handler read loop did not consume Message")
	}
	closeClientAndWaitForRun(t, clientConnection, websocket.StatusNormalClosure, runResult)
	if err := session.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSessionRunWithEchoHandlerContinuesAcrossMessages(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := NewWithHandler(serverConnection, validPrincipal(), "", message.NewEchoHandler())
	if err != nil {
		t.Fatalf("NewWithHandler() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runResult := make(chan error, 1)
	go func() { runResult <- session.Run(context.Background()) }()
	waitForReadLoop(t, session, true)

	for _, expected := range []struct {
		messageType websocket.MessageType
		payload     []byte
	}{
		{messageType: websocket.MessageText, payload: []byte("echo text")},
		{messageType: websocket.MessageBinary, payload: []byte{0x00, 0x01}},
		{messageType: websocket.MessageText, payload: []byte("echo continues")},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		if err := clientConnection.Write(ctx, expected.messageType, expected.payload); err != nil {
			cancel()
			t.Fatalf("client Write() error = %v", err)
		}
		cancel()
		actualType, actualPayload := readClientMessage(t, clientConnection)
		if actualType != expected.messageType || !bytes.Equal(actualPayload, expected.payload) {
			t.Fatalf("echo = (%d, %v), want (%d, %v)", actualType, actualPayload, expected.messageType, expected.payload)
		}
	}
	closeClientAndWaitForRun(t, clientConnection, websocket.StatusNormalClosure, runResult)
	if err := session.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSessionRunReturnsHandlerError(t *testing.T) {
	wantErr := errors.New("handler failed")
	serverConnection, clientConnection := testWebSocketPair(t)
	var receivedSender message.Sender
	handler := messageHandlerFunc(func(_ context.Context, sender message.Sender, _ message.Message) error {
		receivedSender = sender
		return wantErr
	})
	session, err := NewWithHandler(serverConnection, validPrincipal(), "", handler)
	if err != nil {
		t.Fatalf("NewWithHandler() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runResult := make(chan error, 1)
	go func() { runResult <- session.Run(context.Background()) }()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	if err := clientConnection.Write(ctx, websocket.MessageText, []byte("fail")); err != nil {
		cancel()
		t.Fatalf("client Write() error = %v", err)
	}
	cancel()
	select {
	case err := <-runResult:
		if !errors.Is(err, wantErr) {
			t.Fatalf("Run() error = %v, want Handler error", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not return Handler error")
	}
	if receivedSender != session {
		t.Fatal("Handler did not receive current Session Sender")
	}
	clientConnection.CloseNow()
	_ = session.Stop(context.Background())
}

func TestSessionDoesNotStoreMessagePayload(t *testing.T) {
	sessionType := reflect.TypeOf(DefaultSession{})
	bytesType := reflect.TypeOf([]byte(nil))
	for index := 0; index < sessionType.NumField(); index++ {
		if sessionType.Field(index).Type == bytesType {
			t.Fatalf("DefaultSession stores payload-like field %q", sessionType.Field(index).Name)
		}
	}
}

func TestSessionRejectsSecondConcurrentRun(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runResult := make(chan error, 1)
	go func() { runResult <- session.Run(context.Background()) }()
	waitForReadLoop(t, session, true)
	if err := session.Run(context.Background()); !errors.Is(err, ErrSessionReadLoopAlreadyRunning) {
		t.Fatalf("second Run() error = %v, want ErrSessionReadLoopAlreadyRunning", err)
	}
	closeClientAndWaitForRun(t, clientConnection, websocket.StatusNormalClosure, runResult)
	if err := session.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSessionRunTreatsNormalAndGoingAwayAsCleanClosure(t *testing.T) {
	for _, status := range []websocket.StatusCode{websocket.StatusNormalClosure, websocket.StatusGoingAway} {
		t.Run(fmt.Sprintf("status_%d", status), func(t *testing.T) {
			session, clientConnection, runResult := runningSessionReadLoop(t, context.Background())
			closeClientAndWaitForRun(t, clientConnection, status, runResult)
			if session.Running() != true {
				t.Fatal("Run() changed Session lifecycle state")
			}
			if err := session.Stop(context.Background()); err != nil {
				t.Fatalf("Stop() error = %v", err)
			}
		})
	}
}

func TestSessionRunReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	session, _, runResult := runningSessionReadLoop(t, ctx)
	cancel()
	select {
	case err := <-runResult:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not return after context cancellation")
	}
	if err := session.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSessionRunReturnsUnexpectedReadError(t *testing.T) {
	session, clientConnection, runResult := runningSessionReadLoop(t, context.Background())
	if err := clientConnection.CloseNow(); err != nil {
		t.Fatalf("client CloseNow() error = %v", err)
	}
	select {
	case err := <-runResult:
		if err == nil || websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
			t.Fatalf("Run() error = %v, want unexpected read error", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not return after connection failure")
	}
	_ = session.Stop(context.Background())
	if session.Running() {
		t.Fatal("Running() = true after cleanup Stop")
	}
}

func TestSessionStopInterruptsReadLoopWithoutDeadlock(t *testing.T) {
	session, clientConnection, runResult := runningSessionReadLoop(t, context.Background())
	stopResult := make(chan error, 1)
	go func() { stopResult <- session.Stop(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, readErr := clientConnection.Read(ctx)
	if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
		t.Fatalf("client close status = %d, want %d (error %v)", status, websocket.StatusNormalClosure, readErr)
	}
	select {
	case err := <-runResult:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Run() did not stop: %v", ctx.Err())
	}
	select {
	case err := <-stopResult:
		if err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Stop() did not return: %v", ctx.Err())
	}
	waitForReadLoop(t, session, false)
	if session.Running() {
		t.Fatal("Running() = true after Stop")
	}
	if err := session.Start(context.Background()); !errors.Is(err, ErrSessionAlreadyRunning) {
		t.Fatalf("Start() after Stop error = %v, want ErrSessionAlreadyRunning", err)
	}
}

func TestSessionSendTextAndBinaryMessages(t *testing.T) {
	for _, test := range []struct {
		name          string
		messageType   message.Type
		websocketType websocket.MessageType
		payload       []byte
	}{
		{name: "text", messageType: message.TypeText, websocketType: websocket.MessageText, payload: []byte("outbound text")},
		{name: "binary", messageType: message.TypeBinary, websocketType: websocket.MessageBinary, payload: []byte{0x00, 0x01, 0xff}},
	} {
		t.Run(test.name, func(t *testing.T) {
			serverConnection, clientConnection := testWebSocketPair(t)
			session, err := New(serverConnection, validPrincipal(), "")
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if err := session.Start(context.Background()); err != nil {
				t.Fatalf("Start() error = %v", err)
			}
			runtimeMessage, err := message.New(test.messageType, test.payload)
			if err != nil {
				t.Fatalf("message.New() error = %v", err)
			}
			if err := session.Send(context.Background(), runtimeMessage); err != nil {
				t.Fatalf("Send() error = %v", err)
			}
			actualType, actualPayload := readClientMessage(t, clientConnection)
			if actualType != test.websocketType || !bytes.Equal(actualPayload, test.payload) {
				t.Fatalf("received = (%d, %v), want (%d, %v)", actualType, actualPayload, test.websocketType, test.payload)
			}
		})
	}
}

func TestSessionSendUsesCopiedMessagePayload(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	payload := []byte("original")
	runtimeMessage, err := message.New(message.TypeText, payload)
	if err != nil {
		t.Fatalf("message.New() error = %v", err)
	}
	payload[0] = 'X'
	if err := session.Send(context.Background(), runtimeMessage); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	_, actualPayload := readClientMessage(t, clientConnection)
	if string(actualPayload) != "original" {
		t.Fatalf("received payload = %q, want original", actualPayload)
	}
}

func TestSessionConcurrentSendSerializesWrites(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	const messages = 64
	errorsChannel := make(chan error, messages)
	var waitGroup sync.WaitGroup
	for index := range messages {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			runtimeMessage, messageErr := message.New(message.TypeText, []byte(fmt.Sprintf("message-%d", index)))
			if messageErr != nil {
				errorsChannel <- messageErr
				return
			}
			if sendErr := session.Send(context.Background(), runtimeMessage); sendErr != nil {
				errorsChannel <- sendErr
			}
		}()
	}
	received := make(map[string]struct{}, messages)
	for range messages {
		messageType, payload := readClientMessage(t, clientConnection)
		if messageType != websocket.MessageText {
			t.Fatalf("received type = %d, want text", messageType)
		}
		received[string(payload)] = struct{}{}
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Errorf("concurrent Send() error = %v", err)
	}
	if len(received) != messages {
		t.Fatalf("unique received messages = %d, want %d", len(received), messages)
	}
}

func TestSessionSendRequiresRunningLifecycle(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	runtimeMessage, err := message.New(message.TypeText, []byte("payload"))
	if err != nil {
		t.Fatalf("message.New() error = %v", err)
	}
	if err := session.Send(context.Background(), runtimeMessage); !errors.Is(err, ErrSessionNotRunning) {
		t.Fatalf("Send() before Start error = %v, want ErrSessionNotRunning", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	assertNormalSessionStop(t, session, clientConnection)
	if err := session.Send(context.Background(), runtimeMessage); !errors.Is(err, ErrSessionNotRunning) {
		t.Fatalf("Send() after Stop error = %v, want ErrSessionNotRunning", err)
	}
}

func TestSessionSendDuringStopReturnsNotRunning(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	stopResult := make(chan error, 1)
	go func() { stopResult <- session.Stop(context.Background()) }()
	waitForSessionState(t, session, stateStopping)
	runtimeMessage, err := message.New(message.TypeText, []byte("too late"))
	if err != nil {
		t.Fatalf("message.New() error = %v", err)
	}
	if err := session.Send(context.Background(), runtimeMessage); !errors.Is(err, ErrSessionNotRunning) {
		t.Fatalf("Send() during Stop error = %v, want ErrSessionNotRunning", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, _, err := clientConnection.Read(ctx); websocket.CloseStatus(err) != websocket.StatusNormalClosure {
		t.Fatalf("client close error = %v", err)
	}
	select {
	case err := <-stopResult:
		if err != nil {
			t.Fatalf("Stop() error = %v", err)
		}
	case <-ctx.Done():
		t.Fatalf("Stop() did not return: %v", ctx.Err())
	}
}

func TestSessionSendRejectsInvalidMessageType(t *testing.T) {
	session := newTestSession(t)
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := session.Send(context.Background(), message.Message{}); !errors.Is(err, message.ErrInvalidMessageType) {
		t.Fatalf("Send() error = %v, want ErrInvalidMessageType", err)
	}
}

func TestSessionReadLoopContinuesDuringSend(t *testing.T) {
	serverConnection, clientConnection := testWebSocketPair(t)
	received := make(chan message.Message, 1)
	session, err := newWithObserver(serverConnection, validPrincipal(), "", func(runtimeMessage message.Message) {
		received <- runtimeMessage
	})
	if err != nil {
		t.Fatalf("newWithObserver() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runResult := make(chan error, 1)
	go func() { runResult <- session.Run(context.Background()) }()
	waitForReadLoop(t, session, true)

	outbound, err := message.New(message.TypeText, []byte("server message"))
	if err != nil {
		t.Fatalf("message.New() error = %v", err)
	}
	if err := session.Send(context.Background(), outbound); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	messageType, payload := readClientMessage(t, clientConnection)
	if messageType != websocket.MessageText || string(payload) != "server message" {
		t.Fatalf("outbound = (%d, %q)", messageType, payload)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	if err := clientConnection.Write(ctx, websocket.MessageBinary, []byte("client message")); err != nil {
		cancel()
		t.Fatalf("client Write() error = %v", err)
	}
	cancel()
	select {
	case inbound := <-received:
		if inbound.Type() != message.TypeBinary || string(inbound.Data()) != "client message" {
			t.Fatalf("inbound = (%q, %q)", inbound.Type(), inbound.Data())
		}
	case <-time.After(time.Second):
		t.Fatal("read loop did not continue during Send")
	}
	closeClientAndWaitForRun(t, clientConnection, websocket.StatusNormalClosure, runResult)
	if err := session.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func runningSessionReadLoop(
	t *testing.T,
	ctx context.Context,
) (*DefaultSession, *websocket.Conn, <-chan error) {
	t.Helper()
	serverConnection, clientConnection := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := session.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	runResult := make(chan error, 1)
	go func() { runResult <- session.Run(ctx) }()
	waitForReadLoop(t, session, true)
	return session, clientConnection, runResult
}

func closeClientAndWaitForRun(
	t *testing.T,
	clientConnection *websocket.Conn,
	status websocket.StatusCode,
	runResult <-chan error,
) {
	t.Helper()
	if err := clientConnection.Close(status, ""); err != nil {
		t.Fatalf("client Close() error = %v", err)
	}
	select {
	case err := <-runResult:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not return after client closure")
	}
}

func waitForReadLoop(t *testing.T, session *DefaultSession, active bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		session.mu.RLock()
		actual := session.readLoop
		session.mu.RUnlock()
		if actual == active {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("read loop active = %t, want %t", !active, active)
}

func waitForSessionState(t *testing.T, session *DefaultSession, state lifecycleState) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		session.mu.RLock()
		actual := session.state
		session.mu.RUnlock()
		if actual == state {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("Session state did not become %d", state)
}

func readClientMessage(t *testing.T, connection *websocket.Conn) (websocket.MessageType, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	messageType, payload, err := connection.Read(ctx)
	if err != nil {
		t.Fatalf("client Read() error = %v", err)
	}
	return messageType, payload
}

type messageHandlerFunc func(context.Context, message.Sender, message.Message) error

func (handler messageHandlerFunc) Handle(
	ctx context.Context,
	sender message.Sender,
	runtimeMessage message.Message,
) error {
	return handler(ctx, sender, runtimeMessage)
}

func newTestSession(t *testing.T) *DefaultSession {
	t.Helper()
	serverConnection, _ := testWebSocketPair(t)
	session, err := New(serverConnection, validPrincipal(), "192.0.2.1:4321")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return session
}

func validPrincipal() authentication.Principal {
	return authentication.Principal{
		ID:                     "administrator",
		Name:                   "Administrator",
		AuthenticationProvider: "api-key",
		Claims:                 map[string]string{"tenant": "alpha"},
		Roles:                  []string{"admin"},
		Attributes:             map[string]string{"region": "eu"},
		Authenticated:          true,
		Metadata:               map[string]string{"provider": "api-key"},
	}
}

func assertNormalSessionStop(t *testing.T, session Session, clientConnection *websocket.Conn) {
	t.Helper()
	stopResult := make(chan error, 1)
	go func() {
		stopResult <- session.Stop(context.Background())
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, readErr := clientConnection.Read(ctx)
	if status := websocket.CloseStatus(readErr); status != websocket.StatusNormalClosure {
		t.Fatalf("close status = %d, want %d (error %v)", status, websocket.StatusNormalClosure, readErr)
	}
	select {
	case stopErr := <-stopResult:
		if stopErr != nil {
			t.Fatalf("Stop() error = %v", stopErr)
		}
	case <-ctx.Done():
		t.Fatalf("Stop() did not return: %v", ctx.Err())
	}
}

func testWebSocketPair(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()
	accepted := make(chan *websocket.Conn, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		connection, err := websocket.Accept(response, request, nil)
		if err != nil {
			return
		}
		accepted <- connection
		<-release
	}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	clientConnection, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	cancel()
	if err != nil {
		close(release)
		server.Close()
		t.Fatalf("Dial() error = %v", err)
	}
	var serverConnection *websocket.Conn
	select {
	case serverConnection = <-accepted:
	case <-time.After(time.Second):
		clientConnection.CloseNow()
		close(release)
		server.Close()
		t.Fatal("server did not accept WebSocket connection")
	}
	t.Cleanup(func() {
		serverConnection.CloseNow()
		clientConnection.CloseNow()
		close(release)
		server.Close()
	})
	return serverConnection, clientConnection
}
