package event

import (
	"testing"

	"github.com/google/uuid"
)

func BenchmarkNewEvent(b *testing.B) {
	userID := uuid.New().String()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewEvent("2026-06-15", userID, "Team Meeting", "Discuss", StatusActive, "2026-06-15T09:00")
	}
}

func BenchmarkNewEvent_Parallel(b *testing.B) {
	userID := uuid.New().String()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = NewEvent("2026-06-15", userID, "Team Meeting", "Discuss", StatusActive, "2026-06-15T09:00")
		}
	})
}

func BenchmarkEvent_Update(b *testing.B) {
	userID := uuid.New().String()
	ev, _ := NewEvent("2026-06-15", userID, "Original", "Original text", StatusActive, "2026-06-15T09:00")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ev.Update("2026-07-20", "Updated text", "Updated Name", "2026-07-20T14:00")
	}
}
