package config

import (
	"time"
)

type App struct {
	Server             Server             `mapstructure:"server"`
	Logger             Logger             `mapstructure:"logger"`
	DB                 DB                 `mapstructure:"db"`
	Telegram           Telegram           `mapstructure:"telegram"`
	Mail               Mail               `mapstructure:"mail"`
	Gin                Gin                `mapstructure:"gin"`
	Jwt                Jwt                `mapstructure:"jwt"`
	UserValidation     UserValidation     `mapstructure:"username_validation"`
	PasswordValidation PasswordValidation `mapstructure:"password_validation"`
	EventValidation    EventValidation    `mapstructure:"event_validation"`
	Cleaner            Cleaner            `mapstructure:"cleaner"`
}

type Gin struct {
	Mode string `mapstructure:"mode"`
}

type Server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Logger struct {
	Mode     string `mapstructure:"mode"`
	Level    string `mapstructure:"level"`
	BuffSize int    `mapstructure:"buffer_size"`
}

type Postgres struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	LongTimeout     time.Duration `mapstructure:"long_timeout"`
}

type DB struct {
	Postgres Postgres `mapstructure:"postgres"`
}

type Telegram struct {
	BotToken string `mapstructure:"bot_token"`
}

type Mail struct {
	SMTPHost     string `mapstructure:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port"`
	SMTPEmail    string `mapstructure:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password"`
}

type Jwt struct {
	ExpAccessToken  int    `mapstructure:"exp_access_token"`
	ExpRefreshToken int    `mapstructure:"exp_refresh_token"`
	AccessSecret    string `mapstructure:"access_secret"`
	RefreshSecret   string `mapstructure:"refresh_secret"`
}

type UserValidation struct {
	MinLength         int    `mapstructure:"min_length"`
	MaxLength         int    `mapstructure:"max_length"`
	AllowedCharacters string `mapstructure:"allowed_characters"`
	CaseInsensitive   bool   `mapstructure:"case_insensitive"`
}

type PasswordValidation struct {
	MinLength    int  `mapstructure:"min_length"`
	MaxLength    int  `mapstructure:"max_length"`
	RequireUpper bool `mapstructure:"require_upper"`
	RequireLower bool `mapstructure:"require_lower"`
	RequireDigit bool `mapstructure:"require_digit"`
}

type EventValidation struct {
	NameMinLength        int  `mapstructure:"name_min_length"`
	NameMaxLength        int  `mapstructure:"name_max_length"`
	DescriptionMaxLength int  `mapstructure:"description_max_length"`
	DescriptionRequire   bool `mapstructure:"description_required"`
}

type Cleaner struct {
	CheckInterval time.Duration `mapstructure:"check_interval"`
	EventLifetime time.Duration `mapstructure:"event_lifetime"`
}
