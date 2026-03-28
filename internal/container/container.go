package container

import (
	"github.com/zhedevops/shortlink/internal/audit"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/service"
)

type App struct {
	Audit   *audit.AuditService
	Service *service.Service
	Config  *config.Config
}
