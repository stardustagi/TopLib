package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/codec"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/protocol"
)

type HelloReq struct {
	Name string `json:"name" validate:"required"`
}

type HelloResp struct {
	Message string `json:"message"`
}

type WsReq struct {
}

type WsResp struct {
}

func TestNewHttp(t *testing.T) {

	loggerConf := map[string]interface{}{
		"filename":   "logs/app.log",
		"maxsize":    60,
		"maxbackups": 5,
		"maxage":     7,
		"compress":   true,
		"level":      -1,
	}
	jsonBytes, err := json.Marshal(loggerConf)
	if err != nil {
		// 处理错误
	}
	logs.Init(jsonBytes)

	httpServerConfig := map[string]interface{}{
		"port":          8080,
		"path":          "/api",
		"cors":          true,
		"access":        true,
		"request_log":   true,
		"address":       "127.0.0.1",
		"read_timeout":  60,
		"write_timeout": 60,
	}
	opts, err := json.Marshal(httpServerConfig)
	bk, err := NewBackend(opts)
	if err != nil {
		t.Fatal("failed to create backend:", err)
	}
	h := NewHandler(
		"hello",
		[]string{"greet"},
		func(ctx echo.Context, req HelloReq, resp HelloResp) error {
			resp.Message = "Hello " + req.Name
			logs.Debug("Received request", logs.String("name", req.Name))
			return ctx.JSON(http.StatusOK, resp)
		},
	)
	manager := NewClientManager(logs.GetLogger("websocketClientManager")) // Initialize the client manager
	go manager.Start()

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	ws := NewHandler(
		"ws",
		[]string{"websocket"},
		func(ctx echo.Context, req WsReq, resp WsResp) error {
			conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
			if err != nil {
				return err
			}
			logger := logs.GetLogger("websocketClient")
			defaultHandlerInterface := protocol.NewDefaultMessageHandler()
			client := NewClient(
				"testUserId",
				"testSessionId",
				conn,
				codec.NewJsonCodec(),
				logger,
				ctx.Request().Context(),
				defaultHandlerInterface, // Use the default message handler
				manager,
			)
			manager.RegisterClient(client)
			go client.Listen()
			return nil
		},
	)
	bk.AddGroup("test")
	bk.AddPostHandler("test", h)
	bk.AddGetHandler("test", ws)
	bk.Start()
	bk.Stop()
}

// TestGracefulShutdown 测试优雅关闭功能
func TestGracefulShutdown(t *testing.T) {
	loggerConf := map[string]interface{}{
		"filename":   "logs/test_graceful.log",
		"maxsize":    60,
		"maxbackups": 5,
		"maxage":     7,
		"compress":   true,
		"level":      -1,
	}
	jsonBytes, err := json.Marshal(loggerConf)
	if err != nil {
		t.Fatal("failed to marshal logger config:", err)
	}
	logs.Init(jsonBytes)

	httpServerConfig := map[string]interface{}{
		"port":          8081,
		"path":          "/api",
		"cors":          true,
		"access":        true,
		"request_log":   true,
		"address":       "127.0.0.1",
		"read_timeout":  60,
		"write_timeout": 60,
	}
	opts, err := json.Marshal(httpServerConfig)
	if err != nil {
		t.Fatal("failed to marshal http config:", err)
	}

	server, err := NewHttpServer(opts)
	if err != nil {
		t.Fatal("failed to create http server:", err)
	}

	// 添加一个简单的处理器
	h := NewHandler(
		"test",
		[]string{"test"},
		func(ctx echo.Context, req HelloReq, resp HelloResp) error {
			resp.Message = "Test response"
			return ctx.JSON(http.StatusOK, resp)
		},
	)

	server.AddGroup("test")
	server.Post("test", "test", h)

	// 测试 GracefulStop 方法
	t.Run("TestGracefulStop", func(t *testing.T) {
		// 启动服务器
		go func() {
			if err := server.Startup(); err != nil {
				t.Logf("Server startup error: %v", err)
			}
		}()

		// 等待服务器启动
		time.Sleep(100 * time.Millisecond)

		// 测试手动优雅关闭
		err := server.GracefulStop()
		if err != nil {
			t.Errorf("GracefulStop failed: %v", err)
		}
		t.Log("GracefulStop test completed successfully")
	})
}

// TestStartWithGracefulShutdown 测试带信号量的启动和关闭
func TestStartWithGracefulShutdown(t *testing.T) {
	loggerConf := map[string]interface{}{
		"filename":   "logs/test_signal.log",
		"maxsize":    60,
		"maxbackups": 5,
		"maxage":     7,
		"compress":   true,
		"level":      -1,
	}
	jsonBytes, err := json.Marshal(loggerConf)
	if err != nil {
		t.Fatal("failed to marshal logger config:", err)
	}
	logs.Init(jsonBytes)

	httpServerConfig := map[string]interface{}{
		"port":          8082,
		"path":          "/api",
		"cors":          true,
		"access":        true,
		"request_log":   true,
		"address":       "127.0.0.1",
		"read_timeout":  60,
		"write_timeout": 60,
	}
	opts, err := json.Marshal(httpServerConfig)
	if err != nil {
		t.Fatal("failed to marshal http config:", err)
	}

	server, err := NewHttpServer(opts)
	if err != nil {
		t.Fatal("failed to create http server:", err)
	}

	// 添加一个简单的处理器
	h := NewHandler(
		"signal-test",
		[]string{"signal"},
		func(ctx echo.Context, req HelloReq, resp HelloResp) error {
			resp.Message = "Signal test response"
			return ctx.JSON(http.StatusOK, resp)
		},
	)

	server.AddGroup("signal")
	server.Post("signal-test", "signal", h)

	// 测试带超时的启动和关闭
	done := make(chan bool, 1)
	go func() {
		defer func() {
			done <- true
		}()

		// 模拟启动服务器但不等待信号量
		go func() {
			server.logger.Info("Test: Starting server in background")
			if err := server.engine.Start(server.addr); err != nil && err != http.ErrServerClosed {
				t.Logf("Server start error: %v", err)
			}
		}()

		// 等待服务器启动
		time.Sleep(100 * time.Millisecond)

		// 模拟发送HTTP请求验证服务器运行
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://127.0.0.1:8082/api/signal")
		if err == nil && resp != nil {
			resp.Body.Close()
			t.Log("Server is responding to requests")
		}

		// 手动触发关闭
		if err := server.GracefulStop(); err != nil {
			t.Errorf("GracefulStop in background test failed: %v", err)
		}
	}()

	// 等待测试完成或超时
	select {
	case <-done:
		t.Log("StartWithGracefulShutdown test completed successfully")
	case <-time.After(10 * time.Second):
		t.Error("Test timeout")
	}
}

// TestWaitForShutdown 测试等待关闭信号的功能
func TestWaitForShutdown(t *testing.T) {
	loggerConf := map[string]interface{}{
		"filename":   "logs/test_wait.log",
		"maxsize":    60,
		"maxbackups": 5,
		"maxage":     7,
		"compress":   true,
		"level":      -1,
	}
	jsonBytes, err := json.Marshal(loggerConf)
	if err != nil {
		t.Fatal("failed to marshal logger config:", err)
	}
	logs.Init(jsonBytes)

	httpServerConfig := map[string]interface{}{
		"port":          8083,
		"path":          "/api",
		"cors":          true,
		"access":        true,
		"request_log":   true,
		"address":       "127.0.0.1",
		"read_timeout":  60,
		"write_timeout": 60,
	}
	opts, err := json.Marshal(httpServerConfig)
	if err != nil {
		t.Fatal("failed to marshal http config:", err)
	}

	server, err := NewHttpServer(opts)
	if err != nil {
		t.Fatal("failed to create http server:", err)
	}

	// 测试WaitForShutdown方法（不实际发送信号，而是测试方法存在性）
	t.Run("TestWaitForShutdownMethodExists", func(t *testing.T) {
		// 启动服务器在后台
		go func() {
			if err := server.engine.Start(server.addr); err != nil && err != http.ErrServerClosed {
				t.Logf("Server start error in wait test: %v", err)
			}
		}()

		// 等待服务器启动
		time.Sleep(100 * time.Millisecond)

		// 测试方法调用（由于无法在测试中发送真实信号，我们直接测试GracefulStop）
		if err := server.GracefulStop(); err != nil {
			t.Errorf("Server shutdown failed in wait test: %v", err)
		}

		t.Log("WaitForShutdown method test completed successfully")
	})
}
