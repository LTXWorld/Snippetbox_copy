package mysql

import (
	"github.com/LTXWorld/codingOnLinux/GoProject/snippetbox/pkg/models"
	"reflect"
	"testing"
	"time"
)

func TestUserModelGet(t *testing.T) {
	// 检查命令行是否有short标签，short可以用来跳过当前测试
	if testing.Short() {
		t.Skip("mysql: skipping integration test")
	}

	// 设置一个合适的表驱动测试
	tests := []struct {
		name      string
		userID    int
		wantUser  *models.User
		wantError error
	}{
		{
			name:   "Valid ID",
			userID: 1,
			wantUser: &models.User{
				ID:      1,
				Name:    "Alice Jones",
				Email:   "alice@example.com",
				Created: time.Date(2018, 12, 23, 17, 25, 22, 0, time.UTC),
			},
			wantError: nil,
		},
		{
			name:      "Zero ID",
			userID:    0,
			wantUser:  nil,
			wantError: models.ErrNoRecord,
		},
		{
			name:      "Non-existent ID",
			userID:    2,
			wantUser:  nil,
			wantError: models.ErrNoRecord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, teardown := newTestDB(t)
			defer teardown()

			// Create a new instance of the UserModel
			m := UserModel{db}

			// 调用UserModel.Get方法进行测试
			user, err := m.Get(tt.userID)

			if err != tt.wantError {
				t.Errorf("want %v; got %s", tt.wantError, err)
			}

			// 检查任意复杂的自定义类型之间的相等性
			if !reflect.DeepEqual(user, tt.wantUser) {
				t.Errorf("want %v; got %v", tt.wantUser, user)
			}
		})
	}
}
