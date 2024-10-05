package main

import (
	"bytes"
	"fmt"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models"
	"github.com/justinas/nosurf"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

// The serverError helper 写了一个错误信息并且使用栈
// 并返回一个错误响应给客户端 500
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace) // 允许指定日志栈的深度，2正好是可以追溯到错误发生的文件名和行号

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code 400
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status) //StatusText生成用户可读的文本展示
}

// For consistency, we'll also implement a notFound helper
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// This takes a pointer to a templateData struct
// adds the current year to CurrentYear filed
// returns the pointer
// 每次执行渲染render时，自动加入到templateData中的数据：年，临时信息，id, CSRF令牌
func (app *application) addDefaultData(td *templateData, r *http.Request) *templateData {
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()
	// Add the flash message to the template data
	td.Flash = app.session.PopString(r, "flash")
	td.AuthenticatedUser = app.authenticatedUser(r)
	// Add the CSRF token to the template data
	td.CSRFToken = nosurf.Token(r)
	return td
}

// 减少handlers使用模版时的代码复用
func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td *templateData) {
	// 从缓存中根据page name对模版set进行迭代
	ts, ok := app.templateCache[name]
	if !ok {
		app.serverError(w, fmt.Errorf("The template %s does not exist", name))
		return
	}

	// 为了防止模版出错，来一次试渲染
	buf := new(bytes.Buffer)

	// 先绑定数据到buf上，而不是直接在w上绑定数据
	// 暂存渲染后的内容
	err := ts.Execute(buf, app.addDefaultData(td, r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	// 将buffer中的内容写入w中
	buf.WriteTo(w)

	//// 上面加载模版后，将数据绑定在模版上
	//err := ts.Execute(w, td)
	//if err != nil {
	//	app.serverError(w, err)
	//}
}

// authenticatedUser 方法通过当前用户session中的userID返回用户信息(用户对象)
func (app *application) authenticatedUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(contextKeyUser).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// saveUploadedFile 将上传的文件保存在服务器，并返回保存路径
func (app *application) saveUploadedFile(fileHeader *multipart.FileHeader) (string, error) {
	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	destPath := "./uploads/" + fileHeader.Filename

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	// 将上传的文件内容复制到目标文件
	_, err = io.Copy(destFile, file)
	if err != nil {
		return "", err
	}

	return destPath, nil
}
