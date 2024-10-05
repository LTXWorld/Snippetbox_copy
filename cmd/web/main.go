package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models/mysql"
	"github.com/golangcollege/sessions"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // 需要它的init函数自动注册驱动到database/sql包中，但是不会调用这个包的任何函数，所以使用_
)

// 定义请求上下文键类型
type contextKey string

var contextKeyUser = contextKey("user")

// 定义一个application结构体去保存依赖，以便在handlers中使用
type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	// 新加一个依赖来自于pkg的数据库操作
	snippets interface {
		Insert(string, string, string) (int, error)
		Get(int) (*models.Snippet, error)
		Latest() ([]*models.Snippet, error)
	} // 结构体依赖于这个接口，只要实现了这三个方法的任何类型，
	templateCache map[string]*template.Template // 添加依赖来自于html/template包
	session       *sessions.Session             // 添加session依赖管理状态
	// 同理
	users interface {
		Insert(string, string, string) error
		Authenticate(string, string) (int, error)
		Get(int) (*models.User, error)
		UpdatePassword(int, string) error
	}
	//
	snippetsImages interface {
		Insert(int, string) error
		GetImagesBySnippetID(int) ([]string, error)
	}
}

func main() {
	// 定义一个新的命令行标志，默认值+描述
	// 标志值将会被存储在运行时的addr变量中，将所传的值转换为String，为了下方监听所传入参数
	addr := flag.String("addr", "0.0.0.0:4000", "HTTP network address")

	// 定义另一个标志表示数据库连接描述
	dsn := flag.String("dsn", "web:iutaol123@/snippetbox?parseTime=true", "MySQL data source name")

	// 定义一个新的命令行标志为了session secret默认值是一个随机的key
	// 用来封装验证session cookies
	secret := flag.String("secret", "s6Ndh+nzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
	// 扫描命令行参数根据预定义的标志解析参数
	flag.Parse()

	// 创建一个logger来写信息消息，目标输出位置，日志前缀，日志输出的额外信息
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// 创建一个logger来写错误信息
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// 通过传来的命令行参数连接数据库
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	defer db.Close()

	// 初始化一个新的模版缓存,为了下方注入依赖
	templateCache, err := newTemplateCache("./ui/html")
	if err != nil {
		errorLog.Fatal(err)
	}

	// 初始化一个新的session manager,传入密钥当做参数
	// 返回一个session结构体包括了会话的配置信息
	// 比如生命周期，设置12个小时的过期时间
	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	session.Secure = true // 设置安全标志

	// 初始化一个新的application实例包括这些依赖
	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &mysql.SnippetModel{DB: db},
		templateCache:  templateCache,
		session:        session,
		users:          &mysql.UserModel{DB: db},
		snippetsImages: &mysql.SnippetImageModel{DB: db},
	}

	// 初始化一个tls.Config结构体去保存我们想要服务器使用的TLS设置
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// 让服务器的TLSConfig属性使用我们创建的tlsConfig
	// 初始化一个新的http.Server Struct代替简单的http.ListenAndServe
	// 使用和之前相同的路由信息addr,mux
	srv := &http.Server{
		Addr:      *addr,
		ErrorLog:  errorLog,
		Handler:   app.routes(),
		TLSConfig: tlsConfig,
		// Add Idle, Read and Write timeouts to the server
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 使用两个logger来写信息，而不是标准的logger
	infoLog.Printf("Starting server on %s", *addr)
	// 使用ListenAndServeTLS代替ListenAndServe，以便启动一个HTTPS服务器
	// 需要传入TLS certificate and corresponding private key路径
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	// 初始化一个连接池，所有的连接都是lazy等待唤醒的
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// Ping用来创建一个链接来检查错误
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
