package config

import (
	"errors"
	"flag"
	"log"
	"net/url"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

var scheme = "http://"
var defaultAddress = "localhost:8080"

type netAddress struct {
	ServerAddress string
	withScheme    bool
}

type EnvParams struct {
	ServerAddr   string `env:"SERVER_ADDRESS"`
	ResponseAddr string `env:"BASE_URL"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
}

type Config struct {
	ServerAddr   *netAddress
	ResponseAddr *netAddress
	LogLevel     string
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

func GetConfig() *Config {
	SetConfigByFlag()
	parseEnvParams()

	return cfg
}

func parseEnvParams() {
	_ = godotenv.Load(".env")
	var params EnvParams
	err := env.Parse(&params)
	if err != nil {
		log.Fatal(err)
	}

	if params.ServerAddr != "" {
		cfg.ServerAddr.ServerAddress = params.ServerAddr
	}
	if params.ResponseAddr != "" {
		cfg.ResponseAddr.ServerAddress = params.ResponseAddr
	}
	if params.LogLevel != "" {
		cfg.LogLevel = params.LogLevel
	}
}

func SetConfigByFlag() {
	flag.Var(cfg.ServerAddr, "a", "server address host:port")
	flag.Var(cfg.ResponseAddr, "b", "server response base address protocol://host:port")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")
	flag.Parse()
}
