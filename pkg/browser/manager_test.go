package browser

import (
	"context"
	"fmt"
	"testing"
)

func TestBrowserManager_OpenTab(t *testing.T) {
	manager := NewBrowserManager()
	
	if !manager.IsInstalled() {
		fmt.Println("Browser is not installed, skipping test")
	}
	
	browser, err := manager.GetOrCreateBrowser()
	if err != nil {
		t.Fatalf("GetOrCreateBrowser() error = %v", err)
	}
	
	ctx := context.Background()
	err = browser.OpenTab(ctx, "https://www.baidu.com")
	if err != nil {
		t.Errorf("OpenTab() error = %v", err)
	}
	
	// 清理资源
	browser.Close()
}
