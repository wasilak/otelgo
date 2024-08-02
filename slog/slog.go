package slog

import (
	"context"

	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// The `const TraceIDKey = "TraceId"` line is declaring a constant variable named `TraceIDKey` with
// the value `"TraceId"`. This constant is used as a key to add an attribute to a log record.
const TraceIDKey = "TraceId"

// The line `const SpanIDKey = "SpanId"` is declaring a constant variable named `SpanIDKey` with the
// value `"SpanId"`. This constant is used as a key to add an attribute to a log record.
const SpanIDKey = "SpanId"

// The TracingHandler type is a wrapper around a slog.Handler.
// @property handler - The `handler` property is a variable of type `slog.Handler`.
type TracingHandler struct {
	handler slog.Handler
}

// The function NewTracingHandler creates a new TracingHandler by wrapping an existing slog.Handler.
func NewTracingHandler(h slog.Handler) *TracingHandler {
	// avoid chains of handlers.
	if lh, ok := h.(*TracingHandler); ok {
		h = lh.Handler()
	}
	return &TracingHandler{h}
}

// Handler returns the Handler wrapped by h.
func (h *TracingHandler) Handler() slog.Handler {
	return h.handler
}

// The `Enabled` method is a function defined on the `TracingHandler` struct. It takes two parameters:
// `ctx` of type `context.Context` and `level` of type `slog.Level`.
func (h *TracingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// The `Handle` method is a function defined on the `TracingHandler` struct. It takes two parameters:
// `ctx` of type `context.Context` and `r` of type `slog.Record`.
func (h *TracingHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)

	if span.IsRecording() {
		if r.Level >= slog.LevelError {
			span.SetStatus(codes.Error, r.Message)
		}

		if spanCtx := span.SpanContext(); spanCtx.HasTraceID() {
			r.AddAttrs(slog.String(TraceIDKey, spanCtx.TraceID().String()))
			r.AddAttrs(slog.String(SpanIDKey, string(spanCtx.SpanID().String())))
		}
	}

	return h.handler.Handle(ctx, r)
}

// The `func (h *TracingHandler) WithAttrs(attrs []slog.Attr) slog.Handler` method is a function
// defined on the `TracingHandler` struct. It takes a parameter `attrs` of type `[]slog.Attr`, which
// represents a list of log attributes.
func (h *TracingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewTracingHandler(h.handler.WithAttrs(attrs))
}

// The `func (h *TracingHandler) WithGroup(name string) slog.Handler {` method is defining a function
// on the `TracingHandler` struct. This function takes a parameter `name` of type `string`, which
// represents the name of the log group.
func (h *TracingHandler) WithGroup(name string) slog.Handler {
	return NewTracingHandler(h.handler.WithGroup(name))
}
