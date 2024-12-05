package slog

import (
	"context"
	"fmt"

	"log/slog"

	"go.opentelemetry.io/otel/codes"
	otellog "go.opentelemetry.io/otel/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// The TracingHandler type is a wrapper around a slog.Handler.
// @property handler - The `handler` property is a variable of type `slog.Handler`.
type TracingHandler struct {
	handler slog.Handler
}

const sevOffset = slog.Level(otellog.SeverityDebug) - slog.LevelDebug

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

		r = alignWithOTELSpecs(r, span)
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

// https://opentelemetry.io/docs/specs/otel/logs/data-model/#log-and-event-record-definition
// Timestamp	Time when the event occurred.
// ObservedTimestamp	Time when the event was observed.
// TraceId	Request trace id.
// SpanId	Request span id.
// TraceFlags	W3C trace flag.
// SeverityText	The severity text (also known as log level).
// SeverityNumber	Numerical value of the severity.
// Body	The body of the log record.
// Resource	Describes the source of the log.
// InstrumentationScope	Describes the scope that emitted the log.
// Attributes	Additional information about the event.
func alignWithOTELSpecs(r slog.Record, span trace.Span) slog.Record {
	traceId := ""
	spanId := ""
	traceFlags := ""
	if spanCtx := span.SpanContext(); spanCtx.HasTraceID() {
		spanId = spanCtx.SpanID().String()
		traceId = spanCtx.TraceID().String()
		traceFlags = spanCtx.TraceFlags().String()
	}
	r.AddAttrs(slog.String("TraceId", traceId))
	r.AddAttrs(slog.String("SpanId", spanId))
	r.AddAttrs(slog.String("TraceFlags", traceFlags))

	// Convert the span to ReadOnlySpan to access attributes
	roSpan, ok := span.(sdktrace.ReadOnlySpan)
	if !ok {
		fmt.Println("Span is not a ReadOnlySpan")
	} else {
		// Create a group for span attributes
		attributes := make([]any, 0) // Use []any for slog.Group compatibility
		for _, attr := range roSpan.Attributes() {
			attributes = append(attributes, slog.String(string(attr.Key), attr.Value.Emit()))
		}
		r.AddAttrs(slog.Group("Attributes", attributes...))

		// Add InstrumentationScope details as a group
		scope := roSpan.InstrumentationScope()
		scopeAttrs := make([]any, 0) // Use []any for slog.Group compatibility
		iter := scope.Attributes.Iter()
		for iter.Next() {
			attr := iter.Attribute()
			scopeAttrs = append(scopeAttrs, slog.String(string(attr.Key), attr.Value.Emit()))
		}
		r.AddAttrs(slog.Group("InstrumentationScope",
			slog.String("Name", scope.Name),
			slog.String("Version", scope.Version),
			slog.Group("Attributes", scopeAttrs...),
		))

		// Add other span details
		r.AddAttrs(
			slog.String("SpanName", roSpan.Name()),
			slog.String("SpanKind", roSpan.SpanKind().String()),
			slog.String("Resource", roSpan.Resource().String()),
			slog.String("Timestamp", roSpan.StartTime().String()),
			slog.String("ObservedTimestamp", roSpan.StartTime().String()),
		)
	}

	sev := slog.Level(int(r.Level)) + sevOffset

	// Add severity and message details
	r.AddAttrs(
		slog.String("SeverityText", r.Level.String()),
		slog.Int("SeverityNumber", int(sev)),
		slog.String("Body", r.Message),
	)

	return r
}
