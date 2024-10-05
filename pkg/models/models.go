package models

import (
	"errors"
	"time"
)

var (
	ErrNoRecord = errors.New("models: no matching record found")
	// ErrInvalidCredentials 如果用户登录信息有误
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	// ErrDuplicateEmail 如果注册时邮箱地址已经被使用了
	ErrDuplicateEmail = errors.New("models: duplicate email")
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// User 定义一个用户类型
type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type SnippetImages struct {
	ID        int
	SnippetId int
	ImagePath string
}
