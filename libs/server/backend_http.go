package server

import (
	"context"

	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/libs/option"
	"go.uber.org/zap"
)

type Backend struct {
	opts       *option.Options
	Ctx        context.Context
	Logger     *zap.Logger
	httpServer *HttpServer
	handles    []IHandlers
}

func NewBackend(opts *option.Options) (*Backend, error) {
	httpServer, err := NewHttpServer(opts)
	if err != nil {
		return nil, err
	}
	return &Backend{
		opts:       opts,
		Ctx:        context.Background(),
		Logger:     logs.GetLogger("http_backend"),
		httpServer: httpServer,
	}, nil
}

func (m *Backend) AddGroup(grop string) {
	m.httpServer.AddGroup(grop)
}

func (m *Backend) AddPostHandler(group string, h IHandler) {
	m.httpServer.Post(h.GetName(), group, h)
}

func (m *Backend) AddGetHandler(group string, h IHandler) {
	m.httpServer.Get(h.GetName(), group, h)
}

func (m *Backend) Start() error {
	return m.httpServer.Startup()
}
