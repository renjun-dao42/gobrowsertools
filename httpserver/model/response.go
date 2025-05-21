package model

type ResponseList struct {
	Total int64       `json:"total"`
	List  interface{} `json:"list"`
}

type ResponseScreenshot struct {
	ImageType string `json:"image_type"`
	Data      string `json:"data"`
}
