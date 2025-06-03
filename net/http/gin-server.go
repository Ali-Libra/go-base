package http

import (
	"github.com/gin-gonic/gin"
)

type GinServer struct {
	engine *gin.Engine
}

func NewGinServer() *GinServer {
	return &GinServer{
		engine: gin.Default(),
	}
}

func (s *GinServer) Run() error {
	return s.engine.Run()
}
