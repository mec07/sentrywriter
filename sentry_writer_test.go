package sentrywriter_test

import (
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/mec07/sentrywriter"
	"gotest.tools/assert"
)

type mockClient struct {
	sync.Mutex
	messages []string
}

func (m *mockClient) Flush(timeout time.Duration) bool {
	return true
}

func (m *mockClient) CaptureMessage(message string, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID {
	m.Lock()
	defer m.Unlock()

	m.messages = append(m.messages, message)
	return &sentry.NewEvent().EventID
}

func (m *mockClient) getMessages() []string {
	m.Lock()
	defer m.Unlock()

	messages := make([]string, len(m.messages))
	copy(messages, m.messages)
	return messages
}

func TestSentryWriterWrite(t *testing.T) {
	client := &mockClient{}
	writer := sentrywriter.New(sentrywriter.LogLevel{"fatal", sentry.LevelFatal}).WithClient(client).WithUserID("userID").
		WithLogLevel(sentrywriter.LogLevel{"error", sentry.LevelError}).WithBreadcrumbs(20)

	log := `{"level":"error","message":"blah"}`

	n, err := writer.Write([]byte(log))
	if err != nil {
		t.Fatalf("writer.Write: %v", err)
	}
	assert.Equal(t, len(log), n)

	messages := client.getMessages()

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, found: %d", len(messages))
	}
	assert.Equal(t, log, messages[0])
}

func TestSentryWriterWriteFiltersLogs(t *testing.T) {
	client := &mockClient{}
	writer := sentrywriter.New(sentrywriter.LogLevel{"fatal", sentry.LevelFatal}).WithClient(client).
		WithLogLevel(sentrywriter.LogLevel{"error", sentry.LevelError})

	log := `{"level":"info","message":"blah"}`

	n, err := writer.Write([]byte(log))
	if err != nil {
		t.Fatalf("writer.Write: %v", err)
	}
	assert.Equal(t, len(log), n)

	messages := client.getMessages()

	if len(messages) != 0 {
		t.Fatalf("Expected 0 message, found: %d", len(messages))
	}
}

// Although there isn't really a helpful observation to make about breadcrumbs
// we can still ensure nothing goes terribly wrong by logging without
// breadcrumbs turned on, then with breadcrumbs turned on.
func TestSentryWriterBreadcrumbsPaths(t *testing.T) {
	// Add a log before turning on breadcrumbs
	client := &mockClient{}
	writer := sentrywriter.New(sentrywriter.LogLevel{"fatal", sentry.LevelFatal}).WithClient(client).
		WithLogLevel(sentrywriter.LogLevel{"error", sentry.LevelError})

	log := `{"level":"info","message":"blah"}`

	n, err := writer.Write([]byte(log))
	if err != nil {
		t.Fatalf("writer.Write: %v", err)
	}
	assert.Equal(t, len(log), n)

	// Add a log after turning on breadcrumbs
	writer.WithBreadcrumbs(20)
	n, err = writer.Write([]byte(log))
	if err != nil {
		t.Fatalf("writer.Write: %v", err)
	}
	assert.Equal(t, len(log), n)

	// Add final log which should send event with breadcrumb included
	errorLog := `{"level":"error","message":"blah"}`
	n, err = writer.Write([]byte(errorLog))
	if err != nil {
		t.Fatalf("writer.Write: %v", err)
	}
	assert.Equal(t, len(errorLog), n)

	messages := client.getMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, found: %d", len(messages))
	}
	assert.Equal(t, errorLog, messages[0])

}

func TestSentryWriterNonJSONError(t *testing.T) {
	client := &mockClient{}
	writer := sentrywriter.New(sentrywriter.LogLevel{"error", sentry.LevelError}).WithClient(client).
		WithBreadcrumbs(20)

	log := `invalid json`
	_, err := writer.Write([]byte(log))
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestSentryWriterNoFilterByDefault(t *testing.T) {
	// Do not add any filters
	client := &mockClient{}
	writer := sentrywriter.New().WithClient(client)

	log := `just a random log which isn't json formatted`

	// the non-json log can get through fine
	n, err := writer.Write([]byte(log))
	if err != nil {
		t.Fatalf("writer.Write: %v", err)
	}
	assert.Equal(t, len(log), n)

	// Now add filters
	writer = writer.WithLogLevel(sentrywriter.LogLevel{"error", sentry.LevelError})
	_, err = writer.Write([]byte(log))
	if err == nil {
		t.Fatal("expected an error")
	}
}
