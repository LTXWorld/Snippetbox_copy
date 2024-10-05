package mysql

import (
	"database/sql"
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models"
)

type SnippetImageModel struct {
	DB *sql.DB
}

// Insert 用于向数据库中插入图片路径和对应的snippet_id
func (m *SnippetImageModel) Insert(snippetID int, imagePath string) error {
	stmt := `INSERT INTO snippet_images (snippet_id, image_path) VALUES(?, ?)`

	_, err := m.DB.Exec(stmt, snippetID, imagePath)
	return err
}

// GetImagesBySnippetID 获取某个日志的所有图片路径，保存在一个切片中
func (m *SnippetImageModel) GetImagesBySnippetID(snippetID int) ([]string, error) {
	stmt := `SELECT image_path FROM snippet_images WHERE snippet_id = ?`

	rows, err := m.DB.Query(stmt, snippetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 保存读取到的信息里面的路径
	var images []string
	for rows.Next() {
		var imagePath string
		err = rows.Scan(&imagePath)
		if err != nil {
			return nil, err
		}
		images = append(images, imagePath)
	}

	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}

	// 如果一切正常，返回路径
	return images, nil
}
