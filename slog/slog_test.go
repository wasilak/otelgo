package slog

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// mockHandler implements slog.Handler for testing
type mockHandler struct {
	enabled    bool
	lastRecord slog.Record
	lastAttrs  []slog.Attr
	lastGroup  string
}

func (h *mockHandler) Enabled(context.Context, slog.Level) bool { return h.enabled }
func (h *mockHandler) Handle(_ context.Context, r slog.Record) error {
	h.lastRecord = r
	return nil
}
func (h *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.lastAttrs = attrs
	return h
}
func (h *mockHandler) WithGroup(name string) slog.Handler {
	h.lastGroup = name
	return h
}

// mockSpan implements trace.Span for testing
type mockSpan struct {
	trace.Span
	isRecording bool
	statusCode  codes.Code
	statusMsg   string
	traceID     trace.TraceID
	spanID      trace.SpanID
	traceFlags  trace.TraceFlags
}

func (s *mockSpan) IsRecording() bool { return s.isRecording }
func (s *mockSpan) SetStatus(code codes.Code, msg string) {
	s.statusCode = code
	s.statusMsg = msg
}
func (s *mockSpan) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    s.traceID,
		SpanID:     s.spanID,
		TraceFlags: s.traceFlags,
	})
}

func TestNewTracingHandler(t *testing.T) {
	baseHandler := &mockHandler{enabled: true}

	// Test creating new handler
	th := NewTracingHandler(baseHandler)
	if th.handler != baseHandler {
		t.Error("NewTracingHandler did not set the base handler correctly")
	}

	// Test wrapping an existing TracingHandler
	th2 := NewTracingHandler(th)
	if th2.handler != baseHandler {
		t.Error("NewTracingHandler did not unwrap existing TracingHandler")
	}
}

func TestTracingHandler_Handler(t *testing.T) {
	baseHandler := &mockHandler{enabled: true}
	th := NewTracingHandler(baseHandler)

	if got := th.Handler(); got != baseHandler {
		t.Error("Handler() did not return the correct base handler")
	}
}

func TestTracingHandler_Enabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		level    slog.Level
		expected bool
	}{
		{"enabled-info", true, slog.LevelInfo, true},
		{"disabled-info", false, slog.LevelInfo, false},
		{"enabled-error", true, slog.LevelError, true},
		{"disabled-error", false, slog.LevelError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseHandler := &mockHandler{enabled: tt.enabled}
			th := NewTracingHandler(baseHandler)

			if got := th.Enabled(context.Background(), tt.level); got != tt.expected {
				t.Errorf("Enabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTracingHandler_Handle(t *testing.T) {
	tests := []struct {
		name       string
		level      slog.Level
		message    string
		span       *mockSpan
		wantStatus codes.Code
	}{
		{
			name:       "error-level-with-span",
			level:      slog.LevelError,
			message:    "test error",
			span:       &mockSpan{isRecording: true},
			wantStatus: codes.Error,
		},
		{
			name:    "info-level-with-span",
			level:   slog.LevelInfo,
			message: "test info",
			span:    &mockSpan{isRecording: true},
		},
		{
			name:    "no-recording-span",
			level:   slog.LevelError,
			message: "test error",
			span:    &mockSpan{isRecording: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseHandler := &mockHandler{enabled: true}
			th := NewTracingHandler(baseHandler)

			ctx := trace.ContextWithSpan(context.Background(), tt.span)
			record := slog.NewRecord(time.Now(), tt.level, tt.message, 0)

			err := th.Handle(ctx, record)
			if err != nil {
				t.Errorf("Handle() error = %v", err)
			}

			if tt.span.isRecording && tt.level >= slog.LevelError {
				if tt.span.statusCode != tt.wantStatus {
					t.Errorf("Handle() set status code = %v, want %v", tt.span.statusCode, tt.wantStatus)
				}
				if tt.span.statusMsg != tt.message {
					t.Errorf("Handle() set status message = %v, want %v", tt.span.statusMsg, tt.message)
				}
			}
		})
	}
}

func TestTracingHandler_WithAttrs(t *testing.T) {
	baseHandler := &mockHandler{enabled: true}
	th := NewTracingHandler(baseHandler)

	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	newHandler := th.WithAttrs(attrs)
	if _, ok := newHandler.(*TracingHandler); !ok {
		t.Error("WithAttrs() did not return a TracingHandler")
	}

	mockBase := baseHandler
	if !attributesEqual(mockBase.lastAttrs, attrs) {
		t.Error("WithAttrs() did not pass attributes to base handler")
	}
}

func TestTracingHandler_WithGroup(t *testing.T) {
	baseHandler := &mockHandler{enabled: true}
	th := NewTracingHandler(baseHandler)

	groupName := "test_group"
	newHandler := th.WithGroup(groupName)

	if _, ok := newHandler.(*TracingHandler); !ok {
		t.Error("WithGroup() did not return a TracingHandler")
	}

	mockBase := baseHandler
	if mockBase.lastGroup != groupName {
		t.Errorf("WithGroup() set group = %v, want %v", mockBase.lastGroup, groupName)
	}
}

// Helper function to compare slog.Attr slices
func attributesEqual(a, b []slog.Attr) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}
	return true
}
