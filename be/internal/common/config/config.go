// Package config provides configuration management using Viper.
// It supports environment variables and YAML config files with validation at startup.
package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the application.
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	UserService   ServiceConfig       `mapstructure:"user_service"`
	PodService    ServiceConfig       `mapstructure:"pod_service"`
	MatService    ServiceConfig       `mapstructure:"material_service"`
	SearchSvc     ServiceConfig       `mapstructure:"search_service"`
	AISvc         ServiceConfig       `mapstructure:"ai_service"`
	NotifSvc      ServiceConfig       `mapstructure:"notification_service"`
	FileProcSvc   FileProcConfig      `mapstructure:"file_processor"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	OAuth         OAuthConfig         `mapstructure:"oauth"`
	UserDB        DatabaseConfig      `mapstructure:"user_db"`
	PodDB         DatabaseConfig      `mapstructure:"pod_db"`
	MaterialDB    DatabaseConfig      `mapstructure:"material_db"`
	AIDB          DatabaseConfig      `mapstructure:"ai_db"`
	NotifDB       DatabaseConfig      `mapstructure:"notification_db"`
	Redis         RedisConfig         `mapstructure:"redis"`
	NATS          NATSConfig          `mapstructure:"nats"`
	MinIO         MinIOConfig         `mapstructure:"minio"`
	Meilisearch   MeilisearchConfig   `mapstructure:"meilisearch"`
	Qdrant        QdrantConfig        `mapstructure:"qdrant"`
	OpenAI        OpenAIConfig        `mapstructure:"openai"`
	SMTP          SMTPConfig          `mapstructure:"smtp"`
	Email         EmailConfig         `mapstructure:"email"`
	RateLimit     RateLimitConfig     `mapstructure:"rate_limit"`
	CORS          CORSConfig          `mapstructure:"cors"`
	Observability ObservabilityConfig `mapstructure:"observability"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Env      string `mapstructure:"env" validate:"required,oneof=development staging production"`
	Debug    bool   `mapstructure:"debug"`
	LogLevel string `mapstructure:"log_level" validate:"required,oneof=debug info warn error"`
}

// ServiceConfig holds common service configuration.
type ServiceConfig struct {
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	GRPCPort int    `mapstructure:"grpc_port" validate:"omitempty,min=1,max=65535"`
}

// FileProcConfig holds file processor service configuration.
type FileProcConfig struct {
	Host string `mapstructure:"host" validate:"required"`
	Port int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	URL  string `mapstructure:"url" validate:"required,url"`
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret             string        `mapstructure:"secret" validate:"required,min=32"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry" validate:"required"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry" validate:"required"`
}

// OAuthConfig holds OAuth provider settings.
type OAuthConfig struct {
	Google GoogleOAuthConfig `mapstructure:"google"`
}

// GoogleOAuthConfig holds Google OAuth settings.
type GoogleOAuthConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host" validate:"required"`
	Port            int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Name            string        `mapstructure:"name" validate:"required"`
	User            string        `mapstructure:"user" validate:"required"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"min=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"min=0"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// DSN returns the PostgreSQL connection string.
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host         string        `mapstructure:"host" validate:"required"`
	Port         int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db" validate:"min=0,max=15"`
	PoolSize     int           `mapstructure:"pool_size" validate:"min=1"`
	MinIdleConns int           `mapstructure:"min_idle_conns" validate:"min=0"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// Addr returns the Redis address string.
func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// NATSConfig holds NATS JetStream settings.
type NATSConfig struct {
	URL       string `mapstructure:"url" validate:"required,url"`
	ClusterID string `mapstructure:"cluster_id"`
	ClientID  string `mapstructure:"client_id"`
}

// MinIOConfig holds MinIO object storage settings.
type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint" validate:"required"`
	AccessKey       string `mapstructure:"access_key" validate:"required"`
	SecretKey       string `mapstructure:"secret_key" validate:"required"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	BucketMaterials string `mapstructure:"bucket_materials" validate:"required"`
}

// MeilisearchConfig holds Meilisearch settings.
type MeilisearchConfig struct {
	Host   string `mapstructure:"host" validate:"required,url"`
	APIKey string `mapstructure:"api_key"`
}

// QdrantConfig holds Qdrant vector database settings.
type QdrantConfig struct {
	Host       string `mapstructure:"host" validate:"required"`
	Port       int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Collection string `mapstructure:"collection" validate:"required"`
}

// OpenAIConfig holds OpenAI API settings.
type OpenAIConfig struct {
	APIKey         string `mapstructure:"api_key"`
	EmbeddingModel string `mapstructure:"embedding_model" validate:"required"`
	ChatModel      string `mapstructure:"chat_model" validate:"required"`
	ChunkSizeMin   int    `mapstructure:"chunk_size_min" validate:"min=100"`
	ChunkSizeMax   int    `mapstructure:"chunk_size_max" validate:"min=100"`
	ChunkOverlap   int    `mapstructure:"chunk_overlap" validate:"min=0"`
}

// SMTPConfig holds email SMTP settings.
type SMTPConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port" validate:"omitempty,min=1,max=65535"`
	User      string `mapstructure:"user"`
	Password  string `mapstructure:"password"`
	FromEmail string `mapstructure:"from_email" validate:"omitempty,email"`
	FromName  string `mapstructure:"from_name"`
}

// RateLimitConfig holds rate limiting settings (requests per minute).
type RateLimitConfig struct {
	Auth    int `mapstructure:"auth" validate:"min=1"`
	Search  int `mapstructure:"search" validate:"min=1"`
	AIChat  int `mapstructure:"ai_chat" validate:"min=1"`
	Upload  int `mapstructure:"upload" validate:"min=1"`
	General int `mapstructure:"general" validate:"min=1"`
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
	MaxAge         int      `mapstructure:"max_age"`
}

// ObservabilityConfig holds observability settings.
type ObservabilityConfig struct {
	OTLPEndpoint   string `mapstructure:"otlp_endpoint"`
	ServiceName    string `mapstructure:"service_name"`
	MetricsEnabled bool   `mapstructure:"metrics_enabled"`
	MetricsPort    int    `mapstructure:"metrics_port" validate:"omitempty,min=1,max=65535"`
}

// EmailConfig holds email provider settings.
type EmailConfig struct {
	Provider     string `mapstructure:"provider" validate:"required,oneof=sendgrid ses smtp"`
	FromEmail    string `mapstructure:"from_email" validate:"required,email"`
	FromName     string `mapstructure:"from_name" validate:"required"`
	SendGridKey  string `mapstructure:"sendgrid_api_key"`
	SESRegion    string `mapstructure:"ses_region"`
	SESAccessKey string `mapstructure:"ses_access_key"`
	SESSecretKey string `mapstructure:"ses_secret_key"`
}
