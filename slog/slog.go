package slog

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

const TraceIDKey = "trace_id"
const SpanIDKey = "span_id"

type TracingHandler struct {
	handler slog.Handler
}

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

func (h *TracingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

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

func (h *TracingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewTracingHandler(h.handler.WithAttrs(attrs))
}

func (h *TracingHandler) WithGroup(name string) slog.Handler {
	return NewTracingHandler(h.handler.WithGroup(name))
}
