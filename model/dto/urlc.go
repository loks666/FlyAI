package dto

// ShortUrlResponse
// @description: 短链接返回结构体
type ShortUrlResponse struct {
	Error int    `json:"error"` // 错误码，1表示发生错误，0则正常
	Short string `json:"short"` // 短链接
}
