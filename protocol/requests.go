package protocol

type BaseRequest struct {
}

type BasePageRequest struct {
	Page     int    `json:"page"`      //当前面
	PageSize int    `json:"page_size"` // 一页数量
	Sort     string `json:"sort"`      //排序
}
