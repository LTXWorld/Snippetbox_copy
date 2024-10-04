package main

import (
	"fmt"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/forms"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models"
	"net/http"
	"strconv"
)

// 改变了handler的签名为app，所以它成为了app的一个方法
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	//if r.URL.Path != "/" {
	//	app.notFound(w) // 使用helpers中的错误处理
	//	return
	//}
	// 由于Pat的特性不再需要这个判断

	s, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Create an instance of a templateData struct holding the slice of snippets
	data := &templateData{Snippets: s}

	// 使用helper中的render
	app.render(w, r, "home.page.tmpl", data)
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	// Pat不会从命名捕获中移除冒号
	id, err := strconv.Atoi(r.URL.Query().Get(":id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	// 使用Get方法根据指定id获取数据
	s, err := app.snippets.Get(id)
	if err == models.ErrNoRecord {
		app.notFound(w)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// Pass the flash message to the template，将这个逻辑放在了
	// helpers中默认添加（显示flash）
	app.render(w, r, "show.page.tmpl", &templateData{
		Snippet: s,
	})
}

// Add a new createSnippetForm handler, which for now returns a placeholder result
func (app *application) createSnippetForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "create.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// 处理请求体中的数据
func (app *application) createSnippet(w http.ResponseWriter, r *http.Request) {
	// 首先我们调用r.ParseForm()将来自于POST请求体的数据拿到r.PostForm map中
	// 对PUT，PATCH请求同样有效，如果有错返回400错误
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// 创建一个新的表单结构体来保存客户端上传的表单数据
	// 再使用检验方法去检查内容是否有错
	form := forms.New(r.PostForm)
	form.Required("title", "content", "expires")
	form.MaxLength("title", 100)
	form.PermittedValues("expires", "365", "7", "1")

	// 如果表单有错误，重新展示模版内容及其中数据
	if !form.Valid() {
		app.render(w, r, "create.page.tmpl", &templateData{Form: form})
		return
	}

	// 表单数据已经被嵌入在form.Form结构体中，直接使用Get方法获取相应属性的有效的值
	id, err := app.snippets.Insert(form.Get("title"), form.Get("content"), form.Get("expires"))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// 使用Put方法去添加一个字符串值和一个响应key给session data
	// 如果session不存在，session中间件会自动为我们创建一个新的
	app.session.Put(r, "flash", "日志已成功创建!")

	// 创建成功将用户重定向到相关的页面
	http.Redirect(w, r, fmt.Sprintf("/snippet/%d", id), http.StatusSeeOther)
}

// 添加关于用户登录登出等一系列方法
// 展示用户注册表单
func (app *application) signupUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "signup.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// 实现用户注册，创建一个新用户
// 表单被提交时，数据就会发到这个处理器中
func (app *application) signupUser(w http.ResponseWriter, r *http.Request) {
	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form contents using the form helper
	// 与上面createSnippet逻辑一样
	form := forms.New(r.PostForm)
	form.Required("name", "email", "password")
	form.MatchesPattern("email", forms.EmailRX)
	form.MinLength("password", 8)

	// 如果有格式发生，指出错误并重现所填（除了密码）
	if !form.Valid() {
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	}

	// 如果没有格式错误，但有可能出现邮箱重复错误，新建用户记录插入
	err = app.users.Insert(form.Get("name"), form.Get("email"), form.Get("password"))
	if err == models.ErrDuplicateEmail {
		form.Errors.Add("email", "Address is already in use")
		app.render(w, r, "signup.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// 添加一个临时信息到session确定登录成功并要求登录
	app.session.Put(r, "flash", "Your signup was successful. Please log in.")

	// And redirect the user to login page
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// 展示用户登录表单
func (app *application) loginUserForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "login.page.tmpl", &templateData{
		Form: forms.New(nil),
	})
}

// 实现用户登录
func (app *application) loginUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// 调用Usermodel数据库中的验证方法
	// 如果出现凭证错误，返回邮箱或密码错误提示信息
	form := forms.New(r.PostForm)
	id, err := app.users.Authenticate(form.Get("email"), form.Get("password"))
	if err == models.ErrInvalidCredentials {
		form.Errors.Add("generic", "邮箱或密码出现错误")
		app.render(w, r, "login.page.tmpl", &templateData{Form: form})
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	// 将id加入到当前用户的session中
	app.session.Put(r, "userID", id)

	// 重定向到创建日志页面
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

// 实现用户登出
func (app *application) logoutUser(w http.ResponseWriter, r *http.Request) {
	// Remove the userID from the session data 来实现登出
	app.session.Remove(r, "userID")
	// Add a flash message to the session to confirm 已经登出
	app.session.Put(r, "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// 实现展示app介绍页面，只涉及静态信息，不涉及用户输入
func (app *application) about(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "about.page.tmpl", &templateData{})
}

// 用于测试
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
