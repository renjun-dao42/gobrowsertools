package model

type RequestScreenshot struct {
	ImageType   string `json:"image_type" validate:"required"`
	ImageFormat string `json:"image_format" validate:"required"`
}

type RequestBrowserOpenTab struct {
	Url string `json:"url" validate:"required"`
}
