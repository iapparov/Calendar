package user

import "testing"

func BenchmarkNewUser(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewUser("benchuser", "SecurePass1", "bench@example.com", "123456")
	}
}

func BenchmarkNewUser_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = NewUser("benchuser", "SecurePass1", "bench@example.com", "123456")
		}
	})
}
