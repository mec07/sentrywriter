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
	writer := sentrywriter.New().WithClient(client).WithUserID("userID").
		WithLogLevel(sentrywriter.LogLevel{"error", sentry.LevelError})

	log := `{"level":"error","message":"blah"}`

	n, err := writer.Write([]byte(log))
	if err != nil {
		t.Fatalf("writer.Writer: %v", err)
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
	writer := sentrywriter.New().WithClient(client).
		WithLogLevel(sentrywriter.LogLevel{"error", sentry.LevelError})

	log := `{"level":"info","message":"blah"}`

	n, err := writer.Write([]byte(log))
	if err != nil {
		t.Fatalf("writer.Writer: %v", err)
	}
	assert.Equal(t, len(log), n)

	messages := client.getMessages()

	if len(messages) != 0 {
		t.Fatalf("Expected 0 message, found: %d", len(messages))
	}
}

func TestSentryWriterNonJSONError(t *testing.T) {
	client := &mockClient{}
	writer := sentrywriter.New().WithClient(client).
		WithLogLevel(sentrywriter.LogLevel{"error", sentry.LevelError})

	log := `invalid json`
	_, err := writer.Write([]byte(log))
	if err == nil {
		t.Fatal("expected an error")
	}
}
