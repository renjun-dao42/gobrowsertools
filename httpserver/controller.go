package httpserver

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testbrowser/httpserver/model"
	"testbrowser/pkg/browser"
	"testbrowser/pkg/errors"
	"testbrowser/pkg/response"
	"testbrowser/pkg/xgin"
)

type APIController struct {
	manager *browser.BrowserManager
}

func NewController() *APIController {
	return &APIController{
		manager: browser.NewBrowserManager(),
	}
}

func (a *APIController) Screenshot(c *gin.Context) {
	var req model.RequestScreenshot

	xgin.MustBindContext(c, &req)

	b, err := a.manager.GetOrCreateBrowser()
	errors.Check(err, "get browser error")

	bytes, err := b.Screenshot()
	errors.Check(err, "screenshot error")

	if bytes == nil {
		errors.Throw(errors.InternalError)
	}

	data := Base64Encode(bytes)

	c.JSON(http.StatusOK, response.New(model.ResponseScreenshot{ImageType: "png", Data: data}))
}

func (a *APIController) OpenTab(c *gin.Context) {
	var req model.RequestBrowserOpenTab
	xgin.MustBindContext(c, &req)

	b, err := a.manager.GetOrCreateBrowser()
	errors.Check(err, "get browser error")

	err = b.OpenTab(c, req.Url)
	errors.Check(err, "open browser error")

	c.JSON(http.StatusOK, response.New(nil))
}

func (a *APIController) GetConsoleLogs(c *gin.Context) {
	b, err := a.manager.GetOrCreateBrowser()
	errors.Check(err, "get browser error")

	logs := b.GetLogs()
	resp := model.ResponseList{Total: int64(len(logs)), List: logs}

	c.JSON(http.StatusOK, response.New(resp))
}
