package app

import (
	"the-line/backend/internal/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	cfg    config.Config
	engine *gin.Engine
}

func NewServer(cfg config.Config, database *gorm.DB) *Server {
	gin.SetMode(cfg.GinMode)
	return &Server{
		cfg:    cfg,
		engine: NewRouter(cfg, database),
	}
}

func (s *Server) Run() error {
	return s.engine.Run(":" + s.cfg.AppPort)
}
