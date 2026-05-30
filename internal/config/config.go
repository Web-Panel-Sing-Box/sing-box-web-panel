package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string         `yaml:"env" env:"SING_GROK_ENV" env-default:"production"`
	Runtime  RuntimeConfig  `yaml:"runtime"`
	Database DBConfig       `yaml:"database"`
	HTTP     HTTPConfig     `yaml:"http"`
	Frontend FrontendConfig `yaml:"frontend"`
	Auth     AuthConfig     `yaml:"auth"`
	SingBox  SingBoxConfig  `yaml:"sing_box"`
	Stats    StatsConfig    `yaml:"stats"`
	TLS      TLSConfig      `yaml:"tls"`
	Metrics  MetricsConfig  `yaml:"metrics"`
	Logging  LoggingConfig  `yaml:"logging"`
	Sub      SubConfig      `yaml:"subscription"`
}

// TLSConfig controls how the panel's own HTTP server is secured.
//   - off:         plain HTTP (default; sit behind a reverse proxy)
//   - file:        serve cert_file/key_file
//   - self_signed: generate a self-signed cert (works on a bare IP)
//   - acme:        Let's Encrypt via autocert for acme_domains (TLS-ALPN-01)
type TLSConfig struct {
	Mode            string   `yaml:"mode" env:"SING_GROK_TLS_MODE" env-default:"off"`
	CertFile        string   `yaml:"cert_file" env:"SING_GROK_TLS_CERT_FILE" env-default:""`
	KeyFile         string   `yaml:"key_file" env:"SING_GROK_TLS_KEY_FILE" env-default:""`
	ACMEEmail       string   `yaml:"acme_email" env:"SING_GROK_TLS_ACME_EMAIL" env-default:""`
	ACMEDomains     []string `yaml:"acme_domains" env:"SING_GROK_TLS_ACME_DOMAINS" env-separator:","`
	ACMECacheDir    string   `yaml:"acme_cache_dir" env:"SING_GROK_TLS_ACME_CACHE_DIR" env-default:"/var/lib/sing-grok/acme"`
	SelfSignedHosts []string `yaml:"self_signed_hosts" env:"SING_GROK_TLS_SELF_SIGNED_HOSTS" env-separator:","`
	SelfSignedDir   string   `yaml:"self_signed_dir" env:"SING_GROK_TLS_SELF_SIGNED_DIR" env-default:"./storage/tls"`
}

type RuntimeConfig struct {
	GoMemLimit string `yaml:"gomemlimit" env:"SING_GROK_RUNTIME_GOMEMLIMIT" env-default:"180MiB"`
	GoGC       int    `yaml:"gogc" env:"SING_GROK_RUNTIME_GOGC" env-default:"50"`
}

type DBConfig struct {
	Path          string `yaml:"path" env:"SING_GROK_DB_PATH" env-default:"/var/lib/sing-grok/panel.db"`
	JournalMode   string `yaml:"journal_mode" env:"SING_GROK_DB_JOURNAL_MODE" env-default:"wal"`
	Synchronous   string `yaml:"synchronous" env:"SING_GROK_DB_SYNCHRONOUS" env-default:"normal"`
	CacheSizeKB   int    `yaml:"cache_size_kb" env:"SING_GROK_DB_CACHE_SIZE_KB" env-default:"2000"`
	MmapSizeMB    int    `yaml:"mmap_size_mb" env:"SING_GROK_DB_MMAP_SIZE_MB" env-default:"32"`
	BusyTimeoutMS int    `yaml:"busy_timeout_ms" env:"SING_GROK_DB_BUSY_TIMEOUT_MS" env-default:"5000"`
	TempStore     string `yaml:"temp_store" env:"SING_GROK_DB_TEMP_STORE" env-default:"memory"`
	ForeignKeys   bool   `yaml:"foreign_keys" env:"SING_GROK_DB_FOREIGN_KEYS" env-default:"true"`
}

type HTTPConfig struct {
	Address         string        `yaml:"address" env:"SING_GROK_HTTP_ADDRESS" env-default:"127.0.0.1:8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"SING_GROK_HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"SING_GROK_HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"SING_GROK_HTTP_IDLE_TIMEOUT" env-default:"120s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SING_GROK_HTTP_SHUTDOWN_TIMEOUT" env-default:"10s"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes" env:"SING_GROK_HTTP_MAX_HEADER_BYTES" env-default:"1048576"`
	MaxConns        int           `yaml:"max_conns" env:"SING_GROK_HTTP_MAX_CONNS" env-default:"128"`
}

type FrontendConfig struct {
	ServeMode string        `yaml:"serve_mode" env:"SING_GROK_FRONTEND_SERVE_MODE" env-default:"embed"`
	DiskPath  string        `yaml:"disk_path" env:"SING_GROK_FRONTEND_DISK_PATH" env-default:"./frontend/dist"`
	CacheTTL  time.Duration `yaml:"cache_ttl" env:"SING_GROK_FRONTEND_CACHE_TTL" env-default:"360h"`
}

type AuthConfig struct {
	JWTSecret         string        `yaml:"jwt_secret" env:"SING_GROK_AUTH_JWT_SECRET"`
	JWTExpiry         time.Duration `yaml:"jwt_expiry" env:"SING_GROK_AUTH_JWT_EXPIRY" env-default:"24h"`
	AdminUser         string        `yaml:"admin_user" env:"SING_GROK_AUTH_ADMIN_USER" env-default:"admin"`
	AdminPassword     string        `yaml:"admin_password" env:"SING_GROK_AUTH_ADMIN_PASSWORD"`
	Argon2MemoryKB    uint32        `yaml:"argon2_memory_kb" env:"SING_GROK_AUTH_ARGON2_MEMORY_KB" env-default:"65536"`
	Argon2Iterations  uint32        `yaml:"argon2_iterations" env:"SING_GROK_AUTH_ARGON2_ITERATIONS" env-default:"3"`
	Argon2Parallelism uint8         `yaml:"argon2_parallelism" env:"SING_GROK_AUTH_ARGON2_PARALLELISM" env-default:"2"`
	LoginRateLimit    string        `yaml:"login_rate_limit" env:"SING_GROK_AUTH_LOGIN_RATE_LIMIT" env-default:"5/m"`
	APIRateLimit      string        `yaml:"api_rate_limit" env:"SING_GROK_AUTH_API_RATE_LIMIT" env-default:"100/s"`
}

type SingBoxConfig struct {
	BinaryPath   string        `yaml:"binary_path" env:"SING_GROK_SING_BOX_BINARY_PATH" env-default:"/opt/sing-grok/bin/sing-box"`
	ConfigPath   string        `yaml:"config_path" env:"SING_GROK_SING_BOX_CONFIG_PATH" env-default:"/etc/sing-grok/config.json"`
	WorkingDir   string        `yaml:"working_dir" env:"SING_GROK_SING_BOX_WORKING_DIR" env-default:"/etc/sing-grok"`
	APIAddress   string        `yaml:"api_address" env:"SING_GROK_SING_BOX_API_ADDRESS" env-default:"127.0.0.1:9090"`
	APISecret    string        `yaml:"api_secret" env:"SING_GROK_SING_BOX_API_SECRET"`
	CheckTimeout time.Duration `yaml:"check_timeout" env:"SING_GROK_SING_BOX_CHECK_TIMEOUT" env-default:"8s"`
	RestartDelay time.Duration `yaml:"restart_delay" env:"SING_GROK_SING_BOX_RESTART_DELAY" env-default:"2s"`
	MaxRestarts  int           `yaml:"max_restarts" env:"SING_GROK_SING_BOX_MAX_RESTARTS" env-default:"4"`
	ProcessMode  string        `yaml:"process_mode" env:"SING_GROK_SING_BOX_PROCESS_MODE" env-default:"auto"`
	ServiceName  string        `yaml:"service_name" env:"SING_GROK_SING_BOX_SERVICE_NAME" env-default:"sing-box"`
}

type StatsConfig struct {
	Source          string `yaml:"source" env:"SING_GROK_STATS_SOURCE" env-default:"clash"`
	V2RayAPIAddress string `yaml:"v2ray_api_address" env:"SING_GROK_STATS_V2RAY_API_ADDRESS" env-default:"127.0.0.1:8088"`
}

type MetricsConfig struct {
	SystemInterval     time.Duration `yaml:"system_interval" env:"SING_GROK_METRICS_SYSTEM_INTERVAL" env-default:"2s"`
	TrafficInterval    time.Duration `yaml:"traffic_interval" env:"SING_GROK_METRICS_TRAFFIC_INTERVAL" env-default:"1s"`
	HistorySize        int           `yaml:"history_size" env:"SING_GROK_METRICS_HISTORY_SIZE" env-default:"60"`
	BatchFlushInterval time.Duration `yaml:"batch_flush_interval" env:"SING_GROK_METRICS_BATCH_FLUSH_INTERVAL" env-default:"5s"`
}

type LoggingConfig struct {
	Level          string `yaml:"level" env:"SING_GROK_LOG_LEVEL" env-default:"info"`
	Format         string `yaml:"format" env:"SING_GROK_LOG_FORMAT" env-default:"json"`
	FilePath       string `yaml:"file_path" env:"SING_GROK_LOG_FILE_PATH" env-default:""`
	MaxMemoryLines int    `yaml:"max_memory_lines" env:"SING_GROK_LOG_MAX_MEMORY_LINES" env-default:"200"`
	MaxFileSizeMB  int    `yaml:"max_file_size_mb" env:"SING_GROK_LOG_MAX_FILE_SIZE_MB" env-default:"10"`
	MaxFileBackups int    `yaml:"max_file_backups" env:"SING_GROK_LOG_MAX_FILE_BACKUPS" env-default:"2"`
}

type SubConfig struct {
	PublicURL string        `yaml:"public_url" env:"SING_GROK_SUB_PUBLIC_URL" env-default:""`
	TokenTTL  time.Duration `yaml:"token_ttl" env:"SING_GROK_SUB_TOKEN_TTL" env-default:"720h"`
}

func MustLoad() *Config {
	var cfg Config

	configPath := os.Getenv("SING_GROK_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/dev.yaml"
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Errorf("config: %w", err))
	}

	validate(&cfg)
	return &cfg
}

func validate(cfg *Config) {
	if !cfg.isDev() {
		if cfg.Auth.JWTSecret == "" || cfg.Auth.JWTSecret == "change-me" {
			panic("config: SING_GROK_AUTH_JWT_SECRET must be set to a secure value in production")
		}
		if cfg.Auth.AdminPassword == "" {
			panic("config: SING_GROK_AUTH_ADMIN_PASSWORD must be set in production")
		}
		if cfg.SingBox.APISecret == "" || cfg.SingBox.APISecret == "change-me" {
			panic("config: SING_GROK_SING_BOX_API_SECRET must be set to a secure value in production")
		}
	}

	if cfg.Runtime.GoGC < 10 || cfg.Runtime.GoGC > 200 {
		panic(fmt.Sprintf("config: runtime.gogc must be between 10 and 200, got %d", cfg.Runtime.GoGC))
	}

	if cfg.HTTP.MaxConns < 1 || cfg.HTTP.MaxConns > 65535 {
		panic(fmt.Sprintf("config: http.max_conns must be between 1 and 65535, got %d", cfg.HTTP.MaxConns))
	}

	if cfg.Database.CacheSizeKB < 200 {
		panic(fmt.Sprintf("config: database.cache_size_kb too low (%d), minimum 200", cfg.Database.CacheSizeKB))
	}
}

func (c *Config) isDev() bool {
	return c.Env == "dev" || c.Env == "local" || c.Env == "development"
}
