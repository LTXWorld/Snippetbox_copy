package main

import (
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models/mock"
	"github.com/golangcollege/sessions"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"
)

// 定义一个表达式捕获CSRF token value from the HTML for user signup page
var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='{{.CSRFToken}}'`)

// 从HTML body中提取token
func extractCSRFToken(t *testing.T, body []byte) string {
	// 使用FindSubmatch方法从HTML body中提取token，会返回一个数组
	// 数组第一个值是匹配的模式，后面的值才是捕获到的数据
	matches := csrfTokenRX.FindSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}
	// html模版会自动转义所有动态数据（包括CSRF令牌），所以这里拿到结果要转义回去
	return html.UnescapeString(string(matches[1]))
}

// 新建app用于测试（并在其中模拟依赖注入）
func newTestApplication(t *testing.T) *application {
	// 创建一个模版缓存实例
	templateCache, err := newTemplateCache("./../../ui/html/")
	if err != nil {
		t.Fatal(err)
	}

	// 创建一个会话管理实例，设置与生产环境相同
	session := sessions.New([]byte("s6Ndh+nzHbS*+9Pk8qGWhTzbpa@ge"))
	session.Lifetime = 12 * time.Hour
	session.Secure = true

	// 初始化依赖使用模仿的loggers和database models
	return &application{
		errorLog:      log.New(ioutil.Discard, "", 0),
		infoLog:       log.New(ioutil.Discard, "", 0),
		session:       session,
		snippets:      &mock.SnippetModel{},
		templateCache: templateCache,
		users:         &mock.UserModel{},
	}
}

// 自定义类型testServer
type testServer struct {
	*httptest.Server
}

// 创建一个新的TestServer helper初始化并返回一个新的testServer实例
func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	// 初始化新的cookie jar（存储cookie的容器）
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	// 将cookie jar添加到客户端，后续的请求自动发送给服务器
	// cookie存储在客户端
	ts.Client().Jar = jar

	// 禁止客户端的重定向，在客户端接收到3xx响应后被调用，返回http.ErrUseLastResponse
	// 会强制它立即返回接收到的响应，不希望重定向干扰测试结果
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &testServer{ts}
}

// 实现一个get方法通过一个给定的url路径对test服务器进行GET请求
// 并返回请求状态码，请求头和请求体
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, body
}

// 发送POST请求给测试服务器，最后一个参数是url.Values对象模拟表单，可以传入任何数据在请求体中
func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, []byte) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	// Read the response body
	defer rs.Body.Close()
	body, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Return the response status, headers and body
	return rs.StatusCode, rs.Header, body
}
