// Package config Конфигурация сервиса.
package config

import (
	"encoding/json"
	"flag"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/zhedevops/shortlink/internal/model"
)

var (
	scheme         = "http://"
	defaultAddress = "localhost:8080"
)

type netAddress struct {
	ServerAddress string
	withScheme    bool
}

type EnvParams struct {
	ServerAddr      *string `env:"SERVER_ADDRESS"`
	ResponseAddr    *string `env:"BASE_URL"`
	LogLevel        *string `env:"LOG_LEVEL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	DatabaseDsn     *string `env:"DATABASE_DSN"`
	AuditFile       *string `env:"AUDIT_FILE"`
	AuditURL        *string `env:"AUDIT_URL"`
	EnableHTTPS     *bool   `env:"ENABLE_HTTPS"`
	TrustedSubnet   *string `env:"TRUSTED_SUBNET"`
	Key             *string `env:"KEY" envDefault:"kjdfkklsdf932.fjs"`
	GRPCAddress     *string `env:"GRPC_ADDRESS" envDefault:":3200"`
}

// Config Тип конфигурации, содержащий всё необходимую информацю для работы сервиса.
type Config struct {
	Server   ServerConfig
	Log      LogConfig
	Storage  StorageConfig
	Audit    AuditConfig
	Security SecurityConfig
}

type ServerConfig struct {
	ServerAddr    *netAddress
	ResponseAddr  *netAddress
	EnableHTTPS   bool
	TrustedSubnet string
	GRPCAddress   string
}

type StorageConfig struct {
	FileStoragePath string
	DatabaseDsn     string
}

type AuditConfig struct {
	AuditFile string
	AuditURL  string
}

type SecurityConfig struct {
	Key string
}

type LogConfig struct {
	LogLevel string
}

type ConfigFile struct {
	ServerAddress   string `json:"server_address"`
	BaseUrl         string `json:"base_url"`
	LogLevel        string `json:"log_level"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDsn     string `json:"database_dsn"`
	AuditFile       string `json:"audit_file"`
	AuditURL        string `json:"audit_url"`
	EnableHTTPS     bool   `json:"enable_https"`
	TrustedSubnet   string `json:"trusted_subnet"`
}

var serverConfig = ServerConfig{
	ServerAddr:   &netAddress{ServerAddress: defaultAddress, withScheme: false},
	ResponseAddr: &netAddress{ServerAddress: scheme + defaultAddress, withScheme: true},
}

var cfg = &Config{
	Server: serverConfig,
}

func (addr *netAddress) String() string {
	return addr.ServerAddress
}

func (addr *netAddress) Set(flagVal string) error {
	if !strings.Contains(flagVal, "://") {
		flagVal = scheme + flagVal
	}
	u, err := url.Parse(flagVal)
	if err != nil {
		return model.ErrServerAddressFlagValue
	}
	protocol := u.Scheme
	host := u.Hostname()
	port := u.Port()

	if host == "" || port == "" {
		return model.ErrHostPort
	}

	addr.ServerAddress = host + ":" + port
	if addr.withScheme {
		addr.ServerAddress = protocol + "://" + addr.ServerAddress
	}
	return nil
}

// SetConfig Устанавливает конфигурацию.
func SetConfig() error {
	SetConfigByConfigFile()
	SetConfigByFlag()
	return parseEnvParams()
}

// GetConfig Возвращает конфигурацию.
func GetConfig() *Config {
	return cfg
}

func parseEnvParams() error {
	_ = godotenv.Load(".env")
	var params EnvParams
	if err := env.Parse(&params); err != nil {
		return err
	}

	if params.ServerAddr != nil {
		cfg.Server.ServerAddr.ServerAddress = *params.ServerAddr
	}
	if params.ResponseAddr != nil {
		cfg.Server.ResponseAddr.ServerAddress = *params.ResponseAddr
	}
	if params.LogLevel != nil {
		cfg.Log.LogLevel = *params.LogLevel
	}

	if params.FileStoragePath != nil {
		cfg.Storage.FileStoragePath = *params.FileStoragePath
	}
	path, err := filepath.Abs(cfg.Storage.FileStoragePath)
	if err != nil {
		return err
	}
	cfg.Storage.FileStoragePath = path
	dir := filepath.Dir(cfg.Storage.FileStoragePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if params.DatabaseDsn != nil {
		cfg.Storage.DatabaseDsn = *params.DatabaseDsn
	}

	if params.AuditFile != nil {
		cfg.Audit.AuditFile = *params.AuditFile
	}

	if params.AuditURL != nil {
		cfg.Audit.AuditURL = *params.AuditURL
	}

	if params.EnableHTTPS != nil {
		cfg.Server.EnableHTTPS = *params.EnableHTTPS
	}

	if params.TrustedSubnet != nil {
		cfg.Server.TrustedSubnet = *params.TrustedSubnet
	}

	if params.GRPCAddress != nil {
		cfg.Server.GRPCAddress = *params.GRPCAddress
	}

	if params.Key != nil {
		cfg.Security.Key = *params.Key
	}

	return nil
}

func SetConfigByConfigFile() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "config file")
	flag.StringVar(&configPath, "config", "", "config file")
	flag.Parse()

	if configPath == "" {
		configPath, _ = os.LookupEnv("CONFIG")
	}

	if configPath != "" {
		fileCfg, err := loadConfig(configPath)
		if err == nil {
			if fileCfg.ServerAddress != "" {
				cfg.Server.ServerAddr = &netAddress{ServerAddress: fileCfg.ServerAddress, withScheme: false}
			}
			if fileCfg.BaseUrl != "" {
				cfg.Server.ResponseAddr = &netAddress{ServerAddress: fileCfg.BaseUrl, withScheme: true}
			}
			if fileCfg.LogLevel != "" {
				cfg.Log.LogLevel = fileCfg.LogLevel
			}
			if fileCfg.FileStoragePath != "" {
				cfg.Storage.FileStoragePath = fileCfg.FileStoragePath
			}
			if fileCfg.DatabaseDsn != "" {
				cfg.Storage.DatabaseDsn = fileCfg.DatabaseDsn
			}
			if fileCfg.AuditFile != "" {
				cfg.Audit.AuditFile = fileCfg.AuditFile
			}
			if fileCfg.AuditURL != "" {
				cfg.Audit.AuditURL = fileCfg.AuditURL
			}
			if fileCfg.EnableHTTPS {
				cfg.Server.EnableHTTPS = true
			}
			if fileCfg.TrustedSubnet != "" {
				cfg.Server.TrustedSubnet = fileCfg.TrustedSubnet
			}
		}
	}
}

func loadConfig(path string) (ConfigFile, error) {
	var cfg ConfigFile

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

// SetConfigByFlag Осуществляет установку значений конфигурации из переданных флагов.
func SetConfigByFlag() {
	flag.Var(serverConfig.ServerAddr, "a", "server address host:port")
	flag.Var(serverConfig.ResponseAddr, "b", "server response base address protocol://host:port")
	flag.StringVar(&cfg.Log.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.Storage.FileStoragePath, "f", "data/files/defaultpath/store.json", "storage path")
	flag.StringVar(&cfg.Storage.DatabaseDsn, "d", "", "db dsn")
	flag.StringVar(&cfg.Audit.AuditFile, "audit-file", "", "audit-file")
	flag.StringVar(&cfg.Audit.AuditURL, "audit-url", "", "audit-url")
	flag.StringVar(&serverConfig.TrustedSubnet, "t", "", "trusted_subnet")
	flag.BoolVar(&serverConfig.EnableHTTPS, "s", false, "EnableHTTPS")
	flag.Parse()
}
