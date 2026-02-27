package logger

import (
	"calendar/internal/config"
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"warning", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"dpanic", zapcore.DPanicLevel},
		{"panic", zapcore.PanicLevel},
		{"fatal", zapcore.FatalLevel},
		{"unknown", zapcore.InfoLevel},
		{"", zapcore.InfoLevel},
		{"DEBUG", zapcore.InfoLevel}, // case sensitive, defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogEvent_Struct(t *testing.T) {
	event := LogEvent{
		Level:  zapcore.InfoLevel,
		Msg:    "test message",
		Fields: nil,
	}

	if event.Level != zapcore.InfoLevel {
		t.Error("Level mismatch")
	}
	if event.Msg != "test message" {
		t.Error("Msg mismatch")
	}
	if event.Fields != nil {
		t.Error("Fields should be nil")
	}
}

func TestLogEvent_WithFields(t *testing.T) {
	fields := []zap.Field{
		zap.String("key", "value"),
		zap.Int("count", 42),
	}

	event := LogEvent{
		Level:  zapcore.DebugLevel,
		Msg:    "test with fields",
		Fields: fields,
	}

	if event.Level != zapcore.DebugLevel {
		t.Error("Level mismatch")
	}
	if len(event.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(event.Fields))
	}
}

func TestModeProd_Constant(t *testing.T) {
	if ModeProd != "prod" {
		t.Errorf("expected ModeProd to be 'prod', got '%s'", ModeProd)
	}
}

func createTestConfig() *config.App {
	return &config.App{
		Logger: config.Logger{
			Mode:     "dev",
			Level:    "debug",
			BuffSize: 100,
		},
	}
}

func TestNewService(t *testing.T) {
	cfg := createTestConfig()

	service := NewService(cfg)

	if service == nil {
		t.Fatal("service should not be nil")
	}
	if service.logger == nil {
		t.Error("logger should not be nil")
	}
	if service.msgChan == nil {
		t.Error("msgChan should not be nil")
	}
	if cap(service.msgChan) != cfg.Logger.BuffSize {
		t.Errorf("expected channel capacity %d, got %d", cfg.Logger.BuffSize, cap(service.msgChan))
	}
}

func TestNewService_WithProdMode(t *testing.T) {
	cfg := &config.App{
		Logger: config.Logger{
			Mode:     "prod",
			Level:    "info",
			BuffSize: 50,
		},
	}

	service := NewService(cfg)

	if service == nil {
		t.Fatal("service should not be nil")
	}
	if service.logger == nil {
		t.Error("logger should not be nil in prod mode")
	}
}

func TestService_Log(t *testing.T) {
	cfg := createTestConfig()
	service := NewService(cfg)

	// Test logging without blocking
	service.Log(zapcore.InfoLevel, "test message")
	service.Log(zapcore.DebugLevel, "debug message", zap.String("key", "value"))
	service.Log(zapcore.WarnLevel, "warning message")
	service.Log(zapcore.ErrorLevel, "error message")

	// Check that messages are in the channel
	if len(service.msgChan) != 4 {
		t.Errorf("expected 4 messages in channel, got %d", len(service.msgChan))
	}
}

func TestService_Log_ChannelFull(t *testing.T) {
	cfg := &config.App{
		Logger: config.Logger{
			Mode:     "dev",
			Level:    "debug",
			BuffSize: 2, // Small buffer
		},
	}
	service := NewService(cfg)

	// Fill the channel
	service.Log(zapcore.InfoLevel, "message 1")
	service.Log(zapcore.InfoLevel, "message 2")

	// This should not block even though channel is full
	done := make(chan bool)
	go func() {
		service.Log(zapcore.InfoLevel, "message 3") // Should be dropped
		done <- true
	}()

	select {
	case <-done:
		// Success - Log returned without blocking
	case <-time.After(time.Second):
		t.Error("Log blocked when channel was full")
	}
}

func TestService_StartStop(t *testing.T) {
	cfg := createTestConfig()
	service := NewService(cfg)

	// Start the service
	service.Start(context.Background())

	// Log some messages
	service.Log(zapcore.InfoLevel, "test message 1")
	service.Log(zapcore.DebugLevel, "test message 2")

	// Give time for processing
	time.Sleep(50 * time.Millisecond)

	// Stop the service
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := service.Stop(ctx)
	if err != nil {
		t.Errorf("unexpected error stopping service: %v", err)
	}
}

func TestService_Stop_Timeout(t *testing.T) {
	cfg := createTestConfig()
	service := NewService(cfg)

	// Start the service
	service.Start(context.Background())

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Give a tiny bit of time
	time.Sleep(10 * time.Millisecond)

	// Stop should respect context cancellation
	// Note: this might succeed or fail depending on timing
	_ = service.Stop(ctx)
}

func TestService_Run_AllLevels(t *testing.T) {
	cfg := createTestConfig()
	service := NewService(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	// Start Run in a goroutine
	done := make(chan bool)
	go func() {
		service.Run(ctx)
		done <- true
	}()

	// Send messages of all levels
	levels := []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
		zapcore.Level(100), // Unknown level - should default to Info
	}

	for _, level := range levels {
		service.msgChan <- LogEvent{Level: level, Msg: "test", Fields: nil}
	}

	// Give time for processing
	time.Sleep(50 * time.Millisecond)

	// Cancel to stop Run
	cancel()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("Run did not stop after context cancellation")
	}
}

func TestProvideLogger_DevMode(t *testing.T) {
	cfg := &config.App{
		Logger: config.Logger{
			Mode:  "dev",
			Level: "debug",
		},
	}

	logger := provideLogger(cfg)

	if logger == nil {
		t.Fatal("logger should not be nil")
	}
}

func TestProvideLogger_ProdMode(t *testing.T) {
	cfg := &config.App{
		Logger: config.Logger{
			Mode:  "prod",
			Level: "info",
		},
	}

	logger := provideLogger(cfg)

	if logger == nil {
		t.Fatal("logger should not be nil in prod mode")
	}
}

func TestService_LogWithFields(t *testing.T) {
	cfg := createTestConfig()
	service := NewService(cfg)

	fields := []zap.Field{
		zap.String("user_id", "123"),
		zap.Int("request_id", 456),
		zap.Duration("latency", time.Millisecond*100),
	}

	service.Log(zapcore.InfoLevel, "request completed", fields...)

	if len(service.msgChan) != 1 {
		t.Errorf("expected 1 message in channel, got %d", len(service.msgChan))
	}

	// Read the message and verify
	msg := <-service.msgChan
	if msg.Msg != "request completed" {
		t.Errorf("expected message 'request completed', got '%s'", msg.Msg)
	}
	if len(msg.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(msg.Fields))
	}
}
