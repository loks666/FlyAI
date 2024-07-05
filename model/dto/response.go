package dto

// Response
// @description: 基础返回结构体
type Response[T any] struct {
	Code int    `json:"code"` // 状态码
	Data T      `json:"data"` // 数据
	Msg  string `json:"msg"`  // 消息
}
