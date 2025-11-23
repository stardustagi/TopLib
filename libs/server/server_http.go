package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/utils"
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

// NewHttpServer @title StarDust HTTP Server
// @version 1.0
// @description This is the HTTP server for StarDust backend.
// @host localhost:8080
// @BasePath /api
func NewHttpServer(configByte []byte) (*HttpServer, error) {
	engine := echo.New()
	config, err := utils.Bytes2Struct[HttpServerConfig](configByte)
	if err != nil {
		panic("Failed to parse HTTP server configuration: " + err.Error())
	}
	engine.Validator = &CustomValidator{Validator: validator.New()}
	//engine.Use()
	addr := fmt.Sprintf("%s:%d", config.Address, config.Port)
	if config.Cors {
		engine.Use(Cors())
	}
	if config.RequestLog {
		engine.Use(Request())
	}
	if config.Access {
		engine.Use(Access())
	}

	if config.Path != "" && config.Path[0] != '/' {
		return nil, errors.New("the http.path must start with a /")
	}

	srv := &HttpServer{
		ctx:    context.Background(),
		logger: logs.GetLogger("httpServer"),
		engine: engine,
		group:  make(map[string]*StarDustGroup),
		addr:   addr,
		path:   config.Path,
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
	err := m.Engine().Start(m.addr)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		m.logger.Fatal("failed to start server:", zap.Error(err))
		return err
	}
	return nil
}

func (m *HttpServer) Stop() {
	if err := m.engine.Shutdown(m.ctx); err != nil {
		m.logger.Error("shutdown http server:", zap.Error(err))
		return
	}
}

// StartWithGracefulShutdown 启动服务器并监听信号量进行优雅关闭
func (m *HttpServer) StartWithGracefulShutdown() error {
	// 启动服务器
	go func() {
		m.logger.Info("http server starting on:", zap.String("addr", m.addr))
		// 打印路由
		for _, route := range m.engine.Routes() {
			m.logger.Info("http route registered:", logs.String("method", route.Method), logs.String("path", route.Path))
		}

		if err := m.engine.Start(m.addr); err != nil && errors.Is(err, http.ErrServerClosed) {
			m.logger.Fatal("failed to start server:", zap.Error(err))
		}
	}()

	// 等待信号量进行优雅关闭
	return m.WaitForShutdown()
}

// WaitForShutdown 等待关闭信号并执行优雅关闭
func (m *HttpServer) WaitForShutdown() error {
	// 创建信号通道
	quit := make(chan os.Signal, 1)

	// 监听指定的信号量：SIGINT (Ctrl+C), SIGTERM
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-quit
	m.logger.Info("received shutdown signal:", zap.String("signal", sig.String()))

	// 创建带超时的上下文用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m.logger.Info("shutting down server...")

	// 执行优雅关闭
	if err := m.engine.Shutdown(ctx); err != nil {
		m.logger.Error("server forced to shutdown:", zap.Error(err))
		return err
	}

	m.logger.Info("server exited gracefully")
	return nil
}

// GracefulStop 手动触发优雅关闭（可用于程序内部调用）
func (m *HttpServer) GracefulStop() error {
	m.logger.Info("gracefully stopping server...")

	// 创建带超时的上下文用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.engine.Shutdown(ctx); err != nil {
		m.logger.Error("server forced to shutdown:", zap.Error(err))
		return err
	}

	m.logger.Info("server stopped gracefully")
	return nil
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
	urlPath, _ := url.JoinPath(m.path, "api", path)
	m.group[path] = NewStarDustGroup(path, m.engine.Group(urlPath, middleware...))
	m.logger.Info("http group registered:", logs.String("path", urlPath))
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

func (m *HttpServer) AddNativeHandler(method string, path string, handler echo.HandlerFunc) {
	path, _ = url.JoinPath(m.path, "api", path)
	m.engine.Add(method, path, handler)
	m.logger.Info("http native handler registered:", logs.String("method", method), logs.String("path", path))
}
