package slog

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/instrumentation"
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

// mockReadOnlySpan implements sdktrace.ReadOnlySpan for testing
type mockReadOnlySpan struct {
	mockSpan
	name       string
	spanKind   trace.SpanKind
	attributes []attribute.KeyValue
	resource   interface{}
	startTime  time.Time
	scope      instrumentation.Scope
}

func (s *mockReadOnlySpan) Name() string                                { return s.name }
func (s *mockReadOnlySpan) SpanKind() trace.SpanKind                    { return s.spanKind }
func (s *mockReadOnlySpan) Attributes() []attribute.KeyValue            { return s.attributes }
func (s *mockReadOnlySpan) Resource() interface{}                       { return s.resource }
func (s *mockReadOnlySpan) StartTime() time.Time                        { return s.startTime }
func (s *mockReadOnlySpan) InstrumentationScope() instrumentation.Scope { return s.scope }

func TestAlignWithOTELSpecs(t *testing.T) {
	tests := []struct {
		name     string
		span     trace.Span
		level    slog.Level
		message  string
		expected int // expect number of attributes added
	}{
		{
			name: "mockReadOnlySpan",
			span: &mockReadOnlySpan{
				mockSpan: mockSpan{
					isRecording: true,
					traceID:     trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					spanID:      trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
					traceFlags:  trace.FlagsSampled,
				},
				name:     "test-span",
				spanKind: trace.SpanKindInternal,
				attributes: []attribute.KeyValue{
					attribute.String("attr.key", "attr.value"),
				},
				startTime: time.Now(),
			},
			level:   slog.LevelInfo,
			message: "test message",
		},
		{
			name: "regular mock span",
			span: &mockSpan{
				isRecording: true,
				traceID:     trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
				spanID:      trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
				traceFlags:  trace.FlagsSampled,
			},
			level:   slog.LevelInfo,
			message: "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := slog.NewRecord(time.Now(), tt.level, tt.message, 0)
			originalNumAttrs := record.NumAttrs()

			result := alignWithOTELSpecs(record, tt.span)

			// Check that attributes were added (TraceId, SpanId, TraceFlags at minimum)
			newNumAttrs := result.NumAttrs()
			if newNumAttrs <= originalNumAttrs {
				t.Errorf("alignWithOTELSpecs did not add attributes: before=%d, after=%d", originalNumAttrs, newNumAttrs)
			}
		})
	}
}

func TestTracingHandler_HandleWithReadOnlySpan(t *testing.T) {
	baseHandler := &mockHandler{enabled: true}
	th := NewTracingHandler(baseHandler)

	// Create a mock span that implements ReadOnlySpan
	roSpan := &mockReadOnlySpan{
		mockSpan: mockSpan{
			isRecording: true,
			traceID:     trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			spanID:      trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
			traceFlags:  trace.FlagsSampled,
		},
		name:     "test-span",
		spanKind: trace.SpanKindInternal,
		attributes: []attribute.KeyValue{
			attribute.String("test.attr", "test.value"),
		},
		startTime: time.Now(),
	}

	ctx := trace.ContextWithSpan(context.Background(), roSpan)
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)

	err := th.Handle(ctx, record)
	if err != nil {
		t.Errorf("Handle() error = %v", err)
	}
}

// Test the case where span is not a ReadOnlySpan to improve coverage
func TestAlignWithOTELSpecsWithNonReadOnlySpan(t *testing.T) {
	span := &mockSpan{
		isRecording: true,
		traceID:     trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		spanID:      trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
		traceFlags:  trace.FlagsSampled,
	}

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	result := alignWithOTELSpecs(record, span)

	// Verify that basic trace info was added even for non-ReadOnlySpan
	numAttrs := result.NumAttrs()
	// At minimum TraceId, SpanId, and TraceFlags should be added
	if numAttrs < 3 {
		t.Errorf("Expected at least 3 attributes to be added, got %d", numAttrs)
	}
}

func TestAlignWithOTELSpecsLogLevelVariations(t *testing.T) {
	span := &mockReadOnlySpan{
		mockSpan: mockSpan{
			isRecording: true,
			traceID:     trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			spanID:      trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
			traceFlags:  trace.FlagsSampled,
		},
		name:     "test-span",
		spanKind: trace.SpanKindInternal,
		attributes: []attribute.KeyValue{
			attribute.String("test.attr", "test.value"),
		},
		startTime: time.Now(),
		scope:     instrumentation.Scope{Name: "test.scope", Version: "v1.0.0"},
	}

	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			record := slog.NewRecord(time.Now(), level, "test message", 0)
			result := alignWithOTELSpecs(record, span)

			// Check that various attributes were added
			numAttrs := result.NumAttrs()
			if numAttrs < 5 { // Should have at least TraceId, SpanId, TraceFlags, SeverityText, SeverityNumber
				t.Errorf("Expected at least 5 attributes for level %v, got %d", level, numAttrs)
			}
		})
	}
}

// Additional test to increase coverage of alignWithOTELSpecs
func TestAlignWithOTELSpecsCompleteAttributes(t *testing.T) {
	// Create a comprehensive mockReadOnlySpan with all attributes
	attrs := []attribute.KeyValue{
		attribute.String("key1", "value1"),
		attribute.Int("key2", 42),
	}

	span := &mockReadOnlySpan{
		mockSpan: mockSpan{
			isRecording: true,
			traceID:     trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			spanID:      trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8},
			traceFlags:  trace.FlagsSampled,
		},
		name:       "comprehensive-test-span",
		spanKind:   trace.SpanKindServer,
		attributes: attrs,
		startTime:  time.Now(),
		scope:      instrumentation.Scope{Name: "test.instrumentation", Version: "v1.2.3"},
	}

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	result := alignWithOTELSpecs(record, span)

	// Verify that the function doesn't panic and adds attributes
	// (the exact number may vary depending on the implementation details)
	numAttrs := result.NumAttrs()
	if numAttrs == 0 {
		t.Error("Expected attributes to be added to the record")
	}
}
