package app

import (
	"the-line-bridge/internal/client"
	"the-line-bridge/internal/config"
	"the-line-bridge/internal/runtime"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg    config.Config
	engine *gin.Engine
}

func NewServer(cfg config.Config, rt runtime.OpenClawRuntime, thelineClient *client.TheLineClient) *Server {
	return &Server{
		cfg:    cfg,
		engine: NewRouter(rt, thelineClient),
	}
}

func (s *Server) Run() error {
	return s.engine.Run(":" + s.cfg.Port)
}
