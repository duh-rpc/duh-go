package duh

import (
	"bufio"
	"io"
	"sync"
)

type StandardLogger interface {
	Error(msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
}

type NoOpLogger struct{}

func (NoOpLogger) Error(msg string, args ...any) {}
func (NoOpLogger) Info(msg string, args ...any)  {}
func (NoOpLogger) Debug(msg string, args ...any) {}
func (NoOpLogger) Warn(msg string, args ...any)  {}

type HttpLogAdaptor struct {
	writer *io.PipeWriter
	closer func()
}

func (l *HttpLogAdaptor) Write(p []byte) (n int, err error) {
	return l.writer.Write(p)
}

func (l *HttpLogAdaptor) Close() error {
	l.closer()
	return nil
}

var _ io.WriteCloser = &HttpLogAdaptor{}

// NewHttpLogAdaptor creates a new adaptor suitable for forwarding logging from ErrorLog to a standard logger
//
//		srv := &http.Server{
//			ErrorLog:  log.New(duh.NewHttpLogAdaptor(slog.Default(), "", 0),
//			Addr:      "localhost:8080",
//	     .....
//		}
func NewHttpLogAdaptor(log StandardLogger) *HttpLogAdaptor {
	reader, writer := io.Pipe()
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			log.Info(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Error("while reading from Writer",
				"err", err, "category", "HttpLogAdaptor")
		}
		_ = reader.Close()
		wg.Done()
	}()

	return &HttpLogAdaptor{
		writer: writer,
		closer: func() {
			_ = writer.Close()
			wg.Wait()
		},
	}
}
