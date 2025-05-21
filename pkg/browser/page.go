package browser

import (
	"browsertools/log"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

// 常量定义
const (
	// 页面可见性监听脚本
	visibilityScript = `
(() => {
  document.addEventListener('visibilitychange', () => {
    window.onVisibilityChange(document.visibilityState === 'visible');
  });
})();
`
)

type PageListener interface {
	OnClosePage(pageID string)
	OnActivePage(pageID string)
}

type PageHandler struct {
	page         playwright.Page
	pageID       string
	createTime   time.Time
	consoleLogs  []string
	mux          *sync.Mutex
	isClosed     bool
	pageListener PageListener
}

func NewPageHandler(page playwright.Page, pageListener PageListener) *PageHandler {
	id := strconv.FormatInt(time.Now().UnixMilli(), 10)

	handler := &PageHandler{
		page:         page,
		pageID:       id,
		createTime:   time.Now(),
		consoleLogs:  make([]string, 0, maxLogs),
		mux:          &sync.Mutex{},
		isClosed:     false,
		pageListener: pageListener,
	}

	// 注册页面事件
	page.On("console", handler.onConsoleMessage)
	page.On("close", handler.onClose)
	page.On("bringtofront", handler.onBringToFront)

	// 启动可见性监听
	go handler.setupVisibilityTracking()

	return handler
}

func (h *PageHandler) onConsoleMessage(msg playwright.ConsoleMessage) {
	h.mux.Lock()
	defer h.mux.Unlock()

	// 如果页面已关闭，不再处理新的日志
	if h.isClosed {
		return
	}

	if len(h.consoleLogs) >= maxLogs {
		// 移除最旧的日志
		h.consoleLogs = h.consoleLogs[1:]
	}

	h.consoleLogs = append(h.consoleLogs, msg.Text())
}

func (h *PageHandler) onBringToFront() {
	log.Infof("Page %s brought to front", h.pageID)

	if h.pageListener != nil {
		h.pageListener.OnActivePage(h.pageID)
	}
}

func (h *PageHandler) onClose() {
	log.Infof("Page %s close event received", h.pageID)

	h.mux.Lock()
	defer h.mux.Unlock()

	h.isClosed = true

	if h.pageListener != nil {
		h.pageListener.OnClosePage(h.pageID)
	}
}

func (h *PageHandler) Goto(ctx context.Context, url string) error {
	_, err := h.page.Goto(url)
	if err != nil {
		return fmt.Errorf("page %s navigation to %s failed: %w", h.pageID, url, err)
	}
	return nil
}

func (h *PageHandler) Close() {
	h.mux.Lock()
	defer h.mux.Unlock()

	if h.isClosed {
		return
	}

	h.isClosed = true

	if err := h.page.Close(); err != nil {
		log.Errorf("Failed to close page %s: %v", h.pageID, err)
	} else {
		log.Infof("Page %s closed", h.pageID)
	}
}

func (h *PageHandler) Screenshot() ([]byte, error) {
	if h.IsClosed() {
		return nil, fmt.Errorf("page %s is closed, cannot take screenshot", h.pageID)
	}

	data, err := h.page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(true),
		Type:     playwright.ScreenshotTypePng,
	})

	if err != nil {
		return nil, fmt.Errorf("screenshot failed for page %s: %w", h.pageID, err)
	}

	log.Debugf("Screenshot successful for page %s, size: %d bytes", h.pageID, len(data))
	return data, nil
}

func (h *PageHandler) GetLogs() []string {
	h.mux.Lock()
	defer h.mux.Unlock()

	if h.isClosed {
		log.Warnf("Attempted to get logs from closed page %s", h.pageID)
		return []string{}
	}

	return append([]string{}, h.consoleLogs...)
}

func (h *PageHandler) GetPageID() string {
	return h.pageID
}

func (h *PageHandler) GetPage() playwright.Page {
	return h.page
}

func (h *PageHandler) GetCreateTime() time.Time {
	return h.createTime
}

func (h *PageHandler) IsClosed() bool {
	h.mux.Lock()
	defer h.mux.Unlock()
	return h.isClosed
}

// setupVisibilityTracking 设置页面可见性追踪功能
func (h *PageHandler) setupVisibilityTracking() {
	err := h.page.ExposeFunction("onVisibilityChange", func(args ...interface{}) interface{} {
		isVisible := args[0].(bool)
		log.Infof("Page %s visibility changed: %v", h.pageID, isVisible)

		if isVisible && h.pageListener != nil {
			h.pageListener.OnActivePage(h.pageID)
		}

		return nil
	})

	if err != nil {
		log.Errorf("Failed to expose visibility change function for page %s: %v", h.pageID, err)
		return
	}

	_, err = h.page.AddScriptTag(playwright.PageAddScriptTagOptions{
		Content: playwright.String(visibilityScript),
	})

	if err != nil {
		log.Errorf("Failed to add visibility tracking script for page %s: %v", h.pageID, err)
	}
}
