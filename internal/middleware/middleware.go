// Package middleware Промежуточные обработчики запросов.
package middleware

import (
	"errors"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/logger"

	"net/http"
	"strings"
	"time"
)

// RequireContentType Проверяет допустимый тип запроса.
func RequireContentType(rct string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ct := r.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, rct) {
				http.Error(w, fmt.Sprintf("unsupported content type, require: %s", rct), http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Logger Сохраняет информацию о выполнении запроса.
func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := logger.ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := logger.LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   &resp,
		}
		start := time.Now()
		h.ServeHTTP(&lw, r)
		uri := r.RequestURI
		method := r.Method
		duration := time.Since(start)
		status := resp.Status
		size := resp.Size
		log.Info().
			Timestamp().
			Str("uri", uri).
			Str("method", method).
			Str("duration", duration.String()).
			Int("status", status).
			Int("size", size).
			Send()
	})
}

func TrustedSubnet(cnf config.ServerConfig) func(http.Handler) http.Handler {
	var ipNet *net.IPNet
	if cnf.TrustedSubnet != "" {
		_, parsed, err := net.ParseCIDR(cnf.TrustedSubnet)
		if err == nil {
			ipNet = parsed
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ipNet != nil {
				ip, err := extractIP(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				if !ipNet.Contains(ip) {
					http.Error(w, "invalid ip", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractIP(r *http.Request) (net.IP, error) {
	ipStr := r.Header.Get("X-Real-IP")
	ip := net.ParseIP(ipStr)

	if ip != nil {
		return ip, nil
	}

	ips := r.Header.Get("X-Forwarded-For")
	if ips == "" {
		return nil, errors.New("X-Forwarded-For is empty")
	}

	ipStrs := strings.Split(ips, ",")
	ipStr = strings.TrimSpace(ipStrs[0])

	ip = net.ParseIP(ipStr)
	if ip == nil {
		return nil, errors.New("failed parse ip from headers")
	}

	return ip, nil
}
