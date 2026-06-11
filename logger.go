package squad

import "log"

// Logger is the interface for SDK-level request/response logging.
// Implement this to integrate with your application's logging library.
//
// Compatible with log/slog, go.uber.org/zap, sirupsen/logrus, rs/zerolog, and others.
// Fields are passed as key-value pairs: key string, value any, key string, value any, ...
//
//	// log/slog example:
//	type slogAdapter struct{ l *slog.Logger }
//	func (a slogAdapter) Info(msg string, kv ...any)  { a.l.Info(msg, kv...) }
//	func (a slogAdapter) Error(msg string, kv ...any) { a.l.Error(msg, kv...) }
//	client := squad.New(key, squad.WithLogger(slogAdapter{slog.Default()}))
type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
}

// noopLogger discards all output. This is the default logger.
type noopLogger struct{}

func (noopLogger) Info(_ string, _ ...any)  {}
func (noopLogger) Error(_ string, _ ...any) {}

// stdLogger wraps the standard library's log package.
type stdLogger struct{ l *log.Logger }

func (s stdLogger) Info(msg string, fields ...any) {
	s.l.Printf("INFO squad: %s %v", msg, fields)
}

func (s stdLogger) Error(msg string, fields ...any) {
	s.l.Printf("ERROR squad: %s %v", msg, fields)
}

// StdLogger returns a Logger that writes to stderr using the standard log package.
// Suitable for development. For production, use your application's structured logger.
func StdLogger() Logger {
	return stdLogger{l: log.Default()}
}
