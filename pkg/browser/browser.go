package browser

import (
	"browsertools/log"
	"browsertools/pkg/errors"
	"context"
	"sync/atomic"

	"github.com/playwright-community/playwright-go"
)

const maxLogs = 1000

type BrowserHandler struct {
	browser        playwright.Browser
	browserContext playwright.BrowserContext
	pageList       *PageList
	isClosed       atomic.Bool
}

func NewBrowserHandler(browser playwright.Browser) (*BrowserHandler, error) {
	ctx, err := getBrowserContext(browser)
	if err != nil {
		return nil, errors.WithMessage(err, "new context error")
	}

	handler := &BrowserHandler{
		browser:        browser,
		browserContext: ctx,
		pageList:       NewPageList(),
	}

	handler.intExistPageFromContext()

	// 设置浏览器关闭事件监听
	browser.OnDisconnected(handler.OnDisconnected)

	// 设置新页面(标签页)事件监听
	ctx.OnPage(handler.onPage)

	return handler, nil
}

func (h *BrowserHandler) intExistPageFromContext() {
	pages := h.browserContext.Pages()
	for _, page := range pages {
		handler := NewPageHandler(page, h)
		h.pageList.AddPage(handler)
		h.pageList.SetActivePage(handler.GetPageID())
	}
}

func getBrowserContext(browser playwright.Browser) (playwright.BrowserContext, error) {
	contexts := browser.Contexts()
	if len(contexts) > 0 {
		return contexts[0], nil
	}

	opt := playwright.BrowserNewContextOptions{
		NoViewport: playwright.Bool(true),
	}

	return browser.NewContext(opt)

}

func (h *BrowserHandler) OnDisconnected(_ playwright.Browser) {
	log.Infof("Browser disconnected")
	h.isClosed.Store(true)
}

func (h *BrowserHandler) onPage(page playwright.Page) {
	log.Infof("get on page event")

	h.createIfNotExistPageHandler(page)
}

func (h *BrowserHandler) createPage() (*PageHandler, error) {
	page, err := h.browserContext.NewPage()
	if err != nil {
		return nil, errors.WithMessage(err, "new page error")
	}

	log.Infof("create new page successfully!")

	// 可能事件已经推送了, 如果推送了就不需要在创建页面的handler
	handler := h.createIfNotExistPageHandler(page)

	return handler, nil
}
func (h *BrowserHandler) createIfNotExistPageHandler(page playwright.Page) *PageHandler {
	handler := h.pageList.FindPageHandler(page)
	if handler != nil {
		return handler
	}

	handler = NewPageHandler(page, h)
	h.pageList.AddPage(handler)
	h.pageList.SetActivePage(handler.GetPageID())

	log.Infof("New page %s opened: %s", handler.GetPageID(), page.URL())

	return handler
}
func (h *BrowserHandler) IsClosed() bool {
	return h.isClosed.Load()
}

func (h *BrowserHandler) OpenTab(ctx context.Context, url string) error {
	page, err := h.getEmptyPage()
	if err != nil {
		return errors.WithMessage(err, "get page error")
	}

	return page.Goto(ctx, url)
}

func (h *BrowserHandler) getEmptyPage() (*PageHandler, error) {
	pages := h.browserContext.Pages()
	if len(pages) == 1 {
		page := pages[0]
		if page.URL() == "chrome://new-tab-page/" {
			handler := h.pageList.FindPageHandler(page)
			if handler != nil {
				return handler, nil
			}
			// 不需要做操作，不可能发生
		}
	}

	return h.createPage()
}

func (h *BrowserHandler) Close() {
	h.pageList.CloseAll()

	err := h.browserContext.Close()
	if err != nil {
		log.Errorf("close browser context error: %v", err)
	}

	err = h.browser.Close()
	if err != nil {
		log.Errorf("close browser error: %v", err)
	}
}

func (h *BrowserHandler) OnActivePage(pageID string) {
	if h.pageList.SetActivePage(pageID) {
		log.Infof("switched to page %s", pageID)
	}
}

func (h *BrowserHandler) OnClosePage(pageID string) {
	log.Infof("remove page %s from browser", pageID)
	h.pageList.RemovePage(pageID)
}

func (h *BrowserHandler) Screenshot() ([]byte, error) {
	page := h.GetActiveTab()
	if page == nil {
		return nil, errors.ErrCurrentPageEmpty
	}

	return page.Screenshot()
}

func (h *BrowserHandler) GetLogs() []string {
	page := h.GetActiveTab()
	if page == nil {
		return []string{}
	}

	return page.GetLogs()
}

func (h *BrowserHandler) GetActiveTab() *PageHandler {
	if h.isClosed.Load() {
		return nil
	}

	return h.pageList.GetActivePage()
}
