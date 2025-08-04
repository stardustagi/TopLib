package server

import (
	"encoding/json"
	"net/http"
	"testing"

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
}
