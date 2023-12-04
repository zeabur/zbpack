package zeaburpack

import (
	"github.com/containerd/console"
)

// HandledWriter is a console-compatible writer that copies the output
// to the specified log and calls handler with the log.
type HandledWriter struct {
	console.File
	handler func(log string)
}

// NewHandledWriter creates a new HandledWriter.
func NewHandledWriter(w console.File, handler *func(log string)) console.File {
	if handler == nil {
		return w
	}

	return &HandledWriter{
		File:    w,
		handler: *handler,
	}
}

func (h HandledWriter) Write(p []byte) (n int, err error) {
	go h.handler(string(p))
	return h.File.Write(p)
}

var _ console.File = (*HandledWriter)(nil)
