package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/libs/option"
	"go.uber.org/zap"
)

type HttpServer struct {
	ctx    context.Context
	addr   string
	path   string
	logger *zap.Logger
	engine *echo.Echo
	group  map[string]*StarDustGroup
}

func NewHttpServer(opts *option.Options) (*HttpServer, error) {
	engine := echo.New()
	engine.Validator = &CustomValidator{Validator: validator.New()}
	engine.Use()
	addr := fmt.Sprintf("%s:%d", opts.Http.Address, opts.Http.Port)
	if opts.Http.Cors {
		engine.Use(Cors())
	}
	if opts.Http.RequestLog {
		engine.Use(Request())
	}
	if opts.Http.Access {
		engine.Use(Access())
	}

	if opts.Http.Path != "" && opts.Http.Path[0] != '/' {
		return nil, errors.New("the http.path must start with a /")
	}

	srv := &HttpServer{
		ctx:    context.Background(),
		logger: logs.GetLogger("httpServer"),
		engine: engine,
		group:  make(map[string]*StarDustGroup),
		addr:   addr,
		path:   opts.Http.Path,
	}
	return srv, nil
}

func (m *HttpServer) Engine() *echo.Echo {
	return m.engine
}

func (m *HttpServer) Use(middleware ...echo.MiddlewareFunc) *HttpServer {
	m.engine.Use(middleware...)
	return m
}

func (m *HttpServer) Startup() error {
	m.logger.Info("http server listened on:", zap.String("addr", m.addr))
	// 打印路由
	for _, route := range m.engine.Routes() {
		m.logger.Info("http route registered:", logs.String("method", route.Method), logs.String("path", route.Path))
	}
	m.Engine().Start(m.addr)
	go func() {
		<-m.ctx.Done()
		m.Stop()
		m.engine.Close()
	}()
	return nil
}

func (m *HttpServer) Stop() {
	if err := m.engine.Shutdown(m.ctx); err != nil {
		m.logger.Error("shutdown http server:", zap.Error(err))
		return
	}
}

// Handle registers a new route with the HTTP server.
func (m *HttpServer) Handle(method string, path string, handler IHandler) {

	path, _ = url.JoinPath(m.path, "api", path)

	m.engine.Add(method, path, handler.GetFunc())
}

func (m *HttpServer) Internal(method string, path string, handler IHandler) {
	path, _ = url.JoinPath(m.path, "internal", path)
	m.engine.Add(method, path, handler.GetFunc())
}

func (m *HttpServer) AddGroup(path string, middleware ...echo.MiddlewareFunc) {
	url_path, _ := url.JoinPath(m.path, "api", path)
	m.group[path] = NewStarDustGroup(path, m.engine.Group(url_path, middleware...))
	m.logger.Info("http group registered:", logs.String("path", url_path))
}

func (m *HttpServer) Get(path string, group string, handler IHandler) {
	if group != "" {
		if _, exists := m.group[group]; !exists {
			m.logger.Error("group not found", logs.String("group", group))
			return
		}
		m.group[group].Group.GET(fmt.Sprintf("/%s", path), handler.GetFunc())
		m.logger.Info("http handler registered to group:", logs.String("path", path), logs.String("prefix", m.group[group].Prefix))
		return
	}
	m.Handle(http.MethodGet, path, handler)
}

func (m *HttpServer) Post(path string, group string, handler IHandler) {
	if group != "" {
		if _, exists := m.group[group]; !exists {
			m.logger.Error("group not found", zap.String("group", group))
			return
		}
		m.group[group].Group.POST(fmt.Sprintf("/%s", path), handler.GetFunc())
		m.logger.Info("http handler registered to group:", logs.String("path", path), logs.String("prefix", m.group[group].Prefix))
		return
	}
	m.Handle(http.MethodPost, path, handler)
}
