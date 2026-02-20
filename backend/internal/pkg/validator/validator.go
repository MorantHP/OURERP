package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Validator 验证器
type Validator struct {
	errors map[string]string
}

// New 创建验证器
func New() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

// Errors 获取所有错误
func (v *Validator) Errors() map[string]string {
	return v.errors
}

// HasErrors 是否有错误
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// AddError 添加错误
func (v *Validator) AddError(field, message string) {
	if _, exists := v.errors[field]; !exists {
		v.errors[field] = message
	}
}

// Required 验证必填
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, fmt.Sprintf("%s不能为空", field))
	}
	return v
}

// RequiredInt 验证必填整数
func (v *Validator) RequiredInt(field string, value int64) *Validator {
	if value == 0 {
		v.AddError(field, fmt.Sprintf("%s不能为空", field))
	}
	return v
}

// MinLength 最小长度
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if utf8.RuneCountInString(value) < min {
		v.AddError(field, fmt.Sprintf("%s长度不能少于%d个字符", field, min))
	}
	return v
}

// MaxLength 最大长度
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, fmt.Sprintf("%s长度不能超过%d个字符", field, max))
	}
	return v
}

// Range 范围验证
func (v *Validator) Range(field string, value, min, max int) *Validator {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("%s必须在%d到%d之间", field, min, max))
	}
	return v
}

// Email 验证邮箱
func (v *Validator) Email(field, value string) *Validator {
	if value == "" {
		return v
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s格式不正确", field))
	}
	return v
}

// Phone 验证手机号
func (v *Validator) Phone(field, value string) *Validator {
	if value == "" {
		return v
	}
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !phoneRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s格式不正确", field))
	}
	return v
}

// URL 验证URL
func (v *Validator) URL(field, value string) *Validator {
	if value == "" {
		return v
	}
	urlRegex := regexp.MustCompile(`^https?://[\w\-]+(\.[\w\-]+)+[/#?]?.*$`)
	if !urlRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s格式不正确", field))
	}
	return v
}

// Numeric 验证数字
func (v *Validator) Numeric(field, value string) *Validator {
	if value == "" {
		return v
	}
	numericRegex := regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	if !numericRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s必须是数字", field))
	}
	return v
}

// Alphanumeric 验证字母数字
func (v *Validator) Alphanumeric(field, value string) *Validator {
	if value == "" {
		return v
	}
	for _, r := range value {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			v.AddError(field, fmt.Sprintf("%s只能包含字母和数字", field))
			break
		}
	}
	return v
}

// In 验证在指定值中
func (v *Validator) In(field, value string, allowed []string) *Validator {
	if value == "" {
		return v
	}
	for _, a := range allowed {
		if a == value {
			return v
		}
	}
	v.AddError(field, fmt.Sprintf("%s的值无效", field))
	return v
}

// Match 正则匹配
func (v *Validator) Match(field, value string, pattern *regexp.Regexp) *Validator {
	if value == "" {
		return v
	}
	if !pattern.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s格式不正确", field))
	}
	return v
}

// Password 验证密码强度
func (v *Validator) Password(field, value string) *Validator {
	if value == "" {
		return v
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsNumber(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	score := 0
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasNumber {
		score++
	}
	if hasSpecial {
		score++
	}
	if utf8.RuneCountInString(value) >= 8 {
		score++
	}

	if score < 3 {
		v.AddError(field, "密码强度不足，请包含大小写字母、数字或特殊字符，且长度至少8位")
	}
	return v
}

// Username 验证用户名
func (v *Validator) Username(field, value string) *Validator {
	if value == "" {
		return v
	}

	length := utf8.RuneCountInString(value)
	if length < 3 || length > 20 {
		v.AddError(field, "用户名长度必须在3-20个字符之间")
		return v
	}

	usernameRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !usernameRegex.MatchString(value) {
		v.AddError(field, "用户名必须以字母开头，只能包含字母、数字、下划线和连字符")
	}
	return v
}

// Sanitize 清理输入
func Sanitize(input string) string {
	// 移除前后空白
	input = strings.TrimSpace(input)

	// 移除控制字符
	var result strings.Builder
	for _, r := range input {
		if !unicode.IsControl(r) || r == '\n' || r == '\t' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SanitizeHTML 清理HTML（移除标签）
func SanitizeHTML(input string) string {
	// 简单的HTML标签移除
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

// EscapeHTML 转义HTML特殊字符
func EscapeHTML(input string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(input)
}

// Truncate 截断字符串
func Truncate(input string, maxLen int) string {
	if utf8.RuneCountInString(input) <= maxLen {
		return input
	}
	runes := []rune(input)
	return string(runes[:maxLen]) + "..."
}

// ValidateStruct 验证结构体标签 (简化版本)
// 可以集成 go-playground/validator 进行更强大的验证
type StructValidator struct {
	validator *Validator
}

// NewStructValidator 创建结构体验证器
func NewStructValidator() *StructValidator {
	return &StructValidator{
		validator: New(),
	}
}

// Validate 验证结构体
func (sv *StructValidator) Validate(s interface{}) map[string]string {
	// 这里可以扩展为使用反射读取结构体标签进行验证
	// 目前返回空，由业务层手动调用验证器
	return sv.validator.Errors()
}
