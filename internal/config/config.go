package config

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type NetAddress struct {
	Host string
	Port string
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
	Host: "localhost",
	Port: "8080",
}

func (addr *NetAddress) String() string {
	return addr.Host + ":" + addr.Port
}

func (addr *NetAddress) Set(flagVal string) error {
	v := strings.Split(flagVal, ":")
	if len(v) != 2 {
		return errors.New("need address in a form host:port")
	}
	host := strings.TrimSpace(v[0])
	port := strings.TrimSpace(v[1])

	if host == "" || port == "" {
		return errors.New("host or port is empty")
	}
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
	flag.Var(cfg.ResponseAddr, "b", "server response base address host:port")
	flag.Parse()
}
