package main

import (
	"context"
	"fmt"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models"
	"github.com/justinas/nosurf"
	"net/http"
)

// 添加HTTP响应的必要头部信息
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		next.ServeHTTP(w, r)
	})
}

// noSurf 创建了一个CSRF防护中间件
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	// 设置基础的cookie，
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true, // 只能通过HTTP请求访问，不能通过脚本访问
		Path:     "/",  // 应用所有页面都可以访问此Cookie
		Secure:   true,
	})

	return csrfHandler
}

// 记录日志信息,并作为app的方法就可以访问app的依赖项如infoLog
// logRequest - secureHeaders - servemux - app handler
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 这里实现了DI依赖注入和中间件的联动infoLog
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.RequestURI)

		next.ServeHTTP(w, r)
	})
}

// 捕获Panic
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function
		defer func() {
			// Use the builtin recover function to check if there has been panic
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper
				// 将这个string类型的err改为了Error类型传给错误处理
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// 用于检查当前用户的信息中是否有userID，防止当前用户尚未登录
func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.authenticatedUser(r) == nil {
			http.Redirect(w, r, "/user/login", 302)
			return
		}

		next.ServeHTTP(w, r)
	})

}

// 检查当前用户session中的userID是否正确，如果正确将用户信息放入到请求上下文中更新请求
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查userID值是否存在于session中，如果不存在就继续处理链
		exists := app.session.Exists(r, "userID")
		if !exists {
			next.ServeHTTP(w, r)
			return
		}

		// 从数据库中获取当前用户信息，如果不匹配，移除session中的无效userID
		user, err := app.users.Get(app.session.GetInt(r, "userID"))
		if err == models.ErrNoRecord {
			app.session.Remove(r, "userID")
			next.ServeHTTP(w, r)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}

		// 否则，我们可以确定请求来自于一个有效的经过认证的登录用户
		// 将用户信息添加到新创建的请求上下文中，并且调用下一个处理器使用这个请求
		ctx := context.WithValue(r.Context(), contextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx)) // r = r.WithContext(ctx)
	})
}
