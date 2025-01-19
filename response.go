package main

type Response struct {
	SendTime   string `json:"sendTime"`
	CreateTime string `json:"createTime"`
	Data       []byte `json:"data"`
}
