package config

import (
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
)

type (
	Config struct {
		Settings    Settings
		Credentials Credentials
	}

	Settings struct {
		App     AppSettings     `mapstructure:"app"`
		Server  ServerSettings  `mapstructure:"server"`
		DB      DBSettings      `mapstructure:"db"`
		Logger  LoggerSettings  `mapstructure:"logger"`
		CORS    CORSSettings    `mapstructure:"cors"`
		Service ServiceSettings `mapstructure:"service"`
	}

	Credentials struct {
		DB  DBCredentials
		JWT JWTCredentials
	}

	AppSettings struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
		Mode    string `mapstructure:"mode"`
		Port    int    `mapstructure:"port"`
	}

	ServerSettings struct {
		ReadTimeout     time.Duration `mapstructure:"read_timeout"`
		WriteTimeout    time.Duration `mapstructure:"write_timeout"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	}

	DBSettings struct {
		MaxOpenConns    int           `mapstructure:"max_open_conns"`
		MaxIdleConns    int           `mapstructure:"max_idle_conns"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	}

	LoggerSettings struct {
		Level      string `mapstructure:"level"`
		TimeFormat string `mapstructure:"time_format"`
		LogFile    string `mapstructure:"log_file"`
	}

	CORSSettings struct {
		AllowedOrigins   []string `mapstructure:"allowed_origins"`
		AllowedMethods   []string `mapstructure:"allowed_methods"`
		AllowedHeaders   []string `mapstructure:"allowed_headers"`
		AllowCredentials bool     `mapstructure:"allow_credentials"`
		MaxAge           int      `mapstructure:"max_age"`
	}

	ServiceSettings struct {
		InitialCoins int `mapstructure:"initial_coins"`
	}

	DBCredentials struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
	}

	JWTCredentials struct {
		SecretKey string
		ExpiresIn time.Duration
	}
)

func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	if err := cfg.loadSettings(configPath); err != nil {
		return nil, fmt.Errorf("error loading settings: %w", err)
	}

	if err := cfg.loadCredentials(); err != nil {
		return nil, fmt.Errorf("error loading credentials: %w", err)
	}

	return cfg, nil
}

func (c *Config) loadSettings(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := v.Unmarshal(&c.Settings); err != nil {
		return fmt.Errorf("error unmarshaling settings: %w", err)
	}

	return nil
}

func (c *Config) loadCredentials() error {
	c.Credentials.DB = DBCredentials{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  os.Getenv("POSTGRES_SSL_MODE"),
	}

	c.Credentials.JWT = JWTCredentials{
		SecretKey: os.Getenv("JWT_SECRET_KEY"),
		ExpiresIn: time.Hour * 24, // default value
	}

	if envExpiresIn := os.Getenv("JWT_EXPIRES_IN"); envExpiresIn != "" {
		duration, err := time.ParseDuration(envExpiresIn)
		if err != nil {
			return fmt.Errorf("invalid JWT_EXPIRES_IN format: %w", err)
		}
		c.Credentials.JWT.ExpiresIn = duration
	}

	return c.validateCredentials()
}

func (c *Config) validateCredentials() error {
	required := map[string]string{
		"POSTGRES_HOST":     c.Credentials.DB.Host,
		"POSTGRES_PORT":     c.Credentials.DB.Port,
		"POSTGRES_USER":     c.Credentials.DB.User,
		"POSTGRES_PASSWORD": c.Credentials.DB.Password,
		"POSTGRES_DB":       c.Credentials.DB.DBName,
		"JWT_SECRET_KEY":    c.Credentials.JWT.SecretKey,
	}

	for env, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}

	return nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Credentials.DB.Host,
		c.Credentials.DB.Port,
		c.Credentials.DB.User,
		c.Credentials.DB.Password,
		c.Credentials.DB.DBName,
		c.Credentials.DB.SSLMode,
	)
}
