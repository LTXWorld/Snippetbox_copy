package main

import (
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/forms"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models"
	"html/template"
	"path/filepath"
	"time"
)

// 定义一个templateData type 来包裹任何动态数据结构（用来传给HTML）
// 这里既有单个结构体也有结构体切片
type templateData struct {
	CurrentYear       int // 用来存放一般动态数据
	Snippet           *models.Snippet
	Snippets          []*models.Snippet
	Form              *forms.Form  // 引入表单字段（包括具体字段值和错误信息）
	Flash             string       // 临时消息存储机制
	AuthenticatedUser *models.User // 之前通过id判断当前用户是否已经登录，现在通过上下文中包含的用户对象
	CSRFToken         string       // 表示模版中的CSRFToken属性，使每个表单都有一个CSRF令牌
	Images            []string     // 用于存储图片路径
}

// 自定义函数humanDate
func humanDate(t time.Time) string {
	// Return the empty string if time has the zero value
	if t.IsZero() {
		return ""
	}
	// Convert the time to UTC before formatting it
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// FuncMap将自定义函数注册到模版中
// map键为模版中使用时的名称，值是实际的go函数
// 将humanDate函数映射为humanDate这个名称，可以在模版中使用{{huamnDate .Timestamp}}调用
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// 添加缓存方法
func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// 初始化map来充当cache
	cache := map[string]*template.Template{}

	// Use filepath.Glob 函数获取以.page.tmpl结尾的切片
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	// Loop through the pages one-by-one
	for _, page := range pages {
		// 从路径中提取文件名例如 'home.page.tmpl'
		name := filepath.Base(page)

		// Parse the page template file into a template set
		// Funcs将functions函数注册到模版中
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// 这里后面的代码用来添加其他的模版进入模版集中，为每个page.tmpl添加layout,partial
		// 就像handlers中之前的手写files路径的代码一样，完成加载
		// Use the ParseGlob method to add layout模版到set中
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		// 同理，将partial添加到模版中
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		// 最终每一个page形成一个模版ts,将模版set添加到缓存中,使用page的name作为key
		cache[name] = ts
	}

	return cache, nil
}
