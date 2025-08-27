package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/stardustagi/TopLib/libs/logs"

	"github.com/labstack/echo/v4"
)

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Request Response 记录请求日志
func Request() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 取得request body
			requestBody := ""
			b, err := io.ReadAll(c.Request().Body)
			if err != nil {
				requestBody = "failed to request body"
			} else {
				// requestBody = string(b)
				var jsonData interface{}
				if formatted, err := json.MarshalIndent(jsonData, "", "  "); err != nil {
					requestBody = string(b)
				} else {
					requestBody = string(formatted)
				}
				c.Request().Body = io.NopCloser(bytes.NewBuffer(b))
			}
			host := c.Request().Host
			uri := c.Request().RequestURI
			method := c.Request().Method
			agent := c.Request().UserAgent()

			// 取得 response body
			respBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, respBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
			c.Response().Writer = writer
			status := c.Response().Status
			ip := echo.ExtractIPFromXFFHeader()(c.Request())
			logs.Info(fmt.Sprintf("request :'%s %s' %d %s '-' '%s' %s '%s'", method, uri, status, ip, agent, host, requestBody))
			next(c)
			logs.Infof(fmt.Sprintf("response: '%s'", respBody.String()))
			return nil
		}
	}
}
