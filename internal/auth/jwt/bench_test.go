package jwt

import (
	"testing"

	"calendar/internal/domain/user"

	"github.com/google/uuid"
)

var benchService = &Service{
	accessSecretBytes:  []byte("bench-access-secret-key-32bytes!"),
	refreshSecretBytes: []byte("bench-refresh-secret-key-32byte"),
	expAccessToken:     15,
	expRefreshToken:    1440,
}

func BenchmarkGenerateTokens(b *testing.B) {
	u := &user.User{ID: uuid.New(), Login: "benchuser"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = benchService.GenerateTokens(u)
	}
}

func BenchmarkGenerateTokens_Parallel(b *testing.B) {
	u := &user.User{ID: uuid.New(), Login: "benchuser"}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = benchService.GenerateTokens(u)
		}
	})
}

func BenchmarkValidateTokens(b *testing.B) {
	u := &user.User{ID: uuid.New(), Login: "benchuser"}
	tokens, _ := benchService.GenerateTokens(u)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = benchService.ValidateTokens(tokens.AccessToken)
	}
}

func BenchmarkValidateTokens_Parallel(b *testing.B) {
	u := &user.User{ID: uuid.New(), Login: "benchuser"}
	tokens, _ := benchService.GenerateTokens(u)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = benchService.ValidateTokens(tokens.AccessToken)
		}
	})
}

func BenchmarkRefreshTokens(b *testing.B) {
	u := &user.User{ID: uuid.New(), Login: "benchuser"}
	tokens, _ := benchService.GenerateTokens(u)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = benchService.RefreshTokens(tokens.RefreshToken)
	}
}

func BenchmarkValidateTokens_Invalid(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = benchService.ValidateTokens("invalid.token.string")
	}
}
