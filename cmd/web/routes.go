package main

import (
	"github.com/bmizerany/pat"
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// 创建了一个处理链，被用在每个app的请求
	standardMiddleware := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	// 创建动态处理链对于动态应用路由。只会作用域session中间件
	// 将nosurf中间件作用域我们所有的动态路由上
	dynamicMiddleware := alice.New(app.session.Enable, noSurf, app.authenticate)

	// mux也实现了ServeHTTP()方法，是一个请求路由器和多路复用器
	// 这里引入第三方路由框架
	mux := pat.New()
	// 使用动态处理中间件链来解决这些路由问题
	// 最终实现了转为handler注册为路由
	mux.Get("/", dynamicMiddleware.ThenFunc(app.home))
	// Append增加requireAuthenticatedUser中间件来保护create路由
	// 防止未登录的用户进行创建操作
	mux.Get("/snippet/create", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createSnippetForm))
	mux.Post("/snippet/create", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.createSnippet))
	mux.Get("/snippet/:id", dynamicMiddleware.ThenFunc(app.showSnippet))

	// 添加5个新的用户路由
	mux.Get("/user/signup", dynamicMiddleware.ThenFunc(app.signupUserForm))
	mux.Post("/user/signup", dynamicMiddleware.ThenFunc(app.signupUser))
	mux.Get("/user/login", dynamicMiddleware.ThenFunc(app.loginUserForm))
	mux.Post("/user/login", dynamicMiddleware.ThenFunc(app.loginUser))
	// 同理添加中间件保护路由
	mux.Post("/user/logout", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.logoutUser))
	// 添加重置密码的处理路由
	mux.Get("/user/resetpassword", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.resetPasswordForm))
	mux.Post("/user/resetpassword", dynamicMiddleware.Append(app.requireAuthenticatedUser).ThenFunc(app.resetPassword))
	// 添加处理函数为了About界面
	mux.Get("/about", dynamicMiddleware.ThenFunc(app.about))
	// 注册ping处理器为了测试用
	mux.Get("/ping", http.HandlerFunc(ping))

	// fileServer创建一个用于提供静态文件的HTTP文件服务器
	// 将./ui/static目录作为静态文件的根目录，处理对该目录中文件的请求
	fileServer := http.FileServer(http.Dir("./ui/static"))

	// 将文件服务器挂载到路由器，处理/static/开头的请求
	// 去除URL中的/static前缀，将剩余路径交给文件服务器处理
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	// 同理，为uploads创建静态文件服务器，处理/uploads/文件名 路径
	uploadFileServer := http.FileServer(http.Dir("./uploads"))
	mux.Get("/uploads/", http.StripPrefix("/uploads/", uploadFileServer))

	// Pass the servemux as the next param to the secureHeaders middleware
	// 直接给serveMux前面添加中间件，将mux包裹起来
	return standardMiddleware.Then(mux)
}
