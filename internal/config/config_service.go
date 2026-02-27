package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func NewAppConfig() (*App, error) {
	_ = godotenv.Load(".env") // 1️⃣ сначала env

	cfg := viper.New()

	//cfg.SetEnvPrefix("CALENDAR")
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg.AutomaticEnv() // 2️⃣ потом automatic env

	cfg.SetConfigName("local")
	cfg.AddConfigPath("./config")
	cfg.SetConfigType("yaml")

	if err := cfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := envProvider(cfg); err != nil {
		return nil, fmt.Errorf("env provider: %w", err)
	}

	var appCfg App
	if err := cfg.Unmarshal(&appCfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &appCfg, nil
}

func envProvider(cfg *viper.Viper) error {

	if err := cfg.BindEnv("db.postgres.user"); err != nil {
		return err
	}
	if err := cfg.BindEnv("db.postgres.password"); err != nil {
		return err
	}
	if err := cfg.BindEnv("db.postgres.db_name"); err != nil {
		return err
	}
	if err := cfg.BindEnv("telegram.bot_token"); err != nil {
		return err
	}
	if err := cfg.BindEnv("mail.smtp_user"); err != nil {
		return err
	}
	if err := cfg.BindEnv("mail.smtp_password"); err != nil {
		return err
	}
	if err := cfg.BindEnv("jwt.access_secret"); err != nil {
		return err
	}
	if err := cfg.BindEnv("jwt.refresh_secret"); err != nil {
		return err
	}
	return nil
}
