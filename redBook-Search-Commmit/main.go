package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
	"github.com/playwright-community/playwright-go"
)

// ====================== 常量定义 ======================
const (
	browserDataDir = "browser_data"
	dataDir        = "data"
	tsLayout       = "20060102_150405"
)

// ====================== 核心结构体 ======================
type RedBookEngine struct {
	IsLoggedIn     bool
	Page           playwright.Page
	BrowserContext playwright.BrowserContext
	PW             *playwright.Playwright
	BrowserDataDir string
	DataDir        string
}

// ====================== 初始化方法 ======================
func NewRedBookEngine() (*RedBookEngine, error) {
	r := &RedBookEngine{}

	currDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	r.BrowserDataDir = currDir
	r.DataDir = currDir

	os.MkdirAll(browserDataDir, 0755)
	os.MkdirAll(dataDir, 0755)

	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	r.PW = pw

	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		DownloadsPath: playwright.String(r.DataDir),
		Timeout:       playwright.Float(60000),
		Headless:      playwright.Bool(false),
		Viewport: &playwright.Size{
			Width:  1280,
			Height: 800,
		},
	}

	browserContext, err := pw.Chromium.LaunchPersistentContext(r.BrowserDataDir, opts)
	if err != nil {
		return nil, err
	}
	r.BrowserContext = browserContext

	page, err := r.BrowserContext.NewPage()
	if err != nil {
		return nil, err
	}
	page.SetDefaultTimeout(60000)
	r.Page = page

	return r, nil
}

// ====================== 浏览器管理 ======================
func (r *RedBookEngine) EnsureBrowser() (bool, error) {
	if r.BrowserContext == nil || r.Page == nil {
		if err := r.reInitBrowser(); err != nil {
			return false, err
		}
	}

	if !r.IsLoggedIn {
		resp, err := r.Page.Goto("https://www.xiaohongshu.com", playwright.PageGotoOptions{Timeout: playwright.Float(60000)})
		if err != nil {
			return false, err
		}
		log.Printf("页面状态码: %d", resp.Status())

		time.Sleep(3 * time.Second)

		loginElements, err := r.Page.QuerySelectorAll("text=登录")
		if err != nil {
			return false, err
		}
		count := len(loginElements)
		if count > 0 {
			return false, nil // 需要登录
		} else {
			r.IsLoggedIn = true
			return true, nil // 已登录
		}
	}
	return true, nil
}

func (r *RedBookEngine) reInitBrowser() error {
	if r.BrowserContext != nil {
		_ = r.BrowserContext.Close()
	}
	if r.PW == nil {
		pw, err := playwright.Run()
		if err != nil {
			return err
		}
		r.PW = pw
	}

	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		DownloadsPath: playwright.String(r.DataDir),
		Timeout:       playwright.Float(60000),
		Headless:      playwright.Bool(false),
		Viewport: &playwright.Size{
			Width:  1280,
			Height: 800,
		},
	}

	browserContext, err := r.PW.Chromium.LaunchPersistentContext(r.BrowserDataDir, opts)
	if err != nil {
		return err
	}
	r.BrowserContext = browserContext

	page, err := r.BrowserContext.NewPage()
	if err != nil {
		return err
	}
	page.SetDefaultTimeout(60000)
	r.Page = page

	return nil
}

// ====================== 登录功能 ======================
func (r *RedBookEngine) Login() (string, error) {
	ok, err := r.EnsureBrowser()
	if err != nil {
		return "", err
	}
	if ok {
		return "已登录小红书账号", nil
	}

	_, err = r.Page.Goto("https://www.xiaohongshu.com", playwright.PageGotoOptions{Timeout: playwright.Float(60000)})
	if err != nil {
		return "", err
	}
	time.Sleep(3 * time.Second)

	loginElements, err := r.Page.QuerySelectorAll("text=登录")
	if err != nil {
		return "", err
	}
	if len(loginElements) > 0 {
		if err := loginElements[0].Click(); err != nil {
			return "", err
		}

		log.Println("请在打开的浏览器窗口中完成登录操作。登录成功后，系统将自动继续。")

		maxWait := 180 * time.Second
		interval := 5 * time.Second
		timeout := time.After(maxWait)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stillLogin, _ := r.Page.QuerySelectorAll("text=登录")
				if len(stillLogin) == 0 {
					r.IsLoggedIn = true
					time.Sleep(2 * time.Second)
					return "登录成功！", nil
				}
			case <-timeout:
				return "登录等待超时。请重试或手动登录后再使用其他功能。", nil
			}
		}
	} else {
		r.IsLoggedIn = true
		return "已登录小红书账号", nil
	}
}

// ====================== 搜索笔记 ======================
func (r *RedBookEngine) SearchNotes(keywords string, limit int) (string, error) {
	ok, err := r.EnsureBrowser()
	if err != nil {
		return "", err
	}
	if !ok {
		return "请先登录小红书账号", nil
	}

	searchURL := fmt.Sprintf("https://www.xiaohongshu.com/search_result?keyword=%s", keywords)

	resp, err := r.Page.Goto(searchURL, playwright.PageGotoOptions{Timeout: playwright.Float(60000)})
	if err != nil {
		return "", err
	}
	log.Printf("搜索页面状态码: %d", resp.Status())
	time.Sleep(5 * time.Second)

	pageHTML, err := r.Page.InnerHTML("body")
	if err != nil {
		return "", err
	}
	log.Printf("页面HTML片段: %s...", pageHTML[:500])

	log.Println("尝试获取帖子卡片...")
	postCards, err := r.Page.QuerySelectorAll("section.note-item")
	if err != nil {
		return "", err
	}
	count := len(postCards)
	log.Printf("找到 %d 个帖子卡片", count)

	if count == 0 {
		postCards, err = r.Page.QuerySelectorAll("div[data-v-a264b01a]")
		if err != nil {
			return "", err
		}
		count = len(postCards)
		log.Printf("使用备用选择器找到 %d 个帖子卡片", count)
	}

	var posts []map[string]string
	seenURLs := make(map[string]bool)

	for i := 0; i < count && len(posts) < limit; i++ {
		card := postCards[i]

		linkEl, err := card.QuerySelector("a[href*=\"/search_result/\"]")
		if err != nil || linkEl == nil {
			continue
		}

		href, err := linkEl.GetProperty("href")
		if err != nil {
			continue
		}

		hrefStr := href.String()
		if !strings.Contains(hrefStr, "/search_result/") {
			continue
		}

		fullURL := "https://www.xiaohongshu.com" + hrefStr
		if seenURLs[fullURL] {
			continue
		}
		seenURLs[fullURL] = true

		title := "未知标题"
		titleEl, _ := card.QuerySelector("div.footer a.title span")
		if titleEl != nil {
			title, _ = titleEl.TextContent()
		} else {
			titleEl, _ := card.QuerySelector("a.title span")
			if titleEl != nil {
				title, _ = titleEl.TextContent()
			}
		}

		posts = append(posts, map[string]string{
			"url":   fullURL,
			"title": strings.TrimSpace(title),
		})
	}

	if len(posts) > 0 {
		result := "搜索结果：\n"
		for i, post := range posts {
			result += fmt.Sprintf("%d. %s\n   链接: %s\n", i+1, post["title"], post["url"])
		}
		return result, nil
	} else {
		return fmt.Sprintf("未找到与\"%s\"相关的笔记", keywords), nil
	}
}

// ====================== 获取笔记内容 ======================
func (r *RedBookEngine) GetNoteContent(url string) (string, error) {
	if ok, err := r.EnsureBrowser(); !ok || err != nil {
		return "", fmt.Errorf("请先登录小红书账号")
	}

	if _, err := r.Page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(60000)}); err != nil {
		return "", err
	}
	time.Sleep(10 * time.Second)

	// 滚动加载内容
	r.Page.Evaluate(`() => {
        window.scrollTo(0, document.body.scrollHeight);
        setTimeout(() => window.scrollTo(0, 0), 2000);
    }`)

	// 提取内容逻辑
	contentMap := map[string]string{
		"标题":   "未知标题",
		"作者":   "未知作者",
		"发布时间": "未知",
		"内容":   "未能获取内容",
	}

	// 标题提取
	titleEl, _ := r.Page.QuerySelector("#detail-title .note-text")
	if titleEl == nil {
		titleEl, _ = r.Page.QuerySelector("div.title")
	}
	if titleEl != nil {
		title, _ := titleEl.TextContent()
		contentMap["标题"] = strings.TrimSpace(title)
	}

	// 作者提取
	authorEl, _ := r.Page.QuerySelector("span.username")
	if authorEl == nil {
		authorEl, _ = r.Page.QuerySelector("a.name")
	}
	if authorEl != nil {
		author, _ := authorEl.TextContent()
		contentMap["作者"] = strings.TrimSpace(author)
	}

	// 内容提取
	contentEl, _ := r.Page.QuerySelector("#detail-desc .note-text")
	if contentEl == nil {
		contentEl, _ = r.Page.QuerySelector(".note-content")
	}
	if contentEl != nil {
		content, _ := contentEl.TextContent()
		contentMap["内容"] = strings.TrimSpace(content)
	}

	result := fmt.Sprintf(
		"标题: %s\n作者: %s\n发布时间: %s\n链接: %s\n内容:\n%s",
		contentMap["标题"], contentMap["作者"], contentMap["发布时间"], url, contentMap["内容"],
	)
	return result, nil
}

// ====================== 发布评论 ======================
func (r *RedBookEngine) PostComment(url string, comment string) (string, error) {
	if ok, err := r.EnsureBrowser(); !ok || err != nil {
		return "", fmt.Errorf("请先登录小红书账号")
	}

	if _, err := r.Page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(60000)}); err != nil {
		return "", err
	}
	time.Sleep(5 * time.Second)

	inputEl, _ := r.Page.QuerySelector("div[contenteditable='true']")
	if inputEl == nil {
		inputEl, _ = r.Page.QuerySelector("text=说点什么...")
	}
	if inputEl != nil {
		if err := inputEl.Click(); err != nil {
			return "无法点击评论输入框", err
		}

		if err := r.Page.Keyboard().Type(comment); err != nil {
			return "无法输入评论内容", err
		}
		time.Sleep(1 * time.Second)

		sendBtn, _ := r.Page.QuerySelector("button:has-text('发送')")
		if sendBtn != nil {
			if err := sendBtn.Click(); err != nil {
				return "发送按钮点击失败", err
			}
		} else {
			if err := r.Page.Keyboard().Press("Enter"); err != nil {
				return "回车键发送失败", err
			}
		}
		return fmt.Sprintf("已发布评论: %s", comment), nil
	}
	return "无法找到评论输入框", nil
}

// ====================== MCP服务封装 ======================
type MCPService struct {
	Engine *RedBookEngine
}

// 工具调用通用包装器
func wrapTool[T any](fn func(context.Context, T) (string, error)) func(context.Context, *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	return func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		var args T
		if err := protocol.VerifyAndUnmarshal(req.RawArguments, &args); err != nil {
			return nil, err
		}

		content, err := fn(ctx, args)
		if err != nil {
			return &protocol.CallToolResult{
				IsError: true,
				Content: []protocol.Content{},
			}, nil
		}

		return &protocol.CallToolResult{
			Content: []protocol.Content{
				protocol.TextContent{
					Type: "text",
					Text: content,
				},
			},
		}, nil
	}
}

// 登录工具适配
func (m *MCPService) LoginTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	type loginArgs struct{}
	return wrapTool(func(ctx context.Context, args loginArgs) (string, error) {
		return m.Engine.Login()
	})(ctx, req)
}

// 搜索笔记工具适配
func (m *MCPService) SearchNotesTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	type searchArgs struct {
		Keywords string `json:"keywords"`
		Limit    int    `json:"limit"`
	}

	return wrapTool(func(ctx context.Context, args searchArgs) (string, error) {
		if args.Limit <= 0 {
			args.Limit = 5 // 默认值
		}
		return m.Engine.SearchNotes(args.Keywords, args.Limit)
	})(ctx, req)
}

// 获取笔记内容工具适配
func (m *MCPService) GetNoteContentTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	type contentArgs struct {
		URL string `json:"url"`
	}

	return wrapTool(func(ctx context.Context, args contentArgs) (string, error) {
		return m.Engine.GetNoteContent(args.URL)
	})(ctx, req)
}

// 发布评论工具适配
func (m *MCPService) PostCommentTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
	type commentArgs struct {
		URL     string `json:"url"`
		Comment string `json:"comment"`
	}

	return wrapTool(func(ctx context.Context, args commentArgs) (string, error) {
		return m.Engine.PostComment(args.URL, args.Comment)
	})(ctx, req)
}

// ====================== 主函数 ======================
func main() {
	engine, err := NewRedBookEngine()
	if err != nil {
		log.Fatalf("初始化引擎失败: %v", err)
	}
	defer engine.Close()

	service := &MCPService{Engine: engine}
	mcpServer, _ := server.NewServer(
		getTransport(),
		server.WithServerInfo(protocol.Implementation{
			Name:    "login-redBook",
			Version: "1.0.0",
		}),
	)

	// 注册登录工具
	loginTool, _ := protocol.NewTool(
		"login",
		"执行小红书账号登录操作",
		struct{}{},
		//struct {
		//	UserName string `json:"username" description:"登录账号"`
		//	Password string `json:"password" description:"登录密码"`
		//}{},
	)
	mcpServer.RegisterTool(loginTool, service.LoginTool)

	// 注册搜索工具
	searchTool, _ := protocol.NewTool(
		"search_notes",
		"搜索小红书笔记内容",
		struct {
			Keywords string `json:"keywords" description:"要搜索的关键词"`
			Limit    int    `json:"limit" description:"返回结果数量限制"`
		}{},
	)
	mcpServer.RegisterTool(searchTool, service.SearchNotesTool)

	// 注册获取笔记内容工具
	getContentTool, _ := protocol.NewTool(
		"get_note_content",
		"获取指定链接的小红书笔记内容",
		struct {
			URL string `json:"url" description:"要获取内容的笔记URL"`
		}{},
	)
	mcpServer.RegisterTool(getContentTool, service.GetNoteContentTool)

	// 注册发布评论工具
	postCommentTool, _ := protocol.NewTool(
		"post_comment",
		"在指定笔记下发布评论",
		struct {
			URL     string `json:"url" description:"目标笔记URL"`
			Comment string `json:"comment" description:"要发布的评论内容"`
		}{},
	)
	mcpServer.RegisterTool(postCommentTool, service.PostCommentTool)

	log.Println("🚀 启动小红书MCP服务器... http://localhost:8080")
	if err := mcpServer.Run(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}

func getTransport() (t transport.ServerTransport) {
	var (
		mode string
		addr = "127.0.0.1:8080"
	)

	flag.StringVar(&mode, "transport", "sse", "The transport to use, should be \"stdio\" or \"sse\" or \"streamable_http\"")
	flag.Parse()

	switch mode {
	case "stdio":
		log.Println("start current time mcp server with stdio transport")
		t = transport.NewStdioServerTransport()
	case "sse":
		log.Printf("start current time mcp server with sse transport, listen %s", addr)
		t, _ = transport.NewSSEServerTransport(addr)
	case "streamable_http":
		log.Printf("start current time mcp server with streamable_http transport, listen %s", addr)
		t = transport.NewStreamableHTTPServerTransport(addr)
	default:
		panic(fmt.Errorf("unknown mode: %s", mode))
	}

	return t
}

// ====================== 资源清理 ======================
func (r *RedBookEngine) Close() {
	if r.BrowserContext != nil {
		r.BrowserContext.Close()
	}
	if r.PW != nil {
		r.PW.Stop()
	}
}
