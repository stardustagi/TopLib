package server

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/utils"
	"go.uber.org/zap"
)

type Backend struct {
	config     HttpServerConfig
	Ctx        context.Context
	Logger     *zap.Logger
	httpServer *HttpServer
}

func NewBackend(config []byte) (*Backend, error) {
	httpServer, err := NewHttpServer(config)
	if err != nil {
		return nil, err
	}
	configStruct, err := utils.Bytes2Struct[HttpServerConfig](config)
	if err != nil {
		panic("Failed to parse HTTP server configuration: " + err.Error())
	}
	return &Backend{
		config:     configStruct,
		Ctx:        context.Background(),
		Logger:     logs.GetLogger("http_backend"),
		httpServer: httpServer,
	}, nil
}

func (m *Backend) AddGroup(group string, middleware ...echo.MiddlewareFunc) {
	m.httpServer.AddGroup(group, middleware...)
}

func (m *Backend) AddPostHandler(group string, h IHandler) {
	m.httpServer.Post(h.GetName(), group, h)
}

func (m *Backend) AddGetHandler(group string, h IHandler) {
	m.httpServer.Get(h.GetName(), group, h)
}

func (m *Backend) AddHandler(method, path string, h IHandler) {
	m.httpServer.Handle(method, path, h)
}

func (m *Backend) AddNativeHandler(method string, path string, handler echo.HandlerFunc) {
	m.httpServer.AddNativeHandler(method, path, handler)
}

func (m *Backend) Start() error {
	return m.httpServer.Startup()
}

func (m *Backend) Stop() {
	m.httpServer.Stop()
}
