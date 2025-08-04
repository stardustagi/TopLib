package server

import (
	"github.com/labstack/echo/v4"
)

// 定义handler参数的结构
type HandlerParams struct {
	Name    string
	Tags    []string
	Handler interface{}
}

type Handler[Req any, Resp any] struct {
	Path string // 路径
	Name string
	Tags []string
	Func func(echo.Context, Req, Resp) error
	req  Req
	resp Resp
}

// 抽象接口
type IHandler interface {
	GetName() string
	GetTags() []string
	GetFunc() func(echo.Context) error
}

func NewHandler[Req any, Resp any](
	name string,
	tags []string,
	f func(echo.Context, Req, Resp) error,
) *Handler[Req, Resp] {
	var req Req
	var resp Resp
	return &Handler[Req, Resp]{
		Name: name,
		Tags: tags,
		Func: f,
		req:  req,
		resp: resp,
	}
}

func (h *Handler[Req, Resp]) GetName() string {
	return h.Name
}

func (h *Handler[Req, Resp]) GetTags() []string {
	return h.Tags
}

func (h *Handler[Req, Resp]) GetFunc() func(echo.Context) error {
	return func(c echo.Context) error {
		// 绑定
		if err := c.Bind(&h.req); err != nil {
			return err
		}
		// 验证
		if err := c.Validate(&h.req); err != nil {
			return err
		}
		// 执行体
		return h.Func(c, h.req, h.resp)
	}
}

// 句柄管理器抽象接口
type IHandlers interface {
	GetHandlers() []IHandler
	AddHandlers(handler IHandler)
	GetHandlersLen() int
}

// 句柄管理器
type Handlers struct {
	handlers []IHandler
}

func NewHandlers() IHandlers {
	return &Handlers{
		handlers: make([]IHandler, 0),
	}
}

func (h *Handlers) GetHandlers() []IHandler {
	return h.handlers
}

func (h *Handlers) AddHandlers(handler IHandler) {
	// This method can be overridden to add handlers
	h.handlers = append(h.handlers, handler)
}

func (h *Handlers) GetHandlersLen() int {
	return len(h.handlers)
}

type StarDustGroup struct {
	Prefix string
	Group  *echo.Group
}

func NewStarDustGroup(prefix string, group *echo.Group) *StarDustGroup {
	return &StarDustGroup{
		Prefix: prefix,
		Group:  group,
	}
}
