package common

type PageQuery struct {
	Page     int `json:"page" example:"1"`
	PageSize int `json:"page_size" example:"20"`
}

type PageResponse struct {
	Total int `json:"total" example:"100"`
}
