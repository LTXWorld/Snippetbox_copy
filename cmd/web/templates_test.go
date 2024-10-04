package main

import (
	"testing"
	"time"
)

// 测试函数必须以Test开头
func TestHumanDate(t *testing.T) {
	// 创建一个匿名结构体切片包含测试用例
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{
			name: "UTC",
			tm:   time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC),
			want: "17 Dec 2020 at 10:00",
		},
		{
			name: "Empty",
			tm:   time.Time{},
			want: "",
		},
		{
			name: "CET",
			tm:   time.Date(2020, 12, 17, 10, 0, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "17 Dec 2020 at 09:00",
		},
	}
	// 循环测试用例
	for _, tt := range tests {
		// 使用t.Run函数运行测试用例的子例,第一个参数是子例名称，第二个是想要测试的函数
		t.Run(tt.name, func(t *testing.T) {
			hd := humanDate(tt.tm)

			if hd != tt.want {
				t.Errorf("want %q; got %q", tt.want, hd)
			}
		})
	}
}
