package logger

import (
	"calendar/internal/config"
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newBenchLogger(bufSize int) *Service {
	return NewService(&config.App{
		Logger: config.Logger{Mode: "dev", Level: "debug", BuffSize: bufSize},
	})
}

// BenchmarkLog_Unbuffered measures the cost of enqueuing a single log message.
func BenchmarkLog(b *testing.B) {
	svc := newBenchLogger(b.N + 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Log(zapcore.InfoLevel, "benchmark message", zap.Int("iter", i))
	}
}

// BenchmarkLog_Parallel measures contention on the log channel under parallel writes.
func BenchmarkLog_Parallel(b *testing.B) {
	svc := newBenchLogger(1_000_000)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			svc.Log(zapcore.InfoLevel, "parallel bench", zap.Int("i", i))
			i++
		}
	})
}

// BenchmarkLog_WithRunningConsumer benchmarks logging while the consumer goroutine is active.
func BenchmarkLog_WithRunningConsumer(b *testing.B) {
	svc := newBenchLogger(4096)
	svc.Start(context.Background())
	defer func() {
		_ = svc.Stop(context.Background())
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Log(zapcore.InfoLevel, "bench with consumer", zap.Int("iter", i))
	}
}

// BenchmarkLog_ChannelFull measures the fast-path when the channel is full (drop).
func BenchmarkLog_ChannelFull(b *testing.B) {
	svc := newBenchLogger(1) // tiny buffer — always full after 1 message

	// fill the channel
	svc.Log(zapcore.InfoLevel, "filler")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Log(zapcore.InfoLevel, "dropped message")
	}
}
