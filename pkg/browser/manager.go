package browser

import (
	"browsertools/log"
	"browsertools/pkg/errors"
	"os"
	"os/exec"
	"sync"

	"github.com/playwright-community/playwright-go"
)

type BrowserManager struct {
	path           string
	browserHandler *BrowserHandler
	mutex          *sync.Mutex
}

func NewBrowserManager() *BrowserManager {
	return &BrowserManager{path: getBrowserPath(), browserHandler: nil, mutex: &sync.Mutex{}}
}

func getBrowserPath() string {
	path := "/usr/chrome-linux64/chrome"
	_, err := os.Lstat(path)
	if err == nil {
		return path
	}

	path, err = exec.LookPath("google-chrome")
	if err == nil {
		return path
	}

	path, err = exec.LookPath("chromium-browser")
	if err == nil {
		return path
	}

	return ""
}

func (m *BrowserManager) IsInstalled() bool {
	return m.path != ""
}

func (m *BrowserManager) GetOrCreateBrowser() (*BrowserHandler, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.browserHandler != nil && !m.browserHandler.IsClosed() {
		return m.browserHandler, nil
	}

	b, err := m.create()
	if err != nil {
		return nil, errors.WithMessage(err, "could not create browser")
	}

	m.browserHandler = b

	return b, nil
}

func (m *BrowserManager) create() (*BrowserHandler, error) {
	if !m.IsInstalled() {
		return nil, errors.BrowserNotInstalled
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, errors.WithMessage(err, "could not run playwright")
	}

	browser, err := pw.Chromium.ConnectOverCDP("http://localhost:29229")
	if err != nil {
		log.Infof("could not connect to Chromium: %v. we will start chrome server", err)

		opt := playwright.BrowserTypeLaunchOptions{
			ExecutablePath: &m.path,
			Args: []string{
				// 窗口管理
				// 启动时最大化窗口
				"--start-maximized",
				"--process-per-site", "--disable-field-trial-config", "--disable-background-networking",
				"--enable-features=NetworkService,NetworkServiceInProcess", "--disable-background-timer-throttling", "--disable-backgrounding-occluded-windows",
				"--disable-back-forward-cache", "--disable-breakpad", "--disable-client-side-phishing-detection", "--disable-component-extensions-with-background-pages",
				"--disable-component-update", "--no-default-browser-check", "--disable-default-apps", "--disable-dev-shm-usage",
				"--disable-features=ImprovedCookieControls,LazyFrameLoading,GlobalMediaControls,DestroyProfileOnBrowserClose,MediaRouter,DialMediaRouteProvider,AcceptCHFrame,AutoExpandDetailsElement,CertificateTransparencyComponentUpdater,AvoidUnnecessaryBeforeUnloadCheckSync,Translate,HttpsUpgrades,PaintHolding",
				"--allow-pre-commit-input", "--disable-hang-monitor", "--disable-ipc-flooding-protection", "--disable-popup-blocking", "--disable-prompt-on-repost",
				"--disable-renderer-backgrounding", "--force-color-profile=srgb", "--metrics-recording-only", "--no-first-run", "--enable-automation", "--disable-infobars",
				"--password-store=basic", "--use-mock-keychain", "--no-service-autorun", "--export-tagged-pdf", "--disable-search-engine-choice-screen", "--mute-audio",
				"--blink-settings=primaryHoverType=2,availableHoverTypes=2,primaryPointerType=4,availablePointerTypes=4", "--no-sandbox", "--disable-blink-features=AutomationControlled",
				"--use-angle=swiftshader-webgl", "--noerrdialogs", "--disable-gpu",
				//"--remote-debugging-port=29229",
			},
			Timeout:  playwright.Float(30000), // 设置超时时间（毫秒）
			Headless: playwright.Bool(false),  // 禁用无头模式
		}

		browser, err = pw.Chromium.Launch(opt)
		if err != nil {
			return nil, errors.WithMessage(err, "could not launch browser")
		}
	}

	log.Infof("launched browser successfully!")

	return NewBrowserHandler(browser)
}
