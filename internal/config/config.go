package config

import (
	"errors"
	"flag"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type NetAddress struct {
	Protocol string
	Host     string
	Port     string
}

type Config struct {
	Server       *NetAddress
	ResponseAddr *NetAddress
}

var cfg = &Config{
	Server:       defaultServerAddress,
	ResponseAddr: defaultResponseAddress,
}
var defaultServerAddress = &NetAddress{
	Protocol: "http",
	Host:     "localhost",
	Port:     "8080",
}
var defaultResponseAddress = &NetAddress{
	Protocol: "http",
	Host:     "localhost",
	Port:     "8080",
}

func (addr *NetAddress) String() string {
	return addr.Host + ":" + addr.Port
}

func (addr *NetAddress) Set(flagVal string) error {
	if !strings.Contains(flagVal, "://") {
		flagVal = "http://" + flagVal
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

	addr.Protocol = protocol
	addr.Host = host
	addr.Port = port

	return nil
}

func GetConfig() *Config {
	if err := godotenv.Load(".env"); err == nil {
		host, _ := os.LookupEnv("SHORTLINK_HTTP_URL")
		port, _ := os.LookupEnv("SHORTLINK_HTTP_PORT")
		cfg.Server.Host = host
		cfg.Server.Port = port
	}

	SetConfigByFlag()

	return cfg
}

func SetConfigByFlag() {
	flag.Var(cfg.Server, "a", "server address host:port")
	flag.Var(cfg.ResponseAddr, "b", "server response base address protocol://host:port")
	flag.Parse()
}
