// Package config Конфигурация сервиса.
package config

import (
	"errors"
	"flag"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
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
		return errors.New("need url in a form protocol:host:port")
	}
	protocol := u.Scheme
	host := u.Hostname()
	port := u.Port()

	if host == "" || port == "" {
		return errors.New("host or port is empty")
	}

	addr.ServerAddress = host + ":" + port
	if addr.withScheme {
		addr.ServerAddress = protocol + "://" + addr.ServerAddress
	}
	return nil
}

// SetConfig Устанавливает конфигурацию.
func SetConfig() {
	SetConfigByFlag()
	parseEnvParams()
}

// GetConfig Возвращает конфигурацию.
func GetConfig() *Config {
	return cfg
}

func parseEnvParams() {
	_ = godotenv.Load(".env")
	var params EnvParams
	if err := env.Parse(&params); err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}
	cfg.FileStoragePath = path
	dir := filepath.Dir(cfg.FileStoragePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
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
	flag.Parse()
}
