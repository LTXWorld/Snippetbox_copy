package mysql

import (
	"database/sql"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models"
)

// SnippetModel 定义一个Model类型包裹着sql.DB连接池
type SnippetModel struct {
	DB *sql.DB
}

// Insert 插入一个新的snippet到数据库中，并返回对应的id
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	// 书写sql语句
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// 使用Exec方法去执行sql语句，后面的参数用来填充占位符?，不会将内容作为sql语句的一部分，就是简单的值
	// Exec返回一个sql.Result接口
	// 创建了一个prepared statement，数据库提前编译了
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// 获取插入记录后的id，只是mysqlDriver特供方法
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// 将int64类型的ID转换为int类型
	return int(id), nil
}

// Get 根据id返回一个具体的snippet
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	// Write the SQL statement we want to execute
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`

	// 使用QueryRow()方法查询单一的行结果
	row := m.DB.QueryRow(stmt, id)

	// 初始化一个指针指向新的snippet struct
	s := &models.Snippet{}

	// 使用row.Scan从查询到的结果中复制每个属性值给新的结构体
	// driver 自动将原始的SQL数据库中的输出转换为需要的Go类型
	// char,varchar,text->string;time,date,timestamp->time.Time
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	// 如果一切正常，返回新结构体
	return s, nil
}

// Latest 返回10个最近创建的snippet,
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	// 多结果查询ORDER BY DESC LIMIT
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`

	// 使用Query方法在连接池去执行多结果查询
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	// 确保resultset总是在Latest方法返回前正确关闭,但也要确保rows不是nil
	// 因为nil执行Close会产生panic
	// 如果resultset没有正常关闭，会耗尽连接池。
	defer rows.Close()

	// 初始化一个切片来保存Snippets对象
	snippets := []*models.Snippet{}

	// 使用rows.Next来迭代结果(resultset)
	for rows.Next() {
		s := &models.Snippet{}
		// 同单个结果，使用Scan将属性值全部拷贝进s对象
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		// 将结果添加到切片中
		snippets = append(snippets, s)
	}

	// 当循环结束，我们调用rows.Err()来检查是否出错
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// 如果一切正常返回切片
	return snippets, nil
}
