package mysql

import (
	"database/sql"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type UserModel struct {
	DB *sql.DB
}

// Insert 方法将字段信息插入到数据库表中，并且插入哈希后的密码
// 同时出错后检查邮箱是否重复
func (m *UserModel) Insert(name, email, password string) error {
	// Create a bcrypt hash of the plain-text password明文密码的bcr哈希值
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
	VALUES(?, ?, ?, UTC_TIMESTAMP())`

	// Exec执行,如果返回错误，尝试断言为*mysql.MySQLError对象，可以检查错误编号是否为1062
	// 如果是，通过检查错误是否与email重复有关
	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "users_uc_email key") {
				return models.ErrDuplicateEmail
			}
		}
	}
	return err
}

// Authenticate 方法通过给定的邮箱，密码验证用户是否存在，如果存在返回id
func (m *UserModel) Authenticate(email, password string) (int, error) {
	// 通过给定的邮箱搜索id和哈希密码
	// 如果没有查询到结果返回凭证错误
	var id int
	var hashedPassword []byte
	row := m.DB.QueryRow("SELECT id, hashed_password FROM users WHERE email = ?", email)
	err := row.Scan(&id, &hashedPassword) // 将查询结果扫描到变量id,hashedPassword中
	if err == sql.ErrNoRows {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	// 如果根据邮箱查找成功，再来检查密码是否匹配
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	// 如果都没问题，返回用户id
	return id, nil
}

// Get 方法根据特定ID获取特定用户
func (m *UserModel) Get(id int) (*models.User, error) {
	s := &models.User{}

	stmt := `SELECT id, name, email, created FROM users WHERE id = ?`
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Name, &s.Email, &s.Created)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	return s, nil
}
