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
	Key             *string `env:"KEY" envDefault:"kjdfkklsdf932.fjs"`
}

// Config Тип конфигурации, содержащий всё необходимую информацю для работы сервиса.
type Config struct {
	ServerAddr      *netAddress
	ResponseAddr    *netAddress
	LogLevel        string
	FileStoragePath string
	DatabaseDsn     string
	AuditFile       string
	AuditURL        string
	EnableHTTPS     bool
	Key             string
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
}

var cfg = &Config{
	ServerAddr:   &netAddress{ServerAddress: defaultAddress, withScheme: false},
	ResponseAddr: &netAddress{ServerAddress: scheme + defaultAddress, withScheme: true},
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
		cfg.ServerAddr.ServerAddress = *params.ServerAddr
	}
	if params.ResponseAddr != nil {
		cfg.ResponseAddr.ServerAddress = *params.ResponseAddr
	}
	if params.LogLevel != nil {
		cfg.LogLevel = *params.LogLevel
	}

	if params.FileStoragePath != nil {
		cfg.FileStoragePath = *params.FileStoragePath
	}
	path, err := filepath.Abs(cfg.FileStoragePath)
	if err != nil {
		return err
	}
	cfg.FileStoragePath = path
	dir := filepath.Dir(cfg.FileStoragePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if params.DatabaseDsn != nil {
		cfg.DatabaseDsn = *params.DatabaseDsn
	}

	if params.AuditFile != nil {
		cfg.AuditFile = *params.AuditFile
	}

	if params.AuditURL != nil {
		cfg.AuditURL = *params.AuditURL
	}

	if params.EnableHTTPS != nil {
		cfg.EnableHTTPS = *params.EnableHTTPS
	}

	return nil
}

func SetConfigByConfigFile() {
	var configPath string
	flag.StringVar(&configPath, "c", "", "config file")
	flag.StringVar(&configPath, "config", "", "config file")
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if configPath != "" {
		fileCfg, err := loadConfig(configPath)
		if err == nil {
			if fileCfg.ServerAddress != "" {
				cfg.ServerAddr = &netAddress{ServerAddress: fileCfg.ServerAddress, withScheme: false}
			}
			if fileCfg.BaseUrl != "" {
				cfg.ResponseAddr = &netAddress{ServerAddress: fileCfg.BaseUrl, withScheme: true}
			}
			if fileCfg.LogLevel != "" {
				cfg.LogLevel = fileCfg.LogLevel
			}
			if fileCfg.FileStoragePath != "" {
				cfg.FileStoragePath = fileCfg.FileStoragePath
			}
			if fileCfg.DatabaseDsn != "" {
				cfg.DatabaseDsn = fileCfg.DatabaseDsn
			}
			if fileCfg.AuditFile != "" {
				cfg.AuditFile = fileCfg.AuditFile
			}
			if fileCfg.AuditURL != "" {
				cfg.AuditURL = fileCfg.AuditURL
			}
			if fileCfg.EnableHTTPS {
				cfg.EnableHTTPS = true
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
	flag.Var(cfg.ServerAddr, "a", "server address host:port")
	flag.Var(cfg.ResponseAddr, "b", "server response base address protocol://host:port")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.StringVar(&cfg.FileStoragePath, "f", "data/files/defaultpath/store.json", "storage path")
	flag.StringVar(&cfg.DatabaseDsn, "d", "", "db dsn")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "audit-file")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "audit-url")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "EnableHTTPS")
	flag.Parse()
}
