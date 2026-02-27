package config

import (
	"testing"
	"time"
)

func TestApp_Struct(t *testing.T) {
	app := App{
		Server: Server{
			Host: "localhost",
			Port: 8080,
		},
		Logger: Logger{
			Mode:     "dev",
			Level:    "debug",
			BuffSize: 1000,
		},
		Gin: Gin{
			Mode: "release",
		},
	}

	if app.Server.Host != "localhost" {
		t.Error("Server.Host mismatch")
	}
	if app.Server.Port != 8080 {
		t.Error("Server.Port mismatch")
	}
	if app.Logger.Mode != "dev" {
		t.Error("Logger.Mode mismatch")
	}
}

func TestPostgres_Struct(t *testing.T) {
	pg := Postgres{
		Host:            "localhost",
		Port:            5432,
		User:            "user",
		Password:        "pass",
		DBName:          "testdb",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 5,
		ReadTimeout:     time.Second * 3,
		WriteTimeout:    time.Second * 5,
		LongTimeout:     time.Second * 30,
	}

	if pg.Host != "localhost" {
		t.Error("Host mismatch")
	}
	if pg.Port != 5432 {
		t.Error("Port mismatch")
	}
	if pg.SSLMode != "disable" {
		t.Error("SSLMode mismatch")
	}
	if pg.ConnMaxLifetime != time.Minute*5 {
		t.Error("ConnMaxLifetime mismatch")
	}
}

func TestJwt_Struct(t *testing.T) {
	jwt := Jwt{
		ExpAccessToken:  15,
		ExpRefreshToken: 24,
		AccessSecret:    "access-secret",
		RefreshSecret:   "refresh-secret",
	}

	if jwt.ExpAccessToken != 15 {
		t.Error("ExpAccessToken mismatch")
	}
	if jwt.ExpRefreshToken != 24 {
		t.Error("ExpRefreshToken mismatch")
	}
}

func TestUserValidation_Struct(t *testing.T) {
	uv := UserValidation{
		MinLength:         3,
		MaxLength:         20,
		AllowedCharacters: "A-Za-z0-9_-",
		CaseInsensitive:   true,
	}

	if uv.MinLength != 3 {
		t.Error("MinLength mismatch")
	}
	if uv.MaxLength != 20 {
		t.Error("MaxLength mismatch")
	}
	if !uv.CaseInsensitive {
		t.Error("CaseInsensitive should be true")
	}
}

func TestPasswordValidation_Struct(t *testing.T) {
	pv := PasswordValidation{
		MinLength:    8,
		MaxLength:    64,
		RequireUpper: true,
		RequireLower: true,
		RequireDigit: true,
	}

	if pv.MinLength != 8 {
		t.Error("MinLength mismatch")
	}
	if !pv.RequireUpper {
		t.Error("RequireUpper should be true")
	}
}

func TestEventValidation_Struct(t *testing.T) {
	ev := EventValidation{
		NameMinLength:        3,
		NameMaxLength:        40,
		DescriptionMaxLength: 200,
		DescriptionRequire:   true,
	}

	if ev.NameMinLength != 3 {
		t.Error("NameMinLength mismatch")
	}
	if !ev.DescriptionRequire {
		t.Error("DescriptionRequire should be true")
	}
}

func TestCleaner_Struct(t *testing.T) {
	cleaner := Cleaner{
		CheckInterval: time.Minute * 30,
		EventLifetime: time.Hour * 24,
	}

	if cleaner.CheckInterval != time.Minute*30 {
		t.Error("CheckInterval mismatch")
	}
	if cleaner.EventLifetime != time.Hour*24 {
		t.Error("EventLifetime mismatch")
	}
}

func TestDB_Struct(t *testing.T) {
	db := DB{
		Postgres: Postgres{
			Host: "localhost",
			Port: 5432,
		},
	}

	if db.Postgres.Host != "localhost" {
		t.Error("Postgres.Host mismatch")
	}
}

func TestTelegram_Struct(t *testing.T) {
	tg := Telegram{
		BotToken: "test-token",
	}

	if tg.BotToken != "test-token" {
		t.Error("BotToken mismatch")
	}
}

func TestMail_Struct(t *testing.T) {
	mail := Mail{
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPEmail:    "test@gmail.com",
		SMTPPassword: "password",
	}

	if mail.SMTPHost != "smtp.gmail.com" {
		t.Error("SMTPHost mismatch")
	}
	if mail.SMTPPort != 587 {
		t.Error("SMTPPort mismatch")
	}
}
