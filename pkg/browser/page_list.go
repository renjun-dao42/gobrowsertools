package browser

import (
	"github.com/playwright-community/playwright-go"
	"sync"
	"testbrowser/log"
	"time"
)

// PageList 管理多个页面的结构体
type PageList struct {
	pages      map[string]*PageHandler
	mux        *sync.Mutex
	activePage string
}

// NewPageList 创建一个新的PageList实例
func NewPageList() *PageList {
	return &PageList{
		pages:      make(map[string]*PageHandler),
		mux:        &sync.Mutex{},
		activePage: "",
	}
}

// GetActivePage 获取当前活动页面
func (p *PageList) GetActivePage() *PageHandler {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.activePage == "" {
		return nil
	}

	return p.pages[p.activePage]
}

// AddPage 添加页面到列表
func (p *PageList) AddPage(pageHandler *PageHandler) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.pages[pageHandler.GetPageID()] = pageHandler
}

// RemovePage 从列表中移除页面
func (p *PageList) RemovePage(pageID string) {
	if pageID == "" {
		return
	}

	p.mux.Lock()
	defer p.mux.Unlock()

	delete(p.pages, pageID)

	if p.activePage == pageID {
		p.activePage = p.getNextActivePageIDWithoutLock(pageID)
	}

	log.Infof("current active page id: %s", p.activePage)
}

func (p *PageList) getNextActivePageIDWithoutLock(removedPageID string) string {
	// TODO: 目前只取得新的

	nextPageID := ""
	var createTime time.Time

	for id, page := range p.pages {
		if page.GetCreateTime().After(createTime) {
			nextPageID = id
		}
	}

	return nextPageID
}

// SetActivePage 设置当前活动页面
func (p *PageList) SetActivePage(pageID string) bool {
	if pageID == "" {
		return false
	}

	p.mux.Lock()
	defer p.mux.Unlock()

	if _, exists := p.pages[pageID]; exists {
		p.activePage = pageID
		return true
	}

	return false
}

func (p *PageList) CloseAll() {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, pageHandler := range p.pages {
		pageHandler.Close()
	}

	p.pages = make(map[string]*PageHandler)
}

// GetPageByID 通过ID获取页面
func (p *PageList) GetPageByID(pageID string) *PageHandler {
	p.mux.Lock()
	defer p.mux.Unlock()

	return p.pages[pageID]
}

// FindPageHandler 通过Playwright页面对象查找对应的PageHandler
func (p *PageList) FindPageHandler(page playwright.Page) *PageHandler {
	p.mux.Lock()
	defer p.mux.Unlock()

	for _, pageHandler := range p.pages {
		if pageHandler.GetPage() == page {
			return pageHandler
		}
	}

	return nil
}
