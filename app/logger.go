package app

// import (
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/go-chi/chi/middleware"
// 	"github.com/sirupsen/logrus"
// )

// // NewStructuredLogger may be exported
// func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
// 	return middleware.RequestLogger(&StructuredLogger{logger})
// }

// // StructuredLogger exported as we may move this to its own package
// type StructuredLogger struct {
// 	Logger *logrus.Logger
// }

// // NewLogEntry exported as we may move this to its own package
// func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
// 	entry := &StructuredLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
// 	logFields := logrus.Fields{}

// 	logFields["ts"] = time.Now().UTC().Format(time.RFC1123)

// 	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
// 		logFields["req_id"] = reqID
// 	}

// 	scheme := "http"
// 	if r.TLS != nil {
// 		scheme = "https"
// 	}
// 	logFields["http_scheme"] = scheme
// 	logFields["http_proto"] = r.Proto
// 	logFields["http_method"] = r.Method

// 	logFields["remote_addr"] = r.RemoteAddr
// 	logFields["user_agent"] = r.UserAgent()

// 	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

// 	entry.Logger = entry.Logger.WithFields(logFields)

// 	entry.Logger.Infoln("request started")

// 	return entry
// }

// // StructuredLoggerEntry exported as we may move this to its own package
// type StructuredLoggerEntry struct {
// 	Logger logrus.FieldLogger
// }

// // StructuredLoggerEntry exported as we may move this to its own package
// func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
// 	l.Logger = l.Logger.WithFields(logrus.Fields{
// 		"resp_status": status, "resp_bytes_length": bytes,
// 		"resp_elapsed_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
// 	})

// 	l.Logger.Infoln("request complete")
// }

// // Panic exported as we may move this to its own package
// func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
// 	l.Logger = l.Logger.WithFields(logrus.Fields{
// 		"stack": string(stack),
// 		"panic": fmt.Sprintf("%+v", v),
// 	})
// }

// // Helper methods used by the application to get the request-scoped
// // logger entry and set additional fields between handlers.
// //
// // This is a useful pattern to use to set state on the entry as it
// // passes through the handler chain, which at any point can be logged
// // with a call to .Print(), .Info(), etc.

// // GetLogEntry exported as we may move this to its own package
// func GetLogEntry(r *http.Request) logrus.FieldLogger {
// 	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
// 	return entry.Logger
// }

// // LogEntrySetField exported as we may move this to its own package
// func LogEntrySetField(r *http.Request, key string, value interface{}) {
// 	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
// 		entry.Logger = entry.Logger.WithField(key, value)
// 	}
// }

// // LogEntrySetFields exported as we may move this to its own package
// func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
// 	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*StructuredLoggerEntry); ok {
// 		entry.Logger = entry.Logger.WithFields(fields)
// 	}
// }
