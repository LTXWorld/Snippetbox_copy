package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// EmailRX 使用regexp.MustCompile()函数用正则表达式来检查邮箱的格式
// 运行时只编译一次正则表达式
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Form 创建一个自定义的表单结构体，嵌入了url.Values，去保存表单数据和错误属性
// url.Values本身是一个map[string][]string类型，保存表单数据，键为字段名，值为字段真实的值
type Form struct {
	url.Values
	Errors errors
}

// New 初始化这个表单结构体，使用表单data作为参数
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Required 检查表单数据中的多个属性是否存在且不为空
// f调用get方法实际上是调用了url.Values的Get方法
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// MaxLength 检查表单中具体的属性包含一个给定的最大的字符长度
func (f *Form) MaxLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d)", d))
	}
}

// PermittedValues 检查某个具体的属性是否匹配所给的允许的值set
func (f *Form) PermittedValues(field string, opts ...string) {
	value := f.Get(field)
	if value == "" {
		return
	}
	for _, opt := range opts {
		if value == opt {
			return
		}
	}
	f.Errors.Add(field, "This field is invalid")
}

// MinLength 方法检查具体的字段是否满足最小字符长度
func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This feld is too short(minimum is %d)", d))
	}
}

// MatchesPattern 检查具体字段值是否匹配正则表达式
func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}

// Matches 检查两个表单的字段是否匹配
func (f *Form) Matches(filed1, filed2 string) {
	value1 := f.Get(filed1)
	value2 := f.Get(filed2)

	if value1 == "" || value2 == "" {
		return
	}
	if value1 != value2 {
		f.Errors.Add(filed2, "The values do not match")
	}
}

// Valid 如果没有错误发生，返回true
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}
