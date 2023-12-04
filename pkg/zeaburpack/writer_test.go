package zeaburpack

import (
	"testing"
	"time"

	"github.com/containerd/console"
	"github.com/stretchr/testify/assert"
)

type MockWriter struct {
	written string
}

func (m *MockWriter) Read(p []byte) (n int, err error) {
	n = copy(p, m.written)
	return
}

func (m *MockWriter) Close() error {
	return nil
}

func (m *MockWriter) Fd() uintptr {
	return 0
}

func (m *MockWriter) Name() string {
	return "mock"
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	m.written = string(p)
	return len(p), nil
}

func (m *MockWriter) Flush() {}

var _ console.File = &MockWriter{}

func TestHandledWriter_Write(t *testing.T) {
	t.Parallel()

	mockWriter := &MockWriter{}
	receivedLog := ""
	handler := func(log string) {
		receivedLog = log
	}

	handledWriter := NewHandledWriter(mockWriter, &handler)

	_, _ = handledWriter.Write([]byte("test"))
	// wait for 3 ms – `go h.handler(string(p))` is async
	time.Sleep(3 * time.Millisecond)

	assert.Equal(t, "test", mockWriter.written)
	assert.Equal(t, "test", receivedLog)
}
