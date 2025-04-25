package main

import (
	"github.com/playwright-community/playwright-go"
	"log"
)

const (
	BROWSER_DATA_DIR = "browser_data"
	DATA_DIR         = "data"
)

var (
	isLoggedIn bool
	browser    *playwright.Browser
	context    playwright.BrowserContext
	page       *playwright.Page
)

// 初始化浏览器上下文
func initBrowser() {
	pw, err := playwright.Run()
	if err != nil {
		panic(err)
	}
	defer pw.Stop() // 关闭所有浏览器实例

	// 启动 Chromium 浏览器
	browser, err := pw.Chromium.Launch()
	if err != nil {
		panic(err)
	}
	defer browser.Close()

	// 打开新页面并访问 URL
	page, err := browser.NewPage()
	if err != nil {
		panic(err)
	}
	defer page.Close()

	contextOptions := playwright.BrowserNewContextOptions{
		UserAgent:  playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"),
		Locale:     playwright.String("zh-CN"),
		TimezoneId: playwright.String("Asia/Shanghai"),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		RecordVideo: &playwright.RecordVideo{
			Dir: *playwright.String(DATA_DIR),
			Size: &playwright.Size{
				Width:  1920,
				Height: 1080,
			},
		},
		//RecordVideoDir: playwright.String(BROWSER_DATA_DIR),
		//RecordVideoSize: &playwright.ViewportSize{
		//	Width:  1920,
		//	Height: 1080,
		//},
		StorageStatePath: playwright.String(BROWSER_DATA_DIR + "/storage_state.json"),
	}
	context, err = browser.NewContext(contextOptions)
	if err != nil {
		log.Fatal(err)
	}
	page.SetDefaultTimeout(60000)
}

//// 检查登录状态
//func ensureLoggedIn() bool {
//	if isLoggedIn {
//		return true
//	}
//	loginElements, _ := page.QuerySelectorAll("text='登录'")
//	if len(loginElements) == 0 {
//		isLoggedIn = true
//	}
//	return isLoggedIn
//}
//
//// 登录工具
//func loginTool() (string, error) {
//	if ensureLoggedIn() {
//		return "已登录小红书账号", nil
//	}
//
//	// 访问登录页面
//	err := page.Goto("https://www.xiaohongshu.com", playwright.PageGotoOptions{Timeout: playwright.Float64(60000)})
//	if err != nil {
//		return "", err
//	}
//
//	loginElements, _ := page.QuerySelectorAll("text='登录'")
//	if len(loginElements) == 0 {
//		isLoggedIn = true
//		return "登录成功！", nil
//	}
//
//	// 点击登录按钮
//	err = loginElements[0].Click()
//	if err != nil {
//		return "", err
//	}
//
//	// 等待用户登录
//	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
//	defer cancel()
//	for {
//		select {
//		case <-ctx.Done():
//			return "登录超时", nil
//		default:
//			time.Sleep(5 * time.Second)
//			loginCheck, _ := page.QuerySelectorAll("text='登录'")
//			if len(loginCheck) == 0 {
//				isLoggedIn = true
//				return "登录成功！", nil
//			}
//		}
//	}
//}
//
//// 搜索笔记工具
//func searchNotesTool(keywords string, limit int) ([]map[string]string, error) {
//	if !ensureLoggedIn() {
//		return nil, fmt.Errorf("请先登录")
//	}
//
//	searchURL := fmt.Sprintf("https://www.xiaohongshu.com/search_result?keyword=%s", keywords)
//	err := page.Goto(searchURL, playwright.PageGotoOptions{Timeout: playwright.Float64(60000)})
//	if err != nil {
//		return nil, err
//	}
//
//	postCards, _ := page.QuerySelectorAll("section.note-item")
//	if len(postCards) == 0 {
//		postCards, _ = page.QuerySelectorAll("div[data-v-a264b01a]")
//	}
//
//	posts := make([]map[string]string, 0)
//	seenURLs := make(map[string]bool)
//
//	for _, card := range postCards[:limit] {
//		linkEl, _ := card.QuerySelector("a[href*='/search_result/']")
//		if linkEl == nil {
//			continue
//		}
//		href, _ := linkEl.GetAttribute("href")
//		fullURL := "https://www.xiaohongshu.com" + href
//		if seenURLs[fullURL] {
//			continue
//		}
//
//		title := "未知标题"
//		titleEl, _ := card.QuerySelector("div.footer a.title span")
//		if titleEl == nil {
//			titleEl, _ = card.QuerySelector("a.title span")
//		}
//		if titleEl != nil {
//			title, _ = titleEl.TextContent()
//		}
//
//		posts = append(posts, map[string]string{
//			"url":   fullURL,
//			"title": title,
//		})
//		seenURLs[fullURL] = true
//	}
//
//	return posts, nil
//}
//
//// 获取笔记内容工具
//func getNoteContentTool(url string) (map[string]string, error) {
//	if !ensureLoggedIn() {
//		return nil, fmt.Errorf("请先登录")
//	}
//
//	err := page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float64(60000)})
//	if err != nil {
//		return nil, err
//	}
//
//	title := extractTitle(page)
//	author := extractAuthor(page)
//	timeStr := extractTime(page)
//	content := extractContent(page)
//
//	return map[string]string{
//		"title":       title,
//		"author":      author,
//		"publishTime": timeStr,
//		"content":     content,
//		"url":         url,
//	}, nil
//}
//
//// 提取标题
//func extractTitle(page *playwright.Page) string {
//	// 实现与Python版相同的提取逻辑
//}
//
//// 提取作者
//func extractAuthor(page *playwright.Page) string {
//	// 实现作者提取逻辑
//}
//
//// 提取发布时间
//func extractTime(page *playwright.Page) string {
//	// 实现时间提取逻辑
//}
//
//// 提取内容
//func extractContent(page *playwright.Page) string {
//	// 实现内容提取逻辑
//}
//
//func main() {
//	// 初始化目录
//	os.MkdirAll(BROWSER_DATA_DIR, os.ModePerm)
//	os.MkdirAll(DATA_DIR, os.ModePerm)
//
//	initBrowser()
//
//	// 创建MCP服务器
//	server := mcp.NewServer()
//
//	// 注册工具
//	server.RegisterTool("login", loginTool)
//	server.RegisterTool("search_notes", searchNotesTool)
//	server.RegisterTool("get_note_content", getNoteContentTool)
//
//	// 运行MCP服务器
//	log.Fatal(server.ListenAndServe())
//}
