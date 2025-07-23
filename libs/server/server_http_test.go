package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/libs/logs"
	"github.com/stardustagi/TopLib/libs/option"
	"go.uber.org/zap/zapcore"
)

type HelloReq struct {
	Name string `json:"name" validate:"required"`
}

type HelloResp struct {
	Message string `json:"message"`
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
	level := loggerConf["level"]
	jsonBytes, err := json.Marshal(loggerConf)
	if err != nil {
		// 处理错误
	}
	logs.Init(jsonBytes, zapcore.Level(level.(int)))
	opts := &option.Options{
		Http: option.Http{
			Port: 8080,
			Path: "/",
		},
	}

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
	bk.AddGroup("test")
	bk.AddPostHandler("test", h)
	bk.Start()
}
