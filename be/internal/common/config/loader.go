package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Load loads configuration from environment variables and optional YAML file.
// It follows 12-factor app principles with environment variable overrides.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Bind environment variables
	bindEnvVars(v)

	// Load from YAML file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			// Config file is optional, only error if file exists but can't be read
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	}

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadFromEnv loads configuration from environment variables only.
func LoadFromEnv() (*Config, error) {
	return Load("")
}

// setDefaults sets default values for all configuration options.
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.env", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.log_level", "debug")

	// Service defaults
	v.SetDefault("user_service.host", "0.0.0.0")
	v.SetDefault("user_service.port", 8001)
	v.SetDefault("user_service.grpc_port", 9001)

	v.SetDefault("pod_service.host", "0.0.0.0")
	v.SetDefault("pod_service.port", 8002)
	v.SetDefault("pod_service.grpc_port", 9002)

	v.SetDefault("material_service.host", "0.0.0.0")
	v.SetDefault("material_service.port", 8003)
	v.SetDefault("material_service.grpc_port", 9003)

	v.SetDefault("search_service.host", "0.0.0.0")
	v.SetDefault("search_service.port", 8004)

	v.SetDefault("ai_service.host", "0.0.0.0")
	v.SetDefault("ai_service.port", 8005)

	v.SetDefault("notification_service.host", "0.0.0.0")
	v.SetDefault("notification_service.port", 8006)

	v.SetDefault("file_processor.host", "0.0.0.0")
	v.SetDefault("file_processor.port", 8007)
	v.SetDefault("file_processor.url", "http://file-processor:8007")

	// JWT defaults
	v.SetDefault("jwt.access_token_expiry", 15*time.Minute)
	v.SetDefault("jwt.refresh_token_expiry", 7*24*time.Hour)

	// Database defaults (User DB as example, others follow same pattern)
	setDatabaseDefaults(v, "user_db", "ngasihtau_users", 25, 10)
	setDatabaseDefaults(v, "pod_db", "ngasihtau_pods", 20, 8)
	setDatabaseDefaults(v, "material_db", "ngasihtau_materials", 15, 5)
	setDatabaseDefaults(v, "ai_db", "ngasihtau_ai", 10, 5)
	setDatabaseDefaults(v, "notification_db", "ngasihtau_notifications", 10, 5)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.pool_timeout", 4*time.Second)
	v.SetDefault("redis.idle_timeout", 5*time.Minute)

	// NATS defaults
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("nats.cluster_id", "ngasihtau-cluster")
	v.SetDefault("nats.client_id", "ngasihtau-backend")

	// MinIO defaults
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.access_key", "minioadmin")
	v.SetDefault("minio.secret_key", "minioadmin")
	v.SetDefault("minio.use_ssl", false)
	v.SetDefault("minio.bucket_materials", "materials")

	// Meilisearch defaults
	v.SetDefault("meilisearch.host", "http://localhost:7700")

	// Qdrant defaults
	v.SetDefault("qdrant.host", "localhost")
	v.SetDefault("qdrant.port", 6333)
	v.SetDefault("qdrant.collection", "material_chunks")

	// OpenAI defaults
	v.SetDefault("openai.embedding_model", "text-embedding-3-small")
	v.SetDefault("openai.chat_model", "gpt-4")
	v.SetDefault("openai.chunk_size_min", 500)
	v.SetDefault("openai.chunk_size_max", 1000)
	v.SetDefault("openai.chunk_overlap", 100)

	// Rate limit defaults (requests per minute)
	v.SetDefault("rate_limit.auth", 10)
	v.SetDefault("rate_limit.search", 60)
	v.SetDefault("rate_limit.ai_chat", 30)
	v.SetDefault("rate_limit.upload", 10)
	v.SetDefault("rate_limit.general", 100)

	// CORS defaults
	v.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:5173"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Authorization", "Content-Type", "X-Request-ID"})
	v.SetDefault("cors.max_age", 86400)

	// Observability defaults
	v.SetDefault("observability.otlp_endpoint", "http://localhost:4317")
	v.SetDefault("observability.service_name", "ngasihtau")
	v.SetDefault("observability.metrics_enabled", true)
	v.SetDefault("observability.metrics_port", 9090)
}

func setDatabaseDefaults(v *viper.Viper, prefix, dbName string, maxOpen, maxIdle int) {
	v.SetDefault(prefix+".host", "localhost")
	v.SetDefault(prefix+".port", 5432)
	v.SetDefault(prefix+".name", dbName)
	v.SetDefault(prefix+".user", "postgres")
	v.SetDefault(prefix+".password", "postgres")
	v.SetDefault(prefix+".ssl_mode", "disable")
	v.SetDefault(prefix+".max_open_conns", maxOpen)
	v.SetDefault(prefix+".max_idle_conns", maxIdle)
	v.SetDefault(prefix+".conn_max_lifetime", 5*time.Minute)
}

// bindEnvVars binds environment variables to viper keys.
func bindEnvVars(v *viper.Viper) {
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// App
	_ = v.BindEnv("app.env", "APP_ENV")
	_ = v.BindEnv("app.debug", "APP_DEBUG")
	_ = v.BindEnv("app.log_level", "APP_LOG_LEVEL")

	// User Service
	_ = v.BindEnv("user_service.host", "USER_SERVICE_HOST")
	_ = v.BindEnv("user_service.port", "USER_SERVICE_PORT")
	_ = v.BindEnv("user_service.grpc_port", "USER_SERVICE_GRPC_PORT")

	// Pod Service
	_ = v.BindEnv("pod_service.host", "POD_SERVICE_HOST")
	_ = v.BindEnv("pod_service.port", "POD_SERVICE_PORT")
	_ = v.BindEnv("pod_service.grpc_port", "POD_SERVICE_GRPC_PORT")

	// Material Service
	_ = v.BindEnv("material_service.host", "MATERIAL_SERVICE_HOST")
	_ = v.BindEnv("material_service.port", "MATERIAL_SERVICE_PORT")
	_ = v.BindEnv("material_service.grpc_port", "MATERIAL_SERVICE_GRPC_PORT")

	// Search Service
	_ = v.BindEnv("search_service.host", "SEARCH_SERVICE_HOST")
	_ = v.BindEnv("search_service.port", "SEARCH_SERVICE_PORT")

	// AI Service
	_ = v.BindEnv("ai_service.host", "AI_SERVICE_HOST")
	_ = v.BindEnv("ai_service.port", "AI_SERVICE_PORT")

	// Notification Service
	_ = v.BindEnv("notification_service.host", "NOTIFICATION_SERVICE_HOST")
	_ = v.BindEnv("notification_service.port", "NOTIFICATION_SERVICE_PORT")

	// File Processor
	_ = v.BindEnv("file_processor.host", "FILE_PROCESSOR_HOST")
	_ = v.BindEnv("file_processor.port", "FILE_PROCESSOR_PORT")
	_ = v.BindEnv("file_processor.url", "FILE_PROCESSOR_URL")

	// JWT
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("jwt.access_token_expiry", "JWT_ACCESS_TOKEN_EXPIRY")
	_ = v.BindEnv("jwt.refresh_token_expiry", "JWT_REFRESH_TOKEN_EXPIRY")

	// Google OAuth
	_ = v.BindEnv("oauth.google.client_id", "GOOGLE_CLIENT_ID")
	_ = v.BindEnv("oauth.google.client_secret", "GOOGLE_CLIENT_SECRET")
	_ = v.BindEnv("oauth.google.redirect_url", "GOOGLE_REDIRECT_URL")

	// Database bindings
	bindDatabaseEnvVars(v, "user_db", "USER_DB")
	bindDatabaseEnvVars(v, "pod_db", "POD_DB")
	bindDatabaseEnvVars(v, "material_db", "MATERIAL_DB")
	bindDatabaseEnvVars(v, "ai_db", "AI_DB")
	bindDatabaseEnvVars(v, "notification_db", "NOTIFICATION_DB")

	// Redis
	_ = v.BindEnv("redis.host", "REDIS_HOST")
	_ = v.BindEnv("redis.port", "REDIS_PORT")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("redis.db", "REDIS_DB")
	_ = v.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")
	_ = v.BindEnv("redis.min_idle_conns", "REDIS_MIN_IDLE_CONNS")
	_ = v.BindEnv("redis.pool_timeout", "REDIS_POOL_TIMEOUT")
	_ = v.BindEnv("redis.idle_timeout", "REDIS_IDLE_TIMEOUT")

	// NATS
	_ = v.BindEnv("nats.url", "NATS_URL")
	_ = v.BindEnv("nats.cluster_id", "NATS_CLUSTER_ID")
	_ = v.BindEnv("nats.client_id", "NATS_CLIENT_ID")

	// MinIO
	_ = v.BindEnv("minio.endpoint", "MINIO_ENDPOINT")
	_ = v.BindEnv("minio.access_key", "MINIO_ACCESS_KEY")
	_ = v.BindEnv("minio.secret_key", "MINIO_SECRET_KEY")
	_ = v.BindEnv("minio.use_ssl", "MINIO_USE_SSL")
	_ = v.BindEnv("minio.bucket_materials", "MINIO_BUCKET_MATERIALS")

	// Meilisearch
	_ = v.BindEnv("meilisearch.host", "MEILISEARCH_HOST")
	_ = v.BindEnv("meilisearch.api_key", "MEILISEARCH_API_KEY")

	// Qdrant
	_ = v.BindEnv("qdrant.host", "QDRANT_HOST")
	_ = v.BindEnv("qdrant.port", "QDRANT_PORT")
	_ = v.BindEnv("qdrant.collection", "QDRANT_COLLECTION")

	// OpenAI
	_ = v.BindEnv("openai.api_key", "OPENAI_API_KEY")
	_ = v.BindEnv("openai.embedding_model", "OPENAI_EMBEDDING_MODEL")
	_ = v.BindEnv("openai.chat_model", "OPENAI_CHAT_MODEL")
	_ = v.BindEnv("openai.chunk_size_min", "CHUNK_SIZE_MIN")
	_ = v.BindEnv("openai.chunk_size_max", "CHUNK_SIZE_MAX")
	_ = v.BindEnv("openai.chunk_overlap", "CHUNK_OVERLAP")

	// SMTP
	_ = v.BindEnv("smtp.host", "SMTP_HOST")
	_ = v.BindEnv("smtp.port", "SMTP_PORT")
	_ = v.BindEnv("smtp.user", "SMTP_USER")
	_ = v.BindEnv("smtp.password", "SMTP_PASSWORD")
	_ = v.BindEnv("smtp.from_email", "SMTP_FROM_EMAIL")
	_ = v.BindEnv("smtp.from_name", "SMTP_FROM_NAME")

	// Email Provider
	_ = v.BindEnv("email.provider", "EMAIL_PROVIDER")
	_ = v.BindEnv("email.from_email", "EMAIL_FROM_EMAIL")
	_ = v.BindEnv("email.from_name", "EMAIL_FROM_NAME")
	_ = v.BindEnv("email.sendgrid_api_key", "SENDGRID_API_KEY")
	_ = v.BindEnv("email.ses_region", "AWS_SES_REGION")
	_ = v.BindEnv("email.ses_access_key", "AWS_SES_ACCESS_KEY")
	_ = v.BindEnv("email.ses_secret_key", "AWS_SES_SECRET_KEY")

	// Rate Limits
	_ = v.BindEnv("rate_limit.auth", "RATE_LIMIT_AUTH")
	_ = v.BindEnv("rate_limit.search", "RATE_LIMIT_SEARCH")
	_ = v.BindEnv("rate_limit.ai_chat", "RATE_LIMIT_AI_CHAT")
	_ = v.BindEnv("rate_limit.upload", "RATE_LIMIT_UPLOAD")
	_ = v.BindEnv("rate_limit.general", "RATE_LIMIT_GENERAL")

	// CORS
	_ = v.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS")
	_ = v.BindEnv("cors.allowed_methods", "CORS_ALLOWED_METHODS")
	_ = v.BindEnv("cors.allowed_headers", "CORS_ALLOWED_HEADERS")
	_ = v.BindEnv("cors.max_age", "CORS_MAX_AGE")

	// Observability
	_ = v.BindEnv("observability.otlp_endpoint", "OTEL_EXPORTER_OTLP_ENDPOINT")
	_ = v.BindEnv("observability.service_name", "OTEL_SERVICE_NAME")
	_ = v.BindEnv("observability.metrics_enabled", "METRICS_ENABLED")
	_ = v.BindEnv("observability.metrics_port", "METRICS_PORT")
}

func bindDatabaseEnvVars(v *viper.Viper, prefix, envPrefix string) {
	_ = v.BindEnv(prefix+".host", envPrefix+"_HOST")
	_ = v.BindEnv(prefix+".port", envPrefix+"_PORT")
	_ = v.BindEnv(prefix+".name", envPrefix+"_NAME")
	_ = v.BindEnv(prefix+".user", envPrefix+"_USER")
	_ = v.BindEnv(prefix+".password", envPrefix+"_PASSWORD")
	_ = v.BindEnv(prefix+".ssl_mode", envPrefix+"_SSL_MODE")
	_ = v.BindEnv(prefix+".max_open_conns", envPrefix+"_MAX_OPEN_CONNS")
	_ = v.BindEnv(prefix+".max_idle_conns", envPrefix+"_MAX_IDLE_CONNS")
	_ = v.BindEnv(prefix+".conn_max_lifetime", envPrefix+"_CONN_MAX_LIFETIME")
}
