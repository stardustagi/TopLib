package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/utils"
)

// var validate *validator.Validate

// func init() {
// 	validate = validator.New()
// 	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
// 		jsonTag := field.Tag.Get("json")
// 		if jsonTag != "" {
// 			return jsonTag
// 		}
// 		return field.Name
// 	})
// }

type Context struct {
	echo.Context
	RemoteAddr string
	ClientId   string
	Header     http.Header
}

func NewContext(c echo.Context) *Context {
	ctx := &Context{
		Context:    c,
		RemoteAddr: utils.GetRemoteAddr(c.Request()),
		ClientId:   c.Request().Header.Get(ClientIDKey),
		Header:     c.Request().Header,
	}

	return ctx
}
