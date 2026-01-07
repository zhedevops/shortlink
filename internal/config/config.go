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
	Server:       &NetAddress{},
	ResponseAddr: &NetAddress{},
}
var defaultAddress = &NetAddress{
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

func GetConfig(typeAddr string) *NetAddress {
	switch typeAddr {
	case "server":
		if cfg.Server.Host != "" && cfg.Server.Port != "" {
			return cfg.Server
		}
	case "response":
		if cfg.ResponseAddr.Host != "" && cfg.ResponseAddr.Port != "" {
			return cfg.ResponseAddr
		}
	}

	if err := godotenv.Load(".env"); err != nil {
		return defaultAddress
	}

	return &NetAddress{
		Host: os.Getenv("SHORTLINK_HTTP_URL"),
		Port: os.Getenv("SHORTLINK_HTTP_PORT"),
	}
}

func SetConfigByFlag() {
	flag.Var(cfg.Server, "a", "server address host:port")
	flag.Var(cfg.ResponseAddr, "b", "server response base address protocol://host:port")
	flag.Parse()
}
