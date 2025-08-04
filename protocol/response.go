package protocol

import (
	"github.com/labstack/echo/v4"
	"github.com/stardustagi/TopLib/libs/errors"
)

// 返回定义
type BaseResponse struct {
	ErrCode int         `json:"errcode"`
	ErrMsg  string      `json:"errmsg,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type BasePageResponse struct {
	ErrCode int         `json:"errcode"`
	ErrMsg  string      `json:"errmsg,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Page    PageResp    `json:"page"`
}

type PageResp struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort"`
	Total    int64  `json:"total"`
}

func Response(c echo.Context, err *errors.StackError, data any) error {
	if err == nil {
		if data != nil {
			return c.JSON(200, BaseResponse{
				ErrCode: 0,
				ErrMsg:  "操作成功!",
				Data:    data,
			})
		}
		return c.JSON(200, BaseResponse{
			ErrCode: 0,
			ErrMsg:  "操作成功!",
		})
	}

	return c.JSON(200, BaseResponse{
		ErrCode: err.Code(),
		ErrMsg:  err.Msg(),
		Data:    nil,
	})
}
