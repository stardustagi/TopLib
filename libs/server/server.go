package server

import (
	"context"

	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/libs/option"
	"github.com/stardustagi/TopLib/utils"
	"go.uber.org/zap"
)

type Server struct {
	Ctx    context.Context
	opts   *option.Options
	cancel context.CancelFunc
	logger *zap.Logger
	doneCh chan struct{}
}

func NewServer(opts *option.Options) (*Server, error) {

	ctx, cancel := context.WithCancel(context.TODO())

	srv := &Server{
		Ctx:    ctx,
		opts:   opts,
		cancel: cancel,
		logger: logs.GetLogger("Server"),
		doneCh: utils.MakeShutdownCh(),
	}
	return srv, nil
}

func (m *Server) HandleSignal() {
	<-m.doneCh
	m.cancel()
	m.logger.Info("server shutting...")
	m.logger.Info("server shutdown completed")
}
