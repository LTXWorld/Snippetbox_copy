package forms

// 定义一个新的错误类型，将用它来承载检验表单输入时出现的错误信息
// 表单的属性名称将会被用作map的key
type errors map[string][]string

// Add Implement an Add() method to add error messages for a given field
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get 实现一个get方法来获取给定属性的第一个错误信息
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
